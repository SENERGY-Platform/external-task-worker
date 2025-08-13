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
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

// CreateProtocolMessage implements: devicegroups.Callback
func (this *CmdWorker) CreateProtocolMessage(command messages.Command, task messages.CamundaExternalTask) (request *messages.KafkaMessage, event *messages.EventRequest, err error) {
	value, event, err := this.createMessageForProtocolHandler(command, task)
	if err != nil {
		if this.config.Debug {
			log.Println("DEBUG:", task.TenantId)
		}
		log.Println("ERROR: on CreateProtocolMessage createMessageForProtocolHandler(): ", err)
		err = errors.New("internal format error: " + err.Error())
		return
	}
	if value != nil {
		topic := value.Metadata.Protocol.Handler
		var msg []byte
		msg, err = json.Marshal(value)
		request = &messages.KafkaMessage{
			Topic:   topic,
			Key:     value.Metadata.Device.Id,
			Payload: string(msg),
		}
	}

	return request, event, err
}

func (this *CmdWorker) createMessageForProtocolHandler(command messages.Command, task messages.CamundaExternalTask) (taskRequest *messages.ProtocolMsg, eventRequest *messages.EventRequest, err error) {
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
			return nil, nil, err
		}
	}
	if device == nil {
		temp, err := this.repository.GetDevice(token, command.DeviceId)
		if err != nil {
			log.Println("ERROR: unable to load device", command.DeviceId, task.TenantId, token)
			return nil, nil, err
		}
		device = &temp
	}
	if service == nil {
		temp, err := this.repository.GetService(token, *device, command.ServiceId)
		if err != nil {
			log.Println("ERROR: unable to load service", command.ServiceId, task.TenantId, token)
			return nil, nil, err
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
			return nil, nil, err
		}
		protocol = &temp
	}

	var inputCharacteristicId string
	var outputCharacteristicId string

	if isControllingFunction(command.Function) {
		inputCharacteristicId = command.CharacteristicId
	} else {
		outputCharacteristicId = command.CharacteristicId
		aspect := model.AspectNode{}
		if command.Aspect != nil {
			aspect = *command.Aspect
		}
		if service.Interaction == model.EVENT || (service.Interaction == model.EVENT_AND_REQUEST && command.PreferEvent) {
			return nil, &messages.EventRequest{
				Device:           *device,
				Service:          *service,
				Protocol:         *protocol,
				CharacteristicId: command.CharacteristicId,
				FunctionId:       command.Function.Id,
				AspectNode:       aspect,
			}, nil
		}
	}

	marshalledInput := map[string]string{}
	if command.Version < 3 {
		marshalledInput, err = this.marshaller.MarshalFromServiceAndProtocol(inputCharacteristicId, *service, *protocol, command.Input, command.Configurables)
		if err != nil {
			return nil, nil, err
		}
	} else {
		data := []marshaller.MarshallingV2RequestData{}
		somethingToSet := command.Function.ConceptId != "" || command.Input != nil
		if somethingToSet && (len(command.InputPaths) > 0 || isControllingFunction(command.Function)) {
			data = append(data, marshaller.MarshallingV2RequestData{
				Value:            command.Input,
				CharacteristicId: inputCharacteristicId,
				Paths:            command.InputPaths,
				FunctionId:       command.Function.Id,
				AspectNode:       command.Aspect,
			})
		}
		for _, configurable := range command.ConfigurablesV2 {
			var aspect *model.AspectNode
			if configurable.AspectNode.Id != "" {
				temp := configurable.AspectNode
				aspect = &temp
			}
			paths := []string{}
			if configurable.Path != "" {
				paths = []string{configurable.Path}
			}
			data = append(data, marshaller.MarshallingV2RequestData{
				Value:            configurable.Value,
				CharacteristicId: configurable.CharacteristicId,
				Paths:            paths,
				FunctionId:       configurable.FunctionId,
				AspectNode:       aspect,
			})
		}
		marshalStartTime := time.Now()
		marshalledInput, err = this.marshaller.MarshalV2(*service, *protocol, data)
		if err != nil {
			return nil, nil, err
		}

		//log marshal latency
		marshalDuration := time.Since(marshalStartTime)
		functionIds := []string{}
		for _, d := range data {
			if d.FunctionId != "" {
				functionIds = append(functionIds, d.FunctionId)
			}
		}
		sort.Strings(functionIds)
		this.metrics.LogTaskMarshallingLatency("MarshalV2", task.TenantId, service.Id, strings.Join(functionIds, ","), marshalDuration)
	}

	trace = append(trace, messages.Trace{
		Timestamp: time.Now().UnixNano(),
		TimeUnit:  "unix_nano",
		Location:  "github.com/SENERGY-Platform/external-task-worker createMessageForProtocolHandler() end",
	})

	taskRequest = &messages.ProtocolMsg{
		TaskInfo: messages.TaskInfo{
			WorkerId:            this.camunda.GetWorkerId(),
			TaskId:              task.Id,
			ProcessInstanceId:   task.ProcessInstanceId,
			BusinessKey:         task.BusinessKey,
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
		taskRequest.Metadata.Version = command.Version
		if outputCharacteristicId != "" {
			taskRequest.Metadata.OutputPath = command.OutputPath
			taskRequest.Metadata.OutputFunctionId = command.Function.Id
			taskRequest.Metadata.OutputAspectNode = command.Aspect
		}
	}

	return taskRequest, nil, err
}

func isControllingFunction(function model.Function) bool {
	if function.RdfType == model.SES_ONTOLOGY_CONTROLLING_FUNCTION {
		return true
	}
	if strings.HasPrefix(function.Id, model.CONTROLLING_FUNCTION_PREFIX) {
		return true
	}
	return false
}
