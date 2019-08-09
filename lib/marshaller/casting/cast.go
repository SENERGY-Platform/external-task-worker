package casting

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
	_ "github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/example"
)

func Cast(in interface{}, conceptId string, from string, to string) (out interface{}, err error) {
	return Concepts(conceptId)(from)(in)(to)
}

func Concepts(conceptId string) base.FindCastFromCharacteristicToConceptFunction {
	result, ok := base.Concepts[conceptId]
	if !ok {
		return base.GetErrorFindCastFromCharacteristicToConceptFunction(errors.New("concept not found"))
	}
	return result
}
