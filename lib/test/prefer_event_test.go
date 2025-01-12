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
	"github.com/SENERGY-Platform/external-task-worker/lib"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestWorkerPreferEventResponseV2(t *testing.T) {
	mock.CleanKafkaMock()
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.TimescaleWrapperUrl = "placeholder"
	config.GroupScheduler = util.PARALLEL
	config.CompletionStrategy = util.PESSIMISTIC
	config.HttpCommandConsumerPort, err = GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()

	timescaleResponses := mock.TimescaleMockResponses{
		"device_1": {"service_1": {"metrics.level": "#c83200"}},
	}

	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller, mock.GetTimescaleMockFactory(timescaleResponses))

	time.Sleep(1 * time.Second)

	//populate repository
	mock.Repo.RegisterDevice(model.Device{
		Id:           "device_1",
		Name:         "d1",
		DeviceTypeId: "dt1",
		LocalId:      "d1u",
	})

	mock.Repo.RegisterProtocol(model.Protocol{
		Id:               "p1",
		Name:             "protocol1",
		Handler:          "protocol1",
		ProtocolSegments: []model.ProtocolSegment{{Id: "ms1", Name: "body"}},
	})

	mock.Repo.RegisterService(model.Service{
		Id:          "service_1",
		Name:        "s1",
		LocalId:     "s1u",
		ProtocolId:  "p1",
		Interaction: model.EVENT_AND_REQUEST,
		Outputs: []model.Content{
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
							FunctionId:       model.MEASURING_FUNCTION_PREFIX + "f1",
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	})

	cmd1 := messages.Command{
		Version:          2,
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION, Id: model.MEASURING_FUNCTION_PREFIX + "f1"},
		CharacteristicId: example.Rgb,
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
		PreferEvent:      true,
	}

	cmdMsg1, err := json.Marshal(cmd1)
	if err != nil {
		t.Error(err)
		return
	}

	cmd2 := messages.Command{
		Version:          2,
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION, Id: model.MEASURING_FUNCTION_PREFIX + "f1"},
		CharacteristicId: example.Hex,
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
	}

	cmdMsg2, err := json.Marshal(cmd2)
	if err != nil {
		t.Error(err)
		return
	}

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id: "1",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg1),
			},
		},
	})

	time.Sleep(1 * time.Second)

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id: "2",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg2),
			},
		},
	})

	time.Sleep(1 * time.Second)

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	if len(protocolMessageStrings) != 1 {
		t.Error(protocolMessageStrings)
		return
	}

	for _, message := range protocolMessageStrings {
		msg := messages.ProtocolMsg{}
		err = json.Unmarshal([]byte(message), &msg)
		if err != nil {
			t.Error(err)
			return
		}
		msg.Response.Output = map[string]string{
			"body": "{\"level\":\"#c83100\"}",
		}
		resp, err := json.Marshal(msg)
		if err != nil {
			t.Error(err)
			return
		}
		mock.Kafka.Produce(config.ResponseTopic, string(resp))
		time.Sleep(1 * time.Second)
	}

	fetched, completed, failed := mockCamunda.GetStatus()

	if len(fetched) != 0 || len(failed) != 0 || len(completed) != 2 {
		log.Println("fetched:", fetched)
		log.Println("failed:", failed)
		log.Println("completed:", completed)
		log.Println(len(fetched), len(failed), len(completed))
		return
	}

	list := []string{}

	for _, cmd := range completed {
		temp, err := json.Marshal(cmd)
		if err != nil {
			log.Fatal(err)
		}
		list = append(list, string(temp))
	}
	sort.Strings(list)
	if !reflect.DeepEqual(list, []string{
		`["#c83100"]`,
		`[{"b":0,"g":50,"r":200}]`,
	}) {
		t.Error(list)
		return
	}
}

func TestGroupPreferEventResponses(t *testing.T) {
	mock.CleanKafkaMock()
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.TimescaleWrapperUrl = "placeholder"
	config.GroupScheduler = util.PARALLEL
	config.CompletionStrategy = util.PESSIMISTIC
	config.Debug = true
	config.HttpCommandConsumerPort, err = GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()

	timescaleResponses := mock.TimescaleMockResponses{
		"device_1": {"service_1": {"metrics.level": "#c8320{{count}}"}, "service_2": {"metrics.level": "#c8320{{count}}"}},
		"device_2": {"service_3": {"metrics.level": "#c8320{{count}}"}, "service_4": {"metrics.level": "#c8320{{count}}"}},
	}

	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller, mock.GetTimescaleMockFactory(timescaleResponses))

	time.Sleep(1 * time.Second)

	//populate repository
	mock.Repo.RegisterDevice(model.Device{
		Id:           "device_1",
		Name:         "d1",
		DeviceTypeId: "dt1",
		LocalId:      "d1u",
	})

	mock.Repo.RegisterDevice(model.Device{
		Id:           "device_2",
		Name:         "d2",
		DeviceTypeId: "dt2",
		LocalId:      "d1u",
	})

	mock.Repo.RegisterDeviceGroup(model.DeviceGroup{
		Id:   "dg1",
		Name: "dg1",
		Criteria: []model.DeviceGroupFilterCriteria{
			{FunctionId: model.MEASURING_FUNCTION_PREFIX + "f1", AspectId: "a1", Interaction: model.REQUEST},
		},
		DeviceIds: []string{"device_1", "device_2"},
	})

	mock.Repo.RegisterProtocol(model.Protocol{
		Id:               "p1",
		Name:             "protocol1",
		Handler:          "protocol1",
		ProtocolSegments: []model.ProtocolSegment{{Id: "ms1", Name: "body"}},
	})

	mock.Repo.RegisterDeviceType(model.DeviceType{
		Id:            "dt1",
		Name:          "dt1",
		DeviceClassId: "dc1",
		Services: []model.Service{
			{
				Id:          "service_1",
				Name:        "s1",
				LocalId:     "s1u",
				ProtocolId:  "p1",
				Interaction: model.EVENT_AND_REQUEST,
				Outputs: []model.Content{
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
									FunctionId:       model.MEASURING_FUNCTION_PREFIX + "f1",
									AspectId:         "a1",
								},
							},
						},
						Serialization:     "json",
						ProtocolSegmentId: "ms1",
					},
				},
			},
			{
				Id:          "service_2",
				Name:        "s2",
				LocalId:     "s2u",
				ProtocolId:  "p1",
				Interaction: model.EVENT_AND_REQUEST,
				Outputs: []model.Content{
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
									FunctionId:       model.MEASURING_FUNCTION_PREFIX + "f1",
									AspectId:         "a1",
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

	mock.Repo.RegisterDeviceType(model.DeviceType{
		Id:            "dt2",
		Name:          "dt2",
		DeviceClassId: "dc1",
		Services: []model.Service{
			{
				Id:          "service_3",
				Name:        "s3",
				LocalId:     "s3u",
				ProtocolId:  "p1",
				Interaction: model.EVENT_AND_REQUEST,
				Outputs: []model.Content{
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
									FunctionId:       model.MEASURING_FUNCTION_PREFIX + "f1",
									AspectId:         "a1",
								},
							},
						},
						Serialization:     "json",
						ProtocolSegmentId: "ms1",
					},
				},
			},
			{
				Id:          "service_4",
				Name:        "s4",
				LocalId:     "s4u",
				ProtocolId:  "p1",
				Interaction: model.EVENT_AND_REQUEST,
				Outputs: []model.Content{
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
									FunctionId:       model.MEASURING_FUNCTION_PREFIX + "f2",
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
		Version:          2,
		Function:         model.Function{Id: model.MEASURING_FUNCTION_PREFIX + "f1", RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
		Aspect:           &model.AspectNode{Id: "a1"},
		CharacteristicId: example.Rgb,
		DeviceGroupId:    "dg1",
		DeviceId:         "device_1",  //will be ignored because DeviceGroupId is set
		ServiceId:        "service_1", //will be ignored because DeviceGroupId is set
		ProtocolId:       "p1",        //will be ignored because DeviceGroupId is set
		PreferEvent:      true,
	}

	cmdMsg1, err := json.Marshal(cmd1)
	if err != nil {
		t.Error(err)
		return
	}

	cmd2 := messages.Command{
		Version:          2,
		Function:         model.Function{Id: model.MEASURING_FUNCTION_PREFIX + "f1", RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
		Aspect:           &model.AspectNode{Id: "a1"},
		CharacteristicId: example.Hex,
		DeviceGroupId:    "dg1",
	}

	cmdMsg2, err := json.Marshal(cmd2)
	if err != nil {
		t.Error(err)
		return
	}

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id: "1",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg1),
			},
		},
	})

	time.Sleep(1 * time.Second)

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id: "2",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg2),
			},
		},
	})

	time.Sleep(1 * time.Second)

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	for _, message := range protocolMessageStrings {
		msg := messages.ProtocolMsg{}
		err = json.Unmarshal([]byte(message), &msg)
		if err != nil {
			t.Error(err)
			return
		}
		msg.Response.Output = map[string]string{
			"body": "{\"level\":\"#c83100\"}",
		}
		resp, err := json.Marshal(msg)
		if err != nil {
			t.Error(err)
			return
		}
		mock.Kafka.Produce(config.ResponseTopic, string(resp))
		time.Sleep(1 * time.Second)
	}

	fetched, completed, failed := mockCamunda.GetStatus()

	if len(fetched) != 0 || len(failed) != 0 || len(completed) != 2 {
		log.Println("fetched:", fetched)
		log.Println("failed:", failed)
		log.Println("completed:", completed)
		t.Error(len(fetched), len(failed), len(completed))
		return
	}

	list := []string{}

	for _, cmd := range completed {
		temp, err := json.Marshal(cmd)
		if err != nil {
			log.Fatal(err)
		}
		list = append(list, string(temp))
	}
	sort.Strings(list)
	if !reflect.DeepEqual(list, []string{
		`["#c83100","#c83100","#c83100"]`,
		`[{"b":0,"g":50,"r":200},{"b":1,"g":50,"r":200},{"b":2,"g":50,"r":200}]`,
	}) {
		t.Error(list)
		return
	}
}
