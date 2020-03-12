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
	"sort"

	"log"
	"time"
)

func ExampleWorkerResponseWithHints() {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.PESSIMISTIC
	config.CamundaWorkerTimeout = 100

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mock.Camunda, mock.Marshaller)

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
							Id:               "level1",
							Name:             "level1",
							Type:             model.Integer,
							CharacteristicId: example.Hex,
						},
						{
							Id:               "level2",
							Name:             "level2",
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
		Function:             model.Function{RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
		CharacteristicId:     example.Rgb,
		DeviceId:             "device_1",
		ServiceId:            "service_1",
		ProtocolId:           "p1",
		ContentVariableHints: []string{"level1", "foo", "bar"},
	}

	cmdMsg1, err := json.Marshal(cmd1)
	if err != nil {
		log.Fatal(err)
	}

	cmd2 := messages.Command{
		Function:             model.Function{RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
		CharacteristicId:     example.Rgb,
		DeviceId:             "device_1",
		ServiceId:            "service_1",
		ProtocolId:           "p1",
		ContentVariableHints: []string{"level2", "foo", "bar"},
	}

	cmdMsg2, err := json.Marshal(cmd2)
	if err != nil {
		log.Fatal(err)
	}

	mock.Camunda.AddTask(messages.CamundaExternalTask{
		Id: "1",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg1),
			},
		},
	})

	time.Sleep(1 * time.Second)

	mock.Camunda.AddTask(messages.CamundaExternalTask{
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
		log.Fatal(protocolMessageStrings)
	}

	for _, message := range protocolMessageStrings {
		msg := messages.ProtocolMsg{}
		err = json.Unmarshal([]byte(message), &msg)
		if err != nil {
			log.Fatal(err)
		}
		msg.Response.Output = map[string]string{
			"body": "{\"level1\":\"#c83200\",\"level2\":\"#ff00aa\"}",
		}
		resp, err := json.Marshal(msg)
		if err != nil {
			log.Fatal(err)
		}
		mock.Kafka.Produce(config.ResponseTopic, string(resp))
		time.Sleep(1 * time.Second)
	}

	fetched, completed, failed := mock.Camunda.GetStatus()

	if len(fetched) != 0 || len(failed) != 0 || len(completed) != 2 {
		log.Println("fetched:", fetched)
		log.Println("failed:", failed)
		log.Println("completed:", completed)
		log.Fatal(len(fetched), len(failed), len(completed))
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
	for _, cmd := range list {
		fmt.Println(cmd)
	}

	//output:
	//{"b":0,"g":50,"r":200}
	//{"b":170,"g":0,"r":255}
}
