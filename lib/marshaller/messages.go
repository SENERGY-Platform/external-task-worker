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

type MarshallingRequest struct {
	Service          model.Service   `json:"service,omitempty"`           //semi-optional, may be determined by request path
	Protocol         *model.Protocol `json:"protocol,omitempty"`          //semi-optional, may be determined by request path
	CharacteristicId string          `json:"characteristic_id,omitempty"` //semi-optional, may be determined by request path
	Configurables    []Configurable  `json:"configurables,omitempty"`     //optional, may be empty
	Data             interface{}     `json:"data"`
}

type UnmarshallingRequest struct {
	Service              model.Service     `json:"service,omitempty"`           //semi-optional, may be determined by request path
	Protocol             *model.Protocol   `json:"protocol,omitempty"`          //semi-optional, may be determined by service
	CharacteristicId     string            `json:"characteristic_id,omitempty"` //semi-optional, may be determined by request path
	Message              map[string]string `json:"message"`
	ContentVariableHints []string          `json:"content_variable_hints"` //optional
}

type FindConfigurablesRequest struct {
	CharacteristicId string          `json:"characteristic_id"`
	Services         []model.Service `json:"services"`
}

type Configurable struct {
	CharacteristicId string                            `json:"characteristic_id"`
	Values           []ConfigurableCharacteristicValue `json:"values"`
}

type ConfigurableCharacteristicValue struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}
