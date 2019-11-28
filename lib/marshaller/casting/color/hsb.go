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

const Hsb = "urn:infai:ses:characteristic:64928e9f-98ca-42bb-a1e5-adf2a760a2f9"
const HsbH = "urn:infai:ses:characteristic:6ec70e99-8c6a-4909-8d5a-7cc12af76b9a"
const HsbS = "urn:infai:ses:characteristic:a66dc568-c0e0-420f-b513-18e8df405538"
const HsbB = "urn:infai:ses:characteristic:d840607c-c8f9-45d6-b9bd-2c2d444e2899"

func init() {
	conceptToCharacteristic.Set(Hsb, func(concept interface{}) (out interface{}, err error) {
		return map[string]int64{"h": int64(100), "s": int64(100), "b": int64(190)}, nil //TODO
	})

	characteristicToConcept.Set(Hsb, func(in interface{}) (concept interface{}, err error) {
		return "#32a852", nil //TODO
	})
}
