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
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/speedlevel"
	"reflect"
	"testing"
)

func TestSpeedlevelPercentToNumber(t *testing.T) {
	t.Parallel()
	table := []struct {
		Percent int64
		Number  int64
	}{
		{Percent: 0, Number: 1},
		{Percent: 3, Number: 1},
		{Percent: 20, Number: 2},
		{Percent: 77, Number: 7},
		{Percent: 100, Number: 10},
	}
	for _, test := range table {
		out, err := Cast(test.Percent, speedlevel.ConceptId, speedlevel.PercentageId, speedlevel.OneToTenId)
		if err != nil {
			t.Fatal(err)
		}
		result, ok := out.(int64)
		if !ok {
			t.Fatal(out)
		}

		if result != test.Number {
			t.Fatal(result, test.Number)
		}
	}
}

func TestSpeedlevelPercentToNumber2(t *testing.T) {
	t.Parallel()
	table := []struct {
		Percent int
		Number  int64
	}{
		{Percent: 0, Number: 1},
		{Percent: 3, Number: 1},
		{Percent: 20, Number: 2},
		{Percent: 77, Number: 7},
		{Percent: 100, Number: 10},
	}
	for _, test := range table {
		out, err := Cast(test.Percent, speedlevel.ConceptId, speedlevel.PercentageId, speedlevel.OneToTenId)
		if err != nil {
			t.Fatal(err)
		}
		result, ok := out.(int64)
		if !ok {
			t.Fatal(out, reflect.TypeOf(out).Name())
		}

		if result != test.Number {
			t.Fatal(result, test.Number)
		}
	}
}

func TestSpeedlevelPercentToNumber3(t *testing.T) {
	t.Parallel()
	table := []struct {
		Percent float64
		Number  int64
	}{
		{Percent: 0, Number: 1},
		{Percent: 3, Number: 1},
		{Percent: 20, Number: 2},
		{Percent: 77, Number: 7},
		{Percent: 100, Number: 10},
	}
	for _, test := range table {
		out, err := Cast(test.Percent, speedlevel.ConceptId, speedlevel.PercentageId, speedlevel.OneToTenId)
		if err != nil {
			t.Fatal(err)
		}
		result, ok := out.(int64)
		if !ok {
			t.Fatal(out, reflect.TypeOf(out).Name())
		}

		if result != test.Number {
			t.Fatal(result, test.Number)
		}
	}
}

func TestSpeedlevelNumberToPercent(t *testing.T) {
	t.Parallel()
	table := []struct {
		Percent int64
		Number  int64
	}{
		{Percent: 10, Number: 1},
		{Percent: 20, Number: 2},
		{Percent: 70, Number: 7},
		{Percent: 100, Number: 10},
	}
	for _, test := range table {
		out, err := Cast(test.Number, speedlevel.ConceptId, speedlevel.OneToTenId, speedlevel.PercentageId)
		if err != nil {
			t.Fatal(err)
		}
		result, ok := out.(int64)
		if !ok {
			t.Fatal(out)
		}

		if result != test.Percent {
			t.Fatal(result, test.Percent)
		}
	}
}

func TestSpeedlevelNumberToPercent2(t *testing.T) {
	t.Parallel()
	table := []struct {
		Percent int64
		Number  float64
	}{
		{Percent: 10, Number: 1},
		{Percent: 20, Number: 2},
		{Percent: 70, Number: 7},
		{Percent: 100, Number: 10},
	}
	for _, test := range table {
		out, err := Cast(test.Number, speedlevel.ConceptId, speedlevel.OneToTenId, speedlevel.PercentageId)
		if err != nil {
			t.Fatal(err)
		}
		result, ok := out.(int64)
		if !ok {
			t.Fatal(out)
		}

		if result != test.Percent {
			t.Fatal(result, test.Percent)
		}
	}
}
