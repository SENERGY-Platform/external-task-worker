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
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"strconv"
	"strings"
	"time"
)

func (this *CmdWorker) CreateProtocolMessage(command messages.Command, task messages.CamundaExternalTask) (topic string, key string, message string, err error) {
	value, err := this.createMessageForProtocolHandler(command, task)
	if err != nil {
		if this.config.Debug {
			log.Println("DEBUG:", task.TenantId)
		}
		log.Println("ERROR: on CreateProtocolMessage createMessageForProtocolHandler(): ", err)
		err = errors.New("internal format error (" + err.Error() + ") (time: " + util.TimeNow().String() + ")")
		return
	}
	topic = value.Metadata.Protocol.Handler
	msg, err := json.Marshal(value)
	return topic, value.Metadata.Device.Id, string(msg), err
}

func (this *CmdWorker) createMessageForProtocolHandler(command messages.Command, task messages.CamundaExternalTask) (result messages.ProtocolMsg, err error) {
	trace := []messages.Trace{{
		Timestamp: time.Now().UnixNano(),
		TimeUnit:  "unix_nano",
		Location:  "github.com/SENERGY-Platform/external-task-worker createMessageForProtocolHandler() start",
	}}
	device := command.Device
	service := command.Service
	protocol := command.Protocol
	token := devicerepository.Impersonate("")
	if device == nil || service == nil || protocol == nil {
		token, err = this.repository.GetToken(task.TenantId)
		if err != nil {
			return result, err
		}
	}
	if device == nil {
		temp, err := this.repository.GetDevice(token, command.DeviceId)
		if err != nil {
			log.Println("ERROR: unable to load device", command.DeviceId, task.TenantId, token)
			return result, err
		}
		device = &temp
	}
	if service == nil {
		temp, err := this.repository.GetService(token, *device, command.ServiceId)
		if err != nil {
			log.Println("ERROR: unable to load service", command.ServiceId, task.TenantId, token)
			return result, err
		}
		service = &temp
	}
	if protocol == nil {
		if command.ProtocolId == "" {
			command.ProtocolId = service.ProtocolId
		}
		temp, err := this.repository.GetProtocol(token, command.ProtocolId)
		if err != nil {
			log.Println("ERROR: unable to load protocol", command.ProtocolId, task.TenantId, token)
			return result, err
		}
		protocol = &temp
	}

	var inputCharacteristicId string
	var outputCharacteristicId string

	if isControllingFunction(command.Function) {
		inputCharacteristicId = command.CharacteristicId
	} else {
		outputCharacteristicId = command.CharacteristicId
		if service.Interaction == model.EVENT {
			return result, errors.New("command to measuring-function function with event interaction is not possible")
		}
	}

	marshalledInput := map[string]string{}
	if command.Version < 3 {
		marshalledInput, err = this.marshaller.MarshalFromServiceAndProtocol(inputCharacteristicId, *service, *protocol, command.Input, command.Configurables)
		if err != nil {
			return result, err
		}
	} else {
		data := []marshaller.MarshallingV2RequestData{}
		if len(command.InputPaths) > 0 {
			data = append(data, marshaller.MarshallingV2RequestData{
				Value:            command.Input,
				CharacteristicId: inputCharacteristicId,
				Paths:            command.InputPaths,
				FunctionId:       command.Function.Id,
			})
		}
		for _, configurable := range command.ConfigurablesV2 {
			data = append(data, marshaller.MarshallingV2RequestData{
				Value:            configurable.Value,
				CharacteristicId: configurable.CharacteristicId,
				Paths:            []string{configurable.Path},
				FunctionId:       configurable.FunctionId,
			})
		}
		marshalledInput, err = this.marshaller.MarshalV2(*service, *protocol, data)
		if err != nil {
			return result, err
		}
	}

	trace = append(trace, messages.Trace{
		Timestamp: time.Now().UnixNano(),
		TimeUnit:  "unix_nano",
		Location:  "github.com/SENERGY-Platform/external-task-worker createMessageForProtocolHandler() end",
	})

	result = messages.ProtocolMsg{
		TaskInfo: messages.TaskInfo{
			WorkerId:            this.camunda.GetWorkerId(),
			TaskId:              task.Id,
			ProcessInstanceId:   task.ProcessInstanceId,
			ProcessDefinitionId: task.ProcessDefinitionId,
			CompletionStrategy:  this.config.CompletionStrategy,
			Time:                strconv.FormatInt(util.TimeNow().Unix(), 10),
			TenantId:            task.TenantId,
		},
		Request: messages.ProtocolRequest{
			Input: marshalledInput,
		},
		Metadata: messages.Metadata{
			Device:               *device,
			Service:              *service,
			Protocol:             *protocol,
			InputCharacteristic:  inputCharacteristicId,
			OutputCharacteristic: outputCharacteristicId,
			ContentVariableHints: command.ContentVariableHints,
			ResponseTo:           this.config.MetadataResponseTo,
			ErrorTo:              this.config.MetadataErrorTo,
		},
		Trace: trace,
	}

	if command.Version >= 3 {
		result.Metadata.Version = command.Version
		result.Metadata.OutputPath = command.OutputPath
		result.Metadata.OutputFunctionId = command.Function.Id
		result.Metadata.OutputAspectNode = command.Aspect
	}

	return result, err
}

func isControllingFunction(function model.Function) bool {
	if function.RdfType == model.SES_ONTOLOGY_CONTROLLING_FUNCTION {
		return true
	}
	if strings.HasPrefix(function.Id, "urn:infai:ses:controlling-function:") {
		return true
	}
	return false
}
