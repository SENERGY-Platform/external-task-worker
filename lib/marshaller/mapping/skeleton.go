package mapping

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"strconv"
)

func CharacteristicToSkeleton(category model.Characteristic) (out *interface{}, idToPtr map[string]*interface{}, err error) {
	idToPtr = map[string]*interface{}{}
	var t interface{}
	out = &t
	err = categoryToSkeleton(category, out, &idToPtr)
	return
}

const VAR_LEN_PLACEHOLDER = "*"

func categoryToSkeleton(category model.Characteristic, out *interface{}, idToPtr *map[string]*interface{}) (err error) {
	switch category.Type {
	case model.Float, model.Integer:
		if category.Value != nil {
			var ok bool
			*out, ok = category.Value.(float64)
			if !ok {
				return errors.New("unable to interpret value in " + category.Id)
			}
		} else {
			*out = float64(0)
		}
		(*idToPtr)[category.Id] = out
	case model.String:
		if category.Value != nil {
			var ok bool
			*out, ok = category.Value.(string)
			if !ok {
				return errors.New("unable to interpret value in " + category.Id)
			}
		} else {
			*out = ""
		}
		(*idToPtr)[category.Id] = out
	case model.Boolean:
		if category.Value != nil {
			var ok bool
			*out, ok = category.Value.(bool)
			if !ok {
				return errors.New("unable to interpret value in " + category.Id)
			}
		} else {
			*out = false
		}
		(*idToPtr)[category.Id] = out
	case model.Structure:
		if len(category.SubCharacteristics) == 1 && category.SubCharacteristics[0].Name == VAR_LEN_PLACEHOLDER {
			*out = map[string]interface{}{}
			(*idToPtr)[category.Id] = out
		} else {
			*out = map[string]interface{}{}
			for _, sub := range category.SubCharacteristics {
				var subvar interface{}
				err = categoryToSkeleton(sub, &subvar, idToPtr)
				if err != nil {
					return err
				}
				(*out).(map[string]interface{})[sub.Name] = &subvar
			}
		}
	case model.List:
		if len(category.SubCharacteristics) == 1 && category.SubCharacteristics[0].Name == VAR_LEN_PLACEHOLDER {
			*out = []interface{}{}
			(*idToPtr)[category.Id] = out
		} else {
			*out = make([]interface{}, len(category.SubCharacteristics))
			for _, sub := range category.SubCharacteristics {
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
		return errors.New("unknown variable type: " + string(category.Type))
	}
	return nil
}

func ContentToSkeleton(content model.ContentVariable) (out *interface{}, idToPtr map[string]*interface{}, err error) {
	idToPtr = map[string]*interface{}{}
	var t interface{}
	out = &t
	err = contentToSkeleton(content, out, &idToPtr)
	return
}

func contentToSkeleton(content model.ContentVariable, out *interface{}, idToPtr *map[string]*interface{}) (err error) {
	switch content.Type {
	case model.Float, model.Integer:
		if content.Value != nil {
			var ok bool
			*out, ok = content.Value.(float64)
			if !ok {
				return errors.New("unable to interpret value in " + content.Id)
			}
		} else {
			*out = float64(0)
		}
		(*idToPtr)[content.CharacteristicId] = out
	case model.String:
		if content.Value != nil {
			var ok bool
			*out, ok = content.Value.(string)
			if !ok {
				return errors.New("unable to interpret value in " + content.Id)
			}
		} else {
			*out = ""
		}
		(*idToPtr)[content.CharacteristicId] = out
	case model.Boolean:
		if content.Value != nil {
			var ok bool
			*out, ok = content.Value.(bool)
			if !ok {
				return errors.New("unable to interpret value in " + content.Id)
			}
		} else {
			*out = false
		}
		(*idToPtr)[content.CharacteristicId] = out
	case model.Structure:
		if len(content.SubContentVariables) == 1 && content.SubContentVariables[0].Name == VAR_LEN_PLACEHOLDER {
			*out = map[string]interface{}{}
			(*idToPtr)[content.CharacteristicId] = out
		} else {
			*out = map[string]interface{}{}
			for _, sub := range content.SubContentVariables {
				var subvar interface{}
				err = contentToSkeleton(sub, &subvar, idToPtr)
				if err != nil {
					return err
				}
				(*out).(map[string]interface{})[sub.Name] = &subvar
			}
		}
	case model.List:
		if len(content.SubContentVariables) == 1 && content.SubContentVariables[0].Name == VAR_LEN_PLACEHOLDER {
			*out = []interface{}{}
			(*idToPtr)[content.CharacteristicId] = out
		} else {
			*out = make([]interface{}, len(content.SubContentVariables))
			for _, sub := range content.SubContentVariables {
				var subvar interface{}
				err = contentToSkeleton(sub, &subvar, idToPtr)
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
		return errors.New("unknown variable type: " + string(content.Type))
	}
	return nil
}
