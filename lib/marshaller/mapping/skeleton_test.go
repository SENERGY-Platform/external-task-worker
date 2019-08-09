package mapping

import (
	"encoding/json"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"reflect"
	"testing"
)

func TestSimpleStringSkeleton(t *testing.T) {
	t.Parallel()
	out, set, err := CharacteristicToSkeleton(model.Characteristic{Id: "fid", Name: "foo", ValueType: model.String})
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
	out, set, err := CharacteristicToSkeleton(model.Characteristic{Id: "fid", Name: "foo", ValueType: model.Integer})
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
	out, set, err := CharacteristicToSkeleton(model.Characteristic{Id: "fid", Name: "foo", ValueType: model.Float})
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
	out, set, err := CharacteristicToSkeleton(model.Characteristic{Id: "fid", Name: "foo", ValueType: model.Boolean})
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
	out, _, err := CharacteristicToSkeleton(model.Characteristic{Name: "foo", ValueType: model.Structure})
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
		Id:        "rdf",
		Name:      "rdf",
		ValueType: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{Id: "r", Name: "r", ValueType: model.Integer},
			{Id: "g", Name: "g", ValueType: model.Integer},
			{Id: "b", Name: "b", ValueType: model.Integer},
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
		Id:        "rdf",
		Name:      "rdf",
		ValueType: model.Structure,
		SubCharacteristics: []model.Characteristic{
			{Id: "r", Name: "r", ValueType: model.Integer, Value: float64(255)}, // value as float because json transforms every number to float64 if target field is interface{}
			{Id: "g", Name: "g", ValueType: model.Integer, Value: float64(255)},
			{Id: "b", Name: "b", ValueType: model.Integer, Value: float64(255)},
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
	out, _, err := CharacteristicToSkeleton(model.Characteristic{Name: "foo", ValueType: model.List})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := (*out).([]interface{}); !ok {
		t.Fatal()
	}
}

func TestSimpleStringSkeletonContent(t *testing.T) {
	t.Parallel()
	out, set, err := ContentToSkeleton(model.ContentVariable{ExactMatch: "fid", Name: "foo", ValueType: model.String})
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

func TestSimpleIntSkeletonContent(t *testing.T) {
	t.Parallel()
	out, set, err := ContentToSkeleton(model.ContentVariable{ExactMatch: "fid", Name: "foo", ValueType: model.Integer})
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

func TestSimpleFloatSkeletonContent(t *testing.T) {
	t.Parallel()
	out, set, err := ContentToSkeleton(model.ContentVariable{ExactMatch: "fid", Name: "foo", ValueType: model.Float})
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

func TestSimpleBoolSkeletonContent(t *testing.T) {
	t.Parallel()
	out, set, err := ContentToSkeleton(model.ContentVariable{ExactMatch: "fid", Name: "foo", ValueType: model.Boolean})
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

func TestSimpleStructSkeletonContent(t *testing.T) {
	t.Parallel()
	out, _, err := ContentToSkeleton(model.ContentVariable{ExactMatch: "foo", ValueType: model.Structure})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := (*out).(map[string]interface{}); !ok {
		t.Fatal()
	}
}

func TestStructSkeletonContent(t *testing.T) {
	t.Parallel()
	out, set, err := ContentToSkeleton(model.ContentVariable{
		ExactMatch: "rdf",
		Name:       "rdf",
		ValueType:  model.Structure,
		SubContentVariables: []model.ContentVariable{
			{ExactMatch: "r", Name: "r", ValueType: model.Integer},
			{ExactMatch: "g", Name: "g", ValueType: model.Integer},
			{ExactMatch: "b", Name: "b", ValueType: model.Integer},
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

func TestStructDefaultSkeletonContent(t *testing.T) {
	t.Parallel()
	out, set, err := ContentToSkeleton(model.ContentVariable{
		ExactMatch: "rdf",
		Name:       "rdf",
		ValueType:  model.Structure,
		SubContentVariables: []model.ContentVariable{
			{ExactMatch: "r", Name: "r", ValueType: model.Integer, Value: float64(255)}, // value as float because json transforms every number to float64 if target field is interface{}
			{ExactMatch: "g", Name: "g", ValueType: model.Integer, Value: float64(255)},
			{ExactMatch: "b", Name: "b", ValueType: model.Integer, Value: float64(255)},
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

func TestSimpleListSkeletonContent(t *testing.T) {
	t.Parallel()
	out, _, err := ContentToSkeleton(model.ContentVariable{Name: "foo", ValueType: model.List})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := (*out).([]interface{}); !ok {
		t.Fatal()
	}
}
