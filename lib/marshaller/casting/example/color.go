package example

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
)

var characteristicToConcept = &base.CastCharacteristicToConcept{}
var conceptToCharacteristic = &base.CastConceptToCharacteristic{}

const Concept = "example"

func init() {
	base.Concepts[Concept] = base.GetConceptCastFunction(characteristicToConcept, conceptToCharacteristic)
}
