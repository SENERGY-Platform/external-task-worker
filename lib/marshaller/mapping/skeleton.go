/*
 * Copyright 2019 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mapping

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"reflect"
	"strconv"
)

func CharacteristicToSkeleton(characteristic model.Characteristic) (out *interface{}, idToPtr map[string]*interface{}, err error) {
	idToPtr = map[string]*interface{}{}
	var t interface{}
	out = &t
	err = categoryToSkeleton(characteristic, out, &idToPtr)
	return
}

const VAR_LEN_PLACEHOLDER = "*"

func categoryToSkeleton(characteristic model.Characteristic, out *interface{}, idToPtr *map[string]*interface{}) (err error) {
	switch characteristic.Type {
	case model.Float, model.Integer:
		if characteristic.Value != nil {
			var ok bool
			*out, ok = characteristic.Value.(float64)
			if !ok {
				return errors.New("unable to interpret value in " + characteristic.Id)
			}
		} else {
			*out = float64(0)
		}
		(*idToPtr)[characteristic.Id] = out
	case model.String:
		if characteristic.Value != nil {
			var ok bool
			*out, ok = characteristic.Value.(string)
			if !ok {
				return errors.New("unable to interpret value in " + characteristic.Id)
			}
		} else {
			*out = ""
		}
		(*idToPtr)[characteristic.Id] = out
	case model.Boolean:
		if characteristic.Value != nil {
			var ok bool
			*out, ok = characteristic.Value.(bool)
			if !ok {
				return errors.New("unable to interpret value in " + characteristic.Id)
			}
		} else {
			*out = false
		}
		(*idToPtr)[characteristic.Id] = out
	case model.Structure:
		if len(characteristic.SubCharacteristics) == 1 && characteristic.SubCharacteristics[0].Name == VAR_LEN_PLACEHOLDER {
			*out = map[string]interface{}{}
			(*idToPtr)[characteristic.Id] = out
		} else {
			*out = map[string]interface{}{}
			for _, sub := range characteristic.SubCharacteristics {
				var subvar interface{}
				err = categoryToSkeleton(sub, &subvar, idToPtr)
				if err != nil {
					return err
				}
				(*out).(map[string]interface{})[sub.Name] = &subvar
			}
		}
	case model.List:
		if len(characteristic.SubCharacteristics) == 1 && characteristic.SubCharacteristics[0].Name == VAR_LEN_PLACEHOLDER {
			*out = []interface{}{}
			(*idToPtr)[characteristic.Id] = out
		} else {
			*out = make([]interface{}, len(characteristic.SubCharacteristics))
			for _, sub := range characteristic.SubCharacteristics {
				var subvar interface{}
				err = categoryToSkeleton(sub, &subvar, idToPtr)
				if err != nil {
					return err
				}
				index, err := strconv.Atoi(sub.Name)
				if err != nil {
					return errors.New("unable to interpret '" + sub.Name + "' as index for list: " + err.Error())
				}
				(*out).([]interface{})[index] = &subvar
			}
		}
	default:
		return errors.New("unknown variable type: " + string(characteristic.Type))
	}
	return nil
}

func ContentToSkeleton(content model.ContentVariable, partial Partial) (out *interface{}, idToPtr map[string][]*interface{}, err error) {
	idToPtr = map[string][]*interface{}{}
	if partial == nil {
		partial = NewPartial()
	}
	err = contentToSkeleton(content, partial, &idToPtr)
	out = partial.Value
	return
}

func contentToSkeleton(content model.ContentVariable, partial Partial, idToPtr *map[string][]*interface{}) (err error) {
	if partial == nil {
		partial = NewPartial()
	}
	var temp interface{}
	switch content.Type {
	case model.Float, model.Integer:
		if partial.Value == nil {
			if content.Value != nil {
				var ok bool
				temp, ok = content.Value.(float64)
				if !ok {
					return errors.New("unable to interpret value in " + content.Id)
				}
			} else {
				temp = float64(0)
			}
			partial.Value = &temp
		}
		(*idToPtr)[content.CharacteristicId] = append((*idToPtr)[content.CharacteristicId], partial.Value)
	case model.String:
		if partial.Value == nil {
			if content.Value != nil {
				var ok bool
				temp, ok = content.Value.(string)
				if !ok {
					return errors.New("unable to interpret value in " + content.Id)
				}
			} else {
				temp = ""
			}
			partial.Value = &temp
		}
		(*idToPtr)[content.CharacteristicId] = append((*idToPtr)[content.CharacteristicId], partial.Value)
	case model.Boolean:
		if partial.Value == nil || reflect.TypeOf(partial.Value).Kind() != reflect.Bool {
			if content.Value != nil {
				var ok bool
				temp, ok = content.Value.(bool)
				if !ok {
					return errors.New("unable to interpret value in " + content.Id)
				}
			} else {
				temp = false
			}
		}
		partial.Value = &temp
		(*idToPtr)[content.CharacteristicId] = append((*idToPtr)[content.CharacteristicId], partial.Value)
	case model.Structure:
		if len(content.SubContentVariables) == 1 && content.SubContentVariables[0].Name == VAR_LEN_PLACEHOLDER {
			if partial.Value == nil || *partial.Value == nil {
				temp = map[string]*interface{}{}
				partial.Value = &temp
			}
			(*idToPtr)[content.CharacteristicId] = append((*idToPtr)[content.CharacteristicId], partial.Value)
		} else {
			if partial.Value == nil {
				temp = map[string]*interface{}{}
				partial.Value = &temp
			}
			for _, sub := range content.SubContentVariables {
				subvar := NewPartial()
				ok := true
				subvar.Value, ok = (*partial.Value).(map[string]*interface{})[sub.Name]
				err = contentToSkeleton(sub, subvar, idToPtr)
				if err != nil {
					return err
				}
				if !ok {
					(*partial.Value).(map[string]*interface{})[sub.Name] = subvar.Value
				}
			}
		}
	case model.List:
		if len(content.SubContentVariables) == 1 && content.SubContentVariables[0].Name == VAR_LEN_PLACEHOLDER {
			if partial.Value == nil || *partial.Value == nil {
				temp = []*interface{}{}
				partial.Value = &temp
			}
			(*idToPtr)[content.CharacteristicId] = append((*idToPtr)[content.CharacteristicId], partial.Value)
		} else {
			if partial.Value == nil {
				temp = make([]*interface{}, len(content.SubContentVariables))
				partial.Value = &temp
			}
			for _, sub := range content.SubContentVariables {
				subvar := NewPartial()
				index, err := strconv.Atoi(sub.Name)
				if err != nil {
					return errors.New("unable to interpret '" + sub.Name + "' as index for list: " + err.Error())
				}
				subvar.Value = (*partial.Value).([]*interface{})[index]
				ok := subvar.Value != nil
				err = contentToSkeleton(sub, subvar, idToPtr)
				if err != nil {
					return err
				}
				if !ok {
					(*partial.Value).([]*interface{})[index] = subvar.Value
				}
			}
		}
	default:
		return errors.New("unknown variable type: " + string(content.Type))
	}
	return nil
}
