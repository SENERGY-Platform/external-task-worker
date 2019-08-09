package mapping

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"strconv"
)

func MapActuator(in interface{}, category model.Characteristic, content model.ContentVariable) (out interface{}, err error) {
	content, err = completeContentVariableExactMatch(content, category)
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
	(*in)[content.ExactMatch] = content
	for _, sub := range content.SubContentVariables {
		createContentIndex(in, sub)
	}
	return *in
}

func castToContent(in interface{}, variable model.Characteristic, set map[string]*interface{}, content map[string]model.ContentVariable) error {
	switch variable.ValueType {
	case model.String, model.Integer, model.Float, model.Boolean:
		ref, ok := set[variable.Id]
		if ok {
			*ref = in
		} else {
			return errors.New("unable to find target exact_match '" + variable.Id + "' in setter")
		}
	case model.Structure:
		m, ok := in.(map[string]interface{})
		if !ok {
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
			return errors.New("variable '" + variable.Name + "' is not map/structure")
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
