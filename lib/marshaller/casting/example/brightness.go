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

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
)

var brightnessToConcept = &base.CastCharacteristicToConcept{}
var brightnessToCharacteristic = &base.CastConceptToCharacteristic{}

const Brightness = "example_brightness"
const Lux = "example_lux"

func init() {
	brightnessToCharacteristic.Set(Lux, func(concept interface{}) (out interface{}, err error) {
		return concept, nil
	})

	brightnessToConcept.Set(Lux, func(in interface{}) (concept interface{}, err error) {
		return in, nil
	})

	base.Concepts[Brightness] = base.GetConceptCastFunction(brightnessToConcept, brightnessToCharacteristic)
	base.ConceptRepo.Register(model.Concept{Id: Brightness, Name: "example-bri"}, []model.Characteristic{
		{
			Id:   Lux,
			Name: "lux",
			Type: model.Integer,
		},
	})
}
