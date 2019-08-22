package casting

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
	_ "github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/example"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"runtime/debug"
)

var ConceptRepo = base.ConceptRepo

func Cast(in interface{}, conceptId string, from string, to string) (out interface{}, err error) {
	if from == model.NullCharacteristic.Id || to == model.NullCharacteristic.Id || conceptId == model.NullConcept.Id {
		return in, nil
	}
	return Concepts(conceptId)(from)(in)(to)
}

func Concepts(conceptId string) base.FindCastFromCharacteristicToConceptFunction {
	result, ok := base.Concepts[conceptId]
	if !ok {
		debug.PrintStack()
		return base.GetErrorFindCastFromCharacteristicToConceptFunction(errors.New("concept not found"))
	}
	return result
}
