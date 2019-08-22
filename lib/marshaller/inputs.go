package marshaller

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/mapping"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/serialization"
	"reflect"
	"runtime/debug"
)

type ConceptRepo interface {
	GetConceptOfCharacteristic(characteristicId string) (conceptId string, err error)
	GetCharacteristic(id CharacteristicId) (model.Characteristic, error)
	GetRootCharacteristics(ids []CharacteristicId) (result []CharacteristicId)
}

type CharacteristicId = string
type ConceptId = string

func MarshalInputs(protocol model.Protocol, service model.Service, input interface{}, inputCharacteristicId CharacteristicId) (result map[string]string, err error) {
	return MarshalInputsWithRepo(casting.ConceptRepo, protocol, service, input, inputCharacteristicId)
}

func MarshalInputsWithRepo(conceptRepo ConceptRepo, protocol model.Protocol, service model.Service, input interface{}, inputCharacteristicId CharacteristicId) (result map[string]string, err error) {
	inputCharacteristic, err := conceptRepo.GetCharacteristic(inputCharacteristicId)
	if err != nil {
		return result, err
	}
	result = map[string]string{}
	for _, content := range service.Inputs {
		if !reflect.DeepEqual(inputCharacteristic, model.NullCharacteristic) {
			conceptId, variableCharacteristicId, err := getMatchingVariableRootCharacteristic(conceptRepo, content.ContentVariable, inputCharacteristicId)
			if err != nil {
				return result, err
			}
			variableCharacteristic, err := conceptRepo.GetCharacteristic(variableCharacteristicId)
			if err != nil {
				return result, err
			}
			resultPart, err := MarshalInput(input, conceptId, inputCharacteristic, variableCharacteristic, content.ContentVariable, content.Serialization)
			if err != nil {
				return result, err
			}
			for _, segment := range protocol.ProtocolSegments {
				if segment.Id == content.ProtocolSegmentId {
					result[segment.Name] = resultPart
				}
			}
		} else {
			resultPart, err := MarshalInput(input, model.NullConcept.Id, inputCharacteristic, model.NullCharacteristic, content.ContentVariable, content.Serialization)
			if err != nil {
				return result, err
			}
			for _, segment := range protocol.ProtocolSegments {
				if segment.Id == content.ProtocolSegmentId {
					result[segment.Name] = resultPart
				}
			}
		}
	}

	return result, err
}

func getMatchingVariableRootCharacteristic(repo ConceptRepo, variable model.ContentVariable, matchingId CharacteristicId) (conceptId string, matchingVariableRootCharacteristic CharacteristicId, err error) {
	conceptId, err = repo.GetConceptOfCharacteristic(matchingId)
	if err != nil {
		return
	}
	variableCharacteristics := getVariableCharacteristics(variable)
	rootCharacteristics := repo.GetRootCharacteristics(variableCharacteristics)
	for _, candidate := range rootCharacteristics {
		conceptA, err := repo.GetConceptOfCharacteristic(candidate)
		if err != nil {
			return conceptId, matchingVariableRootCharacteristic, err
		}
		if conceptA == conceptId {
			return conceptId, candidate, nil
		}
	}
	return conceptId, matchingVariableRootCharacteristic, errors.New("no match found")
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

func MarshalInput(inputCharacteristicValue interface{}, conceptId string, inputCharacteristic model.Characteristic, serviceCharacteristic model.Characteristic, serviceVariable model.ContentVariable, serializationId string) (result string, err error) {
	serviceCharacteristicValue := inputCharacteristicValue
	serviceCharacteristicValue, err = casting.Cast(inputCharacteristicValue, conceptId, inputCharacteristic.Id, serviceCharacteristic.Id)
	if err != nil {
		return result, err
	}

	normalized, err := normalize(serviceCharacteristicValue)
	if err != nil {
		return result, err
	}

	serviceVariableValue, err := mapping.MapActuator(normalized, serviceCharacteristic, serviceVariable)
	if err != nil {
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
