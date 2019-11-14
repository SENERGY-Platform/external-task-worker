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

package xml

import (
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/serialization/base"
)

func ExampleMarshalPrimitiveInt() {
	value := 24
	marshaller, ok := base.Get(Format)
	if !ok {
		return
	}

	fmt.Println(marshaller.Marshal(value, model.ContentVariable{
		Name: "int",
	}))

	// Output:
	//<int>24</int> <nil>
}

func ExampleMarshalPrimitiveFloat() {
	value := 2.4
	marshaller, ok := base.Get(Format)
	if !ok {
		return
	}

	fmt.Println(marshaller.Marshal(value, model.ContentVariable{
		Name: "f",
	}))

	// Output:
	//<f>2.4</f> <nil>
}

func ExampleMarshalPrimitiveString() {
	value := "foo"
	marshaller, ok := base.Get(Format)
	if !ok {
		return
	}

	fmt.Println(marshaller.Marshal(value, model.ContentVariable{
		Name: "str",
	}))

	// Output:
	//<str>foo</str> <nil>
}

func ExampleMarshalPrimitiveBool() {
	value := true
	marshaller, ok := base.Get(Format)
	if !ok {
		return
	}

	fmt.Println(marshaller.Marshal(value, model.ContentVariable{
		Name: "b",
	}))

	// Output:
	//<b>true</b> <nil>
}

func ExampleMarshal() {
	value := map[string]interface{}{"-attr": "attrVal", "body": "bodyVal"}
	marshaller, ok := base.Get(Format)
	if !ok {
		return
	}

	fmt.Println(marshaller.Marshal(value, model.ContentVariable{
		Name: "example",
		Type: model.Structure,
		SubContentVariables: []model.ContentVariable{
			{Name: "-attr"},
			{Name: "body"},
		},
	}))

	// Output:
	//<example attr="attrVal"><body>bodyVal</body></example> <nil>
}

//TODO: lists
