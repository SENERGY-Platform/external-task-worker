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

package marshaller

import "github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"

type MarshallingV2Request struct {
	Service  model.Service              `json:"service"`  //semi-optional, may be determined by request path
	Protocol model.Protocol             `json:"protocol"` //semi-optional, may be determined by service
	Data     []MarshallingV2RequestData `json:"data"`
}

type MarshallingV2RequestData struct {
	Value            interface{}       `json:"value"`
	CharacteristicId string            `json:"characteristic_id"`
	Paths            []string          `json:"paths"`                 //semi-optional, may be determent by FunctionId
	FunctionId       string            `json:"function_id"`           //semi-optional, to determine Paths if they are not set
	AspectNode       *model.AspectNode `json:"aspect_node,omitempty"` //optional, to determine Paths if they are not set, may be empty if only FunctionId should be searched
}

type UnmarshallingV2Request struct {
	Service          model.Service     `json:"service"`           //semi-optional, may be determined by request path
	Protocol         model.Protocol    `json:"protocol"`          //semi-optional, may be determined by service
	CharacteristicId string            `json:"characteristic_id"` //semi-optional, may be determined by request path
	Message          map[string]string `json:"message"`

	Path         string           `json:"path"`           //semi-optional, may be determent by FunctionId and AspectNode
	FunctionId   string           `json:"function_id"`    //semi-optional, to determine Path if not set
	AspectNode   model.AspectNode `json:"aspect_node"`    //semi-optional, to determine Path if not set, may itself be determent by AspectNodeId
	AspectNodeId string           `json:"aspect_node_id"` //semi-optional, to determine AspectNode if not set
}

type ConfigurableV2 struct {
	Path             string           `json:"path"`
	CharacteristicId string           `json:"characteristic_id"`
	AspectNode       model.AspectNode `json:"aspect_node"`
	FunctionId       string           `json:"function_id"`
	Value            interface{}      `json:"value,omitempty"`
	Type             string           `json:"type,omitempty"`
}
