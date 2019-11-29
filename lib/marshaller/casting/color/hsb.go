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

package color

import (
	"errors"
	"github.com/lucasb-eyer/go-colorful"
	"log"
	"runtime/debug"
)

const Hsb = "urn:infai:ses:characteristic:64928e9f-98ca-42bb-a1e5-adf2a760a2f9"
const HsbH = "urn:infai:ses:characteristic:6ec70e99-8c6a-4909-8d5a-7cc12af76b9a"
const HsbS = "urn:infai:ses:characteristic:a66dc568-c0e0-420f-b513-18e8df405538"
const HsbB = "urn:infai:ses:characteristic:d840607c-c8f9-45d6-b9bd-2c2d444e2899"

func init() {
	conceptToCharacteristic.Set(Hsb, func(concept interface{}) (out interface{}, err error) {
		hexStr, ok := concept.(string)
		if !ok {
			debug.PrintStack()
			return nil, errors.New("unable to interpret value as string")
		}
		hex, err := colorful.Hex(hexStr)
		if err != nil {
			debug.PrintStack()
			return nil, err
		}
		h, s, v := hex.Hsv()
		return map[string]int64{"h": int64(h), "s": int64(s * 100), "b": int64(v * 100)}, nil //TODO
	})

	characteristicToConcept.Set(Hsb, func(in interface{}) (concept interface{}, err error) {
		hsvMap, ok := in.(map[string]interface{})
		if !ok {
			log.Println(in)
			debug.PrintStack()
			return nil, errors.New("unable to interpret value as map[string]interface{}")
		}
		h, ok := hsvMap["h"]
		if !ok {
			debug.PrintStack()
			return nil, errors.New("missing field h")
		}
		hue, ok := h.(float64)
		if !ok {
			debug.PrintStack()
			return nil, errors.New("field h is not a number")
		}
		s, ok := hsvMap["s"]
		if !ok {
			debug.PrintStack()
			return nil, errors.New("missing field s")
		}
		saturation, ok := s.(float64)
		if !ok {
			debug.PrintStack()
			return nil, errors.New("field s is not a number")
		}
		v, ok := hsvMap["b"] //hsb vs hsv
		if !ok {
			debug.PrintStack()
			return nil, errors.New("missing field b")
		}
		value, ok := v.(float64)
		if !ok {
			debug.PrintStack()
			return nil, errors.New("field b is not a number")
		}
		hsv := colorful.Hsv(hue, saturation/100, value/100)
		return hsv.Hex(), nil
	})
}
