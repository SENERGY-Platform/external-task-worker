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

package mapping

import (
	"encoding/json"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"reflect"
	"testing"
)

func TestCastToCategory_simpleStructMapping(t *testing.T) {
	t.Parallel()
	msg := `{
	"r": 255,
	"g": 0,
	"b": 100
}`
	var message interface{}
	err := json.Unmarshal([]byte(msg), &message)
	if err != nil {
		t.Fatal(err)
	}

	out, err := MapSensor(message, model.ContentVariable{
		Id:   "rgb_content",
		Name: "rgb",
		Type: model.Structure,
		SubContentVariables: []model.ContentVariable{
			{
				Id:               "r",
				Name:             "r",
				Type:             model.Integer,
				CharacteristicId: "rgb.r",
			},
			{
				Id:               "g",
				Name:             "g",
				Type:             model.Integer,
				CharacteristicId: "rgb.g",
			},
			{
				Id:               "b",
				Name:             "b",
				Type:             model.Integer,
				CharacteristicId: "rgb.b",
			},
		},
	}, model.Characteristic{
		Id:   "rgb",
		Name: "rgb",
		Type: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{
				Id:   "rgb.r",
				Name: "red",
				Type: model.Integer,
			},
			{
				Id:   "rgb.g",
				Name: "green",
				Type: model.Integer,
			},
			{
				Id:   "rgb.b",
				Name: "blue",
				Type: model.Integer,
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	result, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := json.Marshal(map[string]interface{}{
		"red":   255,
		"green": 0,
		"blue":  100,
	})
	if err != nil {
		t.Fatal(err)
	}

	var resultMsg interface{}
	err = json.Unmarshal(result, &resultMsg)
	if err != nil {
		t.Fatal(err)
	}

	var expectedMsg interface{}
	err = json.Unmarshal(expected, &expectedMsg)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resultMsg, expectedMsg) {
		t.Fatal(string(result), string(expected), resultMsg, expectedMsg)
	}
}

func TestCastToCategory_MapOfStructsMapping(t *testing.T) {
	t.Parallel()
	msg := `{
	"color_a": {
		"r": 255,
		"g": 0,
		"b": 100
	},
    "color_b": {
		"r": 200,
		"g": 200,
		"b": 200
	}
}`
	var message interface{}
	err := json.Unmarshal([]byte(msg), &message)
	if err != nil {
		t.Fatal(err)
	}

	out, err := MapSensor(message, model.ContentVariable{
		Id:               "rgbmap",
		Name:             "rgbmap",
		Type:             model.Structure,
		CharacteristicId: "rgb.map",
		SubContentVariables: []model.ContentVariable{
			{
				Id:               "rgb_content",
				Name:             "*",
				CharacteristicId: "rgb",
				Type:             model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:               "r",
						Name:             "r",
						Type:             model.Integer,
						CharacteristicId: "rgb.r",
					},
					{
						Id:               "g",
						Name:             "g",
						Type:             model.Integer,
						CharacteristicId: "rgb.g",
					},
					{
						Id:               "b",
						Name:             "b",
						Type:             model.Integer,
						CharacteristicId: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:   "rgb.map",
		Name: "rgbmap",
		Type: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{
				Id:   "rgb",
				Name: "*",
				Type: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:   "rgb.r",
						Name: "red",
						Type: model.Integer,
					},
					{
						Id:   "rgb.g",
						Name: "green",
						Type: model.Integer,
					},
					{
						Id:   "rgb.b",
						Name: "blue",
						Type: model.Integer,
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	result, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := json.Marshal(map[string]map[string]interface{}{
		"color_a": {
			"red":   255,
			"green": 0,
			"blue":  100,
		},
		"color_b": {
			"red":   200,
			"green": 200,
			"blue":  200,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var resultMsg interface{}
	err = json.Unmarshal(result, &resultMsg)
	if err != nil {
		t.Fatal(err)
	}

	var expectedMsg interface{}
	err = json.Unmarshal(expected, &expectedMsg)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resultMsg, expectedMsg) {
		t.Fatal(string(result), string(expected), resultMsg, expectedMsg)
	}
}

func TestCastToCategory_MapOfStructsMapping2(t *testing.T) {
	t.Parallel()
	msg := `{
	"color_a": {
		"r": 255,
		"g": 0,
		"b": 100
	},
    "color_b": {
		"r": 200,
		"g": 200,
		"b": 200
	}
}`
	var message interface{}
	err := json.Unmarshal([]byte(msg), &message)
	if err != nil {
		t.Fatal(err)
	}

	out, err := MapSensor(message, model.ContentVariable{
		Id:   "rgbmap",
		Name: "rgbmap",
		Type: model.Structure,
		SubContentVariables: []model.ContentVariable{
			{
				Id:               "rgb_content",
				Name:             "*",
				CharacteristicId: "rgb",
				Type:             model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:               "r",
						Name:             "r",
						Type:             model.Integer,
						CharacteristicId: "rgb.r",
					},
					{
						Id:               "g",
						Name:             "g",
						Type:             model.Integer,
						CharacteristicId: "rgb.g",
					},
					{
						Id:               "b",
						Name:             "b",
						Type:             model.Integer,
						CharacteristicId: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:   "rgb.map",
		Name: "rgbmap",
		Type: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{
				Id:   "rgb",
				Name: "*",
				Type: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:   "rgb.r",
						Name: "red",
						Type: model.Integer,
					},
					{
						Id:   "rgb.g",
						Name: "green",
						Type: model.Integer,
					},
					{
						Id:   "rgb.b",
						Name: "blue",
						Type: model.Integer,
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	result, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := json.Marshal(map[string]map[string]interface{}{
		"color_a": {
			"red":   255,
			"green": 0,
			"blue":  100,
		},
		"color_b": {
			"red":   200,
			"green": 200,
			"blue":  200,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var resultMsg interface{}
	err = json.Unmarshal(result, &resultMsg)
	if err != nil {
		t.Fatal(err)
	}

	var expectedMsg interface{}
	err = json.Unmarshal(expected, &expectedMsg)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resultMsg, expectedMsg) {
		t.Fatal(string(result), string(expected), resultMsg, expectedMsg)
	}
}

func TestCastToCategory_MapOfStructsMapping3(t *testing.T) {
	t.Parallel()
	msg := `{
	"color_a": {
		"r": 255,
		"g": 0,
		"b": 100
	},
    "color_b": {
		"r": 200,
		"g": 200,
		"b": 200
	}
}`
	var message interface{}
	err := json.Unmarshal([]byte(msg), &message)
	if err != nil {
		t.Fatal(err)
	}

	out, err := MapSensor(message, model.ContentVariable{
		Id:               "rgbmap",
		Name:             "rgbmap",
		Type:             model.Structure,
		CharacteristicId: "rgb.map",
		SubContentVariables: []model.ContentVariable{
			{
				Id:   "rgb_content",
				Name: "*",
				Type: model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:               "r",
						Name:             "r",
						Type:             model.Integer,
						CharacteristicId: "rgb.r",
					},
					{
						Id:               "g",
						Name:             "g",
						Type:             model.Integer,
						CharacteristicId: "rgb.g",
					},
					{
						Id:               "b",
						Name:             "b",
						Type:             model.Integer,
						CharacteristicId: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:   "rgb.map",
		Name: "rgbmap",
		Type: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{
				Id:   "rgb",
				Name: "*",
				Type: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:   "rgb.r",
						Name: "red",
						Type: model.Integer,
					},
					{
						Id:   "rgb.g",
						Name: "green",
						Type: model.Integer,
					},
					{
						Id:   "rgb.b",
						Name: "blue",
						Type: model.Integer,
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	result, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := json.Marshal(map[string]map[string]interface{}{
		"color_a": {
			"red":   255,
			"green": 0,
			"blue":  100,
		},
		"color_b": {
			"red":   200,
			"green": 200,
			"blue":  200,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var resultMsg interface{}
	err = json.Unmarshal(result, &resultMsg)
	if err != nil {
		t.Fatal(err)
	}

	var expectedMsg interface{}
	err = json.Unmarshal(expected, &expectedMsg)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resultMsg, expectedMsg) {
		t.Fatal(string(result), string(expected), resultMsg, expectedMsg)
	}
}

func TestCastToCategory_IndexListOfStructsMapping(t *testing.T) {
	t.Parallel()
	msg := `[{
		"r": 255,
		"g": 0,
		"b": 100
	}
]`
	var message interface{}
	err := json.Unmarshal([]byte(msg), &message)
	if err != nil {
		t.Fatal(err)
	}

	out, err := MapSensor(message, model.ContentVariable{
		Id:   "rgbmap",
		Name: "rgbmap",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:   "rgb_content",
				Name: "0",
				Type: model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:               "r",
						Name:             "r",
						Type:             model.Integer,
						CharacteristicId: "rgb.r",
					},
					{
						Id:               "g",
						Name:             "g",
						Type:             model.Integer,
						CharacteristicId: "rgb.g",
					},
					{
						Id:               "b",
						Name:             "b",
						Type:             model.Integer,
						CharacteristicId: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:   "rgb.map",
		Name: "rgbmap",
		Type: model.List,
		SubCharacteristics: []model.Characteristic{
			{
				Id:   "rgb",
				Name: "0",
				Type: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:   "rgb.r",
						Name: "red",
						Type: model.Integer,
					},
					{
						Id:   "rgb.g",
						Name: "green",
						Type: model.Integer,
					},
					{
						Id:   "rgb.b",
						Name: "blue",
						Type: model.Integer,
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	result, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := json.Marshal([]map[string]interface{}{
		{
			"red":   255,
			"green": 0,
			"blue":  100,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var resultMsg interface{}
	err = json.Unmarshal(result, &resultMsg)
	if err != nil {
		t.Fatal(err)
	}

	var expectedMsg interface{}
	err = json.Unmarshal(expected, &expectedMsg)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resultMsg, expectedMsg) {
		t.Fatal(string(result), string(expected), resultMsg, expectedMsg)
	}
}

func TestCastToCategory_ListOfStructsMapping(t *testing.T) {
	t.Parallel()
	msg := `[{
		"r": 255,
		"g": 0,
		"b": 100
	},
    {
		"r": 200,
		"g": 200,
		"b": 200
	}
]`
	var message interface{}
	err := json.Unmarshal([]byte(msg), &message)
	if err != nil {
		t.Fatal(err)
	}

	out, err := MapSensor(message, model.ContentVariable{
		Id:               "rgbmap",
		Name:             "rgbmap",
		Type:             model.List,
		CharacteristicId: "rgb.map",
		SubContentVariables: []model.ContentVariable{
			{
				Id:               "rgb_content",
				Name:             "*",
				CharacteristicId: "rgb",
				Type:             model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:               "r",
						Name:             "r",
						Type:             model.Integer,
						CharacteristicId: "rgb.r",
					},
					{
						Id:               "g",
						Name:             "g",
						Type:             model.Integer,
						CharacteristicId: "rgb.g",
					},
					{
						Id:               "b",
						Name:             "b",
						Type:             model.Integer,
						CharacteristicId: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:   "rgb.map",
		Name: "rgbmap",
		Type: model.List,
		SubCharacteristics: []model.Characteristic{
			{
				Id:   "rgb",
				Name: "*",
				Type: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:   "rgb.r",
						Name: "red",
						Type: model.Integer,
					},
					{
						Id:   "rgb.g",
						Name: "green",
						Type: model.Integer,
					},
					{
						Id:   "rgb.b",
						Name: "blue",
						Type: model.Integer,
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	result, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := json.Marshal([]map[string]interface{}{
		{
			"red":   255,
			"green": 0,
			"blue":  100,
		},
		{
			"red":   200,
			"green": 200,
			"blue":  200,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var resultMsg interface{}
	err = json.Unmarshal(result, &resultMsg)
	if err != nil {
		t.Fatal(err)
	}

	var expectedMsg interface{}
	err = json.Unmarshal(expected, &expectedMsg)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resultMsg, expectedMsg) {
		t.Fatal(string(result), string(expected), resultMsg, expectedMsg)
	}
}

func TestCastToCategory_ListOfStructsMapping2(t *testing.T) {
	t.Parallel()
	msg := `[{
		"r": 255,
		"g": 0,
		"b": 100
	},
    {
		"r": 200,
		"g": 200,
		"b": 200
	}
]`
	var message interface{}
	err := json.Unmarshal([]byte(msg), &message)
	if err != nil {
		t.Fatal(err)
	}

	out, err := MapSensor(message, model.ContentVariable{
		Id:   "rgbmap",
		Name: "rgbmap",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:               "rgb_content",
				Name:             "*",
				CharacteristicId: "rgb",
				Type:             model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:               "r",
						Name:             "r",
						Type:             model.Integer,
						CharacteristicId: "rgb.r",
					},
					{
						Id:               "g",
						Name:             "g",
						Type:             model.Integer,
						CharacteristicId: "rgb.g",
					},
					{
						Id:               "b",
						Name:             "b",
						Type:             model.Integer,
						CharacteristicId: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:   "rgb.map",
		Name: "rgbmap",
		Type: model.List,
		SubCharacteristics: []model.Characteristic{
			{
				Id:   "rgb",
				Name: "*",
				Type: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:   "rgb.r",
						Name: "red",
						Type: model.Integer,
					},
					{
						Id:   "rgb.g",
						Name: "green",
						Type: model.Integer,
					},
					{
						Id:   "rgb.b",
						Name: "blue",
						Type: model.Integer,
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	result, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := json.Marshal([]map[string]interface{}{
		{
			"red":   255,
			"green": 0,
			"blue":  100,
		},
		{
			"red":   200,
			"green": 200,
			"blue":  200,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var resultMsg interface{}
	err = json.Unmarshal(result, &resultMsg)
	if err != nil {
		t.Fatal(err)
	}

	var expectedMsg interface{}
	err = json.Unmarshal(expected, &expectedMsg)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resultMsg, expectedMsg) {
		t.Fatal(string(result), string(expected), resultMsg, expectedMsg)
	}
}

func TestCastToCategory_ListOfStructsMapping3(t *testing.T) {
	t.Parallel()
	msg := `[{
		"r": 255,
		"g": 0,
		"b": 100
	},
    {
		"r": 200,
		"g": 200,
		"b": 200
	}
]`
	var message interface{}
	err := json.Unmarshal([]byte(msg), &message)
	if err != nil {
		t.Fatal(err)
	}

	out, err := MapSensor(message, model.ContentVariable{
		Id:               "rgbmap",
		Name:             "rgbmap",
		Type:             model.List,
		CharacteristicId: "rgb.map",
		SubContentVariables: []model.ContentVariable{
			{
				Id:   "rgb_content",
				Name: "*",
				Type: model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:               "r",
						Name:             "r",
						Type:             model.Integer,
						CharacteristicId: "rgb.r",
					},
					{
						Id:               "g",
						Name:             "g",
						Type:             model.Integer,
						CharacteristicId: "rgb.g",
					},
					{
						Id:               "b",
						Name:             "b",
						Type:             model.Integer,
						CharacteristicId: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:   "rgb.map",
		Name: "rgbmap",
		Type: model.List,
		SubCharacteristics: []model.Characteristic{
			{
				Id:   "rgb",
				Name: "*",
				Type: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:   "rgb.r",
						Name: "red",
						Type: model.Integer,
					},
					{
						Id:   "rgb.g",
						Name: "green",
						Type: model.Integer,
					},
					{
						Id:   "rgb.b",
						Name: "blue",
						Type: model.Integer,
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	result, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := json.Marshal([]map[string]interface{}{
		{
			"red":   255,
			"green": 0,
			"blue":  100,
		},
		{
			"red":   200,
			"green": 200,
			"blue":  200,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var resultMsg interface{}
	err = json.Unmarshal(result, &resultMsg)
	if err != nil {
		t.Fatal(err)
	}

	var expectedMsg interface{}
	err = json.Unmarshal(expected, &expectedMsg)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resultMsg, expectedMsg) {
		t.Fatal(string(result), string(expected), resultMsg, expectedMsg)
	}
}

func TestCastToCategory_IndexListOfStructsToStructMapping(t *testing.T) {
	t.Parallel()
	msg := `[{
		"r": 255,
		"g": 0,
		"b": 100
	}
]`
	var message interface{}
	err := json.Unmarshal([]byte(msg), &message)
	if err != nil {
		t.Fatal(err)
	}

	out, err := MapSensor(message, model.ContentVariable{
		Id:   "rgbmap",
		Name: "rgbmap",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:   "rgb_content",
				Name: "0",
				Type: model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:               "r",
						Name:             "r",
						Type:             model.Integer,
						CharacteristicId: "rgb.r",
					},
					{
						Id:               "g",
						Name:             "g",
						Type:             model.Integer,
						CharacteristicId: "rgb.g",
					},
					{
						Id:               "b",
						Name:             "b",
						Type:             model.Integer,
						CharacteristicId: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:   "rgb.map",
		Name: "rgbmap",
		Type: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{
				Id:   "rgb",
				Name: "first",
				Type: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:   "rgb.r",
						Name: "red",
						Type: model.Integer,
					},
					{
						Id:   "rgb.g",
						Name: "green",
						Type: model.Integer,
					},
					{
						Id:   "rgb.b",
						Name: "blue",
						Type: model.Integer,
					},
				},
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	result, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := json.Marshal(map[string]map[string]interface{}{
		"first": {
			"red":   255,
			"green": 0,
			"blue":  100,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	var resultMsg interface{}
	err = json.Unmarshal(result, &resultMsg)
	if err != nil {
		t.Fatal(err)
	}

	var expectedMsg interface{}
	err = json.Unmarshal(expected, &expectedMsg)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resultMsg, expectedMsg) {
		t.Fatal(string(result), string(expected), resultMsg, expectedMsg)
	}
}

func TestCastToCategory_ListOfStringsMapping(t *testing.T) {
	t.Parallel()
	msg := `["foo", "bar", "batz"]`
	var message interface{}
	err := json.Unmarshal([]byte(msg), &message)
	if err != nil {
		t.Fatal(err)
	}

	out, err := MapSensor(message, model.ContentVariable{
		Id:   "list",
		Name: "list",
		Type: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:               "element",
				Name:             "*",
				CharacteristicId: "str",
				Type:             model.String,
			},
		},
	}, model.Characteristic{
		Id:   "list",
		Name: "list",
		Type: model.List,
		SubCharacteristics: []model.Characteristic{
			{
				Id:   "str",
				Name: "*",
				Type: model.String,
			},
		},
	})

	if err != nil {
		t.Fatal(err)
	}

	result, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := json.Marshal([]string{
		"foo", "bar", "batz",
	})
	if err != nil {
		t.Fatal(err)
	}

	var resultMsg interface{}
	err = json.Unmarshal(result, &resultMsg)
	if err != nil {
		t.Fatal(err)
	}

	var expectedMsg interface{}
	err = json.Unmarshal(expected, &expectedMsg)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(resultMsg, expectedMsg) {
		t.Fatal(string(result), string(expected), resultMsg, expectedMsg)
	}
}
