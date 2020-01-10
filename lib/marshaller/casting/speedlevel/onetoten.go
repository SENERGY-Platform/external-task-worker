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

package speedlevel

import (
	"errors"
	"runtime/debug"
)

const OneToTenId = "urn:infai:ses:characteristic:ac7ad69f-b82e-454a-8b44-5c0440cba787"
const OneToTenName = "1 to 10"

func init() {
	conceptToCharacteristic.Set(OneToTenId, func(concept interface{}) (out interface{}, err error) {
		var percent int64
		switch v := concept.(type) {
		case int64:
			percent = v
		case int:
			percent = int64(v)
		case float64:
			percent = int64(v)
		case float32:
			percent = int64(v)
		default:
			debug.PrintStack()
			return nil, errors.New("unable to interpret concept value")
		}
		result := int64(percent / 10)
		if result < 1 {
			result = 1
		}
		return result, nil
	})

	characteristicToConcept.Set(OneToTenId, func(in interface{}) (concept interface{}, err error) {
		var number int64
		switch v := in.(type) {
		case int64:
			number = v
		case int:
			number = int64(v)
		case float64:
			number = int64(v)
		case float32:
			number = int64(v)
		default:
			debug.PrintStack()
			return nil, errors.New("unable to interpret value as characteristic " + OneToTenName)
		}
		result := number * 10
		return result, nil
	})
}
