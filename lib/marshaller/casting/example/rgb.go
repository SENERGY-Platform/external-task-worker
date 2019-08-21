package example

import (
	"errors"
	"gopkg.in/go-playground/colors.v1"
	"log"
	"runtime/debug"
)

const Rgb = "example_rgb"

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
		return rgb.ToHEX().String(), nil
	})
}
