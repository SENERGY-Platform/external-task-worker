/*
 * Copyright 2026 InfAI (CC SES)
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
	"sort"
	"testing"
	"time"

	"github.com/SENERGY-Platform/external-task-worker/lib"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

//A command may name more than one aspect, and then every named aspect has to be served by
//one content variable. Two aspects only ever match together if they come out of different
//hierarchies, because a content variable carries at most one aspect per hierarchy. The
//nodes below are therefore two trees: an air hierarchy with a room below it, and a light
//aspect of its own. Only a node that is named by a command needs to be spelled out here,
//because the subtree is read off the ChildIds and DescendentIds of the named node.

var aspectInsideAir = model.AspectNode{
	Id:            "inside_air",
	Name:          "inside air",
	RootId:        "air",
	ParentId:      "air",
	AncestorIds:   []string{"air"},
	ChildIds:      []string{"kitchen_air"},
	DescendentIds: []string{"kitchen_air"},
}

var aspectLight = model.AspectNode{
	Id:     "light",
	Name:   "light",
	RootId: "light",
}

const multiAspectFunction = model.CONTROLLING_FUNCTION_PREFIX + "f_set_color"
const multiAspectMeasuringFunction = model.MEASURING_FUNCTION_PREFIX + "f_get_color"

// TestMultiAspectDeviceCommandMarshalsOnlyTheVariableServingEveryAspect sends a command
// without an input path, so the marshaller has to find the path from the function and the
// aspects. Of the three candidates only one serves both named aspects.
func TestMultiAspectDeviceCommandMarshalsOnlyTheVariableServingEveryAspect(t *testing.T) {
	mockKafka := mock.NewKafka()
	mockRepo := mock.NewRepo()
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.CompletionStrategy = util.OPTIMISTIC
	config.HealthCheckPort = ""
	config.HttpCommandConsumerPort, err = GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mockKafka, mockRepo, mockCamunda, mock.Marshaller, mock.Timescale)

	time.Sleep(1 * time.Second)

	mockRepo.RegisterDevice(model.Device{Id: "device_1", Name: "d1", DeviceTypeId: "dt1", LocalId: "d1u"})
	mockRepo.RegisterProtocol(model.Protocol{
		Id:               "p1",
		Name:             "protocol1",
		Handler:          "protocol1",
		ProtocolSegments: []model.ProtocolSegment{{Id: "ms1", Name: "body"}},
	})
	mockRepo.RegisterService(model.Service{
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
							//serves both named aspects
							Id:               "inside_light",
							Name:             "inside_light",
							Type:             model.String,
							CharacteristicId: example.Hex,
							FunctionId:       multiAspectFunction,
							AspectIds:        []string{"inside_air", "light"},
						},
						{
							//light, but a sibling of the named air aspect
							Id:               "outside_light",
							Name:             "outside_light",
							Type:             model.String,
							CharacteristicId: example.Hex,
							FunctionId:       multiAspectFunction,
							AspectIds:        []string{"outside_air", "light"},
						},
						{
							//the named air aspect, but no light
							Id:               "inside_heating",
							Name:             "inside_heating",
							Type:             model.String,
							CharacteristicId: example.Hex,
							FunctionId:       multiAspectFunction,
							AspectIds:        []string{"inside_air"},
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	})

	cmd, err := json.Marshal(messages.Command{
		Version:          3,
		Function:         model.Function{Id: multiAspectFunction},
		Aspects:          []model.AspectNode{aspectInsideAir, aspectLight},
		CharacteristicId: example.Rgb,
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
		Input:            map[string]float64{"r": 200, "g": 50, "b": 0},
	})
	if err != nil {
		t.Error(err)
		return
	}

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:        "1",
		TenantId:  "user",
		Variables: map[string]messages.CamundaVariable{util.CAMUNDA_VARIABLES_PAYLOAD: {Value: string(cmd)}},
	})

	time.Sleep(2 * time.Second)

	body := singleProtocolMessageBody(t, mockKafka.GetProduced("protocol1"))
	if body == nil {
		return
	}

	if body["inside_light"] != "#c83200" {
		t.Error("the variable serving both aspects did not get the value:", body)
	}
	if body["outside_light"] == "#c83200" || body["inside_heating"] == "#c83200" {
		t.Error("a variable that serves only one of the named aspects got the value:", body)
	}
}

// TestMultiAspectDeviceResponseUnmarshalsTheVariableServingEveryAspect answers a command
// without an output path with a message that carries both candidate variables. Only one of
// them serves both named aspects, and its value is the one that has to reach camunda.
func TestMultiAspectDeviceResponseUnmarshalsTheVariableServingEveryAspect(t *testing.T) {
	mockKafka := mock.NewKafka()
	mockRepo := mock.NewRepo()
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.CompletionStrategy = util.PESSIMISTIC
	config.HealthCheckPort = ""
	config.HttpCommandConsumerPort, err = GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mockKafka, mockRepo, mockCamunda, mock.Marshaller, mock.Timescale)

	time.Sleep(1 * time.Second)

	mockRepo.RegisterDevice(model.Device{Id: "device_1", Name: "d1", DeviceTypeId: "dt1", LocalId: "d1u"})
	mockRepo.RegisterProtocol(model.Protocol{
		Id:               "p1",
		Name:             "protocol1",
		Handler:          "protocol1",
		ProtocolSegments: []model.ProtocolSegment{{Id: "ms1", Name: "body"}},
	})
	mockRepo.RegisterService(model.Service{
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
							Id:               "inside_light",
							Name:             "inside_light",
							Type:             model.String,
							CharacteristicId: example.Hex,
							FunctionId:       multiAspectMeasuringFunction,
							AspectIds:        []string{"inside_air", "light"},
						},
						{
							Id:               "outside_light",
							Name:             "outside_light",
							Type:             model.String,
							CharacteristicId: example.Hex,
							FunctionId:       multiAspectMeasuringFunction,
							AspectIds:        []string{"outside_air", "light"},
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	})

	cmd, err := json.Marshal(messages.Command{
		Version:          3,
		Function:         model.Function{Id: multiAspectMeasuringFunction},
		Aspects:          []model.AspectNode{aspectInsideAir, aspectLight},
		CharacteristicId: example.Rgb,
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
	})
	if err != nil {
		t.Error(err)
		return
	}

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:        "1",
		TenantId:  "user",
		Variables: map[string]messages.CamundaVariable{util.CAMUNDA_VARIABLES_PAYLOAD: {Value: string(cmd)}},
	})

	time.Sleep(2 * time.Second)

	produced := mockKafka.GetProduced("protocol1")
	if len(produced) != 1 {
		t.Error("expected exactly one protocol message:", produced)
		return
	}
	msg := messages.ProtocolMsg{}
	if err = json.Unmarshal([]byte(produced[0]), &msg); err != nil {
		t.Error(err)
		return
	}
	//the value of the wrong variable is a different colour, so picking it fails loudly
	msg.Response.Output = map[string]string{
		"body": `{"inside_light":"#c83200","outside_light":"#0000ff"}`,
	}
	response, err := json.Marshal(msg)
	if err != nil {
		t.Error(err)
		return
	}
	if err = mockKafka.Produce(config.ResponseTopic, string(response)); err != nil {
		t.Error(err)
		return
	}

	time.Sleep(2 * time.Second)

	_, completed, _ := mockCamunda.GetStatus()
	actual, err := json.Marshal(completed["1"])
	if err != nil {
		t.Error(err)
		return
	}
	//a task completes with the list of its sub results, here the one device that answered
	if string(actual) != `[{"b":0,"g":50,"r":200}]` {
		t.Error("camunda got the value of a variable that does not serve every named aspect:", string(actual))
	}
}

// TestMultiAspectGroupCommandReachesOnlyDevicesServingEveryAspect sends one command to a
// group of four devices. Two of them serve both named aspects, one directly and one through
// a descendant of the named aspect; the other two miss one of the two.
func TestMultiAspectGroupCommandReachesOnlyDevicesServingEveryAspect(t *testing.T) {
	mockKafka := mock.NewKafka()
	mockRepo := mock.NewRepo()
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.CompletionStrategy = util.OPTIMISTIC
	config.GroupScheduler = util.PARALLEL
	config.HealthCheckPort = ""
	config.HttpCommandConsumerPort, err = GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	mockCamunda := &mock.CamundaMock{}
	mockCamunda.Init()
	go lib.Worker(ctx, config, mockKafka, mockRepo, mockCamunda, mock.Marshaller, mock.Timescale)

	time.Sleep(1 * time.Second)

	mockRepo.RegisterProtocol(model.Protocol{
		Id:               "p1",
		Name:             "protocol1",
		Handler:          "protocol1",
		ProtocolSegments: []model.ProtocolSegment{{Id: "ms1", Name: "body"}},
	})

	//serves inside_air and light directly
	registerMultiAspectDevice(mockRepo, "device_exact", []string{"inside_air", "light"})
	//serves light and a descendant of the named inside_air, so the subtree covers it
	registerMultiAspectDevice(mockRepo, "device_kitchen", []string{"kitchen_air", "light"})
	//serves the named air aspect, but no light
	registerMultiAspectDevice(mockRepo, "device_no_light", []string{"inside_air"})
	//serves light, but a sibling of the named air aspect
	registerMultiAspectDevice(mockRepo, "device_outside", []string{"outside_air", "light"})

	mockRepo.RegisterDeviceGroup(model.DeviceGroup{
		Id:   "dg1",
		Name: "dg1",
		Criteria: []model.DeviceGroupFilterCriteria{
			{FunctionId: multiAspectFunction, AspectIds: []string{"inside_air", "light"}, Interaction: model.REQUEST},
		},
		DeviceIds: []string{"device_exact", "device_kitchen", "device_no_light", "device_outside"},
	})

	cmd, err := json.Marshal(messages.Command{
		Version:          3,
		Function:         model.Function{Id: multiAspectFunction},
		Aspects:          []model.AspectNode{aspectInsideAir, aspectLight},
		CharacteristicId: example.Rgb,
		DeviceGroupId:    "dg1",
		Input:            map[string]float64{"r": 200, "g": 50, "b": 0},
	})
	if err != nil {
		t.Error(err)
		return
	}

	mockCamunda.AddTask(messages.CamundaExternalTask{
		Id:        "1",
		TenantId:  "user",
		Variables: map[string]messages.CamundaVariable{util.CAMUNDA_VARIABLES_PAYLOAD: {Value: string(cmd)}},
	})

	time.Sleep(3 * time.Second)

	actual := []string{}
	for _, message := range mockKafka.GetProduced("protocol1") {
		msg := messages.ProtocolMsg{}
		if err = json.Unmarshal([]byte(message), &msg); err != nil {
			t.Error(err)
			return
		}
		actual = append(actual, msg.Metadata.Device.Id)
	}
	sort.Strings(actual)

	expected := []string{"device_exact", "device_kitchen"}
	if len(actual) != len(expected) {
		t.Error("expected", expected, "got", actual)
		return
	}
	for i := range expected {
		if actual[i] != expected[i] {
			t.Error("expected", expected, "got", actual)
			return
		}
	}
}

// registerMultiAspectDevice registers a device whose one service carries a single input
// variable serving multiAspectFunction and the given aspects.
func registerMultiAspectDevice(repo *mock.RepoMock, deviceId string, aspectIds []string) {
	repo.RegisterDevice(model.Device{
		Id:           deviceId,
		Name:         deviceId,
		DeviceTypeId: "dt_" + deviceId,
		LocalId:      deviceId + "_u",
	})
	repo.RegisterDeviceType(model.DeviceType{
		Id:   "dt_" + deviceId,
		Name: "dt_" + deviceId,
		Services: []model.Service{
			{
				Id:          "service_" + deviceId,
				Name:        "s_" + deviceId,
				LocalId:     "s_" + deviceId + "_u",
				ProtocolId:  "p1",
				Interaction: model.REQUEST,
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
									Type:             model.String,
									CharacteristicId: example.Hex,
									FunctionId:       multiAspectFunction,
									AspectIds:        aspectIds,
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
}

// singleProtocolMessageBody expects exactly one produced protocol message and returns its
// marshalled body, so that a test can assert on the variable the value landed on.
func singleProtocolMessageBody(t *testing.T, produced []string) map[string]string {
	t.Helper()
	if len(produced) != 1 {
		t.Error("expected exactly one protocol message:", produced)
		return nil
	}
	msg := messages.ProtocolMsg{}
	if err := json.Unmarshal([]byte(produced[0]), &msg); err != nil {
		t.Error(err)
		return nil
	}
	body := map[string]string{}
	if err := json.Unmarshal([]byte(msg.Request.Input["body"]), &body); err != nil {
		t.Error("unable to read the marshalled body:", msg.Request.Input, err)
		return nil
	}
	return body
}
