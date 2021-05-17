/*
 * Copyright 2019 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package lib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/converter/lib/converter/base"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda/interfaces"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicegroups"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	"log"
	"reflect"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/SENERGY-Platform/external-task-worker/util"

	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
)

const CheckCamundaDuration = false

type worker struct {
	consumer                  kafka.ConsumerInterface
	producer                  kafka.ProducerInterface
	repository                devicerepository.RepoInterface
	camunda                   interfaces.CamundaInterface
	config                    util.Config
	marshaller                marshaller.Interface
	camundaCallMux            sync.Mutex
	lastSuccessfulCamundaCall time.Time
	producerCallMux           sync.Mutex
	lastProducerCallSuccess   bool
	deviceGroupsHandler       DeviceGroupsHandler
}

type DeviceGroupsHandler interface {
	ProcessCommand(request messages.Command, task messages.CamundaExternalTask) (completed bool, missingRequests []messages.KafkaMessage, finishedResults []interface{}, err error)
	ProcessResponse(taskId string, subResult interface{}) (parent messages.GroupTaskMetadataElement, results []interface{}, finished bool, err error)
}

func Worker(ctx context.Context, config util.Config, kafkaFactory kafka.FactoryInterface, repoFactory devicerepository.FactoryInterface, camundaFactory interfaces.FactoryInterface, marshallerFactory marshaller.FactoryInterface) {
	log.Println("start camunda worker")
	base.DEBUG = config.Debug
	var err error

	w := worker{
		config:                    config,
		marshaller:                marshallerFactory.New(config.MarshallerUrl),
		lastProducerCallSuccess:   true,
		lastSuccessfulCamundaCall: time.Now(),
	}

	StartHealthCheckEndpoint(ctx, config, &w)

	if config.CompletionStrategy != util.OPTIMISTIC {
		w.consumer, err = kafkaFactory.NewConsumer(config, w.HandleTaskResponse)
		if err != nil {
			log.Fatal("ERROR: kafkaFactory.NewConsumer", err)
		}
		defer w.consumer.Stop()
	}
	w.producer, err = kafkaFactory.NewProducer(config)
	if err != nil {
		log.Fatal("ERROR: kafkaFactory.NewProducer", err)
	}
	if config.Debug {
		w.producer.Log(log.New(log.Writer(), "[CONNECTOR-KAFKA] ", 0))
	}
	defer w.producer.Close()
	w.repository = repoFactory.Get(config)
	w.camunda, err = camundaFactory.Get(config, w.producer)
	if err != nil {
		log.Fatal("ERROR: kafkaFactory.NewProducer", err)
	}
	w.deviceGroupsHandler = devicegroups.New(config.GroupScheduler, w.camunda, w.repository, w.CreateProtocolMessage, config.CamundaFetchLockDuration, config.SubResultExpirationInSeconds, config.SubResultDatabaseUrls)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			wait := w.ExecuteNextTasks()
			if wait {
				duration := time.Duration(config.CamundaWorkerTimeout) * time.Millisecond
				time.Sleep(duration)
			}
		}
	}
}

func (this *worker) ExecuteNextTasks() (wait bool) {
	tasks, err := this.camunda.GetTasks()
	if err != nil {
		log.Println("error on ExecuteNextTasks getTask", err)
		return true
	}
	this.logSuccessfulCamundaCall()
	if len(tasks) == 0 {
		return true
	}
	wg := sync.WaitGroup{}
	for _, task := range tasks {
		wg.Add(1)
		go func(asyncTask messages.CamundaExternalTask) {
			defer wg.Done()
			this.ExecuteTask(asyncTask)
		}(task)
	}
	wg.Wait()
	return false
}

func (this *worker) ExecuteTask(task messages.CamundaExternalTask) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR: ", r, "\n", debug.Stack())
		}
	}()
	if this.config.Debug {
		log.Println("Start", task.Id, util.TimeNow().Second())
		log.Println("Get new Task: ", task)
	}
	if task.Error != "" {
		log.Println("WARNING: existing failure in camunda task", task.Error)
	}
	request, err := GetCommandRequest(task)
	if err != nil {
		log.Println("error on GetCommandRequest(): ", err)
		this.camunda.Error(task.Id, task.ProcessInstanceId, task.ProcessDefinitionId, "invalid task format (json)", task.TenantId)
		return
	}

	completed, nextProtocolMessages, results, err := this.deviceGroupsHandler.ProcessCommand(request, task)
	if err != nil {
		this.camunda.Error(task.Id, task.ProcessInstanceId, task.ProcessDefinitionId, err.Error(), task.TenantId)
		return
	}

	if this.config.CompletionStrategy == util.OPTIMISTIC || completed {
		time.Sleep(time.Duration(this.config.OptimisticTaskCompletionTimeout) * time.Millisecond) //prevent completes that are to fast
		err = this.camunda.CompleteTask(messages.TaskInfo{
			WorkerId:            this.camunda.GetWorkerId(),
			TaskId:              task.Id,
			ProcessInstanceId:   task.ProcessInstanceId,
			ProcessDefinitionId: task.ProcessDefinitionId,
			TenantId:            task.TenantId,
		}, this.config.CamundaTaskResultName, handleVersioning(request.Version, results))
		if err != nil {
			log.Println("error on completeCamundaTask(): ", err)
			return
		}
		if this.config.CompletionStrategy != util.OPTIMISTIC {
			//if optimistic: we have still to send the commands
			//else: this is not the first try and now we are finished
			return
		}
	}

	for _, message := range nextProtocolMessages {
		err = this.producer.ProduceWithKey(message.Topic, message.Key, message.Payload)
		if err != nil {
			log.Println("ERROR: unable to produce kafka msg", err)
			this.logProducerError()
		} else {
			this.logProducerSuccess()
		}
	}
}

func (this *worker) HandleTaskResponse(msg string) (err error) {
	var message messages.ProtocolMsg
	err = json.Unmarshal([]byte(msg), &message)
	if err != nil {
		return err
	}
	defer func() {
		if this.config.Debug {
			log.Println("TRACE: ", message.Trace)
		}
	}()

	if message.TaskInfo.CompletionStrategy == util.OPTIMISTIC {
		return nil
	}

	message.Trace = append(message.Trace, messages.Trace{
		Timestamp: time.Now().UnixNano(),
		TimeUnit:  "unix_nano",
		Location:  "github.com/SENERGY-Platform/external-task-worker HandleTaskResponse() message unmarshal",
	})

	if CheckCamundaDuration && this.missesCamundaDuration(message) {
		log.Println("WARNING: drop late response:", msg)
		return
	}

	var output interface{}
	if message.Metadata.OutputCharacteristic != model.NullCharacteristic.Id {
		output, err = this.marshaller.UnmarshalFromServiceAndProtocol(message.Metadata.OutputCharacteristic, message.Metadata.Service, message.Metadata.Protocol, message.Response.Output, message.Metadata.ContentVariableHints)
		if err != nil {
			return err
		}
	}
	message.Trace = append(message.Trace, messages.Trace{
		Timestamp: time.Now().UnixNano(),
		TimeUnit:  "unix_nano",
		Location:  "github.com/SENERGY-Platform/external-task-worker HandleTaskResponse() after this.marshaller.UnmarshalFromServiceAndProtocol()",
	})

	parent, results, finished, err := this.deviceGroupsHandler.ProcessResponse(message.TaskInfo.TaskId, output)
	if err == devicegroups.ErrNotFount {
		return nil //if parent task is not found -> can not be finished (may already be done)
	}
	if err != nil {
		return err
	}

	message.Trace = append(message.Trace, messages.Trace{
		Timestamp: time.Now().UnixNano(),
		TimeUnit:  "unix_nano",
		Location:  "github.com/SENERGY-Platform/external-task-worker deviceGroupsHandler.ProcessResponse()",
	})

	if finished {
		message.TaskInfo = messages.TaskInfo{
			WorkerId:            message.TaskInfo.WorkerId,
			TaskId:              parent.Task.Id,
			ProcessInstanceId:   parent.Task.ProcessInstanceId,
			ProcessDefinitionId: parent.Task.ProcessDefinitionId,
			CompletionStrategy:  message.TaskInfo.CompletionStrategy,
			Time:                message.TaskInfo.Time,
			TenantId:            message.TaskInfo.TenantId,
		}

		err = this.camunda.CompleteTask(
			message.TaskInfo,
			this.config.CamundaTaskResultName,
			handleVersioning(parent.Command.Version, results))

		if this.config.Debug {
			log.Println("Complete", parent.Task.Id, util.TimeNow().Second(), output, msg, err)
		}
		message.Trace = append(message.Trace, messages.Trace{
			Timestamp: time.Now().UnixNano(),
			TimeUnit:  "unix_nano",
			Location:  "github.com/SENERGY-Platform/external-task-worker HandleTaskResponse() after this.camunda.HandleTaskResponse()",
		})
	} else if this.isSequential() {
		this.ExecuteTask(parent.Task)
	}
	return
}

func (this *worker) isSequential() bool {
	return this.config.GroupScheduler == util.SEQUENTIAL || this.config.GroupScheduler == util.ROUND_ROBIN
}

func handleVersioning(version int, results []interface{}) interface{} {
	if version < 2 && len(results) > 0 {
		return results[0]
	} else {
		return results
	}
}

func (this *worker) missesCamundaDuration(msg messages.ProtocolMsg) bool {
	if msg.TaskInfo.Time == "" {
		return true
	}
	unixTime, err := strconv.ParseInt(msg.TaskInfo.Time, 10, 64)
	if err != nil {
		return true
	}
	taskTime := time.Unix(unixTime, 0)
	return util.TimeNow().Sub(taskTime) >= time.Duration(this.config.CamundaFetchLockDuration)*time.Millisecond
}

type WorkerState struct {
	ProducerOk                bool
	LastSuccessfulCamundaCall time.Duration
}

func (this *worker) GetState() WorkerState {
	return WorkerState{
		ProducerOk:                this.LastProducerCallSuccessful(),
		LastSuccessfulCamundaCall: this.SinceLastSuccessfulCamundaCall(),
	}
}

func (this *worker) SinceLastSuccessfulCamundaCall() time.Duration {
	this.camundaCallMux.Lock()
	defer this.camundaCallMux.Unlock()
	return time.Since(this.lastSuccessfulCamundaCall)
}

func (this *worker) logSuccessfulCamundaCall() {
	this.camundaCallMux.Lock()
	defer this.camundaCallMux.Unlock()
	this.lastSuccessfulCamundaCall = time.Now()
}

func (this *worker) LastProducerCallSuccessful() bool {
	this.producerCallMux.Lock()
	defer this.producerCallMux.Unlock()
	return this.lastProducerCallSuccess
}

func (this *worker) logProducerError() {
	this.producerCallMux.Lock()
	defer this.producerCallMux.Unlock()
	this.lastProducerCallSuccess = false
}

func (this *worker) logProducerSuccess() {
	this.producerCallMux.Lock()
	defer this.producerCallMux.Unlock()
	this.lastProducerCallSuccess = true
}

func setVarOnPath(element interface{}, path []string, value interface{}) (result interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR: Recovered in setVarOnPath: ", r)
			err = errors.New(fmt.Sprint("ERROR: Recovered in setVarOnPath: ", r))
		}
	}()
	if len(path) <= 0 {
		switch v := value.(type) {
		case string:
			switch e := element.(type) {
			case string:
				return v, err
			case int:
				return strconv.Atoi(v)
			case bool:
				return strconv.ParseBool(v)
			case float64:
				return strconv.ParseFloat(v, 64)
			default:
				eVal := reflect.ValueOf(element)
				if eVal.Kind() == reflect.Map || eVal.Kind() == reflect.Slice {
					var vInterface interface{}
					err = json.Unmarshal([]byte(v), &vInterface)
					if err == nil {
						eVal.Set(reflect.ValueOf(vInterface))
					}
				} else {
					err = errors.New(fmt.Sprintf("ERROR: getSubelement(), unknown element type %T", e))
				}
			}
		case int:
			switch e := element.(type) {
			case string:
				return strconv.Itoa(v), err
			case int:
				return v, err
			case bool:
				return v >= 1, err
			case float64:
				return float64(v), err
			default:
				err = errors.New(fmt.Sprintf("ERROR: getSubelement(), unknown element type %T", e))
			}
		case bool:
			switch e := element.(type) {
			case string:
				return strconv.FormatBool(v), err
			case int:
				if v {
					return 1, err
				} else {
					return 0, err
				}
			case bool:
				return v, err
			case float64:
				if v {
					return 1.0, err
				} else {
					return 0.0, err
				}
			default:
				err = errors.New(fmt.Sprintf("ERROR: getSubelement(), unknown element type %T", e))
			}
		case float64:
			switch e := element.(type) {
			case string:
				return strconv.FormatFloat(v, 'E', -1, 64), err
			case int:
				return int(v), err
			case bool:
				return v >= 1, err
			case float64:
				return v, err
			default:
				err = errors.New(fmt.Sprintf("ERROR: getSubelement(), unknown element type %T", e))
			}
		default:
			err = errors.New(fmt.Sprintf("ERROR: getSubelement(), unknown value type %T", v))
		}
		return value, err
	}
	key := path[0]
	path = path[1:]
	result = element

	switch {
	case reflect.TypeOf(result).Kind() == reflect.Map:
		keyVal := reflect.ValueOf(key)
		sub := reflect.ValueOf(result).MapIndex(keyVal).Interface()
		val, err := setVarOnPath(sub, path, value)
		if err != nil {
			return result, err
		}
		reflect.ValueOf(result).SetMapIndex(keyVal, reflect.ValueOf(val))
	case reflect.TypeOf(result).Kind() == reflect.Slice:
		index, err := strconv.Atoi(key)
		if err != nil {
			return result, err
		}
		sub := reflect.ValueOf(result).Index(index).Interface()
		val, err := setVarOnPath(sub, path, value)
		if err != nil {
			return result, err
		}
		reflect.ValueOf(result).Index(index).Set(reflect.ValueOf(val))
	default:
		err = errors.New(fmt.Sprintf("ERROR: getSubelement(), unknown result type %T", element))
	}
	return
}
