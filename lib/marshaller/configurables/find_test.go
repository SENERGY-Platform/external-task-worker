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

package configurables

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/color"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/temperature"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"reflect"
	"testing"
)

func TestFindIntersectingConfigurables1(t *testing.T) {
	assert := Assertions{t}
	service1 := model.Service{
		Id:          "s1",
		LocalId:     "s1l",
		Name:        "s1n",
		Description: "s1d",
		ProtocolId:  "p1",
		Inputs: []model.Content{
			{
				Id: "c1",
				ContentVariable: model.ContentVariable{
					Id:   "c1.1",
					Name: "payload",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:               "c1.1.2",
							Name:             "temperature",
							Type:             model.Float,
							CharacteristicId: temperature.Celcius,
						},
						{
							Id:               "c1.1.1",
							Name:             "color",
							Type:             model.Structure,
							CharacteristicId: color.Rgb,
							SubContentVariables: []model.ContentVariable{
								{
									Id:   "c2.1.4",
									Name: "foo",
									Type: model.String,
								},
								{
									Id:               "sr",
									Name:             "red",
									Type:             model.Integer,
									CharacteristicId: color.RgbR,
								},
								{
									Id:               "sg",
									Name:             "green",
									Type:             model.Integer,
									CharacteristicId: color.RgbG,
								},
								{
									Id:               "sb",
									Name:             "blue",
									Type:             model.Integer,
									CharacteristicId: color.RgbB,
								},
							},
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "p1.1",
			},
		},
	}

	service2 := model.Service{
		Id:          "s2",
		LocalId:     "s2l",
		Name:        "s2n",
		Description: "s2d",
		ProtocolId:  "p1",
		Inputs: []model.Content{
			{
				Id: "c2",
				ContentVariable: model.ContentVariable{
					Id:   "c2.1",
					Name: "payload",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:   "c2.1.4",
							Name: "foo",
							Type: model.String,
						},
						{
							Id:   "c2.1.3",
							Name: "bar",
							Type: model.String,
						},
						{
							Id:               "c2.1.2",
							Name:             "temperature",
							Type:             model.Float,
							CharacteristicId: temperature.Celcius,
						},
						{
							Id:               "c2.1.1",
							Name:             "color",
							Type:             model.String,
							CharacteristicId: color.Hex,
						},
						{
							Id:               "c2.1.5",
							Name:             "color_2",
							Type:             model.String,
							CharacteristicId: color.Hex,
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "p1.1",
			},
		},
	}

	configurables, err := FindIntersectingConfigurables(temperature.Celcius, []model.Service{service1, service2})
	if err != nil {
		t.Fatal(err)
	}

	if len(configurables) != 1 {
		t.Fatal(configurables)
	}

	assert.ListContains(configurables, Configurable{
		CharacteristicId: color.Rgb,
		Values: []ConfigurableCharacteristicValue{
			{
				Path:      "r",
				Value:     0,
				ValueType: model.Integer,
			},
			{
				Path:      "g",
				Value:     0,
				ValueType: model.Integer,
			},
			{
				Path:      "b",
				Value:     0,
				ValueType: model.Integer,
			},
		},
	})
}

type Assertions struct {
	*testing.T
}

func (this Assertions) ListContains(list interface{}, element interface{}) {
	this.Helper()
	listValue := reflect.ValueOf(list)
	for i := 0; i < listValue.Len(); i++ {
		if reflect.DeepEqual(listValue.Index(i).Interface(), element) {
			return
		}
	}
	this.Fatal("missing element in list", list, element)
}
