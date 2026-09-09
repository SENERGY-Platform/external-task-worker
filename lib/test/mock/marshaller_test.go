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

package mock

import (
	"context"
	"testing"

	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
)

// TestMarshallerLeavesTheServiceOfTheCallerAlone guards the reason this runs over http
// instead of calling the marshalling in the same process. The marshalling writes the value
// it marshals into the content variables of the service it is given, and the worker puts
// that same service into the message metadata — sharing it would make every message report
// the value of whichever task marshalled last.
func TestMarshallerLeavesTheServiceOfTheCallerAlone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	protocol := model.Protocol{
		Id:               "p1",
		Name:             "protocol1",
		Handler:          "protocol1",
		ProtocolSegments: []model.ProtocolSegment{{Id: "ms1", Name: "body"}},
	}
	service := model.Service{
		Id:         "service_1",
		Name:       "s1",
		LocalId:    "s1u",
		ProtocolId: "p1",
		Inputs: []model.Content{{
			Id: "metrics",
			ContentVariable: model.ContentVariable{
				Id:   "metrics",
				Name: "metrics",
				Type: model.Structure,
				SubContentVariables: []model.ContentVariable{{
					Id:               "level",
					Name:             "level",
					Type:             model.Integer,
					CharacteristicId: "example_hex",
				}},
			},
			Serialization:     "json",
			ProtocolSegmentId: "ms1",
		}},
	}

	message, err := Marshaller.New(ctx, "").MarshalV2(service, protocol, []marshaller.MarshallingV2RequestData{{
		Value:            "#ff00ff",
		CharacteristicId: "example_hex",
		Paths:            []string{"metrics.level"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	if message["body"] != `{"level":"#ff00ff"}` {
		t.Error("unexpected message:", message)
	}
	if value := service.Inputs[0].ContentVariable.SubContentVariables[0].Value; value != nil {
		t.Error("service of the caller carries the marshalled value:", value)
	}
}
