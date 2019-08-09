package mapping

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"strconv"
)

func MapSensor(in interface{}, content model.ContentVariable, category model.Characteristic) (out interface{}, err error) {
	content, err = completeContentVariableExactMatch(content, category)
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

func completeContentVariableExactMatch(variable model.ContentVariable, characteristic model.Characteristic) (model.ContentVariable, error) {
	var err error
	if (variable.ValueType == model.Structure || variable.ValueType == model.List) && len(variable.SubContentVariables) == 1 && variable.SubContentVariables[0].Name == "*" {
		if variable.ExactMatch == "" && variable.SubContentVariables[0].ExactMatch == "" {
			err = errors.New("expect exact_match set in " + variable.Name + " " + variable.Id + " or " + variable.SubContentVariables[0].Name + " " + variable.SubContentVariables[0].Id)
		} else if variable.ExactMatch == "" {
			variable.ExactMatch, err = getContentVariableParentExactMatch(variable.SubContentVariables[0], characteristic)
		} else if variable.SubContentVariables[0].ExactMatch == "" {
			sub := variable.SubContentVariables[0]
			sub.ExactMatch, err = getContentVariableChildExactMatch(variable, characteristic)
			variable.SubContentVariables[0] = sub
		}
	}
	if err != nil {
		return variable, err
	}
	for index, child := range variable.SubContentVariables {
		variable.SubContentVariables[index], err = completeContentVariableExactMatch(child, characteristic)
		if err != nil {
			return variable, err
		}
	}
	return variable, err
}

func getContentVariableChildExactMatch(parent model.ContentVariable, characteristic model.Characteristic) (string, error) {
	if parent.ExactMatch == "" {
		return "", errors.New("expect exact_match set in " + parent.Name + " " + parent.Id + " or its child")
	}
	if parent.ExactMatch == characteristic.Id {
		if len(characteristic.SubCharacteristics) != 1 || characteristic.SubCharacteristics[0].Name != "*" {
			return "", errors.New(characteristic.Name + " " + characteristic.Id + " is used as variable length characteristic without being one")
		}
		return characteristic.SubCharacteristics[0].Id, nil
	}
	for _, sub := range characteristic.SubCharacteristics {
		result, err := getContentVariableChildExactMatch(parent, sub)
		if err != nil {
			return result, err
		}
		if result != "" {
			return result, nil
		}
	}
	return "", nil
}

func getContentVariableParentExactMatch(child model.ContentVariable, characteristic model.Characteristic) (string, error) {
	if child.ExactMatch == "" {
		return "", errors.New("expect exact_match set in " + child.Name + " " + child.Id + " or its parent")
	}
	parent := characteristic
	for _, cchild := range characteristic.SubCharacteristics {
		if cchild.Id == child.ExactMatch {
			return parent.Id, nil
		}
	}
	for _, cchild := range characteristic.SubCharacteristics {
		match, err := getContentVariableParentExactMatch(child, cchild)
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
	switch variable.ValueType {
	case model.String, model.Integer, model.Float, model.Boolean:
		ref, ok := set[variable.ExactMatch]
		if ok {
			*ref = in
		} else {
			return errors.New("unable to find target exact_match '" + variable.ExactMatch + "' in setter")
		}
	case model.Structure:
		m, ok := in.(map[string]interface{})
		if !ok {
			return errors.New("variable '" + variable.Name + "' is not map/structure")
		}
		if len(variable.SubContentVariables) == 1 && variable.SubContentVariables[0].Name == VAR_LEN_PLACEHOLDER && variable.ExactMatch != "" {
			//as map
			category, ok := characteristics[variable.SubContentVariables[0].ExactMatch]
			if !ok {
				return errors.New("unable to find characteristic '" + variable.SubContentVariables[0].ExactMatch + "' (maps need exact match references on the list and element variable)")
			}
			temp := map[string]interface{}{}
			for key, sub := range m {
				out, err := MapSensor(sub, variable.SubContentVariables[0], category)
				if err != nil {
					return err
				}
				temp[key] = out
			}
			ref, ok := set[variable.ExactMatch]
			if ok {
				*ref = temp
			} else {
				return errors.New("unable to find target exact_match '" + variable.ExactMatch + "' in setter")
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
			return errors.New("variable '" + variable.Name + "' is not map/structure")
		}
		if len(variable.SubContentVariables) == 1 && variable.SubContentVariables[0].Name == VAR_LEN_PLACEHOLDER && variable.ExactMatch != "" {
			//as map
			category, ok := characteristics[variable.SubContentVariables[0].ExactMatch]
			if !ok {
				return errors.New("unable to find characteristic '" + variable.SubContentVariables[0].ExactMatch + "' (maps need exact match references on the list and element variable)")
			}
			temp := []interface{}{}
			for _, sub := range l {
				out, err := MapSensor(sub, variable.SubContentVariables[0], category)
				if err != nil {
					return err
				}
				temp = append(temp, out)
			}
			ref, ok := set[variable.ExactMatch]
			if ok {
				*ref = temp
			} else {
				return errors.New("unable to find target exact_match '" + variable.ExactMatch + "' in setter")
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
