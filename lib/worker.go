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
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository"
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"log"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/SENERGY-Platform/external-task-worker/util"

	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
)

type worker struct {
	consumer   kafka.ConsumerInterface
	producer   kafka.ProducerInterface
	repository devicerepository.RepoInterface
	camunda    camunda.CamundaInterface
	config     util.Config
}

func Worker(ctx context.Context, config util.Config, kafkaFactory kafka.FactoryInterface, repoFactory devicerepository.FactoryInterface, camundaFactory camunda.FactoryInterface) {
	log.Println("start camunda worker")
	var err error
	w := worker{config: config}

	if config.CompletionStrategy != util.OPTIMISTIC {
		w.consumer, err = kafkaFactory.NewConsumer(config, w.CompleteTask)
		if err != nil {
			log.Fatal("ERROR: kafkaFactory.NewConsumer", err)
		}
		defer w.consumer.Stop()
	}
	w.producer, err = kafkaFactory.NewProducer(config)
	if err != nil {
		log.Fatal("ERROR: kafkaFactory.NewProducer", err)
	}
	defer w.producer.Close()
	w.repository = repoFactory.Get(config)
	w.camunda = camundaFactory.Get(config, w.producer)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			wait := w.ExecuteNextTask()
			if wait {
				duration := time.Duration(config.CamundaWorkerTimeout) * time.Millisecond
				time.Sleep(duration)
			}
		}
	}
}

func (this *worker) ExecuteNextTask() (wait bool) {
	tasks, err := this.camunda.GetTask()
	if err != nil {
		log.Println("error on ExecuteNextTask getTask", err)
		return true
	}
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
	if this.config.Debug {
		log.Println("Start", task.Id, util.TimeNow().Second())
		log.Println("Get new Task: ", task)
	}
	if task.Error != "" {
		log.Println("WARNING: existing failure in camunda task", task.Error)
	}
	request, err := CreateCommandRequest(task)
	if err != nil {
		log.Println("error on CreateCommandRequest(): ", err)
		this.camunda.Error(task.Id, "invalid task format (json)")
		return
	}

	if request.Retries != -1 && request.Retries < task.Retries {
		this.camunda.Error(task.Id, "communication timeout")
		return
	}

	protocolTopic, message, err := this.CreateProtocolMessage(request, task)
	if err != nil {
		log.Println("error on ExecuteTask CreateProtocolMessage", err)
		this.camunda.Error(task.Id, err.Error())
		return
	}

	this.camunda.SetRetry(task.Id, task.Retries+1)
	this.producer.Produce(protocolTopic, message)

	if this.config.CompletionStrategy == util.OPTIMISTIC {
		err = this.camunda.CompleteTask(task.Id, this.camunda.GetWorkerId(), "", messages.Command{})
		if err != nil {
			log.Println("error on completeCamundaTask(): ", err)
			return
		} else {
			log.Println("Completed task optimistic.")
		}
	}
}

func (this *worker) CompleteTask(msg string) (err error) {
	var message messages.ProtocolMsg
	err = json.Unmarshal([]byte(msg), &message)
	if err != nil {
		return err
	}

	if this.missesCamundaDuration(message) {
		log.Println("WARNING: drop late response:", msg)
		return
	}

	response, err := CreateCommandResult(message)
	if err != nil {
		return err
	}
	err = this.camunda.CompleteTask(message.TaskInfo.TaskId, message.TaskInfo.WorkerId, this.config.CamundaTaskResultName, response)
	log.Println("Complete", message.TaskInfo.TaskId, util.TimeNow().Second())
	return
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
