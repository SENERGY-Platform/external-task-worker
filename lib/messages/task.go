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

package messages

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller"
)

type Command struct {
	Version int64 `json:"version"`

	//modeling time
	Function         model.Function `json:"function"`
	CharacteristicId string         `json:"characteristic_id"`

	//optional modeling time (used to limit/filter device and service selection in deployment)
	DeviceClass *model.DeviceClass `json:"device_class,omitempty"`
	Aspect      *model.AspectNode  `json:"aspect,omitempty"`

	//deployment time
	DeviceGroupId string          `json:"device_group_id"`
	DeviceId      string          `json:"device_id,omitempty"`
	Device        *model.Device   `json:"device,omitempty"`
	ServiceId     string          `json:"service_id,omitempty"`
	Service       *model.Service  `json:"service,omitempty"`
	Protocol      *model.Protocol `json:"protocol,omitempty"`
	ProtocolId    string          `json:"protocol_id,omitempty"`

	//version <= 2
	Configurables        []marshaller.Configurable `json:"configurables,omitempty"`
	ContentVariableHints []string                  `json:"content_variable_hints,omitempty"`

	//version >= 3
	InputPaths      []string                    `json:"input_paths,omitempty"`
	OutputPath      string                      `json:"output_path,omitempty"`
	ConfigurablesV2 []marshaller.ConfigurableV2 `json:"configurables_v2,omitempty"`

	PreferEvent bool `json:"prefer_event"`

	//runtime
	Input  interface{} `json:"input,omitempty"`
	Output interface{} `json:"output,omitempty"`

	Retries int64 `json:"retries,omitempty"`
}

type Overwrite struct {
	DeviceId   string          `json:"device_id,omitempty"`
	Device     *model.Device   `json:"device,omitempty"`
	ServiceId  string          `json:"service_id,omitempty"`
	Service    *model.Service  `json:"service,omitempty"`
	Protocol   *model.Protocol `json:"protocol,omitempty"`
	ProtocolId string          `json:"protocol_id,omitempty"`
}
