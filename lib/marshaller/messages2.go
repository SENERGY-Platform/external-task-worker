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

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/marshaller/lib/api/messages"
	"github.com/SENERGY-Platform/marshaller/lib/configurables"
	marshallermodel "github.com/SENERGY-Platform/marshaller/lib/marshaller/model"
)

//the request types are the ones of the marshaller itself, so that a field added there does
//not have to be repeated here to be usable

type MarshallingV2RequestData = marshallermodel.MarshallingV2RequestData

type UnmarshallingV2Request = messages.UnmarshallingV2Request

type Configurable = configurables.Configurable

type ConfigurableCharacteristicValue = configurables.ConfigurableCharacteristicValue

// ConfigurableV2 is the configurable as the process-deployment writes it into the task
// payload; it mirrors the configurable of a device-repository path option.
type ConfigurableV2 struct {
	Path             string             `json:"path"`
	CharacteristicId string             `json:"characteristic_id"`
	AspectNode       model.AspectNode   `json:"aspect_node"` //deprecated: please use AspectNodes
	AspectNodes      []model.AspectNode `json:"aspect_nodes,omitempty"`
	FunctionId       string             `json:"function_id"`
	Value            interface{}        `json:"value,omitempty"`
	Type             string             `json:"type,omitempty"`
}

// GetAspectNodes returns the aspect nodes of the configurable. The deprecated AspectNode is
// an alias for a list with one element.
func (this ConfigurableV2) GetAspectNodes() []model.AspectNode {
	return marshallermodel.AspectNodesAlias(this.AspectNode, this.AspectNodes)
}
