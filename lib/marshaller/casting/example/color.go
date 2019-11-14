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

var characteristicToConcept = &base.CastCharacteristicToConcept{}
var conceptToCharacteristic = &base.CastConceptToCharacteristic{}

const Color = "example_color"

func init() {
	base.Concepts[Color] = base.GetConceptCastFunction(characteristicToConcept, conceptToCharacteristic)
	base.ConceptRepo.Register(model.Concept{Id: Color, Name: "example"}, []model.Characteristic{
		{
			Id:   Rgb,
			Name: "rgb",
			Type: model.Structure,
			SubCharacteristics: []model.Characteristic{
				{Id: Rgb + ".r", Name: "r", Type: model.Integer},
				{Id: Rgb + ".g", Name: "g", Type: model.Integer},
				{Id: Rgb + ".b", Name: "b", Type: model.Integer},
			},
		},
		{
			Id:   Hex,
			Name: "hex",
			Type: model.String,
		},
	})
}
