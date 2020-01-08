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

func TestSimpleStringSkeleton(t *testing.T) {
	t.Parallel()
	out, set, err := CharacteristicToSkeleton(model.Characteristic{Id: "fid", Name: "foo", Type: model.String})
	if err != nil {
		t.Fatal(err)
	}

	*(set["fid"]) = "bar"

	if v, ok := (*out).(string); !ok {
		t.Fatal()
	} else if v != "bar" {
		t.Fatal(v)
	}
}

func TestSimpleIntSkeleton(t *testing.T) {
	t.Parallel()
	out, set, err := CharacteristicToSkeleton(model.Characteristic{Id: "fid", Name: "foo", Type: model.Integer})
	if err != nil {
		t.Fatal(err)
	}

	*(set["fid"]) = int64(42)

	if v, ok := (*out).(int64); !ok {
		t.Fatal()
	} else if v != int64(42) {
		t.Fatal(v)
	}
}

func TestSimpleFloatSkeleton(t *testing.T) {
	t.Parallel()
	out, set, err := CharacteristicToSkeleton(model.Characteristic{Id: "fid", Name: "foo", Type: model.Float})
	if err != nil {
		t.Fatal(err)
	}

	*(set["fid"]) = float64(4.2)

	if v, ok := (*out).(float64); !ok {
		t.Fatal()
	} else if v != float64(4.2) {
		t.Fatal(v)
	}
}

func TestSimpleBoolSkeleton(t *testing.T) {
	t.Parallel()
	out, set, err := CharacteristicToSkeleton(model.Characteristic{Id: "fid", Name: "foo", Type: model.Boolean})
	if err != nil {
		t.Fatal(err)
	}

	*(set["fid"]) = true

	if v, ok := (*out).(bool); !ok {
		t.Fatal()
	} else if v != true {
		t.Fatal(v)
	}
}

func TestSimpleStructSkeleton(t *testing.T) {
	t.Parallel()
	out, _, err := CharacteristicToSkeleton(model.Characteristic{Name: "foo", Type: model.Structure})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := (*out).(map[string]interface{}); !ok {
		t.Fatal()
	}
}

func TestStructSkeleton(t *testing.T) {
	t.Parallel()
	out, set, err := CharacteristicToSkeleton(model.Characteristic{
		Id:   "rdf",
		Name: "rdf",
		Type: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{Id: "r", Name: "r", Type: model.Integer},
			{Id: "g", Name: "g", Type: model.Integer},
			{Id: "b", Name: "b", Type: model.Integer},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	*(set["r"]) = int64(255)
	*(set["g"]) = int64(100)
	*(set["b"]) = int64(0)

	if m, ok := (*out).(map[string]interface{}); !ok {
		t.Fatal()
	} else {
		if ptr, ok := m["r"].(*interface{}); !ok {
			t.Fatal(reflect.TypeOf(m["r"]).String())
		} else if n, ok := (*ptr).(int64); !ok || n != int64(255) {
			t.Fatal()
		}
		if ptr, ok := m["g"].(*interface{}); !ok {
			t.Fatal(reflect.TypeOf(m["g"]).String())
		} else if n, ok := (*ptr).(int64); !ok || n != int64(100) {
			t.Fatal()
		}
		if ptr, ok := m["b"].(*interface{}); !ok {
			t.Fatal(reflect.TypeOf(m["b"]).String())
		} else if n, ok := (*ptr).(int64); !ok || n != int64(0) {
			t.Fatal()
		}
	}

	msg1, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	msg2, err := json.Marshal(map[string]interface{}{
		"r": 255,
		"g": 100,
		"b": 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	var a interface{}
	err = json.Unmarshal(msg1, &a)
	if err != nil {
		t.Fatal(err)
	}

	var b interface{}
	err = json.Unmarshal(msg2, &b)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(a, b) {
		t.Fatal(string(msg1), string(msg2), a, b)
	}
}

func TestStructDefaultSkeleton(t *testing.T) {
	t.Parallel()
	out, set, err := CharacteristicToSkeleton(model.Characteristic{
		Id:   "rdf",
		Name: "rdf",
		Type: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{Id: "r", Name: "r", Type: model.Integer, Value: float64(255)}, // value as float because json transforms every number to float64 if target field is interface{}
			{Id: "g", Name: "g", Type: model.Integer, Value: float64(255)},
			{Id: "b", Name: "b", Type: model.Integer, Value: float64(255)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	*(set["g"]) = float64(100)

	if m, ok := (*out).(map[string]interface{}); !ok {
		t.Fatal()
	} else {
		if ptr, ok := m["r"].(*interface{}); !ok {
			t.Fatal(reflect.TypeOf(m["r"]).String())
		} else if n, ok := (*ptr).(float64); !ok || n != float64(255) {
			t.Fatal()
		}
		if ptr, ok := m["g"].(*interface{}); !ok {
			t.Fatal(reflect.TypeOf(m["g"]).String())
		} else if n, ok := (*ptr).(float64); !ok || n != float64(100) {
			t.Fatal()
		}
		if ptr, ok := m["b"].(*interface{}); !ok {
			t.Fatal(reflect.TypeOf(m["b"]).String())
		} else if n, ok := (*ptr).(float64); !ok || n != float64(255) {
			t.Fatal()
		}
	}

	msg1, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	msg2, err := json.Marshal(map[string]interface{}{
		"r": 255,
		"g": 100,
		"b": 255,
	})
	if err != nil {
		t.Fatal(err)
	}

	var a interface{}
	err = json.Unmarshal(msg1, &a)
	if err != nil {
		t.Fatal(err)
	}

	var b interface{}
	err = json.Unmarshal(msg2, &b)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(a, b) {
		t.Fatal(string(msg1), string(msg2), a, b)
	}
}

func TestSimpleListSkeleton(t *testing.T) {
	t.Parallel()
	out, _, err := CharacteristicToSkeleton(model.Characteristic{Name: "foo", Type: model.List})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := (*out).([]interface{}); !ok {
		t.Fatal()
	}
}

func TestSimpleStringSkeletonContent(t *testing.T) {
	t.Parallel()
	out, set, err := ContentToSkeleton(model.ContentVariable{CharacteristicId: "fid", Name: "foo", Type: model.String}, nil)
	if err != nil {
		t.Fatal(err)
	}

	*(set["fid"][0]) = "bar"

	if v, ok := (*out).(string); !ok {
		t.Fatal()
	} else if v != "bar" {
		t.Fatal(v)
	}
}

func TestSimpleIntSkeletonContent(t *testing.T) {
	t.Parallel()
	out, set, err := ContentToSkeleton(model.ContentVariable{CharacteristicId: "fid", Name: "foo", Type: model.Integer}, nil)
	if err != nil {
		t.Fatal(err)
	}

	*(set["fid"][0]) = int64(42)

	if v, ok := (*out).(int64); !ok {
		t.Fatal()
	} else if v != int64(42) {
		t.Fatal(v)
	}
}

func TestSimpleFloatSkeletonContent(t *testing.T) {
	t.Parallel()
	out, set, err := ContentToSkeleton(model.ContentVariable{CharacteristicId: "fid", Name: "foo", Type: model.Float}, nil)
	if err != nil {
		t.Fatal(err)
	}

	*(set["fid"][0]) = float64(4.2)

	if v, ok := (*out).(float64); !ok {
		t.Fatal()
	} else if v != float64(4.2) {
		t.Fatal(v)
	}
}

func TestSimpleBoolSkeletonContent(t *testing.T) {
	t.Parallel()
	out, set, err := ContentToSkeleton(model.ContentVariable{CharacteristicId: "fid", Name: "foo", Type: model.Boolean}, nil)
	if err != nil {
		t.Fatal(err)
	}

	*(set["fid"][0]) = true

	if v, ok := (*out).(bool); !ok {
		t.Fatal()
	} else if v != true {
		t.Fatal(v)
	}
}

func TestSimpleStructSkeletonContent(t *testing.T) {
	t.Parallel()
	out, _, err := ContentToSkeleton(model.ContentVariable{CharacteristicId: "foo", Type: model.Structure}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := (*out).(map[string]*interface{}); !ok {
		t.Fatal(*out)
	}
}

func TestStructSkeletonContent(t *testing.T) {
	t.Parallel()
	out, set, err := ContentToSkeleton(model.ContentVariable{
		CharacteristicId: "rdf",
		Name:             "rdf",
		Type:             model.Structure,
		SubContentVariables: []model.ContentVariable{
			{CharacteristicId: "r", Name: "r", Type: model.Integer},
			{CharacteristicId: "g", Name: "g", Type: model.Integer},
			{CharacteristicId: "b", Name: "b", Type: model.Integer},
		},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	*(set["r"][0]) = int64(255)
	*(set["g"][0]) = int64(100)
	*(set["b"][0]) = int64(0)

	if m, ok := (*out).(map[string]*interface{}); !ok {
		t.Fatal()
	} else {
		if n, ok := (*m["r"]).(int64); !ok || n != int64(255) {
			t.Fatal()
		}
		if n, ok := (*m["g"]).(int64); !ok || n != int64(100) {
			t.Fatal()
		}
		if n, ok := (*m["b"]).(int64); !ok || n != int64(0) {
			t.Fatal()
		}
	}

	msg1, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	msg2, err := json.Marshal(map[string]interface{}{
		"r": 255,
		"g": 100,
		"b": 0,
	})
	if err != nil {
		t.Fatal(err)
	}

	var a interface{}
	err = json.Unmarshal(msg1, &a)
	if err != nil {
		t.Fatal(err)
	}

	var b interface{}
	err = json.Unmarshal(msg2, &b)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(a, b) {
		t.Fatal(string(msg1), string(msg2), a, b)
	}
}

func TestStructDefaultSkeletonContent(t *testing.T) {
	t.Parallel()
	out, set, err := ContentToSkeleton(model.ContentVariable{
		CharacteristicId: "rdf",
		Name:             "rdf",
		Type:             model.Structure,
		SubContentVariables: []model.ContentVariable{
			{CharacteristicId: "r", Name: "r", Type: model.Integer, Value: float64(255)}, // value as float because json transforms every number to float64 if target field is interface{}
			{CharacteristicId: "g", Name: "g", Type: model.Integer, Value: float64(255)},
			{CharacteristicId: "b", Name: "b", Type: model.Integer, Value: float64(255)},
		},
	}, nil)
	if err != nil {
		t.Fatal(err)
	}

	*(set["g"][0]) = float64(100)

	if m, ok := (*out).(map[string]*interface{}); !ok {
		t.Fatal()
	} else {
		if n, ok := (*m["r"]).(float64); !ok || n != float64(255) {
			t.Fatal()
		}
		if n, ok := (*m["g"]).(float64); !ok || n != float64(100) {
			t.Fatal()
		}
		if n, ok := (*m["b"]).(float64); !ok || n != float64(255) {
			t.Fatal()
		}
	}

	msg1, err := json.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}

	msg2, err := json.Marshal(map[string]interface{}{
		"r": 255,
		"g": 100,
		"b": 255,
	})
	if err != nil {
		t.Fatal(err)
	}

	var a interface{}
	err = json.Unmarshal(msg1, &a)
	if err != nil {
		t.Fatal(err)
	}

	var b interface{}
	err = json.Unmarshal(msg2, &b)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(a, b) {
		t.Fatal(string(msg1), string(msg2), a, b)
	}
}

func TestSimpleListSkeletonContent(t *testing.T) {
	t.Parallel()
	out, _, err := ContentToSkeleton(model.ContentVariable{Name: "foo", Type: model.List}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := (*out).([]*interface{}); !ok {
		t.Fatal()
	}
}
