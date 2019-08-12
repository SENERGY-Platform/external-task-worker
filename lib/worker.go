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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib/camunda"
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"github.com/SENERGY-Platform/external-task-worker/lib/repo"
	"log"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/SENERGY-Platform/external-task-worker/util"

	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
)

const CAMUNDA_VARIABLES_PAYLOAD = "payload"

type worker struct {
	consumer kafka.ConsumerInterface
	producer kafka.ProducerInterface
	repository repo.RepoInterface
	camunda camunda.CamundaInterface
	config util.ConfigType
}

func Worker(config util.ConfigType, kafkaFactory kafka.FactoryInterface, repoFactory repo.FactoryInterface, camundaFactory camunda.FactoryInterface) {
	log.Println("start camunda worker")
	var err error
	w := worker{config:config}

	w.consumer, err = kafkaFactory.NewConsumer(config, w.CompleteTask)
	if err != nil {
		log.Fatal("ERROR: kafkaFactory.NewConsumer", err)
	}
	defer w.consumer.Stop()
	w.producer, err = kafkaFactory.NewProducer(config)
	if err != nil {
		log.Fatal("ERROR: kafkaFactory.NewProducer", err)
	}
	defer w.producer.Close()
	w.repository = repoFactory.Get(config)
	w.camunda = camundaFactory.Get(config)
	for {
		wait := w.ExecuteNextTask()
		if wait {
			duration := time.Duration(config.CamundaWorkerTimeout) * time.Millisecond
			time.Sleep(duration)
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
		go func(asyncTask messages.CamundaTask) {
			defer wg.Done()
			this.ExecuteTask(asyncTask)
		}(task)
	}
	wg.Wait()
	return false
}

func (this *worker) ExecuteTask(task messages.CamundaTask) {
	log.Println("Start", task.Id, time.Now().Second())
	log.Println("Get new Task: ", task)
	if task.Error != "" {
		log.Println("WARNING: existing failure in camunda task", task.Error)
	}
	if this.config.QosStrategy == "<=" && task.Retries == 1 {
		this.camunda.Error(task, "communication timeout")
		return
	}
	request, err := CreateCommandRequest(task)
	if err != nil {
		log.Println("error on CreateCommandRequest(): ", err)
		this.camunda.Error(task, "invalid task format (json)")
		return
	}

	protocolTopic, message, err := this.CreateProtocolMessage(request, task)
	if err != nil {
		log.Println("error on ExecuteTask CreateProtocolMessage", err)
		this.camunda.Error(task, err.Error())
		return
	}
	if this.config.QosStrategy == "<=" && task.Retries != 1 {
		this.camunda.SetRetry(task.Id)
	}
	this.producer.Produce(protocolTopic, message)

	if this.config.CompletionStrategy == "optimistic" {
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
	var nrMsg messages.ProtocolMsg
	err = json.Unmarshal([]byte(msg), &nrMsg)
	if err != nil {
		return err
	}

	if nrMsg.CompletionStrategy == "optimistic" {
		return
	}

	if util.Config.QosStrategy == ">=" && missesCamundaDuration(nrMsg) {
		return
	}
	response, err := CreateCommandResult(nrMsg)
	if err != nil {
		return err
	}
	err = this.camunda.CompleteTask(nrMsg.TaskId, nrMsg.WorkerId, nrMsg.OutputName, response)
	log.Println("Complete", nrMsg.TaskId, time.Now().Second())
	return
}

func missesCamundaDuration(msg messages.ProtocolMsg) bool {
	if msg.Time == "" {
		return true
	}
	unixTime, err := strconv.ParseInt(msg.Time, 10, 64)
	if err != nil {
		return true
	}
	taskTime := time.Unix(unixTime, 0)
	return time.Since(taskTime) >= time.Duration(util.Config.CamundaFetchLockDuration)*time.Millisecond
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