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

package model

type DeviceClass struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Function struct {
	Id         string   `json:"id"`
	Name       string   `json:"name"`
	ConceptIds []string `json:"concept_ids"`
	Type       string   `json:"type"`
}

type Aspect struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Concept struct {
	Id              string           `json:"id"`
	Name            string           `json:"name"`
	Characteristics []Characteristic `json:"characteristics"`
}

type Characteristic struct {
	Id                 string           `json:"id"`
	Name               string           `json:"name"`
	ValueType          ValueType        `json:"value_type"`
	MinValue           float64          `json:"min_value"`
	MaxValue           float64          `json:"max_value"`
	Value              interface{}      `json:"value"`
	SubCharacteristics []Characteristic `json:"sub_characteristics"`
}
