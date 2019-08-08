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
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/SENERGY-Platform/external-task-worker/util"

	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/formatter-lib"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
)

const CAMUNDA_VARIABLES_PAYLOAD = "payload"

func ExecuteNextCamundaTask() (wait bool) {
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
			ExecuteCamundaTask(asyncTask)
		}(task)
	}
	wg.Wait()
	return false
}

func getPayloadParameter(task messages.CamundaTask) (result map[string]interface{}) {
	result = map[string]interface{}{}
	for key, value := range task.Variables {
		path := strings.SplitN(key, ".", 2)
		if len(path) == 2 && path[0] == "inputs" && path[1] != "" {
			result[path[1]] = value.Value
		}
	}
	return
}

func ToBpmnRequest(task messages.CamundaTask) (request messages.BpmnMsg, err error) {
	payload, ok := task.Variables[CAMUNDA_VARIABLES_PAYLOAD].Value.(string)
	if !ok {
		return request, errors.New(fmt.Sprint("ERROR: payload is not a string, ", task.Variables))
	}
	err = json.Unmarshal([]byte(payload), &request)
	if err != nil {
		return request, err
	}
	parameter := getPayloadParameter(task)
	err = setPayloadParameter(&request, parameter)
	return
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

func setPayloadParameter(msg *messages.BpmnMsg, parameter map[string]interface{}) (err error) {
	for paramName, value := range parameter {
		_, err := setVarOnPath(msg.Inputs, strings.Split(paramName, "."), value)
		if err != nil {
			log.Println("ERROR: setPayloadParameter() -> ignore param", paramName, value, err)
			//return err
		}
	}
	return
}

func ExecuteCamundaTask(task messages.CamundaTask) {
	log.Println("Start", task.Id, time.Now().Second())
	log.Println("Get new Task: ", task)
	if task.Error != "" {
		log.Println("WARNING: existing failure in camunda task", task.Error)
	}
	if util.Config.QosStrategy == "<=" && task.Retries == 1 {
		CamundaError(task, "communication timeout")
		return
	}
	request, err := ToBpmnRequest(task)
	if err != nil {
		log.Println("error on ToBpmnRequest(): ", err)
		CamundaError(task, "invalid task format (json)")
		return
	}

	protocolTopic, message, err := createKafkaCommandMessage(request, task)
	if err != nil {
		log.Println("error on ExecuteCamundaTask createKafkaCommandMessage", err)
		CamundaError(task, err.Error())
		return
	}
	if util.Config.QosStrategy == "<=" && task.Retries != 1 {
		SetCamundaRetry(task.Id)
	}
	Produce(protocolTopic, message)

	if util.Config.CompletionStrategy == "optimistic" {
		err = completeCamundaTask(task.Id, "", "", messages.BpmnMsg{})
		if err != nil {
			log.Println("error on completeCamundaTask(): ", err)
			return
		} else {
			log.Println("Completed task optimistic.")
		}
	}
}

type Envelope struct {
	DeviceId  string      `json:"device_id"`
	ServiceId string      `json:"service_id"`
	Value     interface{} `json:"value"`
}

func (envelope Envelope) Validate() error {
	if envelope.DeviceId == "" {
		return errors.New("missing device id")
	}
	if envelope.ServiceId == "" {
		return errors.New("missing service id")
	}
	return nil
}

func createKafkaCommandMessage(request messages.BpmnMsg, task messages.CamundaTask) (protocolTopic string, message string, err error) {
	instance, service, err := GetDeviceInfo(request.InstanceId, request.ServiceId, task.TenantId)
	if err != nil {
		log.Println("error on createKafkaCommandMessage getDeviceInfo: ", err)
		err = errors.New("unable to find device or service")
		return
	}
	value, err := createMessageForProtocolHandler(instance, service, request.Inputs, task)
	if err != nil {
		log.Println("ERROR: on createKafkaCommandMessage createMessageForProtocolHandler(): ", err)
		err = errors.New("internal format error (inconsistent data?) (time: " + time.Now().String() + ")")
		return
	}
	protocolTopic = service.Protocol.ProtocolHandlerUrl
	if protocolTopic == "" {
		log.Println("ERROR: empty protocol topic")
		log.Println("DEBUG: ", instance, service)
		err = errors.New("empty protocol topic")
		return
	}
	envelope := Envelope{ServiceId: service.Id, DeviceId: instance.Id, Value: value}
	if err := envelope.Validate(); err != nil {
		return protocolTopic, message, err
	}
	msg, err := json.Marshal(envelope)
	return protocolTopic, string(msg), err
}

func createMessageForProtocolHandler(instance model.DeviceInstance, service model.Service, inputs map[string]interface{}, task messages.CamundaTask) (result messages.ProtocolMsg, err error) {
	result = messages.ProtocolMsg{
		WorkerId:           GetWorkerId(),
		CompletionStrategy: util.Config.CompletionStrategy,
		DeviceUrl:          formatter_lib.UseDeviceConfig(instance.Config, instance.Url),
		ServiceUrl:         formatter_lib.UseDeviceConfig(instance.Config, service.Url),
		TaskId:             task.Id,
		DeviceInstanceId:   instance.Id,
		ServiceId:          service.Id,
		OutputName:         "result", //task.ActivityId,
		Time:               strconv.FormatInt(time.Now().Unix(), 10),
		Service:            service,
	}
	for _, serviceInput := range service.Input {
		for name, inputInterface := range inputs {
			if serviceInput.Name == name {
				input, err := formatter_lib.ParseFromJsonInterface(serviceInput.Type, inputInterface)
				if err != nil {
					return result, err
				}
				input.Name = name
				if err := formatter_lib.UseLiterals(&input, serviceInput.Type); err != nil {
					return result, err
				}
				formatedInput, err := formatter_lib.GetFormatedValue(instance.Config, serviceInput.Format, input, serviceInput.AdditionalFormatinfo)
				if err != nil {
					return result, err
				}
				result.ProtocolParts = append(result.ProtocolParts, messages.ProtocolPart{
					Name:  serviceInput.MsgSegment.Name,
					Value: formatedInput,
				})
			}
		}
	}
	return
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
	response, err := createBpmnResponse(nrMsg)
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

func createBpmnResponse(nrMsg messages.ProtocolMsg) (result messages.BpmnMsg, err error) {
	result.Outputs = map[string]interface{}{}
	result.ServiceId = nrMsg.ServiceId
	service := nrMsg.Service
	for _, output := range nrMsg.ProtocolParts {
		for _, serviceOutput := range service.Output {
			if serviceOutput.MsgSegment.Name == output.Name {
				parsedOutput, err := formatter_lib.ParseFormat(serviceOutput.Type, serviceOutput.Format, output.Value, serviceOutput.AdditionalFormatinfo)
				if err != nil {
					log.Println("error on parsing")
					return result, err
				}
				outputInterface, err := formatter_lib.FormatToJsonStruct([]model.ConfigField{}, parsedOutput)
				if err != nil {
					return result, err
				}
				parsedOutput.Name = serviceOutput.Name
				result.Outputs[serviceOutput.Name] = outputInterface
			}
		}
	}
	return
}
