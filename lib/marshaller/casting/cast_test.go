package casting

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/example"
	"testing"
)

func TestColorHexShortF(t *testing.T) {
	t.Parallel()
	out, err := Cast("#FFF", example.Concept, example.Hex, example.Rgb)
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
	out, err := Cast("#FFFFFF", example.Concept, example.Hex, example.Rgb)
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
	out, err := Cast("#000000", example.Concept, example.Hex, example.Rgb)
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
	out, err := Cast("#000", example.Concept, example.Hex, example.Rgb)
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
	out, err := Cast(map[string]int64{"r": 255, "g": 255, "b": 255}, example.Concept, example.Rgb, example.Hex)
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
	out, err := Cast(map[string]int64{"r": 0, "g": 0, "b": 0}, example.Concept, example.Rgb, example.Hex)
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
