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
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/example"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
)

func ExampleMarshalInput1() {
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
							Id:   "c1.1.1",
							Name: "color",
							Type: model.Structure,
							SubContentVariables: []model.ContentVariable{
								{
									Id:               "sr",
									Name:             "red",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".r",
								},
								{
									Id:               "sg",
									Name:             "green",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".g",
								},
								{
									Id:               "sb",
									Name:             "blue",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".b",
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
	input := "#ff0064"
	inputCharacteristic := example.Hex
	result, err := MarshalInputs(protocol, service, input, inputCharacteristic)
	fmt.Println(result, err)

	//output:
	//map[body:{"color":{"blue":100,"green":0,"red":255}}] <nil>
}

func ExampleMarshalInput2() {
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
							Id:               "c1.1.1",
							Name:             "color",
							CharacteristicId: example.Hex,
							Type:             model.String,
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "p1.1",
			},
		},
	}
	input := map[string]interface{}{
		"r": float64(255),
		"g": float64(0),
		"b": float64(100),
	}
	inputCharacteristic := example.Rgb
	result, err := MarshalInputs(protocol, service, input, inputCharacteristic)
	fmt.Println(result, err)

	//output:
	//map[body:{"color":"#ff0064"}] <nil>
}

func ExampleMarshalInputMulti() {
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
							Id:   "c1.1.1",
							Name: "color",
							Type: model.Structure,
							SubContentVariables: []model.ContentVariable{
								{
									Id:               "sr",
									Name:             "red",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".r",
									Value:            float64(255),
								},
								{
									Id:               "sg",
									Name:             "green",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".g",
									Value:            float64(255),
								},
								{
									Id:               "sb",
									Name:             "blue",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".b",
									Value:            float64(255),
								},
							},
						},
						{
							Id:               "c1.1.2",
							Name:             "bri",
							Type:             model.Integer,
							CharacteristicId: example.Lux,
							Value:            float64(100),
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "p1.1",
			},
		},
	}
	fmt.Println(MarshalInputs(protocol, service, "#ff0064", example.Hex))
	fmt.Println(MarshalInputs(protocol, service, float64(25), example.Lux))

	//output:
	//map[body:{"bri":100,"color":{"blue":100,"green":0,"red":255}}] <nil>
	//map[body:{"bri":25,"color":{"blue":255,"green":255,"red":255}}] <nil>
}

func ExampleMarshalInputMultiXml() {
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
							Id:   "c1.1.1",
							Name: "color",
							Type: model.Structure,
							SubContentVariables: []model.ContentVariable{
								{
									Id:               "sr",
									Name:             "-red",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".r",
									Value:            float64(255),
								},
								{
									Id:               "sg",
									Name:             "-green",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".g",
									Value:            float64(255),
								},
								{
									Id:               "sb",
									Name:             "-blue",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".b",
									Value:            float64(255),
								},
							},
						},
						{
							Id:               "c1.1.2",
							Name:             "bri",
							Type:             model.Integer,
							CharacteristicId: example.Lux,
							Value:            float64(100),
						},
					},
				},
				Serialization:     "xml",
				ProtocolSegmentId: "p1.1",
			},
		},
	}
	fmt.Println(MarshalInputs(protocol, service, "#ff0064", example.Hex))
	fmt.Println(MarshalInputs(protocol, service, float64(25), example.Lux))

	//output:
	//map[body:<payload><bri>100</bri><color blue="100" green="0" red="255"/></payload>] <nil>
	//map[body:<payload><bri>25</bri><color blue="255" green="255" red="255"/></payload>] <nil>
}

func ExampleMarshalInputNull() {
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
							Id:   "c1.1.1",
							Name: "color",
							Type: model.Structure,
							SubContentVariables: []model.ContentVariable{
								{
									Id:               "sr",
									Name:             "-red",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".r",
									Value:            float64(255),
								},
								{
									Id:               "sg",
									Name:             "-green",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".g",
									Value:            float64(0),
								},
								{
									Id:               "sb",
									Name:             "-blue",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".b",
									Value:            float64(100),
								},
							},
						},
						{
							Id:               "c1.1.2",
							Name:             "bri",
							Type:             model.Integer,
							CharacteristicId: example.Lux,
							Value:            float64(100),
						},
					},
				},
				Serialization:     "xml",
				ProtocolSegmentId: "p1.1",
			},
		},
	}
	fmt.Println(MarshalInputs(protocol, service, nil, ""))
	fmt.Println(MarshalInputs(protocol, service, nil, model.NullCharacteristic.Id))
	fmt.Println(MarshalInputs(protocol, service, "something", ""))
	fmt.Println(MarshalInputs(protocol, service, "something", model.NullCharacteristic.Id))
	fmt.Println(MarshalInputs(protocol, service, map[string]string{"foo": "bar"}, ""))
	fmt.Println(MarshalInputs(protocol, service, map[string]string{"foo": "bar"}, model.NullCharacteristic.Id))
	fmt.Println(MarshalInputs(protocol, service, "#ff0064", ""))
	fmt.Println(MarshalInputs(protocol, service, "#ff0064", model.NullCharacteristic.Id))

	//output:
	//map[body:<payload><bri>100</bri><color blue="100" green="0" red="255"/></payload>] <nil>
	//map[body:<payload><bri>100</bri><color blue="100" green="0" red="255"/></payload>] <nil>
	//map[body:<payload><bri>100</bri><color blue="100" green="0" red="255"/></payload>] <nil>
	//map[body:<payload><bri>100</bri><color blue="100" green="0" red="255"/></payload>] <nil>
	//map[body:<payload><bri>100</bri><color blue="100" green="0" red="255"/></payload>] <nil>
	//map[body:<payload><bri>100</bri><color blue="100" green="0" red="255"/></payload>] <nil>
	//map[body:<payload><bri>100</bri><color blue="100" green="0" red="255"/></payload>] <nil>
	//map[body:<payload><bri>100</bri><color blue="100" green="0" red="255"/></payload>] <nil>
}

func ExampleMarshalEmptyService() {
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
		Id:          "s1",
		LocalId:     "s1l",
		Name:        "s1n",
		Description: "s1d",
		ProtocolId:  "p1",
	}
	fmt.Println(MarshalInputs(protocol, service, nil, ""))
	fmt.Println(MarshalInputs(protocol, service, nil, model.NullCharacteristic.Id))
	fmt.Println(MarshalInputs(protocol, service, "something", ""))
	fmt.Println(MarshalInputs(protocol, service, "something", model.NullCharacteristic.Id))
	fmt.Println(MarshalInputs(protocol, service, map[string]string{"foo": "bar"}, ""))
	fmt.Println(MarshalInputs(protocol, service, map[string]string{"foo": "bar"}, model.NullCharacteristic.Id))
	fmt.Println(MarshalInputs(protocol, service, "#ff0064", ""))
	fmt.Println(MarshalInputs(protocol, service, "#ff0064", model.NullCharacteristic.Id))

	fmt.Println(MarshalInputs(protocol, service, "#ff0064", example.Hex))
	fmt.Println(MarshalInputs(protocol, service, map[string]interface{}{"r": float64(255), "g": float64(0), "b": float64(100)}, example.Rgb))

	//output:
	//map[] <nil>
	//map[] <nil>
	//map[] <nil>
	//map[] <nil>
	//map[] <nil>
	//map[] <nil>
	//map[] <nil>
	//map[] <nil>
	//map[] <nil>
	//map[] <nil>
}
