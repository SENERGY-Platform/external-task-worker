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

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"

	"log"
	"time"
)

func ExampleConfigurablesCommand1() {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.OPTIMISTIC

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller, mock.Timescale)

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
		Inputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:   "color",
							Name: "color",
							Type: model.Structure,
							SubContentVariables: []model.ContentVariable{
								{
									Id:               "sr",
									Name:             "red",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".r",
								},
								{
									Id:               "sg",
									Name:             "green",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".g",
								},
								{
									Id:               "sb",
									Name:             "blue",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".b",
								},
							},
						},
						{
							Id:               "duration",
							Name:             "duration",
							Type:             model.Float,
							CharacteristicId: duration.Seconds,
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	})

	cmd1 := messages.Command{
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION},
		CharacteristicId: example.Hex,
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
		Input:            "#ff00aa",
		Configurables: []marshaller.Configurable{
			{
				CharacteristicId: duration.Seconds,
				Values: []marshaller.ConfigurableCharacteristicValue{
					{Value: "3"},
				},
			},
		},
	}

	cmdMsg1, err := json.Marshal(cmd1)
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

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	for _, message := range protocolMessageStrings {
		var temp messages.ProtocolMsg
		json.Unmarshal([]byte(message), &temp)
		temp.Trace = nil
		messageWithoutTrace, _ := json.Marshal(temp)
		fmt.Println(string(messageWithoutTrace))
	}

	//output:
	//{"request":{"input":{"body":"{\"color\":{\"blue\":170,\"green\":0,\"red\":255},\"duration\":3}"}},"response":{"output":null},"task_info":{"worker_id":"workerid","task_id":"1","process_instance_id":"","process_definition_id":"","completion_strategy":"optimistic","time":"-62135596800","tenant_id":"user"},"metadata":{"device":{"id":"device_1","local_id":"d1u","name":"d1","attributes":null,"device_type_id":"dt1","owner_id":""},"service":{"id":"service_1","local_id":"s1u","name":"s1","description":"","interaction":"","protocol_id":"p1","inputs":[{"id":"metrics","content_variable":{"id":"metrics","name":"metrics","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"color","name":"color","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"sr","name":"red","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.r","value":null,"serialization_options":null},{"id":"sg","name":"green","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.g","value":null,"serialization_options":null},{"id":"sb","name":"blue","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.b","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},{"id":"duration","name":"duration","is_void":false,"omit_empty":false,"type":"https://schema.org/Float","sub_content_variables":null,"characteristic_id":"urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"ms1"}],"outputs":null,"attributes":null,"service_group_key":""},"protocol":{"id":"p1","name":"protocol1","handler":"protocol1","protocol_segments":[{"id":"ms1","name":"body"}],"constraints":null},"input_characteristic":"example_hex","response_to":"response","error_to":"errors"}}
}

func ExampleConfigurablesCommand2() {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.OPTIMISTIC

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller, mock.Timescale)

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
		Inputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:   "color",
							Name: "color",
							Type: model.Structure,
							SubContentVariables: []model.ContentVariable{
								{
									Id:               "sr",
									Name:             "red",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".r",
								},
								{
									Id:               "sg",
									Name:             "green",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".g",
								},
								{
									Id:               "sb",
									Name:             "blue",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".b",
								},
							},
						},
						{
							Id:               "duration",
							Name:             "duration",
							Type:             model.Float,
							CharacteristicId: duration.Seconds,
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	})

	cmd2 := messages.Command{
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION},
		CharacteristicId: duration.Seconds,
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
		Input:            3,
		Configurables: []marshaller.Configurable{
			{
				CharacteristicId: example.Hex,
				Values: []marshaller.ConfigurableCharacteristicValue{
					{Value: `"#ff00aa"`},
				},
			},
		},
	}

	cmdMsg2, err := json.Marshal(cmd2)
	if err != nil {
		log.Fatal(err)
	}

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "2",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg2),
			},
		},
	})

	time.Sleep(1 * time.Second)

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	for _, message := range protocolMessageStrings {
		var temp messages.ProtocolMsg
		json.Unmarshal([]byte(message), &temp)
		temp.Trace = nil
		messageWithoutTrace, _ := json.Marshal(temp)
		fmt.Println(string(messageWithoutTrace))
	}

	//output:
	//{"request":{"input":{"body":"{\"color\":{\"blue\":170,\"green\":0,\"red\":255},\"duration\":3}"}},"response":{"output":null},"task_info":{"worker_id":"workerid","task_id":"2","process_instance_id":"","process_definition_id":"","completion_strategy":"optimistic","time":"-62135596800","tenant_id":"user"},"metadata":{"device":{"id":"device_1","local_id":"d1u","name":"d1","attributes":null,"device_type_id":"dt1","owner_id":""},"service":{"id":"service_1","local_id":"s1u","name":"s1","description":"","interaction":"","protocol_id":"p1","inputs":[{"id":"metrics","content_variable":{"id":"metrics","name":"metrics","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"color","name":"color","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"sr","name":"red","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.r","value":null,"serialization_options":null},{"id":"sg","name":"green","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.g","value":null,"serialization_options":null},{"id":"sb","name":"blue","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.b","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},{"id":"duration","name":"duration","is_void":false,"omit_empty":false,"type":"https://schema.org/Float","sub_content_variables":null,"characteristic_id":"urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"ms1"}],"outputs":null,"attributes":null,"service_group_key":""},"protocol":{"id":"p1","name":"protocol1","handler":"protocol1","protocol_segments":[{"id":"ms1","name":"body"}],"constraints":null},"input_characteristic":"urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c","response_to":"response","error_to":"errors"}}
}

func ExampleConfigurablesCommand3() {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.OPTIMISTIC

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller, mock.Timescale)

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
		Inputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:   "color",
							Name: "color",
							Type: model.Structure,
							SubContentVariables: []model.ContentVariable{
								{
									Id:               "sr",
									Name:             "red",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".r",
								},
								{
									Id:               "sg",
									Name:             "green",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".g",
								},
								{
									Id:               "sb",
									Name:             "blue",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".b",
								},
							},
						},
						{
							Id:               "duration",
							Name:             "duration",
							Type:             model.Float,
							CharacteristicId: duration.Seconds,
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	})

	cmd3 := messages.Command{
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION},
		CharacteristicId: duration.Seconds,
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
		Input:            3,
		Configurables: []marshaller.Configurable{
			{
				CharacteristicId: example.Rgb,
				Values: []marshaller.ConfigurableCharacteristicValue{
					{
						Path:  "r",
						Value: "255",
					},
					{
						Path:  "g",
						Value: "0",
					},
					{
						Path:  "b",
						Value: "170",
					},
				},
			},
		},
	}

	cmdMsg3, err := json.Marshal(cmd3)
	if err != nil {
		log.Fatal(err)
	}

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "3",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg3),
			},
		},
	})

	time.Sleep(1 * time.Second)

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	for _, message := range protocolMessageStrings {
		var temp messages.ProtocolMsg
		json.Unmarshal([]byte(message), &temp)
		temp.Trace = nil
		messageWithoutTrace, _ := json.Marshal(temp)
		fmt.Println(string(messageWithoutTrace))
	}

	//output:
	//{"request":{"input":{"body":"{\"color\":{\"blue\":170,\"green\":0,\"red\":255},\"duration\":3}"}},"response":{"output":null},"task_info":{"worker_id":"workerid","task_id":"3","process_instance_id":"","process_definition_id":"","completion_strategy":"optimistic","time":"-62135596800","tenant_id":"user"},"metadata":{"device":{"id":"device_1","local_id":"d1u","name":"d1","attributes":null,"device_type_id":"dt1","owner_id":""},"service":{"id":"service_1","local_id":"s1u","name":"s1","description":"","interaction":"","protocol_id":"p1","inputs":[{"id":"metrics","content_variable":{"id":"metrics","name":"metrics","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"color","name":"color","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"sr","name":"red","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.r","value":null,"serialization_options":null},{"id":"sg","name":"green","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.g","value":null,"serialization_options":null},{"id":"sb","name":"blue","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.b","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},{"id":"duration","name":"duration","is_void":false,"omit_empty":false,"type":"https://schema.org/Float","sub_content_variables":null,"characteristic_id":"urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"ms1"}],"outputs":null,"attributes":null,"service_group_key":""},"protocol":{"id":"p1","name":"protocol1","handler":"protocol1","protocol_segments":[{"id":"ms1","name":"body"}],"constraints":null},"input_characteristic":"urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c","response_to":"response","error_to":"errors"}}
}

func ExampleConfigurablesCommand4() {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.OPTIMISTIC

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller, mock.Timescale)

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
		Inputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:   "color",
							Name: "color",
							Type: model.Structure,
							SubContentVariables: []model.ContentVariable{
								{
									Id:               "sr",
									Name:             "red",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".r",
								},
								{
									Id:               "sg",
									Name:             "green",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".g",
								},
								{
									Id:               "sb",
									Name:             "blue",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".b",
								},
							},
						},
						{
							Id:               "duration",
							Name:             "duration",
							Type:             model.Float,
							CharacteristicId: duration.Seconds,
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	})

	cmd := messages.Command{
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION},
		CharacteristicId: duration.Seconds,
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
		Input:            3,
		Configurables: []marshaller.Configurable{
			{
				CharacteristicId: temperature.Celcius,
				Values: []marshaller.ConfigurableCharacteristicValue{
					{
						Value: "255",
					},
				},
			},
		},
	}

	cmdMsg, err := json.Marshal(cmd)
	if err != nil {
		log.Fatal(err)
	}

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "3",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg),
			},
		},
	})

	time.Sleep(1 * time.Second)

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	for _, message := range protocolMessageStrings {
		var temp messages.ProtocolMsg
		json.Unmarshal([]byte(message), &temp)
		temp.Trace = nil
		messageWithoutTrace, _ := json.Marshal(temp)
		fmt.Println(string(messageWithoutTrace))
	}

	//output:
	//{"request":{"input":{"body":"{\"color\":{\"blue\":0,\"green\":0,\"red\":0},\"duration\":3}"}},"response":{"output":null},"task_info":{"worker_id":"workerid","task_id":"3","process_instance_id":"","process_definition_id":"","completion_strategy":"optimistic","time":"-62135596800","tenant_id":"user"},"metadata":{"device":{"id":"device_1","local_id":"d1u","name":"d1","attributes":null,"device_type_id":"dt1","owner_id":""},"service":{"id":"service_1","local_id":"s1u","name":"s1","description":"","interaction":"","protocol_id":"p1","inputs":[{"id":"metrics","content_variable":{"id":"metrics","name":"metrics","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"color","name":"color","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"sr","name":"red","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.r","value":null,"serialization_options":null},{"id":"sg","name":"green","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.g","value":null,"serialization_options":null},{"id":"sb","name":"blue","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.b","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},{"id":"duration","name":"duration","is_void":false,"omit_empty":false,"type":"https://schema.org/Float","sub_content_variables":null,"characteristic_id":"urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"ms1"}],"outputs":null,"attributes":null,"service_group_key":""},"protocol":{"id":"p1","name":"protocol1","handler":"protocol1","protocol_segments":[{"id":"ms1","name":"body"}],"constraints":null},"input_characteristic":"urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c","response_to":"response","error_to":"errors"}}
}

func ExampleConfigurablesCommand5() {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.OPTIMISTIC

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller, mock.Timescale)

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
		Inputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:   "color",
							Name: "color",
							Type: model.Structure,
							SubContentVariables: []model.ContentVariable{
								{
									Id:               "sr",
									Name:             "red",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".r",
								},
								{
									Id:               "sg",
									Name:             "green",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".g",
								},
								{
									Id:               "sb",
									Name:             "blue",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".b",
								},
							},
						},
						{
							Id:               "duration",
							Name:             "duration",
							Type:             model.Float,
							CharacteristicId: duration.Seconds,
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	})

	cmd := messages.Command{
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION},
		CharacteristicId: duration.Seconds,
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
		Input:            3,
		Configurables: []marshaller.Configurable{
			{
				CharacteristicId: temperature.Celcius,
				Values: []marshaller.ConfigurableCharacteristicValue{
					{
						Value: "255",
					},
				},
			},
			{
				CharacteristicId: example.Rgb,
				Values: []marshaller.ConfigurableCharacteristicValue{
					{
						Path:  "r",
						Value: "255",
					},
					{
						Path:  "g",
						Value: "0",
					},
					{
						Path:  "b",
						Value: "170",
					},
				},
			},
		},
	}

	cmdMsg, err := json.Marshal(cmd)
	if err != nil {
		log.Fatal(err)
	}

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "3",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg),
			},
		},
	})

	time.Sleep(1 * time.Second)

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	for _, message := range protocolMessageStrings {
		var temp messages.ProtocolMsg
		json.Unmarshal([]byte(message), &temp)
		temp.Trace = nil
		messageWithoutTrace, _ := json.Marshal(temp)
		fmt.Println(string(messageWithoutTrace))
	}

	//output:
	//{"request":{"input":{"body":"{\"color\":{\"blue\":170,\"green\":0,\"red\":255},\"duration\":3}"}},"response":{"output":null},"task_info":{"worker_id":"workerid","task_id":"3","process_instance_id":"","process_definition_id":"","completion_strategy":"optimistic","time":"-62135596800","tenant_id":"user"},"metadata":{"device":{"id":"device_1","local_id":"d1u","name":"d1","attributes":null,"device_type_id":"dt1","owner_id":""},"service":{"id":"service_1","local_id":"s1u","name":"s1","description":"","interaction":"","protocol_id":"p1","inputs":[{"id":"metrics","content_variable":{"id":"metrics","name":"metrics","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"color","name":"color","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"sr","name":"red","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.r","value":null,"serialization_options":null},{"id":"sg","name":"green","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.g","value":null,"serialization_options":null},{"id":"sb","name":"blue","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.b","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},{"id":"duration","name":"duration","is_void":false,"omit_empty":false,"type":"https://schema.org/Float","sub_content_variables":null,"characteristic_id":"urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"ms1"}],"outputs":null,"attributes":null,"service_group_key":""},"protocol":{"id":"p1","name":"protocol1","handler":"protocol1","protocol_segments":[{"id":"ms1","name":"body"}],"constraints":null},"input_characteristic":"urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c","response_to":"response","error_to":"errors"}}
}

func ExampleConfigurablesCommandWithoutFunctionRdfType() {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.OPTIMISTIC

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mockCamunda, mock.Marshaller, mock.Timescale)

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
		Inputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:   "color",
							Name: "color",
							Type: model.Structure,
							SubContentVariables: []model.ContentVariable{
								{
									Id:               "sr",
									Name:             "red",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".r",
								},
								{
									Id:               "sg",
									Name:             "green",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".g",
								},
								{
									Id:               "sb",
									Name:             "blue",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".b",
								},
							},
						},
						{
							Id:               "duration",
							Name:             "duration",
							Type:             model.Float,
							CharacteristicId: duration.Seconds,
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	})

	cmd := messages.Command{
		Function:         model.Function{Id: "urn:infai:ses:controlling-function:foobar"},
		CharacteristicId: duration.Seconds,
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
		Input:            3,
		Configurables: []marshaller.Configurable{
			{
				CharacteristicId: temperature.Celcius,
				Values: []marshaller.ConfigurableCharacteristicValue{
					{
						Value: "255",
					},
				},
			},
			{
				CharacteristicId: example.Rgb,
				Values: []marshaller.ConfigurableCharacteristicValue{
					{
						Path:  "r",
						Value: "255",
					},
					{
						Path:  "g",
						Value: "0",
					},
					{
						Path:  "b",
						Value: "170",
					},
				},
			},
		},
	}

	cmdMsg, err := json.Marshal(cmd)
	if err != nil {
		log.Fatal(err)
	}

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:       "3",
		TenantId: "user",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg),
			},
		},
	})

	time.Sleep(1 * time.Second)

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	for _, message := range protocolMessageStrings {
		var temp messages.ProtocolMsg
		json.Unmarshal([]byte(message), &temp)
		temp.Trace = nil
		messageWithoutTrace, _ := json.Marshal(temp)
		fmt.Println(string(messageWithoutTrace))
	}

	//output:
	//{"request":{"input":{"body":"{\"color\":{\"blue\":170,\"green\":0,\"red\":255},\"duration\":3}"}},"response":{"output":null},"task_info":{"worker_id":"workerid","task_id":"3","process_instance_id":"","process_definition_id":"","completion_strategy":"optimistic","time":"-62135596800","tenant_id":"user"},"metadata":{"device":{"id":"device_1","local_id":"d1u","name":"d1","attributes":null,"device_type_id":"dt1","owner_id":""},"service":{"id":"service_1","local_id":"s1u","name":"s1","description":"","interaction":"","protocol_id":"p1","inputs":[{"id":"metrics","content_variable":{"id":"metrics","name":"metrics","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"color","name":"color","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"sr","name":"red","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.r","value":null,"serialization_options":null},{"id":"sg","name":"green","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.g","value":null,"serialization_options":null},{"id":"sb","name":"blue","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_rgb.b","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},{"id":"duration","name":"duration","is_void":false,"omit_empty":false,"type":"https://schema.org/Float","sub_content_variables":null,"characteristic_id":"urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"ms1"}],"outputs":null,"attributes":null,"service_group_key":""},"protocol":{"id":"p1","name":"protocol1","handler":"protocol1","protocol_segments":[{"id":"ms1","name":"body"}],"constraints":null},"input_characteristic":"urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c","response_to":"response","error_to":"errors"}}
}
