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
	"github.com/SENERGY-Platform/external-task-worker/lib/test/docker"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"reflect"
	"sort"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestWorkerResponseV2(t *testing.T) {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.GroupScheduler = util.PARALLEL
	config.CompletionStrategy = util.PESSIMISTIC
	config.CamundaWorkerTimeout = 100

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller)

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
		Id:         "service_1",
		Name:       "s1",
		LocalId:    "s1u",
		ProtocolId: "p1",
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
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
		CharacteristicId: example.Rgb,
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
	}

	cmdMsg1, err := json.Marshal(cmd1)
	if err != nil {
		t.Error(err)
		return
	}

	cmd2 := messages.Command{
		Version:          2,
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
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

	if len(protocolMessageStrings) != 2 {
		t.Error(protocolMessageStrings)
		return
	}

	for _, message := range protocolMessageStrings {
		msg := messages.ProtocolMsg{}
		err = json.Unmarshal([]byte(message), &msg)
		if err != nil {
			log.Fatal(err)
		}
		msg.Response.Output = map[string]string{
			"body": "{\"level\":\"#c83200\"}",
		}
		resp, err := json.Marshal(msg)
		if err != nil {
			log.Fatal(err)
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
		`["#c83200"]`,
		`[{"b":0,"g":50,"r":200}]`,
	}) {
		t.Error(list)
		return
	}
}

func TestGroupResponses(t *testing.T) {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	config.GroupScheduler = util.PARALLEL
	config.CompletionStrategy = util.PESSIMISTIC
	config.CamundaWorkerTimeout = 100
	config.Debug = true

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller)

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
				FunctionIds: []string{model.MEASURING_FUNCTION_PREFIX + "f1", model.MEASURING_FUNCTION_PREFIX + "f2"},
				AspectIds:   []string{"a1", "a2"},
				Interaction: model.REQUEST,
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
				FunctionIds: []string{model.MEASURING_FUNCTION_PREFIX + "f1"},
				AspectIds:   []string{"a1"},
				Interaction: model.REQUEST,
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
				FunctionIds: []string{model.MEASURING_FUNCTION_PREFIX + "f1", model.MEASURING_FUNCTION_PREFIX + "f2"},
				AspectIds:   []string{"a1", "a2"},
				Interaction: model.REQUEST,
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
				FunctionIds: []string{model.MEASURING_FUNCTION_PREFIX + "f2"},
				AspectIds:   []string{"a2"},
				Interaction: model.REQUEST,
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
		Aspect:           &model.Aspect{Id: "a1"},
		CharacteristicId: example.Rgb,
		DeviceGroupId:    "dg1",
		DeviceId:         "device_1",  //will be ignored because DeviceGroupId is set
		ServiceId:        "service_1", //will be ignored because DeviceGroupId is set
		ProtocolId:       "p1",        //will be ignored because DeviceGroupId is set
	}

	cmdMsg1, err := json.Marshal(cmd1)
	if err != nil {
		t.Error(err)
		return
	}

	cmd2 := messages.Command{
		Version:          2,
		Function:         model.Function{Id: model.MEASURING_FUNCTION_PREFIX + "f1", RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
		Aspect:           &model.Aspect{Id: "a1"},
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

	if len(protocolMessageStrings) != 6 {
		t.Error(len(protocolMessageStrings), protocolMessageStrings)
		return
	}

	for i, message := range protocolMessageStrings {
		msg := messages.ProtocolMsg{}
		err = json.Unmarshal([]byte(message), &msg)
		if err != nil {
			log.Fatal(err)
		}
		msg.Response.Output = map[string]string{
			"body": "{\"level\":\"#c8320" + strconv.Itoa(i) + "\"}",
		}
		resp, err := json.Marshal(msg)
		if err != nil {
			log.Fatal(err)
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
		`["#c83203","#c83204","#c83205"]`,
		`[{"b":0,"g":50,"r":200},{"b":1,"g":50,"r":200},{"b":2,"g":50,"r":200}]`,
	}) {
		t.Error(list)
		return
	}
}

func TestGroupResponsesWithMemcached(t *testing.T) {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port, _, err := docker.Memcached(ctx, wg)
	if err != nil {
		t.Error(err)
		return
	}

	config.SubResultDatabaseUrls = []string{"localhost:" + port}

	config.GroupScheduler = util.PARALLEL
	config.CompletionStrategy = util.PESSIMISTIC
	config.CamundaWorkerTimeout = 100
	config.Debug = true

	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller)

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
				FunctionIds: []string{model.MEASURING_FUNCTION_PREFIX + "f1", model.MEASURING_FUNCTION_PREFIX + "f2"},
				AspectIds:   []string{"a1", "a2"},
				Interaction: model.REQUEST,
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
				FunctionIds: []string{model.MEASURING_FUNCTION_PREFIX + "f1"},
				AspectIds:   []string{"a1"},
				Interaction: model.REQUEST,
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
				FunctionIds: []string{model.MEASURING_FUNCTION_PREFIX + "f1", model.MEASURING_FUNCTION_PREFIX + "f2"},
				AspectIds:   []string{"a1", "a2"},
				Interaction: model.REQUEST,
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
				FunctionIds: []string{model.MEASURING_FUNCTION_PREFIX + "f2"},
				AspectIds:   []string{"a2"},
				Interaction: model.REQUEST,
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
		Aspect:           &model.Aspect{Id: "a1"},
		CharacteristicId: example.Rgb,
		DeviceGroupId:    "dg1",
		DeviceId:         "device_1",  //will be ignored because DeviceGroupId is set
		ServiceId:        "service_1", //will be ignored because DeviceGroupId is set
		ProtocolId:       "p1",        //will be ignored because DeviceGroupId is set
	}

	cmdMsg1, err := json.Marshal(cmd1)
	if err != nil {
		t.Error(err)
		return
	}

	cmd2 := messages.Command{
		Version:          2,
		Function:         model.Function{Id: model.MEASURING_FUNCTION_PREFIX + "f1", RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
		Aspect:           &model.Aspect{Id: "a1"},
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

	if len(protocolMessageStrings) != 6 {
		t.Error(len(protocolMessageStrings), protocolMessageStrings)
		return
	}

	for i, message := range protocolMessageStrings {
		msg := messages.ProtocolMsg{}
		err = json.Unmarshal([]byte(message), &msg)
		if err != nil {
			log.Fatal(err)
		}
		msg.Response.Output = map[string]string{
			"body": "{\"level\":\"#c8320" + strconv.Itoa(i) + "\"}",
		}
		resp, err := json.Marshal(msg)
		if err != nil {
			log.Fatal(err)
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
		`["#c83203","#c83204","#c83205"]`,
		`[{"b":0,"g":50,"r":200},{"b":1,"g":50,"r":200},{"b":2,"g":50,"r":200}]`,
	}) {
		t.Error(list)
		return
	}
}
