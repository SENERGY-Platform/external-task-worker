package example

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
)

var characteristicToConcept = &base.CastCharacteristicToConcept{}
var conceptToCharacteristic = &base.CastConceptToCharacteristic{}

const Color = "example_color"

func init() {
	base.Concepts[Color] = base.GetConceptCastFunction(characteristicToConcept, conceptToCharacteristic)
	base.ConceptRepo.Register(model.Concept{Id: Color, Name: "example"}, []model.Characteristic{
		{
			Id:   Rgb,
			Name: "rgb",
			Type: model.Structure,
			SubCharacteristics: []model.Characteristic{
				{Id: Rgb + ".r", Name: "r", Type: model.Integer},
				{Id: Rgb + ".g", Name: "g", Type: model.Integer},
				{Id: Rgb + ".b", Name: "b", Type: model.Integer},
			},
		},
		{
			Id:   Hex,
			Name: "hex",
			Type: model.String,
		},
	})
}
