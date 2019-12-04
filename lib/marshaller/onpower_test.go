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

package marshaller

import (
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
)

func ExampleOnPower() {
	protocol := model.Protocol{
		Id:      "p1",
		Name:    "p1",
		Handler: "p1",
		ProtocolSegments: []model.ProtocolSegment{
			{Id: "p1.1", Name: "body"},
			{Id: "p1.2", Name: "head"},
		},
	}
	service := model.Service{
		Id:          "on",
		LocalId:     "power",
		Name:        "on",
		Description: "on",
		ProtocolId:  "p1",
		Inputs: []model.Content{
			{
				Id: "input_content_1",
				ContentVariable: model.ContentVariable{
					Id:    "input_var_1",
					Name:  "state",
					Type:  model.Boolean,
					Value: true,
				},
				Serialization:     "json",
				ProtocolSegmentId: "p1.1",
			},
		},
	}
	result, err := MarshalInputs(protocol, service, nil, "")
	fmt.Println(result, err)

	//output:
	//map[body:true] <nil>
}
