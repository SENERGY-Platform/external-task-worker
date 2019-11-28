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
	"log"
	"runtime/debug"
	"strconv"
)

func MapActuator(in interface{}, category model.Characteristic, content model.ContentVariable) (out interface{}, err error) {
	content, err = completeContentVariableCharacteristicId(content, category)
	if err != nil {
		return nil, err
	}
	temp, set, err := ContentToSkeleton(content)
	if err != nil {
		return nil, err
	}
	err = castToContent(in, category, set, createContentIndex(&map[string]model.ContentVariable{}, content))
	out = *temp
	return

}

func createContentIndex(in *map[string]model.ContentVariable, content model.ContentVariable) map[string]model.ContentVariable {
	(*in)[content.CharacteristicId] = content
	for _, sub := range content.SubContentVariables {
		createContentIndex(in, sub)
	}
	return *in
}

func castToContent(in interface{}, variable model.Characteristic, set map[string]*interface{}, content map[string]model.ContentVariable) error {
	switch variable.Type {
	case model.String, model.Integer, model.Float, model.Boolean:
		ref, ok := set[variable.Id]
		if ok {
			*ref = in
		} else {
			debug.PrintStack()
			return errors.New("unable to find target exact_match '" + variable.Id + "' in setter")
		}
	case model.Structure:
		m, ok := in.(map[string]interface{})
		if !ok {
			debug.PrintStack()
			log.Println(in)
			return errors.New("variable '" + variable.Name + "' is not map/structure")
		}
		if len(variable.SubCharacteristics) == 1 && variable.SubCharacteristics[0].Name == VAR_LEN_PLACEHOLDER && variable.Id != "" {
			//as map
			category, ok := content[variable.SubCharacteristics[0].Id]
			if !ok {
				return errors.New("unable to find characteristic '" + variable.SubCharacteristics[0].Id + "' (maps need exact match references on the list and element variable)")
			}
			temp := map[string]interface{}{}
			for key, sub := range m {
				out, err := MapActuator(sub, variable.SubCharacteristics[0], category)
				if err != nil {
					return err
				}
				temp[key] = out
			}
			ref, ok := set[variable.Id]
			if ok {
				*ref = temp
			} else {
				debug.PrintStack()
				return errors.New("unable to find target exact_match '" + variable.Id + "' in setter")
			}
		} else {
			//as structure
			for _, s := range variable.SubCharacteristics {
				sub, ok := m[s.Name]
				if ok {
					err := castToContent(sub, s, set, content)
					if err != nil {
						return err
					}
				}
			}
		}
	case model.List:
		l, ok := in.([]interface{})
		if !ok {
			return errors.New("variable '" + variable.Name + "' is not list")
		}
		if len(variable.SubCharacteristics) == 1 && variable.SubCharacteristics[0].Name == VAR_LEN_PLACEHOLDER && variable.Id != "" {
			//as map
			category, ok := content[variable.SubCharacteristics[0].Id]
			if !ok {
				return errors.New("unable to find characteristic '" + variable.SubCharacteristics[0].Id + "' (maps need exact match references on the list and element variable)")
			}
			temp := []interface{}{}
			for _, sub := range l {
				out, err := MapActuator(sub, variable.SubCharacteristics[0], category)
				if err != nil {
					return err
				}
				temp = append(temp, out)
			}
			ref, ok := set[variable.Id]
			if ok {
				*ref = temp
			} else {
				debug.PrintStack()
				return errors.New("unable to find target exact_match '" + variable.Id + "' in setter")
			}
		} else {
			//as structure
			for _, s := range variable.SubCharacteristics {
				index, err := strconv.Atoi(s.Name)
				if err != nil {
					if s.Name == VAR_LEN_PLACEHOLDER && len(variable.SubCharacteristics) == 1 {
						return errors.New("expect used exact_match in ContentVariable " + variable.Name + " " + variable.Id)
					}
					return errors.New("unable to interpret '" + s.Name + "' as list index")
				}
				if index < len(l) {
					sub := l[index]
					if ok {
						err := castToContent(sub, s, set, content)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}
