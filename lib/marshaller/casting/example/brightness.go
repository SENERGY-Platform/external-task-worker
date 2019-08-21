package example

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
)

var brightnessToConcept = &base.CastCharacteristicToConcept{}
var brightnessToCharacteristic = &base.CastConceptToCharacteristic{}

const Brightness = "example_brightness"
const Lux = "example_lux"

func init() {
	brightnessToCharacteristic.Set(Lux, func(concept interface{}) (out interface{}, err error) {
		return concept, nil
	})

	brightnessToConcept.Set(Lux, func(in interface{}) (concept interface{}, err error) {
		return in, nil
	})

	base.Concepts[Brightness] = base.GetConceptCastFunction(brightnessToConcept, brightnessToCharacteristic)
	base.ConceptRepo.Register(model.Concept{Id: Brightness, Name: "example-bri", Characteristics: []model.Characteristic{
		{
			Id:   Lux,
			Name: "lux",
			Type: model.Integer,
		},
	}})
}
