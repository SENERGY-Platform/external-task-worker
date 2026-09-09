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

package lib

import (
	"reflect"
	"testing"

	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

func TestGetCommandRequestAspects(t *testing.T) {
	taskWithPayload := func(payload string) messages.CamundaExternalTask {
		return messages.CamundaExternalTask{
			Id:        "task",
			Variables: map[string]messages.CamundaVariable{util.CAMUNDA_VARIABLES_PAYLOAD: {Value: payload}},
		}
	}

	t.Run("folds the deprecated aspect of a pulled task into the aspect list", func(t *testing.T) {
		command, err := GetCommandRequest(taskWithPayload(`{"version":3,"aspect":{"id":"air"}}`))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(command.Aspects, []model.AspectNode{{Id: "air"}}) {
			t.Error(command.Aspects)
		}
	})

	t.Run("keeps the aspect list of a pulled task that carries one", func(t *testing.T) {
		command, err := GetCommandRequest(taskWithPayload(`{"version":3,"aspects":[{"id":"air"},{"id":"humidity"}]}`))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(command.Aspects, []model.AspectNode{{Id: "air"}, {Id: "humidity"}}) {
			t.Error(command.Aspects)
		}
	})

	t.Run("leaves the aspect list empty for a pulled task without an aspect", func(t *testing.T) {
		command, err := GetCommandRequest(taskWithPayload(`{"version":3}`))
		if err != nil {
			t.Fatal(err)
		}
		if len(command.Aspects) != 0 {
			t.Error(command.Aspects)
		}
	})
}
