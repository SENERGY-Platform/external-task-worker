package example

import (
	"errors"
	"gopkg.in/go-playground/colors.v1"
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
		return map[string]int64{"r": int64(rgb.R), "g": int64(rgb.R), "b": int64(rgb.R)}, nil
	})

	characteristicToConcept.Set(Rgb, func(in interface{}) (concept interface{}, err error) {
		rgbMap, ok := in.(map[string]int64)
		if !ok {
			debug.PrintStack()
			return nil, errors.New("unable to interpret value as map[string]int64")
		}
		r, ok := rgbMap["r"]
		if !ok {
			debug.PrintStack()
			return nil, errors.New("missing field r")
		}
		g, ok := rgbMap["g"]
		if !ok {
			debug.PrintStack()
			return nil, errors.New("missing field g")
		}
		b, ok := rgbMap["b"]
		if !ok {
			debug.PrintStack()
			return nil, errors.New("missing field b")
		}
		rgb, err := colors.RGB(uint8(r), uint8(g), uint8(b))
		if err != nil {
			debug.PrintStack()
			return nil, err
		}
		return rgb.ToHEX().String(), nil
	})
}
