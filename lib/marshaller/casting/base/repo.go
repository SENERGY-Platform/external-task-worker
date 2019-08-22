package base

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"runtime/debug"
	"sync"
)

type ConceptRepoType struct {
	concepts                           map[string]model.Concept
	characteristics                    map[string]model.Characteristic
	conceptByCharacteristic            map[string]model.Concept
	rootCharacteristicByCharacteristic map[string]model.Characteristic
	mux                                sync.Mutex
}

var ConceptRepo = &ConceptRepoType{
	concepts:                           map[string]model.Concept{model.NullConcept.Id: model.NullConcept},
	characteristics:                    map[string]model.Characteristic{model.NullCharacteristic.Id: model.NullCharacteristic},
	conceptByCharacteristic:            map[string]model.Concept{},
	rootCharacteristicByCharacteristic: map[string]model.Characteristic{},
}

func (this *ConceptRepoType) Register(concept model.Concept) {
	this.mux.Lock()
	defer this.mux.Unlock()
	this.concepts[concept.Id] = concept
	for _, characteristic := range concept.Characteristics {
		this.characteristics[characteristic.Id] = characteristic
		this.conceptByCharacteristic[characteristic.Id] = concept
		for _, descendent := range getCharacteristicDescendents(characteristic) {
			this.rootCharacteristicByCharacteristic[descendent.Id] = characteristic
		}
	}
}

func getCharacteristicDescendents(characteristic model.Characteristic) (result []model.Characteristic) {
	result = []model.Characteristic{characteristic}
	for _, child := range characteristic.SubCharacteristics {
		result = append(result, getCharacteristicDescendents(child)...)
	}
	return result
}

func (this *ConceptRepoType) GetConceptOfCharacteristic(characteristicId string) (conceptId string, err error) {
	this.mux.Lock()
	defer this.mux.Unlock()
	concept, ok := this.conceptByCharacteristic[characteristicId]
	if !ok {
		debug.PrintStack()
		return conceptId, errors.New("no concept found for characteristic id " + characteristicId)
	}
	return concept.Id, nil
}

func (this *ConceptRepoType) GetCharacteristic(id string) (characteristic model.Characteristic, err error) {
	if id == "" {
		return model.NullCharacteristic, nil
	}
	this.mux.Lock()
	defer this.mux.Unlock()
	characteristic, ok := this.characteristics[id]
	if !ok {
		debug.PrintStack()
		return characteristic, errors.New("no characteristic found for id " + id)
	}
	return characteristic, nil
}

func (this *ConceptRepoType) GetRootCharacteristics(ids []string) (result []string) {
	this.mux.Lock()
	defer this.mux.Unlock()
	for _, id := range ids {
		root, ok := this.rootCharacteristicByCharacteristic[id]
		if ok {
			result = append(result, root.Id)
		}
	}
	return
}
