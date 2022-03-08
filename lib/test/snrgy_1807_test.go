/*
 * Copyright 2020 InfAI (CC SES)
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

package test

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/converter/lib/converter/characteristics"
	"github.com/SENERGY-Platform/external-task-worker/lib"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestCommand(t *testing.T) {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.OPTIMISTIC
	config.CamundaWorkerTimeout = 100

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller)

	time.Sleep(1 * time.Second)

	//populate repository
	device := model.Device{
		Id:           "device_1",
		Name:         "d1",
		DeviceTypeId: "dt1",
		LocalId:      "d1u",
	}
	mock.Repo.RegisterDevice(device)

	protocol := model.Protocol{
		Id:               "p1",
		Name:             "protocol1",
		Handler:          "protocol1",
		ProtocolSegments: []model.ProtocolSegment{{Id: "ms1", Name: "body"}},
	}
	mock.Repo.RegisterProtocol(protocol)

	service := model.Service{
		Id:         "service_1",
		Name:       "s1",
		LocalId:    "s1u",
		ProtocolId: "p1",
		Inputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:               "level",
							Name:             "level",
							Type:             model.Integer,
							CharacteristicId: example.Hex,
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	}
	mock.Repo.RegisterService(service)

	cmd1 := messages.Command{
		Version:          3,
		Function:         model.Function{Id: model.CONTROLLING_FUNCTION_PREFIX + "f"},
		Aspect:           nil,
		CharacteristicId: example.Rgb,
		InputPaths:       []string{"metrics.level"},
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
		Input: map[string]float64{
			"r": 200,
			"g": 50,
			"b": 0,
		},
	}

	cmdMsg1, err := json.Marshal(cmd1)
	if err != nil {
		log.Fatal(err)
	}

	cmd2 := messages.Command{
		Version:          3,
		Function:         model.Function{Id: model.CONTROLLING_FUNCTION_PREFIX + "f"},
		CharacteristicId: example.Hex,
		InputPaths:       []string{"metrics.level"},
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
		Input:            "#ff0064",
	}

	cmdMsg2, err := json.Marshal(cmd2)
	if err != nil {
		log.Fatal(err)
	}

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "1",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg1),
			},
		},
	})

	time.Sleep(1 * time.Second)

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "2",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg1),
			},
			"inputs.b": {Value: "255"},
		},
	})

	time.Sleep(1 * time.Second)

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "3",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg2),
			},
		},
	})

	time.Sleep(1 * time.Second)

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "4",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg2),
			},
			"inputs": {Value: "\"#ff00ff\""},
		},
	})

	time.Sleep(1 * time.Second)

	expectedProtocolMessages := []messages.ProtocolMsg{
		{
			Request: messages.ProtocolRequest{
				Input: map[string]string{"body": "{\"level\":\"#c83200\"}"},
			},
			Response: messages.ProtocolResponse{},
			TaskInfo: messages.TaskInfo{
				WorkerId:            "workerid",
				TaskId:              "1",
				ProcessInstanceId:   "",
				ProcessDefinitionId: "",
				CompletionStrategy:  "optimistic",
				Time:                "-62135596800",
				TenantId:            "user",
			},
			Metadata: messages.Metadata{
				Version:              3,
				Device:               device,
				Service:              service,
				Protocol:             protocol,
				OutputPath:           "",
				OutputFunctionId:     "",
				OutputAspectNode:     nil,
				InputCharacteristic:  example.Rgb,
				OutputCharacteristic: "",
				ContentVariableHints: nil,
				ResponseTo:           "response",
				ErrorTo:              "errors",
			},
			Trace: nil,
		},
		{
			Request: messages.ProtocolRequest{
				Input: map[string]string{"body": "{\"level\":\"#c832ff\"}"},
			},
			Response: messages.ProtocolResponse{},
			TaskInfo: messages.TaskInfo{
				WorkerId:            "workerid",
				TaskId:              "2",
				ProcessInstanceId:   "",
				ProcessDefinitionId: "",
				CompletionStrategy:  "optimistic",
				Time:                "-62135596800",
				TenantId:            "user",
			},
			Metadata: messages.Metadata{
				Version:              3,
				Device:               device,
				Service:              service,
				Protocol:             protocol,
				OutputPath:           "",
				OutputFunctionId:     "",
				OutputAspectNode:     nil,
				InputCharacteristic:  example.Rgb,
				OutputCharacteristic: "",
				ContentVariableHints: nil,
				ResponseTo:           "response",
				ErrorTo:              "errors",
			},
			Trace: nil,
		},
		{
			Request: messages.ProtocolRequest{
				Input: map[string]string{"body": "{\"level\":\"#ff0064\"}"},
			},
			Response: messages.ProtocolResponse{},
			TaskInfo: messages.TaskInfo{
				WorkerId:            "workerid",
				TaskId:              "3",
				ProcessInstanceId:   "",
				ProcessDefinitionId: "",
				CompletionStrategy:  "optimistic",
				Time:                "-62135596800",
				TenantId:            "user",
			},
			Metadata: messages.Metadata{
				Version:              3,
				Device:               device,
				Service:              service,
				Protocol:             protocol,
				OutputPath:           "",
				OutputFunctionId:     "",
				OutputAspectNode:     nil,
				InputCharacteristic:  example.Hex,
				OutputCharacteristic: "",
				ContentVariableHints: nil,
				ResponseTo:           "response",
				ErrorTo:              "errors",
			},
			Trace: nil,
		},
		{
			Request: messages.ProtocolRequest{
				Input: map[string]string{"body": "{\"level\":\"#ff00ff\"}"},
			},
			Response: messages.ProtocolResponse{},
			TaskInfo: messages.TaskInfo{
				WorkerId:            "workerid",
				TaskId:              "4",
				ProcessInstanceId:   "",
				ProcessDefinitionId: "",
				CompletionStrategy:  "optimistic",
				Time:                "-62135596800",
				TenantId:            "user",
			},
			Metadata: messages.Metadata{
				Version:              3,
				Device:               device,
				Service:              service,
				Protocol:             protocol,
				OutputPath:           "",
				OutputFunctionId:     "",
				OutputAspectNode:     nil,
				InputCharacteristic:  example.Hex,
				OutputCharacteristic: "",
				ContentVariableHints: nil,
				ResponseTo:           "response",
				ErrorTo:              "errors",
			},
			Trace: nil,
		},
	}

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	actualProtocolMessages := []messages.ProtocolMsg{}

	for _, message := range protocolMessageStrings {
		var temp messages.ProtocolMsg
		json.Unmarshal([]byte(message), &temp)
		temp.Trace = nil
		actualProtocolMessages = append(actualProtocolMessages, temp)
	}

	sort.Slice(actualProtocolMessages, func(i, j int) bool {
		return actualProtocolMessages[i].TaskInfo.TaskId < actualProtocolMessages[j].TaskInfo.TaskId
	})

	if !reflect.DeepEqual(actualProtocolMessages, expectedProtocolMessages) {
		actualJson, _ := json.Marshal(actualProtocolMessages)
		expectedJson, _ := json.Marshal(expectedProtocolMessages)
		t.Error("\n", string(actualJson), "\n", string(expectedJson))
	}
}

func TestCommandWithConfigurables(t *testing.T) {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.OPTIMISTIC
	config.CamundaWorkerTimeout = 100

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller)

	time.Sleep(1 * time.Second)

	//populate repository
	device := model.Device{
		Id:           "device_1",
		Name:         "d1",
		DeviceTypeId: "dt1",
		LocalId:      "d1u",
	}
	mock.Repo.RegisterDevice(device)

	protocol := model.Protocol{
		Id:               "p1",
		Name:             "protocol1",
		Handler:          "protocol1",
		ProtocolSegments: []model.ProtocolSegment{{Id: "ms1", Name: "body"}},
	}
	mock.Repo.RegisterProtocol(protocol)

	service := model.Service{
		Id:         "service_1",
		Name:       "s1",
		LocalId:    "s1u",
		ProtocolId: "p1",
		Inputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:               "level",
							Name:             "level",
							Type:             model.String,
							CharacteristicId: example.Hex,
						},
						{
							Id:               "duration",
							Name:             "duration",
							Type:             model.Integer,
							CharacteristicId: characteristics.Seconds,
							FunctionId:       "f-duration",
							AspectId:         "a-duration",
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	}
	mock.Repo.RegisterService(service)

	cmd1 := messages.Command{
		Version:          3,
		Function:         model.Function{Id: model.CONTROLLING_FUNCTION_PREFIX + "f"},
		Aspect:           nil,
		CharacteristicId: example.Rgb,
		InputPaths:       []string{"metrics.level"},
		ConfigurablesV2: []marshaller.ConfigurableV2{
			{
				Path:             "metrics.duration",
				CharacteristicId: characteristics.Minutes,
				Value:            2,
			},
		},
		DeviceId:   "device_1",
		ServiceId:  "service_1",
		ProtocolId: "p1",
		Input: map[string]float64{
			"r": 200,
			"g": 50,
			"b": 0,
		},
	}

	cmdMsg1, err := json.Marshal(cmd1)
	if err != nil {
		log.Fatal(err)
	}

	cmd2 := messages.Command{
		Version:          3,
		Function:         model.Function{Id: model.CONTROLLING_FUNCTION_PREFIX + "f"},
		CharacteristicId: example.Hex,
		InputPaths:       []string{"metrics.level"},
		ConfigurablesV2: []marshaller.ConfigurableV2{
			{
				Path:             "",
				CharacteristicId: characteristics.Minutes,
				Value:            2,
				FunctionId:       "f-duration",
				AspectNode:       model.AspectNode{Id: "a-duration-parent", DescendentIds: []string{"a-duration"}},
			},
			{
				Path:             "metrics.unknown",
				CharacteristicId: characteristics.Minutes,
				Value:            5,
			},
		},
		DeviceId:   "device_1",
		ServiceId:  "service_1",
		ProtocolId: "p1",
		Input:      "#ff0064",
	}

	cmdMsg2, err := json.Marshal(cmd2)
	if err != nil {
		log.Fatal(err)
	}

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "1",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg1),
			},
		},
	})

	time.Sleep(1 * time.Second)

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "2",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg1),
			},
			"inputs.b": {Value: "255"},
		},
	})

	time.Sleep(1 * time.Second)

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "3",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg2),
			},
		},
	})

	time.Sleep(1 * time.Second)

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "4",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg2),
			},
			"inputs": {Value: "\"#ff00ff\""},
		},
	})

	time.Sleep(1 * time.Second)

	expectedProtocolMessages := []messages.ProtocolMsg{
		{
			Request: messages.ProtocolRequest{
				Input: map[string]string{"body": "{\"duration\":120,\"level\":\"#c83200\"}"},
			},
			Response: messages.ProtocolResponse{},
			TaskInfo: messages.TaskInfo{
				WorkerId:            "workerid",
				TaskId:              "1",
				ProcessInstanceId:   "",
				ProcessDefinitionId: "",
				CompletionStrategy:  "optimistic",
				Time:                "-62135596800",
				TenantId:            "user",
			},
			Metadata: messages.Metadata{
				Version:              3,
				Device:               device,
				Service:              service,
				Protocol:             protocol,
				OutputPath:           "",
				OutputFunctionId:     "",
				OutputAspectNode:     nil,
				InputCharacteristic:  example.Rgb,
				OutputCharacteristic: "",
				ContentVariableHints: nil,
				ResponseTo:           "response",
				ErrorTo:              "errors",
			},
			Trace: nil,
		},
		{
			Request: messages.ProtocolRequest{
				Input: map[string]string{"body": "{\"duration\":120,\"level\":\"#c832ff\"}"},
			},
			Response: messages.ProtocolResponse{},
			TaskInfo: messages.TaskInfo{
				WorkerId:            "workerid",
				TaskId:              "2",
				ProcessInstanceId:   "",
				ProcessDefinitionId: "",
				CompletionStrategy:  "optimistic",
				Time:                "-62135596800",
				TenantId:            "user",
			},
			Metadata: messages.Metadata{
				Version:              3,
				Device:               device,
				Service:              service,
				Protocol:             protocol,
				OutputPath:           "",
				OutputFunctionId:     "",
				OutputAspectNode:     nil,
				InputCharacteristic:  example.Rgb,
				OutputCharacteristic: "",
				ContentVariableHints: nil,
				ResponseTo:           "response",
				ErrorTo:              "errors",
			},
			Trace: nil,
		},
		{
			Request: messages.ProtocolRequest{
				Input: map[string]string{"body": "{\"duration\":120,\"level\":\"#ff0064\"}"},
			},
			Response: messages.ProtocolResponse{},
			TaskInfo: messages.TaskInfo{
				WorkerId:            "workerid",
				TaskId:              "3",
				ProcessInstanceId:   "",
				ProcessDefinitionId: "",
				CompletionStrategy:  "optimistic",
				Time:                "-62135596800",
				TenantId:            "user",
			},
			Metadata: messages.Metadata{
				Version:              3,
				Device:               device,
				Service:              service,
				Protocol:             protocol,
				OutputPath:           "",
				OutputFunctionId:     "",
				OutputAspectNode:     nil,
				InputCharacteristic:  example.Hex,
				OutputCharacteristic: "",
				ContentVariableHints: nil,
				ResponseTo:           "response",
				ErrorTo:              "errors",
			},
			Trace: nil,
		},
		{
			Request: messages.ProtocolRequest{
				Input: map[string]string{"body": "{\"duration\":120,\"level\":\"#ff00ff\"}"},
			},
			Response: messages.ProtocolResponse{},
			TaskInfo: messages.TaskInfo{
				WorkerId:            "workerid",
				TaskId:              "4",
				ProcessInstanceId:   "",
				ProcessDefinitionId: "",
				CompletionStrategy:  "optimistic",
				Time:                "-62135596800",
				TenantId:            "user",
			},
			Metadata: messages.Metadata{
				Version:              3,
				Device:               device,
				Service:              service,
				Protocol:             protocol,
				OutputPath:           "",
				OutputFunctionId:     "",
				OutputAspectNode:     nil,
				InputCharacteristic:  example.Hex,
				OutputCharacteristic: "",
				ContentVariableHints: nil,
				ResponseTo:           "response",
				ErrorTo:              "errors",
			},
			Trace: nil,
		},
	}

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	actualProtocolMessages := []messages.ProtocolMsg{}

	for _, message := range protocolMessageStrings {
		var temp messages.ProtocolMsg
		json.Unmarshal([]byte(message), &temp)
		temp.Trace = nil
		actualProtocolMessages = append(actualProtocolMessages, temp)
	}

	sort.Slice(actualProtocolMessages, func(i, j int) bool {
		return actualProtocolMessages[i].TaskInfo.TaskId < actualProtocolMessages[j].TaskInfo.TaskId
	})

	if !reflect.DeepEqual(actualProtocolMessages, expectedProtocolMessages) {
		actualJson, _ := json.Marshal(actualProtocolMessages)
		expectedJson, _ := json.Marshal(expectedProtocolMessages)
		t.Error("\n", string(actualJson), "\n", string(expectedJson))
	}
}

func TestGroupCommand(t *testing.T) {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.OPTIMISTIC
	config.CamundaWorkerTimeout = 100

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller)

	time.Sleep(1 * time.Second)

	//populate repository
	device1 := model.Device{
		Id:           "device_1",
		Name:         "d1",
		DeviceTypeId: "dt1",
		LocalId:      "d1u",
	}
	mock.Repo.RegisterDevice(device1)

	device2 := model.Device{
		Id:           "device_2",
		Name:         "d1",
		DeviceTypeId: "dt1",
		LocalId:      "d1u",
	}
	mock.Repo.RegisterDevice(device2)

	device3 := model.Device{
		Id:           "device_3",
		Name:         "d1",
		DeviceTypeId: "dt2",
		LocalId:      "d1u",
	}
	mock.Repo.RegisterDevice(device3)

	mock.Repo.RegisterDeviceGroup(model.DeviceGroup{
		Id:   "dg1",
		Name: "dg1",
		Criteria: []model.DeviceGroupFilterCriteria{
			{FunctionId: model.MEASURING_FUNCTION_PREFIX + "f1", AspectId: "a1", Interaction: model.REQUEST},
		},
		DeviceIds: []string{"device_1", "device_2", "device_3"},
	})

	protocol := model.Protocol{
		Id:               "p1",
		Name:             "protocol1",
		Handler:          "protocol1",
		ProtocolSegments: []model.ProtocolSegment{{Id: "ms1", Name: "body"}},
	}
	mock.Repo.RegisterProtocol(protocol)

	service1 := model.Service{
		Id:          "service_1",
		Name:        "s1",
		LocalId:     "s1u",
		ProtocolId:  "p1",
		Interaction: model.REQUEST,
		Inputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:               "level",
							Name:             "level",
							Type:             model.String,
							CharacteristicId: example.Hex,
							FunctionId:       model.CONTROLLING_FUNCTION_PREFIX + "f1",
							AspectId:         "a1",
						},
						{
							Id:               "duration",
							Name:             "duration",
							Type:             model.Integer,
							CharacteristicId: characteristics.Seconds,
							FunctionId:       model.CONTROLLING_FUNCTION_PREFIX + "f3",
							AspectId:         "a3",
							Value:            10,
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	}

	service2 := model.Service{
		Id:          "service_2",
		Name:        "s2",
		LocalId:     "s2u",
		ProtocolId:  "p1",
		Interaction: model.REQUEST,
		Inputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:               "level",
							Name:             "level",
							Type:             model.String,
							CharacteristicId: example.Hex,
							FunctionId:       model.CONTROLLING_FUNCTION_PREFIX + "f1",
							AspectId:         "a1",
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	}

	service3 := model.Service{
		Id:          "service_3",
		Name:        "s3",
		LocalId:     "s3u",
		ProtocolId:  "p1",
		Interaction: model.REQUEST,
		Inputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:               "level",
							Name:             "level",
							Type:             model.String,
							CharacteristicId: example.Hex,
							FunctionId:       model.CONTROLLING_FUNCTION_PREFIX + "f1",
							AspectId:         "a1",
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	}

	mock.Repo.RegisterDeviceType(model.DeviceType{
		Id:            "dt1",
		Name:          "dt1",
		DeviceClassId: "dc1",
		Services: []model.Service{
			service1,
			service2,
		},
	})

	mock.Repo.RegisterDeviceType(model.DeviceType{
		Id:            "dt2",
		Name:          "dt2",
		DeviceClassId: "dc1",
		Services: []model.Service{
			service3,
			{
				Id:          "service_4",
				Name:        "s4",
				LocalId:     "s4u",
				ProtocolId:  "p1",
				Interaction: model.REQUEST,
				Inputs: []model.Content{
					{
						Id: "metrics",
						ContentVariable: model.ContentVariable{
							Id:   "metrics",
							Name: "metrics",
							Type: model.Structure,
							SubContentVariables: []model.ContentVariable{
								{
									Id:               "level",
									Name:             "level",
									Type:             model.Integer,
									CharacteristicId: example.Hex,
									FunctionId:       model.CONTROLLING_FUNCTION_PREFIX + "f2",
									AspectId:         "a2",
								},
							},
						},
						Serialization:     "json",
						ProtocolSegmentId: "ms1",
					},
				},
			},
		},
	})

	cmd1 := messages.Command{
		Version:          3,
		Function:         model.Function{Id: model.CONTROLLING_FUNCTION_PREFIX + "f1"},
		Aspect:           nil,
		CharacteristicId: example.Rgb,
		DeviceGroupId:    "dg1",
		Input: map[string]float64{
			"r": 200,
			"g": 50,
			"b": 0,
		},
	}

	cmdMsg1, err := json.Marshal(cmd1)
	if err != nil {
		log.Fatal(err)
	}

	cmd2 := messages.Command{
		Version:          3,
		Function:         model.Function{Id: model.CONTROLLING_FUNCTION_PREFIX + "f1"},
		CharacteristicId: example.Hex,
		DeviceGroupId:    "dg1",
		Input:            "#ff0064",
	}

	cmdMsg2, err := json.Marshal(cmd2)
	if err != nil {
		log.Fatal(err)
	}

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "1",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg1),
			},
		},
	})

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "2",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg1),
			},
			"inputs.b": {Value: "255"},
		},
	})

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "3",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg2),
			},
		},
	})

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "4",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg2),
			},
			"inputs": {Value: "\"#ff00ff\""},
		},
	})

	time.Sleep(1 * time.Second)

	createExpected := func(taskIdPrefix string, characteristicId string, level string) []messages.ProtocolMsg {
		return []messages.ProtocolMsg{
			{
				Request: messages.ProtocolRequest{
					Input: map[string]string{"body": "{\"duration\":10,\"level\":\"" + level + "\"}"},
				},
				Response: messages.ProtocolResponse{},
				TaskInfo: messages.TaskInfo{
					WorkerId:            "workerid",
					TaskId:              taskIdPrefix + "_0_0",
					ProcessInstanceId:   "",
					ProcessDefinitionId: "",
					CompletionStrategy:  "optimistic",
					Time:                "-62135596800",
					TenantId:            "user",
				},
				Metadata: messages.Metadata{
					Version:              3,
					Device:               device1,
					Service:              service1,
					Protocol:             protocol,
					OutputPath:           "",
					OutputFunctionId:     "",
					OutputAspectNode:     nil,
					InputCharacteristic:  characteristicId,
					OutputCharacteristic: "",
					ContentVariableHints: nil,
					ResponseTo:           "response",
					ErrorTo:              "errors",
				},
				Trace: nil,
			},
			{
				Request: messages.ProtocolRequest{
					Input: map[string]string{"body": "{\"level\":\"" + level + "\"}"},
				},
				Response: messages.ProtocolResponse{},
				TaskInfo: messages.TaskInfo{
					WorkerId:            "workerid",
					TaskId:              taskIdPrefix + "_0_1",
					ProcessInstanceId:   "",
					ProcessDefinitionId: "",
					CompletionStrategy:  "optimistic",
					Time:                "-62135596800",
					TenantId:            "user",
				},
				Metadata: messages.Metadata{
					Version:              3,
					Device:               device1,
					Service:              service2,
					Protocol:             protocol,
					OutputPath:           "",
					OutputFunctionId:     "",
					OutputAspectNode:     nil,
					InputCharacteristic:  characteristicId,
					OutputCharacteristic: "",
					ContentVariableHints: nil,
					ResponseTo:           "response",
					ErrorTo:              "errors",
				},
				Trace: nil,
			},
			{
				Request: messages.ProtocolRequest{
					Input: map[string]string{"body": "{\"duration\":10,\"level\":\"" + level + "\"}"},
				},
				Response: messages.ProtocolResponse{},
				TaskInfo: messages.TaskInfo{
					WorkerId:            "workerid",
					TaskId:              taskIdPrefix + "_1_0",
					ProcessInstanceId:   "",
					ProcessDefinitionId: "",
					CompletionStrategy:  "optimistic",
					Time:                "-62135596800",
					TenantId:            "user",
				},
				Metadata: messages.Metadata{
					Version:              3,
					Device:               device2,
					Service:              service1,
					Protocol:             protocol,
					OutputPath:           "",
					OutputFunctionId:     "",
					OutputAspectNode:     nil,
					InputCharacteristic:  characteristicId,
					OutputCharacteristic: "",
					ContentVariableHints: nil,
					ResponseTo:           "response",
					ErrorTo:              "errors",
				},
				Trace: nil,
			},
			{
				Request: messages.ProtocolRequest{
					Input: map[string]string{"body": "{\"level\":\"" + level + "\"}"},
				},
				Response: messages.ProtocolResponse{},
				TaskInfo: messages.TaskInfo{
					WorkerId:            "workerid",
					TaskId:              taskIdPrefix + "_1_1",
					ProcessInstanceId:   "",
					ProcessDefinitionId: "",
					CompletionStrategy:  "optimistic",
					Time:                "-62135596800",
					TenantId:            "user",
				},
				Metadata: messages.Metadata{
					Version:              3,
					Device:               device2,
					Service:              service2,
					Protocol:             protocol,
					OutputPath:           "",
					OutputFunctionId:     "",
					OutputAspectNode:     nil,
					InputCharacteristic:  characteristicId,
					OutputCharacteristic: "",
					ContentVariableHints: nil,
					ResponseTo:           "response",
					ErrorTo:              "errors",
				},
				Trace: nil,
			},
			{
				Request: messages.ProtocolRequest{
					Input: map[string]string{"body": "{\"level\":\"" + level + "\"}"},
				},
				Response: messages.ProtocolResponse{},
				TaskInfo: messages.TaskInfo{
					WorkerId:            "workerid",
					TaskId:              taskIdPrefix + "_2_0",
					ProcessInstanceId:   "",
					ProcessDefinitionId: "",
					CompletionStrategy:  "optimistic",
					Time:                "-62135596800",
					TenantId:            "user",
				},
				Metadata: messages.Metadata{
					Version:              3,
					Device:               device3,
					Service:              service3,
					Protocol:             protocol,
					OutputPath:           "",
					OutputFunctionId:     "",
					OutputAspectNode:     nil,
					InputCharacteristic:  characteristicId,
					OutputCharacteristic: "",
					ContentVariableHints: nil,
					ResponseTo:           "response",
					ErrorTo:              "errors",
				},
				Trace: nil,
			},
		}
	}

	expectedProtocolMessages := []messages.ProtocolMsg{}

	expectedProtocolMessages = append(expectedProtocolMessages, createExpected("1", example.Rgb, "#c83200")...)
	expectedProtocolMessages = append(expectedProtocolMessages, createExpected("2", example.Rgb, "#c832ff")...)
	expectedProtocolMessages = append(expectedProtocolMessages, createExpected("3", example.Hex, "#ff0064")...)
	expectedProtocolMessages = append(expectedProtocolMessages, createExpected("4", example.Hex, "#ff00ff")...)

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	actualProtocolMessages := []messages.ProtocolMsg{}

	for _, message := range protocolMessageStrings {
		var temp messages.ProtocolMsg
		json.Unmarshal([]byte(message), &temp)
		temp.Trace = nil
		actualProtocolMessages = append(actualProtocolMessages, temp)
	}

	sort.Slice(actualProtocolMessages, func(i, j int) bool {
		return actualProtocolMessages[i].TaskInfo.TaskId < actualProtocolMessages[j].TaskInfo.TaskId
	})

	if !reflect.DeepEqual(normalize(actualProtocolMessages), normalize(expectedProtocolMessages)) {
		actualJson, _ := json.Marshal(actualProtocolMessages)
		expectedJson, _ := json.Marshal(expectedProtocolMessages)
		t.Error("\n", string(actualJson), "\n", string(expectedJson))
	}
}

func TestResponse(t *testing.T) {
	t.Error("TODO")
}

func TestResponseWithConfigurables(t *testing.T) {
	t.Error("TODO")
}

func TestGroupResponse(t *testing.T) {
	t.Error("TODO")
}

func normalize(in interface{}) (out interface{}) {
	temp, err := json.Marshal(in)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(temp, &out)
	if err != nil {
		panic(err)
	}
	return
}
