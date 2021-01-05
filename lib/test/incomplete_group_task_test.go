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
	"strconv"
	"testing"
	"time"
)

func TestConstraintGroup(t *testing.T) {
	t.Run("parallel, no missing", testForConstraintGroup(false, map[int]bool{}))
	t.Run("sequential, no missing", testForConstraintGroup(false, map[int]bool{}))

	t.Run("parallel, first missing", testForConstraintGroup(false, map[int]bool{0: true}))
	t.Run("sequential, first missing", testForConstraintGroup(false, map[int]bool{0: true}))

	t.Run("parallel, second missing", testForConstraintGroup(false, map[int]bool{1: true}))
	t.Run("sequential, second missing", testForConstraintGroup(false, map[int]bool{1: true}))

	t.Run("parallel, last missing", testForConstraintGroup(false, map[int]bool{2: true}))
	t.Run("sequential, last missing", testForConstraintGroup(false, map[int]bool{2: true}))

	t.Run("parallel, fist and last missing", testForConstraintGroup(false, map[int]bool{0: true, 2: true}))
	t.Run("sequential, first last missing", testForConstraintGroup(false, map[int]bool{0: true, 2: true}))
}

func testForConstraintGroup(sequential bool, missingResponseForRequestIndex map[int]bool) func(t *testing.T) {
	return func(t *testing.T) {
		mock.CleanKafkaMock()
		util.TimeNow = func() time.Time {
			return time.Time{}
		}
		config, err := util.LoadConfig("../../config.json")
		if err != nil {
			t.Error(err)
			return
		}

		config.CompletionStrategy = util.PESSIMISTIC
		config.CamundaWorkerTimeout = 100
		config.Debug = true

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		mock.Camunda = &mock.CamundaMock{}
		go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mock.Camunda, mock.Marshaller)

		time.Sleep(1 * time.Second)

		counter := 0
		mock.Kafka.Subscribe("protocol1", func(message string) error {
			if !missingResponseForRequestIndex[counter] {
				msg := messages.ProtocolMsg{}
				err = json.Unmarshal([]byte(message), &msg)
				if err != nil {
					log.Fatal(err)
				}
				msg.Response.Output = map[string]string{
					"body": "{\"level\":\"#c8320" + strconv.Itoa(counter) + "\"}",
				}
				resp, err := json.Marshal(msg)
				if err != nil {
					log.Fatal(err)
				}
				mock.Kafka.Produce(config.ResponseTopic, string(resp))
			}
			counter = counter + 1
			return nil
		})

		//populate repository
		mock.Repo.RegisterDevice(model.Device{
			Id:           "device_1",
			Name:         "d1",
			DeviceTypeId: "dt2",
			LocalId:      "d1u",
		})

		mock.Repo.RegisterDevice(model.Device{
			Id:           "device_2",
			Name:         "d2",
			DeviceTypeId: "dt2",
			LocalId:      "d2u",
		})

		mock.Repo.RegisterDevice(model.Device{
			Id:           "device_3",
			Name:         "d3",
			DeviceTypeId: "dt2",
			LocalId:      "d3u",
		})

		mock.Repo.RegisterDeviceGroup(model.DeviceGroup{
			Id:   "dg1",
			Name: "dg1",
			Criteria: []model.DeviceGroupFilterCriteria{
				{FunctionId: model.MEASURING_FUNCTION_PREFIX + "f1", AspectId: "a1", Interaction: model.REQUEST},
			},
			DeviceIds: []string{"device_1", "device_2", "device_3"},
		})

		mock.Repo.RegisterProtocol(model.Protocol{
			Id:               "p1",
			Name:             "protocol1",
			Handler:          "protocol1",
			ProtocolSegments: []model.ProtocolSegment{{Id: "ms1", Name: "body"}},
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

		cmd := messages.Command{
			Version:          2,
			Function:         model.Function{Id: model.MEASURING_FUNCTION_PREFIX + "f1", RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
			Aspect:           &model.Aspect{Id: "a1"},
			CharacteristicId: example.Rgb,
			DeviceGroupId:    "dg1",
		}

		cmdMsg, err := json.Marshal(cmd)
		if err != nil {
			t.Error(err)
			return
		}

		mock.Camunda.AddTask(messages.CamundaExternalTask{
			Id: "1",
			Variables: map[string]messages.CamundaVariable{
				util.CAMUNDA_VARIABLES_PAYLOAD: {
					Value: string(cmdMsg),
				},
			},
		})

		time.Sleep(1 * time.Second)

		fetched, completed, failed := mock.Camunda.GetStatus()

		if len(fetched) != 0 || len(failed) != 0 || len(completed) != 1 {
			log.Println("fetched:", fetched)
			log.Println("failed:", failed)
			log.Println("completed:", completed)
			t.Error(len(fetched), len(failed), len(completed))
			return
		}

		if len(completed) != 1 {
			t.Error(completed)
			return
		}

		expectedResult := []interface{}{}
		for i := 0; i < 3; i++ {
			if !missingResponseForRequestIndex[i] {
				expectedResult = append(expectedResult, map[string]interface{}{
					"b": float64(i),
					"g": float64(50),
					"r": float64(200),
				})
			}
		}

		if !reflect.DeepEqual(completed["1"], expectedResult) {
			t.Error(completed, expectedResult)
		}

	}
}
