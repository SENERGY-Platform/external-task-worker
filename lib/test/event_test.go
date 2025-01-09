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
	"sync"
	"testing"
	"time"
)

func TestWorkerEventResponseV2(t *testing.T) {
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
		Interaction: model.EVENT,
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

func TestGroupEventResponses(t *testing.T) {
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
				Interaction: model.EVENT,
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
				Interaction: model.EVENT,
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
				Interaction: model.EVENT,
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
				Interaction: model.EVENT,
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

func TestGroupEventResponsesWithMemcached(t *testing.T) {
	mock.CleanKafkaMock()
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

	config.TimescaleWrapperUrl = "placeholder"
	config.GroupScheduler = util.PARALLEL
	config.CompletionStrategy = util.PESSIMISTIC
	config.Debug = true

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
				Id:         "service_1",
				Name:       "s1",
				LocalId:    "s1u",
				ProtocolId: "p1",
				//FunctionIds: []string{model.MEASURING_FUNCTION_PREFIX + "f1", model.MEASURING_FUNCTION_PREFIX + "f2"},
				//AspectIds:   []string{"a1", "a2"},
				Interaction: model.EVENT,
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
				Interaction: model.EVENT,
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
				Id:         "service_3",
				Name:       "s3",
				LocalId:    "s3u",
				ProtocolId: "p1",
				//FunctionIds: []string{model.MEASURING_FUNCTION_PREFIX + "f1", model.MEASURING_FUNCTION_PREFIX + "f2"},
				//AspectIds:   []string{"a1", "a2"},
				Interaction: model.EVENT,
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
				Interaction: model.EVENT,
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

func TestWorkerEventDeviceWithoutServiceCommand(t *testing.T) {
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
		"urn:infai:ses:device:423a2718-dea0-4f69-85a3-626c52de175b": {
			"urn:infai:ses:service:7ac752b5-8d25-48b0-aa8a-07dffeb55347": {"struct.kelvin": 5641},
		},
	}

	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller, mock.GetTimescaleMockFactory(timescaleResponses))

	time.Sleep(1 * time.Second)

	//populate repository
	mock.Repo.RegisterDevice(model.Device{
		Id:           "urn:infai:ses:device:423a2718-dea0-4f69-85a3-626c52de175b",
		Name:         "lampe 1",
		DeviceTypeId: "urn:infai:ses:device-type:57871169-38a8-40cd-871b-184b99776ca3",
		LocalId:      "618dfabb-c6a8-4d59-a338-ad9d82735ea2",
	})

	mock.Repo.RegisterProtocol(model.Protocol{
		Id:               "urn:infai:ses:protocol:3b59ea31-da98-45fd-a354-1b9bd06b837e",
		Name:             "protocol1",
		Handler:          "protocol1",
		ProtocolSegments: []model.ProtocolSegment{{Id: "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a", Name: "body"}},
	})

	deviceTypeStr := `{
    "id": "urn:infai:ses:device-type:57871169-38a8-40cd-871b-184b99776ca3",
    "name": "Philips Extended Color Light (moses)",
    "description": "Philips Hue Extended Color Light (moses)",
    "service_groups": [],
    "services": [
        {
            "id": "urn:infai:ses:service:163fa8dc-d919-43d0-aee0-4c4b50e406f8",
            "local_id": "getBrightness",
            "name": "Get Brightness",
            "description": "",
            "interaction": "event",
            "protocol_id": "urn:infai:ses:protocol:3b59ea31-da98-45fd-a354-1b9bd06b837e",
            "inputs": [],
            "outputs": [
                {
                    "id": "urn:infai:ses:content:941a6838-84e1-4521-a5d8-c6b83f2c4844",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:871efa99-163d-4a0f-9b36-7cb346bb8565",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:0e8c968a-141e-43a3-9412-d943082648e2",
                                "name": "brightness",
                                "is_void": false,
                                "type": "https://schema.org/Float",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:46f808f4-bb9e-4cc2-bd50-dc33ca74f273",
                                "value": null,
                                "serialization_options": null,
                                "function_id": "urn:infai:ses:measuring-function:c51a6ea5-90c3-4223-9052-6fe4136386cd",
                                "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                            },
                            {
                                "id": "urn:infai:ses:content-variable:c7f55305-62f1-4cfc-826e-d5d81b9ee22a",
                                "name": "status",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:c0353532-a8fb-4553-a00b-418cb8a80a65",
                                "value": null,
                                "serialization_options": null
                            },
                            {
                                "id": "urn:infai:ses:content-variable:6423a99f-fb43-4024-81f6-fc56ecbd5b97",
                                "name": "time",
                                "is_void": false,
                                "type": "https://schema.org/Text",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:6bc41b45-a9f3-4d87-9c51-dd3e11257800",
                                "value": null,
                                "serialization_options": null,
                                "function_id": "urn:infai:ses:measuring-function:3b4e0766-0d67-4658-b249-295902cd3290",
                                "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "attributes": [],
            "service_group_key": ""
        },
        {
            "id": "urn:infai:ses:service:38af42fa-7e89-4b41-826c-2c851eb6a7e8",
            "local_id": "getColor",
            "name": "Get Color",
            "description": "",
            "interaction": "event",
            "protocol_id": "urn:infai:ses:protocol:3b59ea31-da98-45fd-a354-1b9bd06b837e",
            "inputs": [],
            "outputs": [
                {
                    "id": "urn:infai:ses:content:0b4d2ede-1d5f-4b0b-bb38-5d4c3bb5306d",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:64bae8d4-2508-4407-85fc-82b51de00eac",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:eb0a4432-439d-4af5-af73-db476e151c3f",
                                "name": "red",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:dfe6be4a-650c-4411-8d87-062916b48951",
                                "value": null,
                                "serialization_options": null
                            },
                            {
                                "id": "urn:infai:ses:content-variable:a858090f-f4a5-4d19-94a8-81f59bf937f1",
                                "name": "green",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:5ef27837-4aca-43ad-b8f6-4d95cf9ed99e",
                                "value": null,
                                "serialization_options": null
                            },
                            {
                                "id": "urn:infai:ses:content-variable:55b9c48f-78a8-443d-a042-842ad8d38879",
                                "name": "blue",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:590af9ef-3a5e-4edb-abab-177cb1320b17",
                                "value": null,
                                "serialization_options": null
                            },
                            {
                                "id": "urn:infai:ses:content-variable:5056535c-5bcd-4f6e-bab3-2a7e958fdd34",
                                "name": "status",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:c0353532-a8fb-4553-a00b-418cb8a80a65",
                                "value": null,
                                "serialization_options": null
                            },
                            {
                                "id": "urn:infai:ses:content-variable:b7b69993-ae11-4039-9591-ba2e0fde7611",
                                "name": "time",
                                "is_void": false,
                                "type": "https://schema.org/Text",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:6bc41b45-a9f3-4d87-9c51-dd3e11257800",
                                "value": null,
                                "serialization_options": null,
                                "function_id": "urn:infai:ses:measuring-function:3b4e0766-0d67-4658-b249-295902cd3290",
                                "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                            }
                        ],
                        "characteristic_id": "urn:infai:ses:characteristic:5b4eea52-e8e5-4e80-9455-0382f81a1b43",
                        "value": null,
                        "serialization_options": null,
                        "function_id": "urn:infai:ses:measuring-function:bdb6a7c8-4a3d-4fe0-bab3-ce02e09b5869",
                        "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "attributes": [],
            "service_group_key": ""
        },
        {
            "id": "urn:infai:ses:service:7ac752b5-8d25-48b0-aa8a-07dffeb55347",
            "local_id": "getKelvin",
            "name": "Get Color Temperature",
            "description": "",
            "interaction": "event",
            "protocol_id": "urn:infai:ses:protocol:3b59ea31-da98-45fd-a354-1b9bd06b837e",
            "inputs": [],
            "outputs": [
                {
                    "id": "urn:infai:ses:content:854acb4f-2cd4-48ae-b23b-85c32090e9d6",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:40ab57d5-201f-40d8-9c9e-f9b0f0f5f57d",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:a04e23c7-2292-40db-a9d1-6c826e970dd4",
                                "name": "kelvin",
                                "is_void": false,
                                "type": "https://schema.org/Float",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:75b2d113-1d03-4ef8-977a-8dbcbb31a683",
                                "value": null,
                                "serialization_options": null,
                                "function_id": "urn:infai:ses:measuring-function:fb0f474f-c1d7-4a90-971b-a0b9915968f0",
                                "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                            },
                            {
                                "id": "urn:infai:ses:content-variable:9ea10499-50d5-4dc6-821a-49893e15a849",
                                "name": "status",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:c0353532-a8fb-4553-a00b-418cb8a80a65",
                                "value": null,
                                "serialization_options": null
                            },
                            {
                                "id": "urn:infai:ses:content-variable:80181cb5-bbcb-4c68-8ad8-fc2f4bfd9a7a",
                                "name": "time",
                                "is_void": false,
                                "type": "https://schema.org/Text",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:6bc41b45-a9f3-4d87-9c51-dd3e11257800",
                                "value": null,
                                "serialization_options": null,
                                "function_id": "urn:infai:ses:measuring-function:3b4e0766-0d67-4658-b249-295902cd3290",
                                "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "attributes": [],
            "service_group_key": ""
        },
        {
            "id": "urn:infai:ses:service:4565b9a6-68c9-4d66-9c5e-67e6fc5e989e",
            "local_id": "getPower",
            "name": "Get Power",
            "description": "",
            "interaction": "event",
            "protocol_id": "urn:infai:ses:protocol:3b59ea31-da98-45fd-a354-1b9bd06b837e",
            "inputs": [],
            "outputs": [
                {
                    "id": "urn:infai:ses:content:a0d22df3-1998-43bf-b48a-84e8c95fa1e9",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:b12505dc-aa57-4cae-b6b9-2bcfbf64b8db",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:4ce88010-cb10-4443-8f0e-d7a1beba61c0",
                                "name": "power",
                                "is_void": false,
                                "type": "https://schema.org/Boolean",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:7dc1bb7e-b256-408a-a6f9-044dc60fdcf5",
                                "value": null,
                                "serialization_options": null,
                                "function_id": "urn:infai:ses:measuring-function:20d3c1d3-77d7-4181-a9f3-b487add58cd0",
                                "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                            },
                            {
                                "id": "urn:infai:ses:content-variable:008b09ad-04ca-42be-aadf-49e4f9809ac8",
                                "name": "status",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:c0353532-a8fb-4553-a00b-418cb8a80a65",
                                "value": null,
                                "serialization_options": null
                            },
                            {
                                "id": "urn:infai:ses:content-variable:30a042d7-6cfb-48ab-9a82-138745b2964f",
                                "name": "time",
                                "is_void": false,
                                "type": "https://schema.org/Text",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:6bc41b45-a9f3-4d87-9c51-dd3e11257800",
                                "value": null,
                                "serialization_options": null,
                                "function_id": "urn:infai:ses:measuring-function:3b4e0766-0d67-4658-b249-295902cd3290",
                                "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "attributes": [],
            "service_group_key": ""
        },
        {
            "id": "urn:infai:ses:service:c8751122-ec97-4a84-82e5-fd0341a294e5",
            "local_id": "setBrightness",
            "name": "Set Brightness",
            "description": "",
            "interaction": "event",
            "protocol_id": "urn:infai:ses:protocol:3b59ea31-da98-45fd-a354-1b9bd06b837e",
            "inputs": [
                {
                    "id": "urn:infai:ses:content:32b69aa9-1fcf-4045-b1c6-67e142862cfc",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:acc1b599-08b2-4691-9126-c2a7b0be36f6",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:cbf646e1-9a21-4cac-9ab3-6b89b7ea6fe6",
                                "name": "brightness",
                                "is_void": false,
                                "type": "https://schema.org/Float",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:46f808f4-bb9e-4cc2-bd50-dc33ca74f273",
                                "value": null,
                                "serialization_options": null,
                                "function_id": "urn:infai:ses:controlling-function:6ce74d6d-7acb-40b7-aac7-49daca214e85",
                                "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                            },
                            {
                                "id": "urn:infai:ses:content-variable:462db592-797d-43df-8059-ab4d06272729",
                                "name": "duration",
                                "is_void": false,
                                "type": "https://schema.org/Float",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c",
                                "value": null,
                                "serialization_options": null
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "outputs": [
                {
                    "id": "urn:infai:ses:content:14b0e0b6-efad-4684-a8cc-fe8871ec2b0b",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:149eed98-a02b-46f7-be11-d04177873cc4",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:74ef2588-f54d-46c7-add8-d9bef536a0ca",
                                "name": "status",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:c0353532-a8fb-4553-a00b-418cb8a80a65",
                                "value": null,
                                "serialization_options": null
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "attributes": [],
            "service_group_key": ""
        },
        {
            "id": "urn:infai:ses:service:0f43f156-cd03-4b81-8196-57673e49b15c",
            "local_id": "setColor",
            "name": "Set Color",
            "description": "",
            "interaction": "event",
            "protocol_id": "urn:infai:ses:protocol:3b59ea31-da98-45fd-a354-1b9bd06b837e",
            "inputs": [
                {
                    "id": "urn:infai:ses:content:80111071-fc44-4d06-b443-14dc33685a67",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:05a0f4a4-4cf9-4e17-8e1a-113776856a67",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:1b522486-0764-486f-9ba6-deecd63e0af0",
                                "name": "red",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:dfe6be4a-650c-4411-8d87-062916b48951",
                                "value": null,
                                "serialization_options": null
                            },
                            {
                                "id": "urn:infai:ses:content-variable:bd617a0c-8761-4e43-8548-38d381871f48",
                                "name": "green",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:5ef27837-4aca-43ad-b8f6-4d95cf9ed99e",
                                "value": null,
                                "serialization_options": null
                            },
                            {
                                "id": "urn:infai:ses:content-variable:7112c185-f473-4923-ade4-761ebf2c8c97",
                                "name": "blue",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:590af9ef-3a5e-4edb-abab-177cb1320b17",
                                "value": null,
                                "serialization_options": null
                            },
                            {
                                "id": "urn:infai:ses:content-variable:4df18031-8902-4437-a63c-4ea90b89536f",
                                "name": "duration",
                                "is_void": false,
                                "type": "https://schema.org/Float",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c",
                                "value": null,
                                "serialization_options": null
                            }
                        ],
                        "characteristic_id": "urn:infai:ses:characteristic:5b4eea52-e8e5-4e80-9455-0382f81a1b43",
                        "value": null,
                        "serialization_options": null,
                        "function_id": "urn:infai:ses:controlling-function:c54e2a89-1fb8-4ecb-8993-a7b40b355599",
                        "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "outputs": [
                {
                    "id": "urn:infai:ses:content:37c36a9d-0250-4eeb-a568-480ca692be18",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:12b59e29-adad-49d2-8a8b-37134b4f1e41",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:5513ec3a-47c8-4657-b15b-10bc72c8cfc9",
                                "name": "status",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:c0353532-a8fb-4553-a00b-418cb8a80a65",
                                "value": null,
                                "serialization_options": null
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "attributes": [],
            "service_group_key": ""
        },
        {
            "id": "urn:infai:ses:service:6add6a31-c5f3-41e4-8224-3004d4407ec1",
            "local_id": "setKelvin",
            "name": "Set Color Temperature",
            "description": "",
            "interaction": "event",
            "protocol_id": "urn:infai:ses:protocol:3b59ea31-da98-45fd-a354-1b9bd06b837e",
            "inputs": [
                {
                    "id": "urn:infai:ses:content:6fd06e48-eeb4-4935-a8f9-8eab8fff5306",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:7e39c04f-203b-49d2-b540-c8c4870efdfd",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:73caed73-3224-4861-b543-ffd70d70eca0",
                                "name": "kelvin",
                                "is_void": false,
                                "type": "https://schema.org/Float",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:75b2d113-1d03-4ef8-977a-8dbcbb31a683",
                                "value": null,
                                "serialization_options": null,
                                "function_id": "urn:infai:ses:controlling-function:11ede745-afb3-41a6-98fc-6942d0e9cb33",
                                "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                            },
                            {
                                "id": "urn:infai:ses:content-variable:7bde920a-ca2f-4c57-bf0d-dd866fce0589",
                                "name": "duration",
                                "is_void": false,
                                "type": "https://schema.org/Float",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c",
                                "value": null,
                                "serialization_options": null
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "outputs": [
                {
                    "id": "urn:infai:ses:content:ca782d35-8ae9-413b-bac9-ad73450b1211",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:bd9becc2-a8fe-42b1-9de0-b26cd58e5e20",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:fdde22dd-60c5-4967-904b-ddecb9dab100",
                                "name": "status",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:c0353532-a8fb-4553-a00b-418cb8a80a65",
                                "value": null,
                                "serialization_options": null
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "attributes": [],
            "service_group_key": ""
        },
{
            "id": "urn:infai:ses:service:6add6a31-c5f3-41e4-8224-3004d4407ec2",
            "local_id": "setKelvin2",
            "name": "Set Color Temperature 2",
            "description": "",
            "interaction": "event",
            "protocol_id": "urn:infai:ses:protocol:3b59ea31-da98-45fd-a354-1b9bd06b837e",
            "inputs": [
                {
                    "id": "urn:infai:ses:content:6fd06e48-eeb4-4935-a8f9-8eab8fff5302",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:7e39c04f-203b-49d2-b540-c8c4870efdf2",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:73caed73-3224-4861-b543-ffd70d70eca2",
                                "name": "kelvin",
                                "is_void": false,
                                "type": "https://schema.org/Float",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:75b2d113-1d03-4ef8-977a-8dbcbb31a683",
                                "value": null,
                                "serialization_options": null,
                                "function_id": "urn:infai:ses:controlling-function:11ede745-afb3-41a6-98fc-6942d0e9cb33",
                                "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                            },
                            {
                                "id": "urn:infai:ses:content-variable:7bde920a-ca2f-4c57-bf0d-dd866fce0582",
                                "name": "duration",
                                "is_void": false,
                                "type": "https://schema.org/Float",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c",
                                "value": null,
                                "serialization_options": null
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "outputs": [
                {
                    "id": "urn:infai:ses:content:ca782d35-8ae9-413b-bac9-ad73450b1212",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:bd9becc2-a8fe-42b1-9de0-b26cd58e5e22",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:fdde22dd-60c5-4967-904b-ddecb9dab102",
                                "name": "status",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:c0353532-a8fb-4553-a00b-418cb8a80a65",
                                "value": null,
                                "serialization_options": null
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "attributes": [],
            "service_group_key": ""
        },
        {
            "id": "urn:infai:ses:service:59dd05fc-cd67-4f66-98de-bbed8257a868",
            "local_id": "setPower",
            "name": "Set Power Off",
            "description": "",
            "interaction": "event",
            "protocol_id": "urn:infai:ses:protocol:3b59ea31-da98-45fd-a354-1b9bd06b837e",
            "inputs": [
                {
                    "id": "urn:infai:ses:content:9f6ac32e-5f26-423b-947a-8a769773239a",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:b3cd09ff-8d5b-4b35-abad-d998bd46a05f",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:18c19fcf-9ded-4393-995f-6ff751c73d83",
                                "name": "power",
                                "is_void": false,
                                "type": "https://schema.org/Boolean",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:7dc1bb7e-b256-408a-a6f9-044dc60fdcf5",
                                "value": false,
                                "serialization_options": null,
                                "function_id": "urn:infai:ses:controlling-function:2f35150b-9df7-4cad-95bc-165fa00219fd",
                                "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "outputs": [
                {
                    "id": "urn:infai:ses:content:77e18f9c-27f7-4755-82b5-9baba3196666",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:cbab719c-5105-49f0-9419-e3733ffae1b9",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:82d51640-9907-4303-b97c-514e8de7c7ad",
                                "name": "status",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:c0353532-a8fb-4553-a00b-418cb8a80a65",
                                "value": null,
                                "serialization_options": null
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "attributes": [],
            "service_group_key": ""
        },
        {
            "id": "urn:infai:ses:service:37ec6222-252a-4025-9f0b-18d2598c47c3",
            "local_id": "setPower",
            "name": "Set Power On",
            "description": "",
            "interaction": "event",
            "protocol_id": "urn:infai:ses:protocol:3b59ea31-da98-45fd-a354-1b9bd06b837e",
            "inputs": [
                {
                    "id": "urn:infai:ses:content:afaf90a4-f76b-44e5-8c81-008b4189cdcb",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:fe0fa8e2-304c-4087-83a5-3f01dcb36b64",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:796bd355-0b42-44de-acef-0c79b912a1b2",
                                "name": "power",
                                "is_void": false,
                                "type": "https://schema.org/Boolean",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:7dc1bb7e-b256-408a-a6f9-044dc60fdcf5",
                                "value": true,
                                "serialization_options": null,
                                "function_id": "urn:infai:ses:controlling-function:79e7914b-f303-4a7d-90af-dee70db05fd9",
                                "aspect_id": "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "outputs": [
                {
                    "id": "urn:infai:ses:content:d00dd025-e7e7-4e15-bb80-5050f74a1f9a",
                    "content_variable": {
                        "id": "urn:infai:ses:content-variable:638b96e7-8b06-4b47-93b4-bbc983133724",
                        "name": "struct",
                        "is_void": false,
                        "type": "https://schema.org/StructuredValue",
                        "sub_content_variables": [
                            {
                                "id": "urn:infai:ses:content-variable:8a8c558c-a168-4bf0-b7a5-5035d4c88d20",
                                "name": "status",
                                "is_void": false,
                                "type": "https://schema.org/Integer",
                                "sub_content_variables": null,
                                "characteristic_id": "urn:infai:ses:characteristic:c0353532-a8fb-4553-a00b-418cb8a80a65",
                                "value": null,
                                "serialization_options": null
                            }
                        ],
                        "characteristic_id": "",
                        "value": null,
                        "serialization_options": null
                    },
                    "serialization": "json",
                    "protocol_segment_id": "urn:infai:ses:protocol-segment:05f1467c-95c8-4a83-a1ed-1c8369fd158a"
                }
            ],
            "attributes": [],
            "service_group_key": ""
        }
    ],
    "device_class_id": "urn:infai:ses:device-class:14e56881-16f9-4120-bb41-270a43070c86",
    "attributes": []
}`
	dt := model.DeviceType{}
	err = json.Unmarshal([]byte(deviceTypeStr), &dt)
	if err != nil {
		t.Error(err)
		return
	}
	mock.Repo.RegisterDeviceType(dt)

	cmd1 := messages.Command{
		Version:  3,
		Function: model.Function{RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION, Id: "urn:infai:ses:measuring-function:fb0f474f-c1d7-4a90-971b-a0b9915968f0"},
		DeviceId: "urn:infai:ses:device:423a2718-dea0-4f69-85a3-626c52de175b",
		Aspect:   &model.AspectNode{Id: "urn:infai:ses:aspect:a7470d73-dde3-41fc-92bd-f16bb28f2da6"},
	}

	cmdMsg1, err := json.Marshal(cmd1)
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

	fetched, completed, failed := mockCamunda.GetStatus()

	if len(fetched) != 0 || len(failed) != 0 || len(completed) != 1 {
		log.Println("fetched:", fetched)
		log.Println("failed:", failed)
		log.Println("completed:", completed)
		log.Println(len(fetched), len(failed), len(completed))
		return
	}
}

func TestDeviceEventWithoutServiceResponses(t *testing.T) {
	mock.CleanKafkaMock()
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

	config.TimescaleWrapperUrl = "placeholder"

	config.SubResultDatabaseUrls = []string{"localhost:" + port}

	config.GroupScheduler = util.PARALLEL
	config.CompletionStrategy = util.PESSIMISTIC
	config.Debug = true

	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()

	timescaleResponses := mock.TimescaleMockResponses{
		"device_1": {
			"service_1": {"metrics.level": "#c8320{{count}}"},
			"service_2": {"metrics.level": "#c8320{{count}}"},
		},
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
				Interaction: model.EVENT,
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
				Interaction: model.EVENT,
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

	cmd1 := messages.Command{
		Version:          2,
		Function:         model.Function{Id: model.MEASURING_FUNCTION_PREFIX + "f1", RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
		Aspect:           &model.AspectNode{Id: "a1"},
		CharacteristicId: example.Rgb,
		DeviceId:         "device_1",
	}

	cmdMsg1, err := json.Marshal(cmd1)
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

	fetched, completed, failed := mockCamunda.GetStatus()

	if len(fetched) != 0 || len(failed) != 0 || len(completed) != 1 {
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
		`[{"b":0,"g":50,"r":200},{"b":1,"g":50,"r":200}]`,
	}) {
		t.Error(list)
		return
	}
}
