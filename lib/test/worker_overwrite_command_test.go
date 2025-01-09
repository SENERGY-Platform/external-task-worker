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
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"

	"log"
	"time"
)

func ExampleWorkerOverwriteCommand() {
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

	cmdOverwrite := messages.Overwrite{
		DeviceId:   "device_1",
		ServiceId:  "service_1",
		ProtocolId: "p1",
	}

	cmdOverwriteStr, err := json.Marshal(cmdOverwrite)
	if err != nil {
		log.Fatal(err)
	}

	cmd1 := messages.Command{
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION},
		CharacteristicId: example.Rgb,
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
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION},
		CharacteristicId: example.Hex,
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
			util.CAMUNDA_VARIABLES_OVERWRITE: {
				Value: string(cmdOverwriteStr),
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
			util.CAMUNDA_VARIABLES_OVERWRITE: {
				Value: string(cmdOverwriteStr),
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
			util.CAMUNDA_VARIABLES_OVERWRITE: {
				Value: string(cmdOverwriteStr),
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
			util.CAMUNDA_VARIABLES_OVERWRITE: {
				Value: string(cmdOverwriteStr),
			},
			"inputs": {Value: "\"#ff00ff\""},
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
	//{"request":{"input":{"body":"{\"level\":\"#c83200\"}"}},"response":{"output":null},"task_info":{"worker_id":"workerid","task_id":"1","process_instance_id":"","process_definition_id":"","completion_strategy":"optimistic","time":"-62135596800","tenant_id":"user"},"metadata":{"device":{"id":"device_1","local_id":"d1u","name":"d1","attributes":null,"device_type_id":"dt1","owner_id":""},"service":{"id":"service_1","local_id":"s1u","name":"s1","description":"","interaction":"","protocol_id":"p1","inputs":[{"id":"metrics","content_variable":{"id":"metrics","name":"metrics","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"level","name":"level","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_hex","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"ms1"}],"outputs":null,"attributes":null,"service_group_key":""},"protocol":{"id":"p1","name":"protocol1","handler":"protocol1","protocol_segments":[{"id":"ms1","name":"body"}],"constraints":null},"input_characteristic":"example_rgb","response_to":"response","error_to":"errors"}}
	//{"request":{"input":{"body":"{\"level\":\"#c832ff\"}"}},"response":{"output":null},"task_info":{"worker_id":"workerid","task_id":"2","process_instance_id":"","process_definition_id":"","completion_strategy":"optimistic","time":"-62135596800","tenant_id":"user"},"metadata":{"device":{"id":"device_1","local_id":"d1u","name":"d1","attributes":null,"device_type_id":"dt1","owner_id":""},"service":{"id":"service_1","local_id":"s1u","name":"s1","description":"","interaction":"","protocol_id":"p1","inputs":[{"id":"metrics","content_variable":{"id":"metrics","name":"metrics","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"level","name":"level","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_hex","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"ms1"}],"outputs":null,"attributes":null,"service_group_key":""},"protocol":{"id":"p1","name":"protocol1","handler":"protocol1","protocol_segments":[{"id":"ms1","name":"body"}],"constraints":null},"input_characteristic":"example_rgb","response_to":"response","error_to":"errors"}}
	//{"request":{"input":{"body":"{\"level\":\"#ff0064\"}"}},"response":{"output":null},"task_info":{"worker_id":"workerid","task_id":"3","process_instance_id":"","process_definition_id":"","completion_strategy":"optimistic","time":"-62135596800","tenant_id":"user"},"metadata":{"device":{"id":"device_1","local_id":"d1u","name":"d1","attributes":null,"device_type_id":"dt1","owner_id":""},"service":{"id":"service_1","local_id":"s1u","name":"s1","description":"","interaction":"","protocol_id":"p1","inputs":[{"id":"metrics","content_variable":{"id":"metrics","name":"metrics","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"level","name":"level","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_hex","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"ms1"}],"outputs":null,"attributes":null,"service_group_key":""},"protocol":{"id":"p1","name":"protocol1","handler":"protocol1","protocol_segments":[{"id":"ms1","name":"body"}],"constraints":null},"input_characteristic":"example_hex","response_to":"response","error_to":"errors"}}
	//{"request":{"input":{"body":"{\"level\":\"#ff00ff\"}"}},"response":{"output":null},"task_info":{"worker_id":"workerid","task_id":"4","process_instance_id":"","process_definition_id":"","completion_strategy":"optimistic","time":"-62135596800","tenant_id":"user"},"metadata":{"device":{"id":"device_1","local_id":"d1u","name":"d1","attributes":null,"device_type_id":"dt1","owner_id":""},"service":{"id":"service_1","local_id":"s1u","name":"s1","description":"","interaction":"","protocol_id":"p1","inputs":[{"id":"metrics","content_variable":{"id":"metrics","name":"metrics","is_void":false,"omit_empty":false,"type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"level","name":"level","is_void":false,"omit_empty":false,"type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_hex","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"ms1"}],"outputs":null,"attributes":null,"service_group_key":""},"protocol":{"id":"p1","name":"protocol1","handler":"protocol1","protocol_segments":[{"id":"ms1","name":"body"}],"constraints":null},"input_characteristic":"example_hex","response_to":"response","error_to":"errors"}}

}
