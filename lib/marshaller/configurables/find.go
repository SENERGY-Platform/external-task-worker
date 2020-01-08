/*
 * Copyright 2020 InfAI (CC SES)
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

package configurables

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/mapping"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"strings"
)

func FindConfigurables(characteristicId string, services []model.Service) (result Configurables, err error) {
	return FindIntersectingConfigurables(characteristicId, services)
}

func FindIntersectingConfigurables(notCharacteristicId string, services []model.Service) (result Configurables, err error) {
	notConcept, err := base.ConceptRepo.GetConceptOfCharacteristic(notCharacteristicId)
	if err != nil {
		return nil, err
	}
	invertedIndex := map[string]map[string]bool{}
	for _, service := range services {
		characteristics := []string{}
		for _, content := range service.Inputs {
			characteristics = append(characteristics, characteristicsInContentVariable(content.ContentVariable)...)
		}
		for _, characteristic := range characteristics {
			concept, err := base.ConceptRepo.GetConceptOfCharacteristic(characteristic)
			if err != nil {
				return nil, err
			}
			if _, ok := invertedIndex[concept]; !ok {
				invertedIndex[concept] = map[string]bool{}
			}
			invertedIndex[concept][service.Id] = true
		}
	}
	serviceCount := len(services)
	for conceptId, servicesUsingConcept := range invertedIndex {
		if conceptId != notConcept && len(servicesUsingConcept) == serviceCount {
			configurable, err := createConfigurable(conceptId)
			if err != nil {
				return nil, err
			}
			result = append(result, configurable)
		}
	}
	return result, nil
}

func createConfigurable(conceptId string) (result Configurable, err error) {
	concept, err := base.ConceptRepo.GetConcept(conceptId)
	if err != nil {
		return result, err
	}
	if len(concept.CharacteristicIds) == 0 {
		return result, errors.New("expect at least one characteristic for concept " + concept.Name + " " + conceptId)
	}
	characteristicId := concept.BaseCharacteristicId
	if characteristicId == "" {
		characteristicId = concept.CharacteristicIds[0]
	}
	characteristic, err := base.ConceptRepo.GetCharacteristic(characteristicId)
	if err != nil {
		return result, err
	}
	return Configurable{
		CharacteristicId: concept.BaseCharacteristicId,
		Values:           createConfigurableValues(characteristic),
	}, nil
}

func createConfigurableValues(characteristic model.Characteristic, pathSegments ...string) (result []ConfigurableCharacteristicValue) {
	switch characteristic.Type {
	case model.Integer, model.Float:
		return []ConfigurableCharacteristicValue{
			{
				Path:      strings.Join(pathSegments, "."),
				Value:     0,
				ValueType: characteristic.Type,
			},
		}
	case model.Boolean:
		return []ConfigurableCharacteristicValue{
			{
				Path:      strings.Join(pathSegments, "."),
				Value:     false,
				ValueType: characteristic.Type,
			},
		}
	case model.String:
		return []ConfigurableCharacteristicValue{
			{
				Path:      strings.Join(pathSegments, "."),
				Value:     "",
				ValueType: characteristic.Type,
			},
		}
	case model.List:
		for _, sub := range characteristic.SubCharacteristics {
			index := "0"
			if sub.Name != mapping.VAR_LEN_PLACEHOLDER {
				index = sub.Name
			}
			result = append(result, createConfigurableValues(sub, append(pathSegments, index)...)...)
		}
		return result
	case model.Structure:
		for _, sub := range characteristic.SubCharacteristics {
			result = append(result, createConfigurableValues(sub, append(pathSegments, sub.Name)...)...)
		}
		return result
	default:
		return
	}
}

func characteristicsInContentVariable(variable model.ContentVariable) (result []string) {
	if variable.CharacteristicId != "" {
		result = append(result, variable.CharacteristicId)
	}
	for _, sub := range variable.SubContentVariables {
		result = append(result, characteristicsInContentVariable(sub)...)
	}
	return result
}
