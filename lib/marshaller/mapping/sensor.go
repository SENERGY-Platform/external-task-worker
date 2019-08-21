package mapping

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"runtime/debug"
	"strconv"
)

func MapSensors(in map[string]interface{}, contents map[string]model.ContentVariable, category model.Characteristic) (out interface{}, err error) {
	temp, set, err := CharacteristicToSkeleton(category)
	if err != nil {
		return nil, err
	}
	for key, value := range in {
		content, ok := contents[key]
		if ok {
			content, err = completeContentVariableCharacteristicId(content, category)
			if err != nil {
				return nil, err
			}
			err = castToCategory(value, content, set, createCharacteristicIndex(&map[string]model.Characteristic{}, category))
			if err != nil {
				return out, err
			}
		}
	}
	out = *temp
	return
}

func MapSensor(in interface{}, content model.ContentVariable, category model.Characteristic) (out interface{}, err error) {
	content, err = completeContentVariableCharacteristicId(content, category)
	if err != nil {
		return nil, err
	}
	temp, set, err := CharacteristicToSkeleton(category)
	if err != nil {
		return nil, err
	}
	err = castToCategory(in, content, set, createCharacteristicIndex(&map[string]model.Characteristic{}, category))
	out = *temp
	return
}

func completeContentVariableCharacteristicId(variable model.ContentVariable, characteristic model.Characteristic) (model.ContentVariable, error) {
	var err error
	if (variable.Type == model.Structure || variable.Type == model.List) && len(variable.SubContentVariables) == 1 && variable.SubContentVariables[0].Name == "*" {
		if variable.CharacteristicId == "" && variable.SubContentVariables[0].CharacteristicId == "" {
			err = errors.New("expect exact_match set in " + variable.Name + " " + variable.Id + " or " + variable.SubContentVariables[0].Name + " " + variable.SubContentVariables[0].Id)
		} else if variable.CharacteristicId == "" {
			variable.CharacteristicId, err = getContentVariableParentCharacteristicId(variable.SubContentVariables[0], characteristic)
		} else if variable.SubContentVariables[0].CharacteristicId == "" {
			sub := variable.SubContentVariables[0]
			sub.CharacteristicId, err = getContentVariableChildCharacteristicId(variable, characteristic)
			variable.SubContentVariables[0] = sub
		}
	}
	if err != nil {
		return variable, err
	}
	for index, child := range variable.SubContentVariables {
		variable.SubContentVariables[index], err = completeContentVariableCharacteristicId(child, characteristic)
		if err != nil {
			return variable, err
		}
	}
	return variable, err
}

func getContentVariableChildCharacteristicId(parent model.ContentVariable, characteristic model.Characteristic) (string, error) {
	if parent.CharacteristicId == "" {
		return "", errors.New("expect exact_match set in " + parent.Name + " " + parent.Id + " or its child")
	}
	if parent.CharacteristicId == characteristic.Id {
		if len(characteristic.SubCharacteristics) != 1 || characteristic.SubCharacteristics[0].Name != "*" {
			return "", errors.New(characteristic.Name + " " + characteristic.Id + " is used as variable length characteristic without being one")
		}
		return characteristic.SubCharacteristics[0].Id, nil
	}
	for _, sub := range characteristic.SubCharacteristics {
		result, err := getContentVariableChildCharacteristicId(parent, sub)
		if err != nil {
			return result, err
		}
		if result != "" {
			return result, nil
		}
	}
	return "", nil
}

func getContentVariableParentCharacteristicId(child model.ContentVariable, characteristic model.Characteristic) (string, error) {
	if child.CharacteristicId == "" {
		return "", errors.New("expect exact_match set in " + child.Name + " " + child.Id + " or its parent")
	}
	parent := characteristic
	for _, cchild := range characteristic.SubCharacteristics {
		if cchild.Id == child.CharacteristicId {
			return parent.Id, nil
		}
	}
	for _, cchild := range characteristic.SubCharacteristics {
		match, err := getContentVariableParentCharacteristicId(child, cchild)
		if err != nil {
			return "", err
		}
		if match != "" {
			return match, nil
		}
	}
	return "", nil
}

func createCharacteristicIndex(in *map[string]model.Characteristic, characteristic model.Characteristic) map[string]model.Characteristic {
	(*in)[characteristic.Id] = characteristic
	for _, sub := range characteristic.SubCharacteristics {
		createCharacteristicIndex(in, sub)
	}
	return *in
}

func castToCategory(in interface{}, variable model.ContentVariable, set map[string]*interface{}, characteristics map[string]model.Characteristic) error {
	switch variable.Type {
	case model.String, model.Integer, model.Float, model.Boolean:
		ref, ok := set[variable.CharacteristicId]
		if ok {
			*ref = in
		} else {
			//return errors.New("unable to find target exact_match '" + variable.CharacteristicId + "' in setter")
		}
	case model.Structure:
		m, ok := in.(map[string]interface{})
		if !ok {
			debug.PrintStack()
			return errors.New("variable '" + variable.Name + "' is not map/structure")
		}
		if len(variable.SubContentVariables) == 1 && variable.SubContentVariables[0].Name == VAR_LEN_PLACEHOLDER && variable.CharacteristicId != "" {
			//as map
			category, ok := characteristics[variable.SubContentVariables[0].CharacteristicId]
			if !ok {
				return errors.New("unable to find characteristic '" + variable.SubContentVariables[0].CharacteristicId + "' (maps need exact match references on the list and element variable)")
			}
			temp := map[string]interface{}{}
			for key, sub := range m {
				out, err := MapSensor(sub, variable.SubContentVariables[0], category)
				if err != nil {
					return err
				}
				temp[key] = out
			}
			ref, ok := set[variable.CharacteristicId]
			if ok {
				*ref = temp
			} else {
				//return errors.New("unable to find target exact_match '" + variable.CharacteristicId + "' in setter")
			}
		} else {
			//as structure
			for _, s := range variable.SubContentVariables {
				sub, ok := m[s.Name]
				if ok {
					err := castToCategory(sub, s, set, characteristics)
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
		if len(variable.SubContentVariables) == 1 && variable.SubContentVariables[0].Name == VAR_LEN_PLACEHOLDER && variable.CharacteristicId != "" {
			//as map
			category, ok := characteristics[variable.SubContentVariables[0].CharacteristicId]
			if !ok {
				return errors.New("unable to find characteristic '" + variable.SubContentVariables[0].CharacteristicId + "' (maps need exact match references on the list and element variable)")
			}
			temp := []interface{}{}
			for _, sub := range l {
				out, err := MapSensor(sub, variable.SubContentVariables[0], category)
				if err != nil {
					return err
				}
				temp = append(temp, out)
			}
			ref, ok := set[variable.CharacteristicId]
			if ok {
				*ref = temp
			} else {
				//return errors.New("unable to find target exact_match '" + variable.CharacteristicId + "' in setter")
			}
		} else {
			//as structure
			for _, s := range variable.SubContentVariables {
				index, err := strconv.Atoi(s.Name)
				if err != nil {
					if s.Name == VAR_LEN_PLACEHOLDER && len(variable.SubContentVariables) == 1 {
						return errors.New("expect used exact_match in ContentVariable " + variable.Name + " " + variable.Id)
					}
					return errors.New("unable to interpret '" + s.Name + "' as list index")
				}
				if index < len(l) {
					sub := l[index]
					if ok {
						err := castToCategory(sub, s, set, characteristics)
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
