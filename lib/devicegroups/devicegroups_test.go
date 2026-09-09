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

package devicegroups

import (
	"errors"
	"reflect"
	"testing"

	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository"
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

const testFunctionId = model.MEASURING_FUNCTION_PREFIX + "f1"

var testAspectNodes = map[string]model.AspectNode{
	"air":      {Id: "air", ChildIds: []string{"inside_air"}, DescendentIds: []string{"inside_air"}},
	"humidity": {Id: "humidity"},
}

// TestGetGroupSubTasksAspects checks which services of a group member a command reaches,
// now that a command may name several aspects.
func TestGetGroupSubTasksAspects(t *testing.T) {
	outputWith := func(aspectIds []string, deprecatedAspectId string) []model.Content {
		return []model.Content{{
			ContentVariable: model.ContentVariable{
				Name:       "value",
				FunctionId: testFunctionId,
				AspectId:   deprecatedAspectId,
				AspectIds:  aspectIds,
			},
		}}
	}
	repo := &repoStub{
		group:  model.DeviceGroup{Id: "g1", DeviceIds: []string{"d1"}},
		device: model.Device{Id: "d1", DeviceTypeId: "dt1"},
		deviceType: model.DeviceType{Id: "dt1", Services: []model.Service{
			{Id: "s_air", Interaction: model.REQUEST, Outputs: outputWith([]string{"air"}, "")},
			{Id: "s_air_humidity", Interaction: model.REQUEST, Outputs: outputWith([]string{"air", "humidity"}, "")},
			{Id: "s_humidity", Interaction: model.REQUEST, Outputs: outputWith([]string{"humidity"}, "")},
			{Id: "s_inside_air", Interaction: model.REQUEST, Outputs: outputWith([]string{"inside_air"}, "")},
			{Id: "s_legacy_air", Interaction: model.REQUEST, Outputs: outputWith(nil, "air")},
		}},
	}
	handler := NewWithKeyValueStore(util.PARALLEL, nil, repo, nil, 0, 0, NewLocalDb(), false)

	serviceIdsOf := func(t *testing.T, command messages.Command) (result []string) {
		t.Helper()
		subTasks, err := handler.GetGroupSubTasks(command, messages.CamundaExternalTask{Id: "task"})
		if err != nil {
			t.Fatal(err)
		}
		for _, subTask := range subTasks {
			result = append(result, subTask.Command.ServiceId)
		}
		return result
	}

	command := func(aspects []model.AspectNode, deprecatedAspect *model.AspectNode) messages.Command {
		return messages.Command{
			DeviceGroupId: "g1",
			Function:      model.Function{Id: testFunctionId},
			Aspect:        deprecatedAspect,
			Aspects:       aspects,
		}
	}

	t.Run("reaches every service serving the aspect or one of its descendants", func(t *testing.T) {
		expected := []string{"s_air", "s_air_humidity", "s_inside_air", "s_legacy_air"}
		if actual := serviceIdsOf(t, command([]model.AspectNode{testAspectNodes["air"]}, nil)); !reflect.DeepEqual(actual, expected) {
			t.Error(actual)
		}
	})

	t.Run("reaches only services serving every named aspect", func(t *testing.T) {
		expected := []string{"s_air_humidity"}
		aspects := []model.AspectNode{testAspectNodes["air"], testAspectNodes["humidity"]}
		if actual := serviceIdsOf(t, command(aspects, nil)); !reflect.DeepEqual(actual, expected) {
			t.Error(actual)
		}
	})

	t.Run("reads the deprecated single aspect of a command like a list with one element", func(t *testing.T) {
		expected := []string{"s_air", "s_air_humidity", "s_inside_air", "s_legacy_air"}
		aspect := testAspectNodes["air"]
		if actual := serviceIdsOf(t, command(nil, &aspect)); !reflect.DeepEqual(actual, expected) {
			t.Error(actual)
		}
	})

	t.Run("reaches every service of the function without a named aspect", func(t *testing.T) {
		expected := []string{"s_air", "s_air_humidity", "s_humidity", "s_inside_air", "s_legacy_air"}
		if actual := serviceIdsOf(t, command(nil, nil)); !reflect.DeepEqual(actual, expected) {
			t.Error(actual)
		}
	})
}

type repoStub struct {
	group      model.DeviceGroup
	device     model.Device
	deviceType model.DeviceType
}

func (this *repoStub) GetDevice(token devicerepository.Impersonate, id string) (model.Device, error) {
	if id != this.device.Id {
		return model.Device{}, errors.New("device not found")
	}
	return this.device, nil
}

func (this *repoStub) GetService(token devicerepository.Impersonate, device model.Device, serviceId string) (model.Service, error) {
	for _, service := range this.deviceType.Services {
		if service.Id == serviceId {
			return service, nil
		}
	}
	return model.Service{}, errors.New("service not found")
}

func (this *repoStub) GetProtocol(token devicerepository.Impersonate, id string) (model.Protocol, error) {
	return model.Protocol{}, errors.New("protocol not found")
}

func (this *repoStub) GetToken(user string) (devicerepository.Impersonate, error) {
	return "", nil
}

func (this *repoStub) GetDeviceType(token devicerepository.Impersonate, id string) (model.DeviceType, error) {
	if id != this.deviceType.Id {
		return model.DeviceType{}, errors.New("device-type not found")
	}
	return this.deviceType, nil
}

func (this *repoStub) GetDeviceGroup(token devicerepository.Impersonate, id string) (model.DeviceGroup, error) {
	if id != this.group.Id {
		return model.DeviceGroup{}, errors.New("device-group not found")
	}
	return this.group, nil
}
