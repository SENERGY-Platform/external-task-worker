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
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/binary"
	"testing"
)

func TestBoolToStatusTrue(t *testing.T) {
	t.Parallel()
	out, err := Cast(true, binary.Binary, binary.Boolean, binary.BinaryStatusCode)
	if err != nil {
		t.Fatal(err)
	}
	status, ok := out.(int)
	if !ok {
		t.Fatal(out)
	}

	if status != 1 {
		t.Fatal(status)
	}
}

func TestBoolToStatusFalse(t *testing.T) {
	t.Parallel()
	out, err := Cast(false, binary.Binary, binary.Boolean, binary.BinaryStatusCode)
	if err != nil {
		t.Fatal(err)
	}
	status, ok := out.(int)
	if !ok {
		t.Fatal(out)
	}

	if status != 0 {
		t.Fatal(status)
	}
}

func TestStatusToBool1(t *testing.T) {
	t.Parallel()
	out, err := Cast(1, binary.Binary, binary.BinaryStatusCode, binary.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	status, ok := out.(bool)
	if !ok {
		t.Fatal(out)
	}

	if !status {
		t.Fatal(status)
	}
}

func TestStatusToBool0(t *testing.T) {
	t.Parallel()
	out, err := Cast(0, binary.Binary, binary.BinaryStatusCode, binary.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	status, ok := out.(bool)
	if !ok {
		t.Fatal(out)
	}

	if status {
		t.Fatal(status)
	}
}

func TestStatusToBool42(t *testing.T) {
	t.Parallel()
	out, err := Cast(42, binary.Binary, binary.BinaryStatusCode, binary.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	status, ok := out.(bool)
	if !ok {
		t.Fatal(out)
	}

	if !status {
		t.Fatal(status)
	}
}

func TestStatusToBoolM42(t *testing.T) {
	t.Parallel()
	out, err := Cast(-42, binary.Binary, binary.BinaryStatusCode, binary.Boolean)
	if err != nil {
		t.Fatal(err)
	}
	status, ok := out.(bool)
	if !ok {
		t.Fatal(out)
	}

	if status {
		t.Fatal(status)
	}
}
