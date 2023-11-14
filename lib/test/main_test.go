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

package test

import (
	"context"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/docker"
	"io"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"sync"
	"testing"
)

var example = struct {
	Brightness string
	Lux        string
	Color      string
	Rgb        string
	Hex        string
}{
	Brightness: "example_brightness",
	Lux:        "example_lux",
	Color:      "example_color",
	Rgb:        "example_rgb",
	Hex:        "example_hex",
}

var duration = struct {
	Seconds string
}{
	Seconds: "urn:infai:ses:characteristic:9e1024da-3b60-4531-9f29-464addccb13c",
}

var temperature = struct {
	Celcius string
}{
	Celcius: "urn:infai:ses:characteristic:5ba31623-0ccb-4488-bfb7-f73b50e03b5a",
}

var color = struct {
	Rgb  string
	RgbR string
	RgbG string
	RgbB string
	Hex  string
}{
	Rgb:  "urn:infai:ses:characteristic:5b4eea52-e8e5-4e80-9455-0382f81a1b43",
	RgbR: "urn:infai:ses:characteristic:dfe6be4a-650c-4411-8d87-062916b48951",
	RgbG: "urn:infai:ses:characteristic:5ef27837-4aca-43ad-b8f6-4d95cf9ed99e",
	RgbB: "urn:infai:ses:characteristic:590af9ef-3a5e-4edb-abab-177cb1320b17",
	Hex:  "urn:infai:ses:characteristic:0fc343ce-4627-4c88-b1e0-d3ed29754af8",
}

// helper to find log creation
// example: log.SetOutput(StackWriter{Out: log.Writer(), Compare: "ERROR:  json: Unmarshal(nil *model.AspectNode)"})
type StackWriter struct {
	Out     io.Writer
	Compare string
}

func (s StackWriter) Write(p []byte) (n int, err error) {
	if strings.Contains(string(p), s.Compare) {
		debug.PrintStack()
	}
	return s.Out.Write(p)
}

// TODO: remove as soon as https://github.com/testcontainers/testcontainers-go/issues/1671 is fixed
func TestMain(m *testing.M) {
	//this main is needed to keep ryuk from closing
	//which is needed because ryuk is currently (github.com/testcontainers/testcontainers-go v0.26.0) not able to restart
	//https://github.com/testcontainers/testcontainers-go/issues/1671
	var code int
	defer os.Exit(code)
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//keep ryuk alive by keeping a container for the context of all tests
	_, _, err := docker.Memcached(ctx, wg)
	if err != nil {
		log.Fatal(err)
	}
	code = m.Run()
}
