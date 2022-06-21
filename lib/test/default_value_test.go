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
	"testing"
	"time"
)

func TestDefaultValue(t *testing.T) {
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
	mock.Repo.RegisterDevice(model.Device{
		Id:           "device_1",
		Name:         "d1",
		DeviceTypeId: "dt1",
		LocalId:      "d1u",
	})

	mock.Repo.RegisterProtocol(model.Protocol{
		Id:               "urn:infai:ses:protocol:f3a63aeb-187e-4dd9-9ef5-d97a6eb6292b",
		Name:             "protocol1",
		Handler:          "protocol1",
		ProtocolSegments: []model.ProtocolSegment{{Id: "urn:infai:ses:protocol-segment:0d211842-cef8-41ec-ab6b-9dbc31bc3a65", Name: "body"}},
	})

	serviceJson := `{
   "id":"urn:infai:ses:service:494be526-d9df-459a-887d-b427ef6cc50e",
   "local_id":"37-0-targetValue",
   "name":"Set Device On",
   "description":"",
   "interaction":"request",
   "protocol_id":"urn:infai:ses:protocol:f3a63aeb-187e-4dd9-9ef5-d97a6eb6292b",
   "inputs":[
      {
         "id":"urn:infai:ses:content:b4449aba-7edd-43b5-9305-c71c8c701899",
         "content_variable":{
            "id":"urn:infai:ses:content-variable:18ea8ab0-f59c-4214-baf2-f2caf299241a",
            "name":"value",
            "is_void":false,
            "type":"https://schema.org/Boolean",
            "sub_content_variables":null,
            "characteristic_id":"",
            "value":true,
            "serialization_options":null,
            "function_id":"urn:infai:ses:controlling-function:79e7914b-f303-4a7d-90af-dee70db05fd9",
            "aspect_id":"urn:infai:ses:aspect:861227f6-1523-46a7-b8ab-a4e76f0bdd32"
         },
         "serialization":"json",
         "protocol_segment_id":"urn:infai:ses:protocol-segment:0d211842-cef8-41ec-ab6b-9dbc31bc3a65"
      }
   ],
   "outputs":[
      
   ],
   "attributes":[
      
   ],
   "service_group_key":""
}`

	service := model.Service{}
	err = json.Unmarshal([]byte(serviceJson), &service)
	if err != nil {
		t.Error(err)
		return
	}
	mock.Repo.RegisterService(service)

	cmd1 := messages.Command{
		Version:  3,
		Function: model.Function{RdfType: model.SES_ONTOLOGY_CONTROLLING_FUNCTION, Id: "urn:infai:ses:controlling-function:79e7914b-f303-4a7d-90af-dee70db05fd9"},
		DeviceClass: &model.DeviceClass{
			Id: "urn:infai:ses:device-class:79de1bd9-b933-412d-b98e-4cfe19aa3250",
		},
		Aspect:     nil,
		DeviceId:   "device_1",
		ServiceId:  "urn:infai:ses:service:494be526-d9df-459a-887d-b427ef6cc50e",
		ProtocolId: "urn:infai:ses:protocol:f3a63aeb-187e-4dd9-9ef5-d97a6eb6292b",
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

	expected := `{"request":{"input":{"body":"true"}},"response":{"output":null},"task_info":{"worker_id":"workerid","task_id":"1","process_instance_id":"","process_definition_id":"","completion_strategy":"optimistic","time":"-62135596800","tenant_id":"user"},"metadata":{"version":3,"device":{"id":"device_1","local_id":"d1u","name":"d1","attributes":null,"device_type_id":"dt1"},"service":{"id":"urn:infai:ses:service:494be526-d9df-459a-887d-b427ef6cc50e","local_id":"37-0-targetValue","name":"Set Device On","description":"","interaction":"request","protocol_id":"urn:infai:ses:protocol:f3a63aeb-187e-4dd9-9ef5-d97a6eb6292b","inputs":[{"id":"urn:infai:ses:content:b4449aba-7edd-43b5-9305-c71c8c701899","content_variable":{"id":"urn:infai:ses:content-variable:18ea8ab0-f59c-4214-baf2-f2caf299241a","name":"value","is_void":false,"type":"https://schema.org/Boolean","sub_content_variables":null,"characteristic_id":"","value":true,"serialization_options":null,"function_id":"urn:infai:ses:controlling-function:79e7914b-f303-4a7d-90af-dee70db05fd9","aspect_id":"urn:infai:ses:aspect:861227f6-1523-46a7-b8ab-a4e76f0bdd32"},"serialization":"json","protocol_segment_id":"urn:infai:ses:protocol-segment:0d211842-cef8-41ec-ab6b-9dbc31bc3a65"}],"outputs":[],"attributes":[],"service_group_key":""},"protocol":{"id":"urn:infai:ses:protocol:f3a63aeb-187e-4dd9-9ef5-d97a6eb6292b","name":"protocol1","handler":"protocol1","protocol_segments":[{"id":"urn:infai:ses:protocol-segment:0d211842-cef8-41ec-ab6b-9dbc31bc3a65","name":"body"}]},"response_to":"response","error_to":"errors"}}`

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	if len(protocolMessageStrings) != 1 {
		t.Error(len(protocolMessageStrings), protocolMessageStrings)
		return
	}
	message := protocolMessageStrings[0]
	var temp messages.ProtocolMsg
	json.Unmarshal([]byte(message), &temp)
	temp.Trace = nil
	messageWithoutTrace, _ := json.Marshal(temp)
	if string(messageWithoutTrace) != expected {
		t.Error(string(messageWithoutTrace))
	}
}
