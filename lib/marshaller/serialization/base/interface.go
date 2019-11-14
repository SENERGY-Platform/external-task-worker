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

package base

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"sync"
)

type Marshaller interface {
	Marshal(in interface{}, variable model.ContentVariable) (out string, err error)
	Unmarshal(in string, variable model.ContentVariable) (out interface{}, err error)
}

var Marshallers = map[string]Marshaller{}

var mux = sync.Mutex{}

func Register(key string, marshaller Marshaller) {
	mux.Lock()
	defer mux.Unlock()
	Marshallers[key] = marshaller
}

func Get(key string) (marshaller Marshaller, ok bool) {
	marshaller, ok = Marshallers[key]
	return
}
