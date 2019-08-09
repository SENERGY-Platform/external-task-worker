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
		Id:        "rgb_content",
		Name:      "rgb",
		ValueType: model.Structure,
		SubContentVariables: []model.ContentVariable{
			{
				Id:         "r",
				Name:       "r",
				ValueType:  model.Integer,
				ExactMatch: "rgb.r",
			},
			{
				Id:         "g",
				Name:       "g",
				ValueType:  model.Integer,
				ExactMatch: "rgb.g",
			},
			{
				Id:         "b",
				Name:       "b",
				ValueType:  model.Integer,
				ExactMatch: "rgb.b",
			},
		},
	}, model.Characteristic{
		Id:        "rgb",
		Name:      "rgb",
		ValueType: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{
				Id:        "rgb.r",
				Name:      "red",
				ValueType: model.Integer,
			},
			{
				Id:        "rgb.g",
				Name:      "green",
				ValueType: model.Integer,
			},
			{
				Id:        "rgb.b",
				Name:      "blue",
				ValueType: model.Integer,
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
		Id:         "rgbmap",
		Name:       "rgbmap",
		ValueType:  model.Structure,
		ExactMatch: "rgb.map",
		SubContentVariables: []model.ContentVariable{
			{
				Id:         "rgb_content",
				Name:       "*",
				ExactMatch: "rgb",
				ValueType:  model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:         "r",
						Name:       "r",
						ValueType:  model.Integer,
						ExactMatch: "rgb.r",
					},
					{
						Id:         "g",
						Name:       "g",
						ValueType:  model.Integer,
						ExactMatch: "rgb.g",
					},
					{
						Id:         "b",
						Name:       "b",
						ValueType:  model.Integer,
						ExactMatch: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:        "rgb.map",
		Name:      "rgbmap",
		ValueType: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{
				Id:        "rgb",
				Name:      "*",
				ValueType: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:        "rgb.r",
						Name:      "red",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.g",
						Name:      "green",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.b",
						Name:      "blue",
						ValueType: model.Integer,
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
		Id:        "rgbmap",
		Name:      "rgbmap",
		ValueType: model.Structure,
		SubContentVariables: []model.ContentVariable{
			{
				Id:         "rgb_content",
				Name:       "*",
				ExactMatch: "rgb",
				ValueType:  model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:         "r",
						Name:       "r",
						ValueType:  model.Integer,
						ExactMatch: "rgb.r",
					},
					{
						Id:         "g",
						Name:       "g",
						ValueType:  model.Integer,
						ExactMatch: "rgb.g",
					},
					{
						Id:         "b",
						Name:       "b",
						ValueType:  model.Integer,
						ExactMatch: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:        "rgb.map",
		Name:      "rgbmap",
		ValueType: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{
				Id:        "rgb",
				Name:      "*",
				ValueType: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:        "rgb.r",
						Name:      "red",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.g",
						Name:      "green",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.b",
						Name:      "blue",
						ValueType: model.Integer,
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
		Id:         "rgbmap",
		Name:       "rgbmap",
		ValueType:  model.Structure,
		ExactMatch: "rgb.map",
		SubContentVariables: []model.ContentVariable{
			{
				Id:        "rgb_content",
				Name:      "*",
				ValueType: model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:         "r",
						Name:       "r",
						ValueType:  model.Integer,
						ExactMatch: "rgb.r",
					},
					{
						Id:         "g",
						Name:       "g",
						ValueType:  model.Integer,
						ExactMatch: "rgb.g",
					},
					{
						Id:         "b",
						Name:       "b",
						ValueType:  model.Integer,
						ExactMatch: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:        "rgb.map",
		Name:      "rgbmap",
		ValueType: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{
				Id:        "rgb",
				Name:      "*",
				ValueType: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:        "rgb.r",
						Name:      "red",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.g",
						Name:      "green",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.b",
						Name:      "blue",
						ValueType: model.Integer,
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
		Id:        "rgbmap",
		Name:      "rgbmap",
		ValueType: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:        "rgb_content",
				Name:      "0",
				ValueType: model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:         "r",
						Name:       "r",
						ValueType:  model.Integer,
						ExactMatch: "rgb.r",
					},
					{
						Id:         "g",
						Name:       "g",
						ValueType:  model.Integer,
						ExactMatch: "rgb.g",
					},
					{
						Id:         "b",
						Name:       "b",
						ValueType:  model.Integer,
						ExactMatch: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:        "rgb.map",
		Name:      "rgbmap",
		ValueType: model.List,
		SubCharacteristics: []model.Characteristic{
			{
				Id:        "rgb",
				Name:      "0",
				ValueType: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:        "rgb.r",
						Name:      "red",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.g",
						Name:      "green",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.b",
						Name:      "blue",
						ValueType: model.Integer,
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
		Id:         "rgbmap",
		Name:       "rgbmap",
		ValueType:  model.List,
		ExactMatch: "rgb.map",
		SubContentVariables: []model.ContentVariable{
			{
				Id:         "rgb_content",
				Name:       "*",
				ExactMatch: "rgb",
				ValueType:  model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:         "r",
						Name:       "r",
						ValueType:  model.Integer,
						ExactMatch: "rgb.r",
					},
					{
						Id:         "g",
						Name:       "g",
						ValueType:  model.Integer,
						ExactMatch: "rgb.g",
					},
					{
						Id:         "b",
						Name:       "b",
						ValueType:  model.Integer,
						ExactMatch: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:        "rgb.map",
		Name:      "rgbmap",
		ValueType: model.List,
		SubCharacteristics: []model.Characteristic{
			{
				Id:        "rgb",
				Name:      "*",
				ValueType: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:        "rgb.r",
						Name:      "red",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.g",
						Name:      "green",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.b",
						Name:      "blue",
						ValueType: model.Integer,
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
		Id:        "rgbmap",
		Name:      "rgbmap",
		ValueType: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:         "rgb_content",
				Name:       "*",
				ExactMatch: "rgb",
				ValueType:  model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:         "r",
						Name:       "r",
						ValueType:  model.Integer,
						ExactMatch: "rgb.r",
					},
					{
						Id:         "g",
						Name:       "g",
						ValueType:  model.Integer,
						ExactMatch: "rgb.g",
					},
					{
						Id:         "b",
						Name:       "b",
						ValueType:  model.Integer,
						ExactMatch: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:        "rgb.map",
		Name:      "rgbmap",
		ValueType: model.List,
		SubCharacteristics: []model.Characteristic{
			{
				Id:        "rgb",
				Name:      "*",
				ValueType: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:        "rgb.r",
						Name:      "red",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.g",
						Name:      "green",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.b",
						Name:      "blue",
						ValueType: model.Integer,
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
		Id:         "rgbmap",
		Name:       "rgbmap",
		ValueType:  model.List,
		ExactMatch: "rgb.map",
		SubContentVariables: []model.ContentVariable{
			{
				Id:        "rgb_content",
				Name:      "*",
				ValueType: model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:         "r",
						Name:       "r",
						ValueType:  model.Integer,
						ExactMatch: "rgb.r",
					},
					{
						Id:         "g",
						Name:       "g",
						ValueType:  model.Integer,
						ExactMatch: "rgb.g",
					},
					{
						Id:         "b",
						Name:       "b",
						ValueType:  model.Integer,
						ExactMatch: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:        "rgb.map",
		Name:      "rgbmap",
		ValueType: model.List,
		SubCharacteristics: []model.Characteristic{
			{
				Id:        "rgb",
				Name:      "*",
				ValueType: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:        "rgb.r",
						Name:      "red",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.g",
						Name:      "green",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.b",
						Name:      "blue",
						ValueType: model.Integer,
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
		Id:        "rgbmap",
		Name:      "rgbmap",
		ValueType: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:        "rgb_content",
				Name:      "0",
				ValueType: model.Structure,
				SubContentVariables: []model.ContentVariable{
					{
						Id:         "r",
						Name:       "r",
						ValueType:  model.Integer,
						ExactMatch: "rgb.r",
					},
					{
						Id:         "g",
						Name:       "g",
						ValueType:  model.Integer,
						ExactMatch: "rgb.g",
					},
					{
						Id:         "b",
						Name:       "b",
						ValueType:  model.Integer,
						ExactMatch: "rgb.b",
					},
				},
			},
		},
	}, model.Characteristic{
		Id:        "rgb.map",
		Name:      "rgbmap",
		ValueType: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{
				Id:        "rgb",
				Name:      "first",
				ValueType: model.Structure,
				SubCharacteristics: []model.Characteristic{
					{
						Id:        "rgb.r",
						Name:      "red",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.g",
						Name:      "green",
						ValueType: model.Integer,
					},
					{
						Id:        "rgb.b",
						Name:      "blue",
						ValueType: model.Integer,
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
		Id:        "list",
		Name:      "list",
		ValueType: model.List,
		SubContentVariables: []model.ContentVariable{
			{
				Id:         "element",
				Name:       "*",
				ExactMatch: "str",
				ValueType:  model.String,
			},
		},
	}, model.Characteristic{
		Id:        "list",
		Name:      "list",
		ValueType: model.List,
		SubCharacteristics: []model.Characteristic{
			{
				Id:        "str",
				Name:      "*",
				ValueType: model.String,
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
