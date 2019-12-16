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
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/base"
	"gopkg.in/go-playground/colors.v1"
	"log"
	"runtime/debug"
)

const Rgb = "urn:infai:ses:characteristic:5b4eea52-e8e5-4e80-9455-0382f81a1b43"

func init() {
	conceptToCharacteristic.Set(Rgb, func(concept interface{}) (out interface{}, err error) {
		hexStr, ok := concept.(string)
		if !ok {
			debug.PrintStack()
			return nil, errors.New("unable to interpret value as string")
		}
		hex, err := colors.ParseHEX(hexStr)
		if err != nil {
			debug.PrintStack()
			return nil, err
		}
		rgb := hex.ToRGB()
		if base.DEBUG {
			log.Println("hex to rgb:", hexStr, map[string]int64{"r": int64(rgb.R), "g": int64(rgb.G), "b": int64(rgb.B)})
		}
		return map[string]int64{"r": int64(rgb.R), "g": int64(rgb.G), "b": int64(rgb.B)}, nil
	})

	characteristicToConcept.Set(Rgb, func(in interface{}) (concept interface{}, err error) {
		rgbMap, ok := in.(map[string]interface{})
		if !ok {
			log.Println(in)
			debug.PrintStack()
			return nil, errors.New("unable to interpret value as map[string]interface{}")
		}
		r, ok := rgbMap["r"]
		if !ok {
			debug.PrintStack()
			return nil, errors.New("missing field r")
		}
		red, ok := r.(float64)
		if !ok {
			debug.PrintStack()
			return nil, errors.New("field r is not a number")
		}
		g, ok := rgbMap["g"]
		if !ok {
			debug.PrintStack()
			return nil, errors.New("missing field g")
		}
		green, ok := g.(float64)
		if !ok {
			debug.PrintStack()
			return nil, errors.New("field g is not a number")
		}
		b, ok := rgbMap["b"]
		if !ok {
			debug.PrintStack()
			return nil, errors.New("missing field b")
		}
		blue, ok := b.(float64)
		if !ok {
			debug.PrintStack()
			return nil, errors.New("field b is not a number")
		}
		rgb, err := colors.RGB(uint8(red), uint8(green), uint8(blue))
		if err != nil {
			debug.PrintStack()
			return nil, err
		}
		if base.DEBUG {
			log.Println("hsb to hex:", red, green, blue, rgb.ToHEX().String())
		}
		return rgb.ToHEX().String(), nil
	})
}
