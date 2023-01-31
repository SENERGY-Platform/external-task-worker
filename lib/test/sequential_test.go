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
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"reflect"
	"testing"
	"time"
)

func TestSequentialExecution(t *testing.T) {
	mock.CleanKafkaMock()
	t.Run("0 retries nil lost", testSequentialExecution(0, nil, []float64{0, 1, 2}, false))
	t.Run("0 retries 0 lost", testSequentialExecution(0, []int{0}, []float64{1, 2}, false))
	t.Run("0 retries 1 lost", testSequentialExecution(0, []int{1}, []float64{0, 2}, false))
	t.Run("0 retries 2 lost", testSequentialExecution(0, []int{2}, []float64{0, 1}, false))
	t.Run("0 retries 0 1 lost", testSequentialExecution(0, []int{0, 1}, []float64{2}, false))
	t.Run("0 retries 0 2 lost", testSequentialExecution(0, []int{0, 2}, []float64{1}, false))
	t.Run("0 retries 1 2 lost", testSequentialExecution(0, []int{1, 2}, []float64{0}, false))
	t.Run("0 retries 0 1 2 lost", testSequentialExecution(0, []int{0, 1, 2}, nil, true))

	t.Run("1 retries nil lost", testSequentialExecution(1, nil, []float64{0, 1, 2}, false))
	t.Run("1 retries 0 lost", testSequentialExecution(1, []int{0}, []float64{1, 2, 3}, false))
	t.Run("1 retries 1 lost", testSequentialExecution(1, []int{1}, []float64{0, 2, 3}, false))

	//may result in fail of first sub-task -> []float64{2, 3, 4} becomes []float64{2, 3}
	t.Run("1 retries 0 1 lost", testSequentialExecution(1, []int{0, 1}, []float64{2, 3}, false))

	t.Run("1 retries 2 lost", testSequentialExecution(1, []int{2}, []float64{0, 1, 3}, false))
	t.Run("1 retries 0 2 lost", testSequentialExecution(1, []int{0, 2}, []float64{1, 3, 4}, false))

	//may result in fail of second sub-task -> []float64{0, 3, 4} becomes []float64{0, 3}
	t.Run("1 retries 1 2 lost", testSequentialExecution(1, []int{1, 2}, []float64{0, 3}, false))

	//may result in fail of first sub-task -> []float64{3, 4, 5} becomes []float64{3, 4}
	t.Run("1 retries 0 1 2 lost", testSequentialExecution(1, []int{0, 1, 2}, []float64{3, 4}, false))

	t.Run("1 retries 0 3 lost", testSequentialExecution(1, []int{0, 3}, []float64{1, 2, 4}, false))
	t.Run("1 retries 1 3 lost", testSequentialExecution(1, []int{1, 3}, []float64{0, 2, 4}, false))

	//may result in fail of last sub-task -> []float64{0, 1, 4} becomes []float64{0, 1}
	t.Run("1 retries 2 3 lost", testSequentialExecution(1, []int{2, 3}, []float64{0, 1}, false))

	//may result in fail of second sub-task -> []float64{1, 4, 5} becomes []float64{1, 4}
	t.Run("1 retries 0 2 3 lost", testSequentialExecution(1, []int{0, 2, 3}, []float64{1, 4}, false))

	//may result in fail of second sub-task -> []float64{0, 4, 5} becomes []float64{0, 4}
	t.Run("1 retries 1 2 3 lost", testSequentialExecution(1, []int{1, 2, 3}, []float64{0, 4}, false))

	//may result in fail of first and second sub-task -> []float64{4, 5, 6} becomes []float64{4}
	t.Run("1 retries 0 1 2 3 lost", testSequentialExecution(1, []int{0, 1, 2, 3}, []float64{4}, false))

	//may result in fail of first and second sub-task -> []float64{5, 6, 7} becomes []float64{5}
	t.Run("1 retries 0 1 2 3 4 lost", testSequentialExecution(1, []int{0, 1, 2, 3, 4}, []float64{5}, false))

	t.Run("1 retries 0 1 2 3 4 5 6 lost", testSequentialExecution(1, []int{0, 1, 2, 3, 4, 5, 6}, nil, true))

	//may result in fail of first and second sub-task -> []float64{7, 8, 9} becomes []float64{7}
	t.Run("2 retries 0 1 2 3 4 5 6 lost", testSequentialExecution(2, []int{0, 1, 2, 3, 4, 5, 6}, []float64{7}, false))
}

func testSequentialExecution(retries int64, lostResponseFor []int, expectedResultsBlueValues []float64, expectFailed bool) func(t *testing.T) {
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
		config.CamundaFetchLockDuration = 300
		config.CamundaWorkerTimeout = 100
		config.Debug = true
		config.GroupScheduler = util.SEQUENTIAL

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		mockCamunda := &mock.CamundaMock{}
		mockCamunda.Init()
		kafkamock := &mock.KafkaMock{}
		go lib.Worker(ctx, config, kafkamock, mock.Repo, mockCamunda, mock.Marshaller, mock.Timescale)

		time.Sleep(100 * time.Millisecond)

		lostResponseForIndex := map[int]bool{}
		for _, v := range lostResponseFor {
			lostResponseForIndex[v] = true
		}

		counter := 0
		kafkamock.Subscribe("protocol1", func(message string) error {
			msg := messages.ProtocolMsg{}
			err = json.Unmarshal([]byte(message), &msg)
			if err != nil {
				log.Fatal(err)
			}
			msg.Response.Output = map[string]string{
				"body": "{\"level\":\"#c832" + fmt.Sprintf("%02x", counter) + "\"}",
			}
			log.Println("DEBUG:", lostResponseForIndex, lostResponseFor)
			ignore := lostResponseForIndex[counter]
			counter = counter + 1
			if !ignore {
				log.Println("RESPOND WITH", counter-1, msg.Response.Output)
				resp, err := json.Marshal(msg)
				if err != nil {
					log.Fatal(err)
				}
				kafkamock.Produce(config.ResponseTopic, string(resp))
			} else {
				log.Println("IGNORE", counter-1, msg.Response.Output)
			}
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
					Id:         "service_3",
					Name:       "s3",
					LocalId:    "s3u",
					ProtocolId: "p1",
					//FunctionIds: []string{model.MEASURING_FUNCTION_PREFIX + "f1", model.MEASURING_FUNCTION_PREFIX + "f2"},
					//AspectIds:   []string{"a1", "a2"},
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

		cmd := messages.Command{
			Version:          2,
			Function:         model.Function{Id: model.MEASURING_FUNCTION_PREFIX + "f1", RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
			Aspect:           &model.AspectNode{Id: "a1"},
			CharacteristicId: example.Rgb,
			DeviceGroupId:    "dg1",
			Retries:          retries,
		}

		cmdMsg, err := json.Marshal(cmd)
		if err != nil {
			t.Error(err)
			return
		}

		mockCamunda.AddTask(messages.CamundaExternalTask{
			Id: "1",
			Variables: map[string]messages.CamundaVariable{
				util.CAMUNDA_VARIABLES_PAYLOAD: {
					Value: string(cmdMsg),
				},
			},
		})

		sleepDuration := time.Duration(int64(len(lostResponseFor))*config.CamundaFetchLockDuration)*time.Millisecond + time.Second
		log.Println("sleep for" + sleepDuration.String())
		time.Sleep(sleepDuration)

		fetched, completed, failed := mockCamunda.GetStatus()

		if len(failed) > 0 != expectFailed {
			log.Println("fetched:", fetched)
			log.Println("failed:", failed)
			log.Println("completed:", completed)
			t.Error(len(fetched), len(failed), len(completed))
			return
		}
		if expectFailed {
			return
		}

		if len(fetched) != 0 || len(completed) != 1 {
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
		for _, b := range expectedResultsBlueValues {
			expectedResult = append(expectedResult, map[string]interface{}{
				"b": b,
				"g": float64(50),
				"r": float64(200),
			})
		}

		if !reflect.DeepEqual(completed["1"], expectedResult) {
			t.Error(completed, expectedResult)
			rjson, _ := json.Marshal(completed["1"])
			ejson, _ := json.Marshal(expectedResult)
			t.Log("\n", string(rjson), "\n", string(ejson))
		}

	}
}
