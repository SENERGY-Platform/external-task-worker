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
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
)

var characteristicToConcept = &base.CastCharacteristicToConcept{}
var conceptToCharacteristic = &base.CastConceptToCharacteristic{}

const Binary = "urn:infai:ses:concept:ebfeabb3-50f0-44bd-b06e-95eb52df484e"

func init() {
	base.Concepts[Binary] = base.GetConceptCastFunction(characteristicToConcept, conceptToCharacteristic)
	base.ConceptRepo.Register(model.Concept{Id: Binary, Name: "binary state"}, []model.Characteristic{
		{
			Id:   Boolean,
			Name: "boolean",
			Type: model.Boolean,
		},
		{
			Id:   BinaryStatusCode,
			Name: "binary status code",
			Type: model.Integer,
		},
		{
			Id:   OnOff,
			Name: "on/off",
			Type: model.String,
		},
	})
}
