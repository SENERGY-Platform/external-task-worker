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

package marshaller

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/configurables"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/mapping"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/serialization"
	"log"
	"reflect"
	"runtime/debug"
	"strings"
)

type ConceptRepo interface {
	GetConceptOfCharacteristic(characteristicId string) (conceptId string, err error)
	GetCharacteristic(id CharacteristicId) (model.Characteristic, error)
	GetRootCharacteristics(ids []CharacteristicId) (result []CharacteristicId)
}

type CharacteristicId = string
type ConceptId = string

func MarshalInputs(protocol model.Protocol, service model.Service, input interface{}, inputCharacteristicId CharacteristicId, configurables ...configurables.Configurable) (result map[string]string, err error) {
	return MarshalInputsWithRepo(casting.ConceptRepo, protocol, service, input, inputCharacteristicId, configurables...)
}

func MarshalInputsWithRepo(conceptRepo ConceptRepo, protocol model.Protocol, service model.Service, input interface{}, inputCharacteristicId CharacteristicId, configurables ...configurables.Configurable) (result map[string]string, err error) {
	result = map[string]string{}
	for _, content := range service.Inputs {
		partial := mapping.NewPartial()
		var resultPart = ""
		for _, configurable := range configurables {
			characteristic, err := base.ConceptRepo.GetCharacteristic(configurable.CharacteristicId)
			if err != nil {
				return result, err
			}
			skeleton, setter, err := mapping.CharacteristicToSkeleton(characteristic)
			if err != nil {
				return result, err
			}
			assignConfigurableValues(setter, characteristic, configurable.Values)
			value, err := normalize(skeleton)
			if err != nil {
				return result, err
			}
			resultPart, err = partialInputMarshalling(conceptRepo, content.ContentVariable, partial, configurable.CharacteristicId, value, content.Serialization)
		}
		resultPart, err = partialInputMarshalling(conceptRepo, content.ContentVariable, partial, inputCharacteristicId, input, content.Serialization)
		if err != nil {
			return result, err
		}
		for _, segment := range protocol.ProtocolSegments {
			if segment.Id == content.ProtocolSegmentId {
				result[segment.Name] = resultPart
			}
		}
	}

	return result, err
}

func partialInputMarshalling(repo ConceptRepo, variable model.ContentVariable, partial mapping.Partial, inputCharacteristicId string, input interface{}, serializationId string) (result string, err error) {
	inputCharacteristic, err := repo.GetCharacteristic(inputCharacteristicId)
	if err != nil {
		return result, err
	}
	if !reflect.DeepEqual(inputCharacteristic, model.NullCharacteristic) {
		conceptId, variableCharacteristicIds, err := getMatchingVariableRootCharacteristic(repo, variable, inputCharacteristicId)
		if err != nil {
			return result, err
		}
		for _, variableCharacteristicId := range variableCharacteristicIds {
			variableCharacteristic, err := repo.GetCharacteristic(variableCharacteristicId)
			if err != nil {
				return result, err
			}
			result, err = MarshalInput(partial, input, conceptId, inputCharacteristic, variableCharacteristic, variable, serializationId)
			if err != nil {
				return result, err
			}
		}
	} else {
		result, err = MarshalInput(partial, input, model.NullConcept.Id, inputCharacteristic, model.NullCharacteristic, variable, serializationId)
	}
	return
}

func assignConfigurableValues(setter map[string]*interface{}, characteristic model.Characteristic, values []configurables.ConfigurableCharacteristicValue) {
	for _, value := range values {
		assignConfigurableValue(setter, characteristic, strings.Split(value.Path, "."), value.Value)
	}
}

func assignConfigurableValue(setter map[string]*interface{}, characteristic model.Characteristic, path []string, value interface{}) {
	if len(path) == 0 || (len(path) == 1 && path[0] == "") {
		set, ok := setter[characteristic.Id]
		if ok {
			*set = value
		}
	} else {
		next, rest := path[0], path[1:]
		for _, sub := range characteristic.SubCharacteristics {
			if sub.Name == next {
				assignConfigurableValue(setter, sub, rest, value)
			}
		}
	}
}

func getMatchingVariableRootCharacteristic(repo ConceptRepo, variable model.ContentVariable, matchingId CharacteristicId) (conceptId string, matchingVariableRootCharacteristic []CharacteristicId, err error) {
	conceptId, err = repo.GetConceptOfCharacteristic(matchingId)
	if err != nil {
		return
	}
	variableCharacteristics := getVariableCharacteristics(variable)
	rootCharacteristics := repo.GetRootCharacteristics(variableCharacteristics)
	resultSet := map[string]bool{}
	for _, candidate := range rootCharacteristics {
		conceptA, err := repo.GetConceptOfCharacteristic(candidate)
		if err != nil {
			return conceptId, matchingVariableRootCharacteristic, err
		}
		if conceptA == conceptId {
			resultSet[candidate] = true
		}
	}
	for characteristic, _ := range resultSet {
		matchingVariableRootCharacteristic = append(matchingVariableRootCharacteristic, characteristic)
	}
	if len(matchingVariableRootCharacteristic) == 0 {
		return conceptId, matchingVariableRootCharacteristic, errors.New("no match found between " + matchingId + " and characteristics of " + variable.Id + " (" + strings.Join(variableCharacteristics, ",") + ") => (" + strings.Join(rootCharacteristics, ",") + ")")
	}
	return conceptId, matchingVariableRootCharacteristic, nil
}

func getVariableCharacteristics(variable model.ContentVariable) (result []CharacteristicId) {
	if variable.CharacteristicId != "" {
		result = []CharacteristicId{variable.CharacteristicId}
	}
	for _, sub := range variable.SubContentVariables {
		result = append(result, getVariableCharacteristics(sub)...)
	}
	return result
}

func MarshalInput(partial mapping.Partial, inputCharacteristicValue interface{}, conceptId string, inputCharacteristic model.Characteristic, serviceCharacteristic model.Characteristic, serviceVariable model.ContentVariable, serializationId string) (result string, err error) {
	serviceCharacteristicValue := inputCharacteristicValue
	serviceCharacteristicValue, err = casting.Cast(inputCharacteristicValue, conceptId, inputCharacteristic.Id, serviceCharacteristic.Id)
	if err != nil {
		return result, err
	}

	normalized, err := normalize(serviceCharacteristicValue)
	if err != nil {
		return result, err
	}

	serviceVariableValue, err := mapping.MapActuator(normalized, serviceCharacteristic, serviceVariable, partial)
	if err != nil {
		log.Println("ERROR: unable to map actuator", serviceCharacteristic.Id, serviceCharacteristic.Value, "-->", serviceVariable.Id, serviceVariable.Name)
		return result, err
	}

	marshaller, ok := serialization.Get(serializationId)
	if !ok {
		return result, errors.New("unknown serialization " + serializationId)
	}

	normalized, err = normalize(serviceVariableValue)
	if err != nil {
		return result, err
	}

	result, err = marshaller.Marshal(normalized, serviceVariable)
	return result, err
}

func normalize(value interface{}) (result interface{}, err error) {
	temp, err := json.Marshal(value)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	err = json.Unmarshal(temp, &result)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	return
}
