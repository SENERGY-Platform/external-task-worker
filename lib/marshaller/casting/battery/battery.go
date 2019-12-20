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

package battery

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
)

var characteristicToConcept = &base.CastCharacteristicToConcept{}
var conceptToCharacteristic = &base.CastConceptToCharacteristic{}

const ConceptId = "urn:infai:ses:concept:f917e146-9192-4796-8bbe-6be14292346f"
const ConceptName = "battery"

func init() {
	base.Concepts[ConceptId] = base.GetConceptCastFunction(characteristicToConcept, conceptToCharacteristic)
	base.ConceptRepo.Register(model.Concept{Id: ConceptId, Name: ConceptName}, []model.Characteristic{
		{
			Id:   PercentageId,
			Name: PercentageName,
			Type: model.Float,
		},
	})
}
