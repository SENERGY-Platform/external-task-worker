package marshaller

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting"
	castingbase "github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/mapping"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	marshalling "github.com/SENERGY-Platform/external-task-worker/lib/marshaller/serialization/base"
	"runtime/debug"
)

func UnmarshalOutputs(protocol model.Protocol, service model.Service, output map[string]string, outputCharacteristicId CharacteristicId) (result interface{}, err error) {
	return UnmarshalOutputsWithRepo(castingbase.ConceptRepo, protocol, service, output, outputCharacteristicId)
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
					marshaller, ok := marshalling.Get(content.Serialization)
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
