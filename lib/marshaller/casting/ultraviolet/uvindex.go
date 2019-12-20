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

package ultraviolet

const UvIndexId = "urn:infai:ses:characteristic:0a61343d-c0d1-4af8-9329-3829c30ba59f"
const UvIndexName = "uv index"

func init() {
	conceptToCharacteristic.Set(UvIndexId, func(concept interface{}) (out interface{}, err error) {
		return concept, nil
	})

	characteristicToConcept.Set(UvIndexId, func(in interface{}) (concept interface{}, err error) {
		return in, nil
	})
}
