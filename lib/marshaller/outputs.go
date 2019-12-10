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
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/mapping"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/serialization"
	"runtime/debug"
)

func UnmarshalOutputs(protocol model.Protocol, service model.Service, output map[string]string, outputCharacteristicId CharacteristicId) (result interface{}, err error) {
	if outputCharacteristicId == "" {
		return nil, nil
	}
	if len(output) == 0 {
		return nil, nil
	}
	return UnmarshalOutputsWithRepo(casting.ConceptRepo, protocol, service, output, outputCharacteristicId)
}

func UnmarshalOutputsWithRepo(conceptRepo ConceptRepo, protocol model.Protocol, service model.Service, outputMap map[string]string, outputCharacteristicId CharacteristicId) (result interface{}, err error) {
	outputObjectMap, err := serializeOutput(outputMap, service, protocol)
	if err != nil {
		return result, err
	}
	contentMap := map[string]model.ContentVariable{}
	for _, content := range service.Outputs {
		for _, segment := range protocol.ProtocolSegments {
			if segment.Id == content.ProtocolSegmentId {
				contentMap[segment.Name] = content.ContentVariable
			}
		}
	}

	matchingServiceCharacteristicId, conceptId, err := getMatchingOutputRootCharacteristic(conceptRepo, service.Outputs, outputCharacteristicId)
	if err != nil {
		return result, err
	}

	serviceCharacteristic, err := conceptRepo.GetCharacteristic(matchingServiceCharacteristicId)
	if err != nil {
		return result, err
	}

	serviceCharacteristicValue, err := mapping.MapSensors(outputObjectMap, contentMap, serviceCharacteristic)
	if err != nil {
		return result, err
	}

	normalized, err := normalize(serviceCharacteristicValue)

	result, err = casting.Cast(normalized, conceptId, serviceCharacteristic.Id, outputCharacteristicId)
	return
}

func getMatchingOutputRootCharacteristic(repo ConceptRepo, contents []model.Content, matchingId CharacteristicId) (matchingServiceCharacteristicId CharacteristicId, conceptId string, err error) {
	conceptId, err = repo.GetConceptOfCharacteristic(matchingId)
	if err != nil {
		return
	}
	for _, content := range contents {
		variableCharacteristics := getVariableCharacteristics(content.ContentVariable)
		rootCharacteristics := repo.GetRootCharacteristics(variableCharacteristics)
		for _, candidate := range rootCharacteristics {
			candidateConcept, err := repo.GetConceptOfCharacteristic(candidate)
			if err != nil {
				return matchingServiceCharacteristicId, conceptId, err
			}
			if candidateConcept == conceptId {
				return candidate, conceptId, err
			}
		}
	}
	return matchingServiceCharacteristicId, conceptId, errors.New("no match found")
}

func serializeOutput(output map[string]string, service model.Service, protocol model.Protocol) (result map[string]interface{}, err error) {
	result = map[string]interface{}{}
	for _, content := range service.Outputs {
		for _, segment := range protocol.ProtocolSegments {
			if segment.Id == content.ProtocolSegmentId {
				output, ok := output[segment.Name]
				if ok {
					marshaller, ok := serialization.Get(content.Serialization)
					if !ok {
						debug.PrintStack()
						return result, errors.New("unknown serialization " + content.Serialization)
					}
					value, err := marshaller.Unmarshal(output, content.ContentVariable)
					if err != nil {
						return result, err
					}
					result[segment.Name] = value
				}
			}
		}
	}
	return
}
