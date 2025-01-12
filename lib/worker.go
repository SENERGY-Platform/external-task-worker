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
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda/interfaces"
	"github.com/SENERGY-Platform/external-task-worker/lib/com"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicegroups"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	"github.com/SENERGY-Platform/external-task-worker/lib/prometheus"
	"github.com/SENERGY-Platform/external-task-worker/lib/timescale"
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

type CmdWorker struct {
	producer                  com.ProducerInterface
	repository                devicerepository.RepoInterface
	camunda                   interfaces.CamundaInterface
	config                    util.Config
	marshaller                marshaller.Interface
	camundaCallMux            sync.Mutex
	lastSuccessfulCamundaCall time.Time
	producerCallMux           sync.Mutex
	lastProducerCallSuccess   bool
	deviceGroupsHandler       DeviceGroupsHandler
	timescale                 Timescale
	metrics                   Metrics
}

type Metrics interface {
	interfaces.Metrics
	HandleResponseTrace(trace []messages.Trace)
}

type Timescale interface {
	Query(token string, request []messages.TimescaleRequest, timeout time.Duration) (result []messages.TimescaleResponse, err error)
}

type DeviceGroupsHandler interface {
	ProcessCommand(request messages.Command, task messages.CamundaExternalTask, caller string) (completed bool, missingRequests messages.RequestInfoList, finishedResults []interface{}, err error)
	ProcessResponse(taskId string, subResult interface{}) (parent messages.GroupTaskMetadataElement, results []interface{}, finished bool, err error)
}

func Worker(ctx context.Context, config util.Config, comFactory com.FactoryInterface, repoFactory devicerepository.FactoryInterface, camundaFactory interfaces.FactoryInterface, marshallerFactory marshaller.FactoryInterface, timescaleFactory timescale.FactoryInterface) {
	log.Println("start camunda worker")
	w, err := New(ctx, config, comFactory, repoFactory, camundaFactory, marshallerFactory, timescaleFactory)
	if err != nil {
		log.Fatal("FATAL-ERROR:", err)
	}
	StartHealthCheckEndpoint(ctx, config, w)
	err = w.Loop(ctx)
	if err != nil {
		log.Fatal("FATAL-ERROR:", err)
	}
}

func New(ctx context.Context, config util.Config, comFactory com.FactoryInterface, repoFactory devicerepository.FactoryInterface, camundaFactory interfaces.FactoryInterface, marshallerFactory marshaller.FactoryInterface, timescaleFactory timescale.FactoryInterface) (w *CmdWorker, err error) {
	metrics, err := prometheus.Start(ctx, config)
	if err != nil {
		return w, fmt.Errorf("prometheus.Start %w", err)
	}

	w = &CmdWorker{
		config:                    config,
		marshaller:                marshallerFactory.New(ctx, config.MarshallerUrl),
		lastProducerCallSuccess:   true,
		lastSuccessfulCamundaCall: time.Now(),
		metrics:                   metrics,
	}

	w.timescale, err = timescaleFactory(ctx, config)
	if err != nil {
		return w, fmt.Errorf("comFactory.NewProducer %w", err)
	}

	if config.CompletionStrategy != util.OPTIMISTIC {
		if config.ResponseWorkerCount > 1 {
			err = comFactory.NewConsumer(ctx, config, w.GetQueuedResponseHandler(ctx, config.ResponseWorkerCount, config.ResponseWorkerCount), w.ErrorMessageHandler)
		} else {
			err = comFactory.NewConsumer(ctx, config, w.HandleTaskResponse, w.ErrorMessageHandler)
		}
		if err != nil {
			return w, fmt.Errorf("comFactory.NewConsumer %w", err)
		}
	}
	w.producer, err = comFactory.NewProducer(ctx, config)
	if err != nil {
		return w, fmt.Errorf("comFactory.NewProducer %w", err)
	}
	w.repository, err = repoFactory.Get(config)
	if err != nil {
		return w, fmt.Errorf("repoFactory.Get %w", err)
	}
	w.camunda, err = camundaFactory.Get(config, w.producer, metrics)
	if err != nil {
		return w, fmt.Errorf("camundaFactory.Get %w", err)
	}
	filterEvents := config.TimescaleWrapperUrl == "" || config.TimescaleWrapperUrl == "-"
	w.deviceGroupsHandler = devicegroups.New(config.GroupScheduler, w.camunda, w.repository, w.CreateProtocolMessage, config.CamundaFetchLockDuration, config.SubResultExpirationInSeconds, config.SubResultDatabaseUrls, config.MemcachedTimeout, config.MemcachedMaxIdleConns, filterEvents)

	return w, nil
}

func (w *CmdWorker) Loop(ctx context.Context) error {
	tasks, errChan, err := w.camunda.ProvideTasks(ctx)
	if err != nil {
		return err
	}
	go func() {
		for err := range errChan {
			log.Println("error from ProvideTasks():", err)
		}
	}()
	workerCount := 10
	wg := &sync.WaitGroup{}
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range tasks {
				w.ExecuteTaskBatch(batch)
			}
		}()
	}
	wg.Wait()
	return nil
}

const DuplicateActivitySleep = 100 * time.Millisecond

func (this *CmdWorker) ExecuteTaskBatch(tasks []messages.CamundaExternalTask) {
	this.logSuccessfulCamundaCall()
	if len(tasks) == 0 {
		return
	}

	for _, task := range tasks {
		this.metrics.LogTaskReceived(task)
	}

	tasksWithDuplicateIndex := []messages.TaskWithDuplicateIndex{}
	activityIndex := map[string][]string{}
	for _, task := range tasks {
		taskWithStrategy := messages.TaskWithDuplicateIndex{
			Task:           task,
			DuplicateIndex: len(activityIndex[task.ActivityId]),
		}
		activityIndex[task.ActivityId] = append(activityIndex[task.ActivityId], task.Id)
		tasksWithDuplicateIndex = append(tasksWithDuplicateIndex, taskWithStrategy)
	}

	wg := sync.WaitGroup{}
	for _, task := range tasksWithDuplicateIndex {
		wg.Add(1)
		go func(asyncTask messages.CamundaExternalTask, duplicateIndex int) {
			defer wg.Done()
			if duplicateIndex > 0 {
				time.Sleep(time.Duration(duplicateIndex) * DuplicateActivitySleep)
			}
			this.ExecuteTask(asyncTask, util.CALLER_CAMUNDA_LOOP)
		}(task.Task, task.DuplicateIndex)
	}
	wg.Wait()
	return
}

func (this *CmdWorker) ExecuteTask(task messages.CamundaExternalTask, caller string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR: ", r, "\n", string(debug.Stack()))
		}
	}()
	if this.config.Debug {
		log.Println("Start", task.Id, util.TimeNow().Second())
		log.Println("Get new Task: ", task)
	}
	if task.Error != "" {
		log.Println("WARNING: existing failure in camunda task", task.Error)
	}
	incident := GetIncident(task)
	if incident != nil {
		this.camunda.Error(task.Id, task.ProcessInstanceId, task.ProcessDefinitionId, "user triggered incident: "+*incident, task.TenantId)
		return
	}
	request, err := GetCommandRequest(task)
	if err != nil {
		log.Println("error on GetCommandRequest(): ", err)
		this.camunda.Error(task.Id, task.ProcessInstanceId, task.ProcessDefinitionId, "invalid task format (json)", task.TenantId)
		return
	}

	this.ExecuteCommand(request, task, caller)
}

func (this *CmdWorker) ExecuteCommand(command messages.Command, task messages.CamundaExternalTask, caller string) {
	if this.config.CompletionStrategy == util.OPTIMISTIC {
		command.Retries = -1
	}
	completed, nextProtocolMessages, results, err := this.deviceGroupsHandler.ProcessCommand(command, task, caller)
	if err != nil {
		this.camunda.Error(task.Id, task.ProcessInstanceId, task.ProcessDefinitionId, err.Error(), task.TenantId)
		return
	}

	if this.config.CompletionStrategy == util.OPTIMISTIC || completed {
		time.Sleep(time.Duration(this.config.OptimisticTaskCompletionTimeout) * time.Millisecond) //prevent completes that are too fast
		taskInfo := messages.TaskInfo{
			WorkerId:            this.camunda.GetWorkerId(),
			TaskId:              task.Id,
			ProcessInstanceId:   task.ProcessInstanceId,
			ProcessDefinitionId: task.ProcessDefinitionId,
			TenantId:            task.TenantId,
		}
		this.metrics.LogTaskCompleted(taskInfo)
		err = this.camunda.CompleteTask(taskInfo, this.config.CamundaTaskResultName, handleVersioning(command.Version, results))
		if err != nil {
			log.Println("error on completeCamundaTask(): ", err)
			return
		}
		if this.config.CompletionStrategy != util.OPTIMISTIC {
			//if optimistic: we have still to send the commands
			//else: this is not the first try, and now we are finished
			return
		}
	}

	if this.config.Debug {
		temp, _ := json.Marshal(nextProtocolMessages)
		log.Println("DEBUG: nextProtocolMessages:\n", string(temp))
	}
	for _, message := range nextProtocolMessages {
		if message.Request != nil {
			this.metrics.LogTaskCommandSend(message.Metadata)
			err = this.producer.ProduceWithKey(message.Request.Topic, message.Request.Key, message.Request.Payload)
			if err != nil {
				log.Println("ERROR: unable to produce kafka msg", err)
				this.logProducerError()
			} else {
				this.logProducerSuccess()
			}
		} else if message.Event != nil && this.config.TimescaleWrapperUrl != "" && this.config.TimescaleWrapperUrl != "-" {
			this.metrics.LogTaskLastEventValueRequest(message.Metadata)
			token, err := this.repository.GetToken(message.Metadata.Task.TenantId)
			if err != nil {
				log.Println("ERROR: unable to get token", err)
				continue
			}
			code, result := this.GetLastEventValue(string(token), message.Metadata.Task.TenantId, message.Event.Device, message.Event.Service, message.Event.Protocol, message.Event.CharacteristicId, message.Event.FunctionId, message.Event.AspectNode, 10*time.Second)
			if code == 200 {
				err = this.handleTaskResponse(messages.TaskInfo{
					WorkerId:            this.camunda.GetWorkerId(),
					TaskId:              message.Metadata.Task.Id,
					ProcessInstanceId:   task.ProcessInstanceId,
					ProcessDefinitionId: task.ProcessDefinitionId,
					CompletionStrategy:  this.config.CompletionStrategy,
					Time:                strconv.FormatInt(util.TimeNow().Unix(), 10),
					TenantId:            task.TenantId,
				}, result)
				if err != nil {
					log.Println("ERROR: unable to handle GetLastEventValue() as task response", err)
				}
			} else {
				log.Println("ERROR: GetLastEventValue()", code, result)
			}
		} else {
			temp, _ := json.Marshal(message)
			log.Println("WARNING: unhandled command", string(temp))
		}
	}
}

func (this *CmdWorker) GetQueuedResponseHandler(ctx context.Context, workerCount int64, queueSize int64) func(msg string) (err error) {
	queue := make(chan string, queueSize)
	for i := int64(0); i < workerCount; i++ {
		go func() {
			for msg := range queue {
				err := this.HandleTaskResponse(msg)
				if err != nil {
					log.Println("ERROR: ", err)
				}
			}
		}()
	}
	go func() {
		<-ctx.Done()
		close(queue)
	}()
	return func(msg string) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = errors.New(fmt.Sprint(r))
			}
		}()
		queue <- msg
		return err
	}
}

func (this *CmdWorker) HandleTaskResponse(msg string) (err error) {
	var message messages.ProtocolMsg
	err = json.Unmarshal([]byte(msg), &message)
	if err != nil {
		return err
	}
	defer func() {
		this.metrics.HandleResponseTrace(message.Trace)
		if this.config.Debug {
			log.Println("TRACE: ", message.Trace)
			if len(message.Trace) >= 2 {
				log.Println("TRACE-DURATION:", time.Unix(0, message.Trace[len(message.Trace)-1].Timestamp).Sub(time.Unix(0, message.Trace[0].Timestamp)).String())
			}
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
	if message.Metadata.OutputCharacteristic != model.NullCharacteristic.Id && message.Metadata.OutputCharacteristic != "" {
		if message.Metadata.Version < 3 {
			output, err = this.marshaller.UnmarshalFromServiceAndProtocol(message.Metadata.OutputCharacteristic, message.Metadata.Service, message.Metadata.Protocol, message.Response.Output, message.Metadata.ContentVariableHints)
		} else {
			aspect := model.AspectNode{}
			if message.Metadata.OutputAspectNode != nil {
				aspect = *message.Metadata.OutputAspectNode
			}
			marshalStartTime := time.Now()
			output, err = this.marshaller.UnmarshalV2(marshaller.UnmarshallingV2Request{
				Service:          message.Metadata.Service,
				Protocol:         message.Metadata.Protocol,
				CharacteristicId: message.Metadata.OutputCharacteristic,
				Message:          message.Response.Output,
				Path:             message.Metadata.OutputPath,
				FunctionId:       message.Metadata.OutputFunctionId,
				AspectNode:       aspect,
			})
			if err == nil {
				//log marshal latency
				marshalDuration := time.Since(marshalStartTime)
				this.metrics.LogTaskMarshallingLatency("UnmarshalV2", message.TaskInfo.TenantId, message.Metadata.Service.Id, message.Metadata.OutputFunctionId, marshalDuration)
			}
		}
	}
	if err != nil {
		return err
	}
	return this.handleTaskResponse(message.TaskInfo, output)
}

func (this *CmdWorker) handleTaskResponse(taskInfo messages.TaskInfo, output interface{}) (err error) {
	this.metrics.LogTaskCommandResponseReceived(taskInfo)

	parent, results, finished, err := this.deviceGroupsHandler.ProcessResponse(taskInfo.TaskId, output)
	if err == devicegroups.ErrNotFount {
		log.Println("WARNING: no parent task found --> unable to complete task (tasks may already be done):", taskInfo.TaskId, output)
		return nil //if parent task is not found -> can not be finished (may already be done)
	}
	if err != nil {
		return err
	}

	if finished {
		taskInfo = messages.TaskInfo{
			WorkerId:            taskInfo.WorkerId,
			TaskId:              parent.Task.Id,
			ProcessInstanceId:   parent.Task.ProcessInstanceId,
			ProcessDefinitionId: parent.Task.ProcessDefinitionId,
			CompletionStrategy:  taskInfo.CompletionStrategy,
			Time:                taskInfo.Time,
			TenantId:            taskInfo.TenantId,
		}

		this.metrics.LogTaskCompleted(taskInfo)

		err = this.camunda.CompleteTask(
			taskInfo,
			this.config.CamundaTaskResultName,
			handleVersioning(parent.Command.Version, results))
		if err != nil {
			log.Println("ERROR: handleTaskResponse()", err.Error())
		}
		if this.config.Debug {
			log.Println("Complete", parent.Task.Id, util.TimeNow().Second(), output, err)
		}
	} else if this.isSequential() {
		this.ExecuteTask(parent.Task, util.CALLER_RESPONSE)
	}
	return
}

func (this *CmdWorker) isSequential() bool {
	return this.config.GroupScheduler == util.SEQUENTIAL || this.config.GroupScheduler == util.ROUND_ROBIN
}

func handleVersioning(version int64, results []interface{}) interface{} {
	if version < 2 && len(results) > 0 {
		return results[0]
	} else {
		return results
	}
}

func (this *CmdWorker) missesCamundaDuration(msg messages.ProtocolMsg) bool {
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

func (this *CmdWorker) GetState() WorkerState {
	return WorkerState{
		ProducerOk:                this.LastProducerCallSuccessful(),
		LastSuccessfulCamundaCall: this.SinceLastSuccessfulCamundaCall(),
	}
}

func (this *CmdWorker) SinceLastSuccessfulCamundaCall() time.Duration {
	this.camundaCallMux.Lock()
	defer this.camundaCallMux.Unlock()
	return time.Since(this.lastSuccessfulCamundaCall)
}

func (this *CmdWorker) logSuccessfulCamundaCall() {
	this.camundaCallMux.Lock()
	defer this.camundaCallMux.Unlock()
	this.lastSuccessfulCamundaCall = time.Now()
}

func (this *CmdWorker) LastProducerCallSuccessful() bool {
	this.producerCallMux.Lock()
	defer this.producerCallMux.Unlock()
	return this.lastProducerCallSuccess
}

func (this *CmdWorker) logProducerError() {
	this.producerCallMux.Lock()
	defer this.producerCallMux.Unlock()
	this.lastProducerCallSuccess = false
}

func (this *CmdWorker) logProducerSuccess() {
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
