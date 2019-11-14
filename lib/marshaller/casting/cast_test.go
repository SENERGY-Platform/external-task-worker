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

package casting

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/example"
	"testing"
)

func TestColorHexShortF(t *testing.T) {
	t.Parallel()
	out, err := Cast("#FFF", example.Color, example.Hex, example.Rgb)
	if err != nil {
		t.Fatal(err)
	}
	rgb, ok := out.(map[string]int64)
	if !ok {
		t.Fatal(out)
	}

	if rgb["r"] != 255 {
		t.Fatal(rgb)
	}

	if rgb["g"] != 255 {
		t.Fatal(rgb)
	}

	if rgb["b"] != 255 {
		t.Fatal(rgb)
	}
}

func TestColorHexF(t *testing.T) {
	t.Parallel()
	out, err := Cast("#FFFFFF", example.Color, example.Hex, example.Rgb)
	if err != nil {
		t.Fatal(err)
	}
	rgb, ok := out.(map[string]int64)
	if !ok {
		t.Fatal(out)
	}

	if rgb["r"] != 255 {
		t.Fatal(rgb)
	}

	if rgb["g"] != 255 {
		t.Fatal(rgb)
	}

	if rgb["b"] != 255 {
		t.Fatal(rgb)
	}
}

func TestColorHex0(t *testing.T) {
	t.Parallel()
	out, err := Cast("#000000", example.Color, example.Hex, example.Rgb)
	if err != nil {
		t.Fatal(err)
	}
	rgb, ok := out.(map[string]int64)
	if !ok {
		t.Fatal(out)
	}

	if rgb["r"] != 0 {
		t.Fatal(rgb)
	}

	if rgb["g"] != 0 {
		t.Fatal(rgb)
	}

	if rgb["b"] != 0 {
		t.Fatal(rgb)
	}
}

func TestColorHexShort0(t *testing.T) {
	t.Parallel()
	out, err := Cast("#000", example.Color, example.Hex, example.Rgb)
	if err != nil {
		t.Fatal(err)
	}
	rgb, ok := out.(map[string]int64)
	if !ok {
		t.Fatal(out)
	}

	if rgb["r"] != 0 {
		t.Fatal(rgb)
	}

	if rgb["g"] != 0 {
		t.Fatal(rgb)
	}

	if rgb["b"] != 0 {
		t.Fatal(rgb)
	}
}

func TestColorRgb255(t *testing.T) {
	t.Parallel()
	out, err := Cast(map[string]interface{}{"r": float64(255), "g": float64(255), "b": float64(255)}, example.Color, example.Rgb, example.Hex)
	if err != nil {
		t.Fatal(err)
	}
	hex, ok := out.(string)
	if !ok {
		t.Fatal(out)
	}
	if !ok {
		t.Fatal(out)
	}
	if hex != "#ffffff" {
		t.Fatal(hex)
	}
}

func TestColorRgb0(t *testing.T) {
	t.Parallel()
	out, err := Cast(map[string]interface{}{"r": float64(0), "g": float64(0), "b": float64(0)}, example.Color, example.Rgb, example.Hex)
	if err != nil {
		t.Fatal(err)
	}
	hex, ok := out.(string)
	if !ok {
		t.Fatal(out)
	}
	if !ok {
		t.Fatal(out)
	}
	if hex != "#000000" {
		t.Fatal(hex)
	}
}
