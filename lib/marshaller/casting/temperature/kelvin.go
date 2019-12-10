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

package temperature

import (
	"errors"
	"log"
	"reflect"
	"runtime/debug"
)

const Kelvin = "urn:infai:ses:characteristic:75b2d113-1d03-4ef8-977a-8dbcbb31a683"

func init() {
	conceptToCharacteristic.Set(Kelvin, func(concept interface{}) (out interface{}, err error) {
		switch celcius := concept.(type) {
		case int:
			return celcius + 273, nil
		case int32:
			return celcius + 273, nil
		case int64:
			return celcius + 273, nil
		case float32:
			return celcius + 273.15, nil
		case float64:
			return celcius + 273.15, nil
		default:
			debug.PrintStack()
			log.Println("ERROR: ", reflect.TypeOf(concept).String(), concept)
			return nil, errors.New("unable to interpret value; input type is " + reflect.TypeOf(concept).String())
		}
	})

	characteristicToConcept.Set(Kelvin, func(in interface{}) (concept interface{}, err error) {
		switch kelvin := in.(type) {
		case int:
			return kelvin - 273, nil
		case int32:
			return kelvin - 273, nil
		case int64:
			return kelvin - 273, nil
		case float32:
			return kelvin - 273.15, nil
		case float64:
			return kelvin - 273.15, nil
		default:
			debug.PrintStack()
			log.Println("ERROR: ", reflect.TypeOf(in).String(), in)
			return nil, errors.New("unable to interpret value; input type is " + reflect.TypeOf(in).String())
		}
	})
}
