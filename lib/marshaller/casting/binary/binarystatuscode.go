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
	"runtime/debug"
)

const BinaryStatusCode = "urn:infai:ses:characteristic:c0353532-a8fb-4553-a00b-418cb8a80a65"

//color concept uses hex -> do nothing
func init() {
	conceptToCharacteristic.Set(BinaryStatusCode, func(concept interface{}) (out interface{}, err error) {
		b, ok := concept.(bool)
		if !ok {
			debug.PrintStack()
			return nil, errors.New("unable to interpret value as string")
		}
		if b {
			return int(1), nil
		} else {
			return int(0), nil
		}
	})

	characteristicToConcept.Set(BinaryStatusCode, func(in interface{}) (concept interface{}, err error) {
		i, ok := in.(int)
		if !ok {
			debug.PrintStack()
			return nil, errors.New("unable to interpret value as string")
		}
		if i > 0 {
			return true, nil
		} else {
			return false, nil
		}
	})
}
