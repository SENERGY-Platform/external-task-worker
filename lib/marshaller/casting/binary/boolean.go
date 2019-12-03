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

package binary

import (
	"errors"
	"log"
	"reflect"
	"runtime/debug"
)

const Boolean = "urn:infai:ses:characteristic:7dc1bb7e-b256-408a-a6f9-044dc60fdcf5"

func init() {
	conceptToCharacteristic.Set(Boolean, func(concept interface{}) (out interface{}, err error) {
		b, ok := concept.(bool)
		if !ok {
			debug.PrintStack()
			log.Println("ERROR: ", reflect.TypeOf(concept).String(), concept)
			return nil, errors.New("unable to interpret value as boolean; input type is " + reflect.TypeOf(concept).String())
		}
		return b, nil
	})

	characteristicToConcept.Set(Boolean, func(in interface{}) (concept interface{}, err error) {
		b, ok := in.(bool)
		if !ok {
			debug.PrintStack()
			log.Println("ERROR: ", reflect.TypeOf(in).String(), in)
			return nil, errors.New("unable to interpret value as boolean; input type is " + reflect.TypeOf(in).String())
		}
		return b, nil
	})
}
