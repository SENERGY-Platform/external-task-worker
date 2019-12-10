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
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
)

var characteristicToConcept = &base.CastCharacteristicToConcept{}
var conceptToCharacteristic = &base.CastConceptToCharacteristic{}

const Temperature = "urn:infai:ses:concept:0bc81398-3ed6-4e2b-a6c4-b754583aac37"

func init() {
	base.Concepts[Temperature] = base.GetConceptCastFunction(characteristicToConcept, conceptToCharacteristic)
	base.ConceptRepo.Register(model.Concept{Id: Temperature, Name: "temperature"}, []model.Characteristic{
		{
			Id:   Celcius,
			Name: "celcius",
			Type: model.Float,
		},
		{
			Id:   Kelvin,
			Name: "kelvin",
			Type: model.Integer,
		},
	})
}
