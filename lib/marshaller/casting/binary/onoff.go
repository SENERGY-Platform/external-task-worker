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

const OnOff = "urn:infai:ses:characteristic:7621686a-56bc-402d-b4cc-5b266d39736f"

func init() {
	conceptToCharacteristic.Set(OnOff, func(concept interface{}) (out interface{}, err error) {
		b, ok := concept.(bool)
		if !ok {
			debug.PrintStack()
			log.Println("ERROR: ", reflect.TypeOf(concept).String(), concept)
			return nil, errors.New("unable to interpret value as boolean; input type is " + reflect.TypeOf(concept).String())
		}
		if b {
			return "on", nil
		} else {
			return "off", nil
		}
	})

	characteristicToConcept.Set(OnOff, func(in interface{}) (concept interface{}, err error) {
		switch v := in.(type) {
		case string:
			if v == "on" {
				return true, nil
			} else {
				return false, nil
			}
		default:
			debug.PrintStack()
			return nil, errors.New("unable to interpret value as string; input type is " + reflect.TypeOf(in).String())
		}
	})
}
