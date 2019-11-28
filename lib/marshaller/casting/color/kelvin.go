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

package example

const Kelvin = "urn:infai:ses:characteristic:7c0b7bf3-7039-4bf2-8e9e-703d19bec0ed"

func init() {
	conceptToCharacteristic.Set(Kelvin, func(concept interface{}) (out interface{}, err error) {
		return 2700, nil //TODO
	})

	characteristicToConcept.Set(Kelvin, func(in interface{}) (concept interface{}, err error) {
		return "#32a852", nil //TODO
	})
}
