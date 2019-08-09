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
	"github.com/SENERGY-Platform/external-task-worker/lib/kafka"
	"log"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/SENERGY-Platform/external-task-worker/util"

	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
)

const CAMUNDA_VARIABLES_PAYLOAD = "payload"

func CamundaWorker(factory kafka.FactoryInterface) {
	log.Println("start camunda worker")
	consumer, err := factory.NewConsumer(util.Config, CompleteCamundaTask)
	if err != nil {
		log.Fatal("ERROR: factory.NewConsumer", err)
	}
	defer consumer.Stop()
	producer, err := factory.NewProducer(util.Config)
	if err != nil {
		log.Fatal("ERROR: factory.NewProducer", err)
	}
	defer producer.Close()
	for {
		wait := ExecuteNextCamundaTask(producer)
		if wait {
			duration := time.Duration(util.Config.CamundaWorkerTimeout) * time.Millisecond
			time.Sleep(duration)
		}
	}
}

func ExecuteNextCamundaTask(producer kafka.ProducerInterface) (wait bool) {
	tasks, err := GetCamundaTask()
	if err != nil {
		log.Println("error on ExecuteNextCamundaTask getTask", err)
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
			ExecuteCamundaTask(asyncTask, producer)
		}(task)
	}
	wg.Wait()
	return false
}

func ExecuteCamundaTask(task messages.CamundaTask, producer kafka.ProducerInterface) {
	log.Println("Start", task.Id, time.Now().Second())
	log.Println("Get new Task: ", task)
	if task.Error != "" {
		log.Println("WARNING: existing failure in camunda task", task.Error)
	}
	if util.Config.QosStrategy == "<=" && task.Retries == 1 {
		CamundaError(task, "communication timeout")
		return
	}
	request, err := SenergyRequestTask(task)
	if err != nil {
		log.Println("error on SenergyRequestTask(): ", err)
		CamundaError(task, "invalid task format (json)")
		return
	}

	protocolTopic, message, err := CreateProtocolMessage(request, task)
	if err != nil {
		log.Println("error on ExecuteCamundaTask CreateProtocolMessage", err)
		CamundaError(task, err.Error())
		return
	}
	if util.Config.QosStrategy == "<=" && task.Retries != 1 {
		SetCamundaRetry(task.Id)
	}
	producer.Produce(protocolTopic, message)

	if util.Config.CompletionStrategy == "optimistic" {
		err = completeCamundaTask(task.Id, "", "", messages.SenergyTask{})
		if err != nil {
			log.Println("error on completeCamundaTask(): ", err)
			return
		} else {
			log.Println("Completed task optimistic.")
		}
	}
}

func CompleteCamundaTask(msg string) (err error) {
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
	response, err := SenergyResultTask(nrMsg)
	if err != nil {
		return err
	}
	err = completeCamundaTask(nrMsg.TaskId, nrMsg.WorkerId, nrMsg.OutputName, response)
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