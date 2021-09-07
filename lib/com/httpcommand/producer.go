/*
 * Copyright 2020 InfAI (CC SES)
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

package httpcommand

import (
	"errors"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"io"
	"log"
	"net/http"
	"strings"
)

type Producer struct {
	config util.Config
}

func (this *Producer) Produce(topic string, message string) (err error) {
	resp, err := http.Post(topic, "application/json", strings.NewReader(message))
	if err != nil {
		log.Println("ERROR:", topic, err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		respMsg, _ := io.ReadAll(resp.Body)
		err = errors.New("http producer: " + resp.Status + " " + string(respMsg))
		log.Println("ERROR:", topic, err)
		return err
	}
	return nil
}

func (this *Producer) ProduceWithKey(topic string, key string, message string) (err error) {
	return this.Produce(topic, message)
}
