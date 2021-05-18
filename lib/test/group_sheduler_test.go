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
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestGroupScheduler(t *testing.T) {
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.PESSIMISTIC
	config.CamundaWorkerTimeout = 300      //in ms
	config.CamundaFetchLockDuration = 1000 //in ms
	config.HealthCheckPort = ""

	config.GroupScheduler = util.SEQUENTIAL
	t.Run("simple sequential", getGroupShedullerTest(config, GroupSimConfig{
		Retries:       1,
		CheckAfter:    1 * time.Minute,
		ResponseTimes: [][]time.Duration{{100 * time.Millisecond}, {-1}, {-1, 100 * time.Millisecond}, {100 * time.Millisecond}, {100 * time.Millisecond}},
		Responses:     []string{"#c83200", "#c83201", "#c83202", "#c83203", "#c83204"},
		ExpectedResult: []interface{}{
			map[string]interface{}{"b": float64(0), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(2), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(3), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(4), "g": float64(50), "r": float64(200)},
		},
	}))

	config.GroupScheduler = util.PARALLEL
	t.Run("simple parallel", getGroupShedullerTest(config, GroupSimConfig{
		Retries:       1,
		CheckAfter:    1 * time.Minute,
		ResponseTimes: [][]time.Duration{{100 * time.Millisecond}, {-1}, {-1, 100 * time.Millisecond}, {100 * time.Millisecond}, {100 * time.Millisecond}},
		Responses:     []string{"#c83200", "#c83201", "#c83202", "#c83203", "#c83204"},
		ExpectedResult: []interface{}{
			map[string]interface{}{"b": float64(0), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(2), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(3), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(4), "g": float64(50), "r": float64(200)},
		},
	}))

	config.GroupScheduler = util.SEQUENTIAL
	t.Run("slow sequential", getGroupShedullerTest(config, GroupSimConfig{
		Retries:       1,
		CheckAfter:    1 * time.Minute,
		ResponseTimes: [][]time.Duration{{1000 * time.Millisecond}, {-1}, {-1, 1000 * time.Millisecond}, {1000 * time.Millisecond}, {1000 * time.Millisecond}},
		Responses:     []string{"#c83200", "#c83201", "#c83202", "#c83203", "#c83204"},
		ExpectedResult: []interface{}{
			map[string]interface{}{"b": float64(0), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(2), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(3), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(4), "g": float64(50), "r": float64(200)},
		},
	}))

	config.GroupScheduler = util.PARALLEL
	t.Run("slow parallel", getGroupShedullerTest(config, GroupSimConfig{
		Retries:       1,
		CheckAfter:    1 * time.Minute,
		ResponseTimes: [][]time.Duration{{1000 * time.Millisecond}, {-1}, {-1, 1000 * time.Millisecond}, {1000 * time.Millisecond}, {1000 * time.Millisecond}},
		Responses:     []string{"#c83200", "#c83201", "#c83202", "#c83203", "#c83204"},
		ExpectedResult: []interface{}{
			map[string]interface{}{"b": float64(0), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(2), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(3), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(4), "g": float64(50), "r": float64(200)},
		},
	}))

	config.GroupScheduler = util.SEQUENTIAL
	t.Run("very slow sequential flip", getGroupShedullerTest(config, GroupSimConfig{
		Retries:       1,
		CheckAfter:    1 * time.Minute,
		ResponseTimes: [][]time.Duration{{1500 * time.Millisecond}, {-1}, {1500 * time.Millisecond, -1}, {1500 * time.Millisecond}, {1500 * time.Millisecond}},
		Responses:     []string{"#c83200", "#c83201", "#c83202", "#c83203", "#c83204"},
		ExpectedResult: []interface{}{
			map[string]interface{}{"b": float64(0), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(2), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(3), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(4), "g": float64(50), "r": float64(200)},
		},
	}))

	config.GroupScheduler = util.SEQUENTIAL
	t.Run("very slow sequential", getGroupShedullerTest(config, GroupSimConfig{
		Retries:       1,
		CheckAfter:    1 * time.Minute,
		ResponseTimes: [][]time.Duration{{1500 * time.Millisecond}, {-1}, {-1, 1500 * time.Millisecond}, {1500 * time.Millisecond}, {1500 * time.Millisecond}},
		Responses:     []string{"#c83200", "#c83201", "#c83202", "#c83203", "#c83204"},
		ExpectedResult: []interface{}{
			map[string]interface{}{"b": float64(0), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(2), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(3), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(4), "g": float64(50), "r": float64(200)},
		},
	}))

	config.GroupScheduler = util.PARALLEL
	t.Run("very slow parallel", getGroupShedullerTest(config, GroupSimConfig{
		Retries:       1,
		CheckAfter:    1 * time.Minute,
		ResponseTimes: [][]time.Duration{{1500 * time.Millisecond}, {-1}, {1500 * time.Millisecond, -1}, {1500 * time.Millisecond}, {1500 * time.Millisecond}},
		Responses:     []string{"#c83200", "#c83201", "#c83202", "#c83203", "#c83204"},
		ExpectedResult: []interface{}{
			map[string]interface{}{"b": float64(0), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(2), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(3), "g": float64(50), "r": float64(200)},
			map[string]interface{}{"b": float64(4), "g": float64(50), "r": float64(200)},
		},
	}))
}

type DeviceSimInterface interface {
	HandleRequest(resp func(value string))
}

type DeviceSim struct {
	mux           sync.Mutex
	currentCall   int
	Value         string
	ResponseTimes []time.Duration
}

func (this *DeviceSim) HandleRequest(resp func(value string)) {
	this.mux.Lock()
	defer this.mux.Unlock()
	i := this.currentCall % len(this.ResponseTimes)
	this.currentCall = this.currentCall + 1
	timeout := this.ResponseTimes[i]
	if timeout >= 0 {
		go func() {
			if timeout > 0 {
				time.Sleep(timeout)
			}
			log.Println("TEST: respond", this.Value, timeout)
			resp(this.Value)
		}()
	}
}

type GroupSim struct {
	Devices map[string]DeviceSimInterface
}

type GroupSimConfig struct {
	Retries        int64
	ResponseTimes  [][]time.Duration
	Responses      []string
	ExpectedResult []interface{}
	CheckAfter     time.Duration
}

func (this *GroupSim) Handle(msg messages.ProtocolMsg, resp func(value string)) {
	this.Devices[msg.Metadata.Device.Id].HandleRequest(resp)
}

func (this *GroupSim) GetDevices() (devices []model.Device) {
	for id, _ := range this.Devices {
		devices = append(devices, model.Device{
			Id:           id,
			Name:         id,
			DeviceTypeId: "dt1",
			LocalId:      id,
		})
	}
	return devices
}

func getGroupShedullerTest(config util.Config, simConfig GroupSimConfig) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		mockRepo := mock.Repo.New()
		mockCamunda := &mock.CamundaMock{}
		mockCamunda.Init()
		mockKafka := mock.Kafka.New()

		group := NewGroupSim(config, simConfig, mockKafka)

		//populate repository
		deviceIds := []string{}
		for _, device := range group.GetDevices() {
			mockRepo.RegisterDevice(device)
			deviceIds = append(deviceIds, device.Id)
		}

		mockRepo.RegisterDeviceGroup(model.DeviceGroup{
			Id:   "dg1",
			Name: "dg1",
			Criteria: []model.DeviceGroupFilterCriteria{
				{FunctionId: model.MEASURING_FUNCTION_PREFIX + "f1", AspectId: "a1", Interaction: model.REQUEST},
			},
			DeviceIds: deviceIds,
		})

		mockRepo.RegisterProtocol(model.Protocol{
			Id:               "p1",
			Name:             "protocol1",
			Handler:          "protocol1",
			ProtocolSegments: []model.ProtocolSegment{{Id: "ms1", Name: "body"}},
		})

		mockRepo.RegisterDeviceType(model.DeviceType{
			Id:            "dt1",
			Name:          "dt1",
			DeviceClassId: "dc1",
			Services: []model.Service{
				{
					Id:          "service_3",
					Name:        "s3",
					LocalId:     "s3u",
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

		cmd := messages.Command{
			Version:          2,
			Function:         model.Function{Id: model.MEASURING_FUNCTION_PREFIX + "f1", RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
			Aspect:           &model.Aspect{Id: "a1"},
			CharacteristicId: example.Rgb,
			DeviceGroupId:    "dg1",
			Retries:          simConfig.Retries,
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

		go lib.Worker(ctx, config, mockKafka, mockRepo, mockCamunda, mock.Marshaller)

		time.Sleep(simConfig.CheckAfter)

		fetched, completed, failed := mockCamunda.GetStatus()

		if len(fetched) != 0 || len(failed) != 0 || len(completed) != 1 {
			log.Println("fetched:", fetched)
			log.Println("failed:", failed)
			log.Println("completed:", completed)
			t.Error(len(fetched), len(failed), len(completed))
			return
		}

		expectedResult := simConfig.ExpectedResult
		actualResult := completed["1"].([]interface{})

		sort.Slice(expectedResult, func(i, j int) bool {
			a, ok := expectedResult[i].(map[string]interface{})
			if !ok {
				t.Error("unexpected type in expected result")
				return true
			}
			b, ok := expectedResult[j].(map[string]interface{})
			if !ok {
				t.Error("unexpected type in expected result")
				return true
			}
			return a["b"].(float64) < b["b"].(float64)
		})

		sort.Slice(actualResult, func(i, j int) bool {
			a, ok := actualResult[i].(map[string]interface{})
			if !ok {
				t.Error("unexpected type in expected result")
				return true
			}
			b, ok := actualResult[j].(map[string]interface{})
			if !ok {
				t.Error("unexpected type in expected result")
				return true
			}
			return a["b"].(float64) < b["b"].(float64)
		})

		if !reflect.DeepEqual(actualResult, expectedResult) {
			t.Error(actualResult, expectedResult)
		}
	}
}

func NewGroupSim(config util.Config, simConfig GroupSimConfig, kafka *mock.KafkaMock) (group *GroupSim) {
	group = &GroupSim{Devices: map[string]DeviceSimInterface{}}

	for i, times := range simConfig.ResponseTimes {
		group.Devices[strconv.Itoa(i)] = &DeviceSim{
			Value:         simConfig.Responses[i],
			ResponseTimes: times,
		}
	}

	kafka.Subscribe("protocol1", func(message string) error {
		msg := messages.ProtocolMsg{}
		err := json.Unmarshal([]byte(message), &msg)
		if err != nil {
			log.Println("ERROR:", err)
			return err
		}
		group.Handle(msg, func(value string) {
			msg.Response.Output = map[string]string{
				"body": "{\"level\":\"" + value + "\"}",
			}
			log.Println("RESPOND WITH", value, msg.Response.Output)
			resp, err := json.Marshal(msg)
			if err != nil {
				log.Println("ERROR:", err)
				return
			}
			err = kafka.Produce(config.ResponseTopic, string(resp))
			if err != nil {
				log.Println("ERROR:", err)
				return
			}
		})
		return nil
	})

	return group
}
