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

//TODO: replace with keycloak
const adminjwt = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICIzaUtabW9aUHpsMmRtQnBJdS1vSkY4ZVVUZHh4OUFIckVOcG5CcHM5SjYwIn0.eyJqdGkiOiJiOGUyNGZkNy1jNjJlLTRhNWQtOTQ4ZC1mZGI2ZWVkM2JmYzYiLCJleHAiOjE1MzA1MzIwMzIsIm5iZiI6MCwiaWF0IjoxNTMwNTI4NDMyLCJpc3MiOiJodHRwczovL2F1dGguc2VwbC5pbmZhaS5vcmcvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiZnJvbnRlbmQiLCJzdWIiOiJkZDY5ZWEwZC1mNTUzLTQzMzYtODBmMy03ZjQ1NjdmODVjN2IiLCJ0eXAiOiJCZWFyZXIiLCJhenAiOiJmcm9udGVuZCIsIm5vbmNlIjoiMjJlMGVjZjgtZjhhMS00NDQ1LWFmMjctNGQ1M2JmNWQxOGI5IiwiYXV0aF90aW1lIjoxNTMwNTI4NDIzLCJzZXNzaW9uX3N0YXRlIjoiMWQ3NWE5ODQtNzM1OS00MWJlLTgxYjktNzMyZDgyNzRjMjNlIiwiYWNyIjoiMCIsImFsbG93ZWQtb3JpZ2lucyI6WyIqIl0sInJlYWxtX2FjY2VzcyI6eyJyb2xlcyI6WyJjcmVhdGUtcmVhbG0iLCJhZG1pbiIsImRldmVsb3BlciIsInVtYV9hdXRob3JpemF0aW9uIiwidXNlciJdfSwicmVzb3VyY2VfYWNjZXNzIjp7Im1hc3Rlci1yZWFsbSI6eyJyb2xlcyI6WyJ2aWV3LWlkZW50aXR5LXByb3ZpZGVycyIsInZpZXctcmVhbG0iLCJtYW5hZ2UtaWRlbnRpdHktcHJvdmlkZXJzIiwiaW1wZXJzb25hdGlvbiIsImNyZWF0ZS1jbGllbnQiLCJtYW5hZ2UtdXNlcnMiLCJxdWVyeS1yZWFsbXMiLCJ2aWV3LWF1dGhvcml6YXRpb24iLCJxdWVyeS1jbGllbnRzIiwicXVlcnktdXNlcnMiLCJtYW5hZ2UtZXZlbnRzIiwibWFuYWdlLXJlYWxtIiwidmlldy1ldmVudHMiLCJ2aWV3LXVzZXJzIiwidmlldy1jbGllbnRzIiwibWFuYWdlLWF1dGhvcml6YXRpb24iLCJtYW5hZ2UtY2xpZW50cyIsInF1ZXJ5LWdyb3VwcyJdfSwiYWNjb3VudCI6eyJyb2xlcyI6WyJtYW5hZ2UtYWNjb3VudCIsIm1hbmFnZS1hY2NvdW50LWxpbmtzIiwidmlldy1wcm9maWxlIl19fSwicm9sZXMiOlsidW1hX2F1dGhvcml6YXRpb24iLCJhZG1pbiIsImNyZWF0ZS1yZWFsbSIsImRldmVsb3BlciIsInVzZXIiLCJvZmZsaW5lX2FjY2VzcyJdLCJuYW1lIjoiZGYgZGZmZmYiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJzZXBsIiwiZ2l2ZW5fbmFtZSI6ImRmIiwiZmFtaWx5X25hbWUiOiJkZmZmZiIsImVtYWlsIjoic2VwbEBzZXBsLmRlIn0.eOwKV7vwRrWr8GlfCPFSq5WwR_p-_rSJURXCV1K7ClBY5jqKQkCsRL2V4YhkP1uS6ECeSxF7NNOLmElVLeFyAkvgSNOUkiuIWQpMTakNKynyRfH0SrdnPSTwK2V1s1i4VjoYdyZWXKNjeT2tUUX9eCyI5qOf_Dzcai5FhGCSUeKpV0ScUj5lKrn56aamlW9IdmbFJ4VwpQg2Y843Vc0TqpjK9n_uKwuRcQd9jkKHkbwWQ-wyJEbFWXHjQ6LnM84H0CQ2fgBqPPfpQDKjGSUNaCS-jtBcbsBAWQSICwol95BuOAqVFMucx56Wm-OyQOuoQ1jaLt2t-Uxtr-C9wKJWHQ"

func (this *worker) createMessageForProtocolHandler(command messages.Command, task messages.CamundaExternalTask) (result messages.ProtocolMsg, err error) {
	device := command.Device
	service := command.Service
	protocol := command.Protocol
	token := devicerepository.Impersonate(adminjwt)
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
