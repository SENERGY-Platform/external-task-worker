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

package lib

import (
	"encoding/json"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"log"
	"runtime/debug"
)

func (this *worker) ErrorMessageHandler(msg string) error {
	var message messages.ProtocolMsg
	err := json.Unmarshal([]byte(msg), &message)
	if err != nil {
		log.Println("ERROR:", err)
		debug.PrintStack()
		return nil
	}
	err = this.camunda.UnlockTask(message.TaskInfo)
	if err != nil {
		log.Println("ERROR: unable to unlock task", err)
		return nil
	}
	return nil
}
