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

package messages

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
)

func TestCommandGetAspects(t *testing.T) {
	t.Run("reads the deprecated single aspect as a list with one element", func(t *testing.T) {
		command := Command{Aspect: &model.AspectNode{Id: "air"}}
		if actual := command.GetAspects(); !reflect.DeepEqual(actual, []model.AspectNode{{Id: "air"}}) {
			t.Error(actual)
		}
	})

	t.Run("returns the aspect list of a command that carries one", func(t *testing.T) {
		command := Command{Aspects: []model.AspectNode{{Id: "air"}, {Id: "humidity"}}}
		if actual := command.GetAspects(); !reflect.DeepEqual(actual, []model.AspectNode{{Id: "air"}, {Id: "humidity"}}) {
			t.Error(actual)
		}
	})

	t.Run("adds a deprecated aspect the list does not name", func(t *testing.T) {
		command := Command{Aspect: &model.AspectNode{Id: "humidity"}, Aspects: []model.AspectNode{{Id: "air"}}}
		if actual := command.GetAspects(); !reflect.DeepEqual(actual, []model.AspectNode{{Id: "air"}, {Id: "humidity"}}) {
			t.Error(actual)
		}
	})

	t.Run("does not repeat a deprecated aspect the list already names", func(t *testing.T) {
		command := Command{Aspect: &model.AspectNode{Id: "air"}, Aspects: []model.AspectNode{{Id: "air"}}}
		if actual := command.GetAspects(); !reflect.DeepEqual(actual, []model.AspectNode{{Id: "air"}}) {
			t.Error(actual)
		}
	})

	t.Run("returns no aspect for a command without one", func(t *testing.T) {
		if actual := (Command{}).GetAspects(); len(actual) != 0 {
			t.Error(actual)
		}
	})
}

func TestCommandSetAspects(t *testing.T) {
	t.Run("folds the aspect of a payload that predates the list", func(t *testing.T) {
		command := Command{}
		err := json.Unmarshal([]byte(`{"aspect":{"id":"air"}}`), &command)
		if err != nil {
			t.Fatal(err)
		}
		command.SetAspects()
		if !reflect.DeepEqual(command.Aspects, []model.AspectNode{{Id: "air"}}) {
			t.Error(command.Aspects)
		}
	})

	t.Run("keeps the aspects of a payload that carries the list", func(t *testing.T) {
		command := Command{}
		err := json.Unmarshal([]byte(`{"aspects":[{"id":"air"},{"id":"humidity"}]}`), &command)
		if err != nil {
			t.Fatal(err)
		}
		command.SetAspects()
		if !reflect.DeepEqual(command.Aspects, []model.AspectNode{{Id: "air"}, {Id: "humidity"}}) {
			t.Error(command.Aspects)
		}
	})
}

func TestMetadataOutputAspectNodes(t *testing.T) {
	t.Run("sets the list and the deprecated node with the alphabetically first id", func(t *testing.T) {
		metadata := Metadata{}
		metadata.SetOutputAspectNodes([]model.AspectNode{{Id: "humidity"}, {Id: "air"}})
		if !reflect.DeepEqual(metadata.OutputAspectNodes, []model.AspectNode{{Id: "humidity"}, {Id: "air"}}) {
			t.Error(metadata.OutputAspectNodes)
		}
		if metadata.OutputAspectNode == nil || metadata.OutputAspectNode.Id != "air" {
			t.Error(metadata.OutputAspectNode)
		}
	})

	t.Run("reports no aspect without one", func(t *testing.T) {
		metadata := Metadata{}
		metadata.SetOutputAspectNodes(nil)
		if metadata.OutputAspectNode != nil {
			t.Error(metadata.OutputAspectNode)
		}
		if len(metadata.GetOutputAspectNodes()) != 0 {
			t.Error(metadata.GetOutputAspectNodes())
		}
	})

	t.Run("reads a response that only carries the deprecated node", func(t *testing.T) {
		msg := ProtocolMsg{}
		err := json.Unmarshal([]byte(`{"metadata":{"output_aspect_node":{"id":"air"}}}`), &msg)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(msg.Metadata.GetOutputAspectNodes(), []model.AspectNode{{Id: "air"}}) {
			t.Error(msg.Metadata.GetOutputAspectNodes())
		}
	})
}
