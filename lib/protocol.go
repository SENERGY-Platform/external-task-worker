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
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"strconv"
)

func (this *worker) CreateProtocolMessage(command messages.Command, task messages.CamundaExternalTask) (topic string, message string, err error) {
	value, err := this.createMessageForProtocolHandler(command, task)
	if err != nil {
		log.Println("ERROR: on CreateProtocolMessage createMessageForProtocolHandler(): ", err)
		err = errors.New("internal format error (inconsistent data?) (time: " + util.TimeNow().String() + ")")
		return
	}
	topic = value.Metadata.Protocol.Handler
	msg, err := json.Marshal(value)
	return topic, string(msg), err
}

func (this *worker) createMessageForProtocolHandler(command messages.Command, task messages.CamundaExternalTask) (result messages.ProtocolMsg, err error) {
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
			return result, err
		}
		device = &temp
	}
	if service == nil {
		temp, err := this.repository.GetService(token, *device, command.ServiceId)
		if err != nil {
			return result, err
		}
		service = &temp
	}
	if protocol == nil {
		temp, err := this.repository.GetProtocol(token, command.ProtocolId)
		if err != nil {
			return result, err
		}
		protocol = &temp
	}

	var inputCharacteristicId string
	var outputCharacteristicId string

	if command.Function.RdfType == model.SES_ONTOLOGY_CONTROLLING_FUNCTION {
		inputCharacteristicId = command.CharacteristicId
	} else {
		outputCharacteristicId = command.CharacteristicId
	}

	marshalledInput, err := marshaller.MarshalInputs(*protocol, *service, command.Input, inputCharacteristicId)
	if err != nil {
		return result, err
	}

	result = messages.ProtocolMsg{
		TaskInfo: messages.TaskInfo{
			WorkerId:            this.camunda.GetWorkerId(),
			TaskId:              task.Id,
			ProcessInstanceId:   task.ProcessInstanceId,
			ProcessDefinitionId: task.ProcessDefinitionId,
			CompletionStrategy:  this.config.CompletionStrategy,
			Time:                strconv.FormatInt(util.TimeNow().Unix(), 10),
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
		},
	}

	return result, err
}
