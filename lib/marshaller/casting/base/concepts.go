package base

import (
	"errors"
)

var Concepts = map[string]FindCastFromCharacteristicToConceptFunction{}

type CastFromConceptToCharacteristicFunction func(characteristicId string) (out interface{}, err error)

type CastFromCharacteristicToConceptFunction func(in interface{}) CastFromConceptToCharacteristicFunction

type FindCastFromCharacteristicToConceptFunction = func(characteristicId string) CastFromCharacteristicToConceptFunction

func GetErrorCastConceptToCharacteristicFunction(err error) CastFromConceptToCharacteristicFunction {
	return func(conceptId string) (interface{}, error) {
		return nil, err
	}
}

func GetErrorCastFromCharacteristicToConceptFunction(err error) CastFromCharacteristicToConceptFunction {
	return func(in interface{}) CastFromConceptToCharacteristicFunction {
		return GetErrorCastConceptToCharacteristicFunction(err)
	}
}

func GetErrorFindCastFromCharacteristicToConceptFunction(err error) FindCastFromCharacteristicToConceptFunction {
	return func(characteristicId string) CastFromCharacteristicToConceptFunction {
		return GetErrorCastFromCharacteristicToConceptFunction(err)
	}
}

type CastCharacteristicToConceptFunction func(in interface{}) (concept interface{}, err error)
type CastConceptToCharacteristicFunction func(concept interface{}) (out interface{}, err error)

type CastCharacteristicToConcept struct {
	values map[string]CastCharacteristicToConceptFunction
}

func (this *CastCharacteristicToConcept) Set(key string, value CastCharacteristicToConceptFunction) {
	if this.values == nil {
		this.values = map[string]CastCharacteristicToConceptFunction{}
	}
	this.values[key] = value
}

func (this *CastCharacteristicToConcept) Get(key string) (value CastCharacteristicToConceptFunction, ok bool) {
	if this.values == nil {
		this.values = map[string]CastCharacteristicToConceptFunction{}
	}
	value, ok = this.values[key]
	return
}

type CastConceptToCharacteristic struct {
	values map[string]CastConceptToCharacteristicFunction
}

func (this *CastConceptToCharacteristic) Set(key string, value CastConceptToCharacteristicFunction) {
	if this.values == nil {
		this.values = map[string]CastConceptToCharacteristicFunction{}
	}
	this.values[key] = value
}

func (this *CastConceptToCharacteristic) Get(key string) (value CastConceptToCharacteristicFunction, ok bool) {
	if this.values == nil {
		this.values = map[string]CastConceptToCharacteristicFunction{}
	}
	value, ok = this.values[key]
	return
}

func GetConceptCastFunction(characteristicToConcept *CastCharacteristicToConcept, conceptToCharacteristic *CastConceptToCharacteristic) FindCastFromCharacteristicToConceptFunction {
	return func(characteristicId string) CastFromCharacteristicToConceptFunction {
		castToConcept, ok := characteristicToConcept.Get(characteristicId)
		if !ok {
			return GetErrorCastFromCharacteristicToConceptFunction(errors.New("unknown characteristic " + characteristicId + " in concept color"))
		}
		return func(in interface{}) CastFromConceptToCharacteristicFunction {
			concept, err := castToConcept(in)
			if err != nil {
				return GetErrorCastConceptToCharacteristicFunction(err)
			}
			return func(characteristicId string) (out interface{}, err error) {
				castToCharacteristic, ok := conceptToCharacteristic.Get(characteristicId)
				if !ok {
					return nil, errors.New("unknown characteristic " + characteristicId + " in concept color")
				}
				return castToCharacteristic(concept)
			}
		}
	}
}
