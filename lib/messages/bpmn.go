/*
 * Copyright 2018 InfAI (CC SES)
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

import "github.com/SENERGY-Platform/iot-device-repository/lib/model"

type BpmnAbstractMsg struct {
	TaskId     string                  `json:"-"`
	Parameter  []AbstractTaskParameter `json:"-"`
	ReuseId    string                  `json:"-"`
	DeviceType string                  `json:"device_type"`
	Service    string                  `json:"service"`
	Label      string                  `json:"label"`
	Values     BpmnValueSkeleton       `json:"values"`
}

type BpmnValueSkeleton struct {
	Inputs  map[string]interface{} `json:"inputs,omitempty"`
	Outputs map[string]interface{} `json:"outputs,omitempty"`
}

type InputOutput struct {
	Name    string        `json:"name,omitempty"`
	FieldId string        `json:"field_id"`
	Type    Type          `json:"type"`
	Value   string        `json:"value,omitempty"`
	Values  []InputOutput `json:"values,omitempty"`
}

type Type struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
	Base string `json:"base"`
}

type AbstractProcessParameter struct {
	Task         []AbstractTask
	DeviceTypeId string
	Options      []model.DeviceInstance
	Selected     model.DeviceInstance
}

type AbstractTaskParameter struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AbstractTask struct {
	Id            string                  `json:"id"`
	Values        BpmnValueSkeleton       `json:"values"`
	TaskParameter []AbstractTaskParameter `json:"task_parameter"`
	ServiceId     string                  `json:"service_id"`
	Label         string                  `json:"label"`
}

type EventParameter struct {
	Id       string
	ShapeId  string
	Event    RuleSet
	Options  []model.DeviceInstance
	Selected model.DeviceInstance
}

type AbstractProcess struct {
	Xml       string
	Name      string
	Parameter []AbstractProcessParameter
	Events    []EventParameter
}

type RuleSet struct {
	Id        string `json:"id" bson:"id"`
	Name      string `json:"name" bson:"name"`
	ServiceId string `json:"service_id" bson:"service_id"`
}
