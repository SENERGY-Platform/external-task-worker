package marshaller

import (
	"encoding/json"
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting"
	castingBase "github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/mapping"
	marshalling "github.com/SENERGY-Platform/external-task-worker/lib/marshaller/marshalling/base"
	_ "github.com/SENERGY-Platform/external-task-worker/lib/marshaller/marshalling/json"
	_ "github.com/SENERGY-Platform/external-task-worker/lib/marshaller/marshalling/xml"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"runtime/debug"
)

type ConceptRepo interface {
	GetConceptOfCharacteristic(characteristicId string) (conceptId string, err error)
	GetCharacteristic(id CharacteristicId) (model.Characteristic, error)
	GetRootCharacteristics(ids []CharacteristicId) (result []CharacteristicId)
}

type CharacteristicId = string
type ConceptId = string

func MarshalInputs(protocol model.Protocol, service model.Service, input interface{}, inputCharacteristicId CharacteristicId) (result string, err error) {
	return MarshalInputsWithRepo(castingBase.ConceptRepo, protocol, service, input, inputCharacteristicId)
}

func MarshalInputsWithRepo(conceptRepo ConceptRepo, protocol model.Protocol, service model.Service, input interface{}, inputCharacteristicId CharacteristicId) (result string, err error) {
	inputCharacteristic, err := conceptRepo.GetCharacteristic(inputCharacteristicId)
	if err != nil {
		return result, err
	}
	resultObj := map[string]string{}
	for _, content := range service.Inputs {
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
				resultObj[segment.Name] = resultPart
			}
		}
	}
	resultBuffer, err := json.Marshal(resultObj)
	return string(resultBuffer), err
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

func MarshalInput(inputCharacteristicValue interface{}, conceptId string, inputCharacteristic model.Characteristic, serviceCharacteristic model.Characteristic, serviceVariable model.ContentVariable, serialization string) (result string, err error) {
	serviceCharacteristicValue, err := casting.Cast(inputCharacteristicValue, conceptId, inputCharacteristic.Id, serviceCharacteristic.Id)
	if err != nil {
		return result, err
	}

	//normalize
	temp, err := json.Marshal(serviceCharacteristicValue)
	if err != nil {
		debug.PrintStack()
		return result, err
	}
	var normalized interface{}
	err = json.Unmarshal(temp, &normalized)
	if err != nil {
		debug.PrintStack()
		return result, err
	}

	serviceVariableValue, err := mapping.MapActuator(normalized, serviceCharacteristic, serviceVariable)
	if err != nil {
		return result, err
	}

	marshaller, ok := marshalling.Get(serialization)
	if !ok {
		return result, errors.New("unknown serialization " + serialization)
	}
	result, err = marshaller.Marshal(serviceVariableValue, serviceVariable)
	return result, err
}

func UnmarshalOutput(serviceCharacteristicValueString string, conceptId string, outputCharacteristic model.Characteristic, serviceCharacteristic model.Characteristic, serviceVariable model.ContentVariable, serialization string) (result interface{}, err error) {
	marshaller, ok := marshalling.Get(serialization)
	if !ok {
		return result, errors.New("unknown serialization " + serialization)
	}
	serviceVariableValue, err := marshaller.Unmarshal(serviceCharacteristicValueString, serviceVariable)
	if err != nil {
		return result, err
	}

	serviceCharacteristicValue, err := mapping.MapSensor(serviceVariableValue, serviceVariable, serviceCharacteristic)
	if err != nil {
		return result, err
	}

	result, err = casting.Cast(serviceCharacteristicValue, conceptId, serviceCharacteristic.Id, outputCharacteristic.Id)
	return
}
