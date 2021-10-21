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

package devicegroups

import (
	"encoding/json"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/bradfitz/gomemcache/memcache"
	"time"
)

const METADATA_KEY_PREFIX = "meta."
const RESULT_KEY_PREFIX = "result."
const SUB_TASK_STATE_KEY_PREFIX = "ststate."
const LAST_CALL_INFO = "lastcall."

var ErrNotFount = memcache.ErrCacheMiss

type DbInterface interface {
	Get(key string) (*memcache.Item, error)
	Set(item *memcache.Item) error
	Delete(key string) error
}

type SubResultWrapper struct {
	Failed bool
	Value  interface{}
}

type SubTaskState struct {
	LastTry  time.Time
	TryCount int64
}

var MaxRetries = 10

func (this *DeviceGroups) dbGet(key string, value interface{}) (err error) {
	var item *memcache.Item
	for i := 0; i < MaxRetries; i++ {
		item, err = this.db.Get(key)
		if err == nil || err == ErrNotFount {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err != nil {
		return err
	}
	err = json.Unmarshal(item.Value, value)
	return err
}

func (this *DeviceGroups) dbSet(key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	for i := 0; i < MaxRetries; i++ {
		err = this.db.Set(&memcache.Item{Value: jsonValue, Expiration: this.expirationInSeconds, Key: key})
		if err == nil {
			return err
		}
		time.Sleep(200 * time.Millisecond)
	}
	return err
}

func (this *DeviceGroups) getGroupMetadata(taskId string) (metadata messages.GroupTaskMetadata, err error) {
	err = this.dbGet(METADATA_KEY_PREFIX+taskId, &metadata)
	return
}

func (this *DeviceGroups) setGroupMetadata(taskId string, metadata messages.GroupTaskMetadata) (err error) {
	err = this.dbSet(METADATA_KEY_PREFIX+taskId, metadata)
	return
}

func (this *DeviceGroups) getSubResult(subTaskId string) (result SubResultWrapper, err error) {
	err = this.dbGet(RESULT_KEY_PREFIX+subTaskId, &result)
	return
}

func (this *DeviceGroups) setSubResult(subTaskId string, value SubResultWrapper) (err error) {
	err = this.dbSet(RESULT_KEY_PREFIX+subTaskId, value)
	return
}

func (this *DeviceGroups) getSubTaskState(subTaskId string) (result SubTaskState, err error) {
	err = this.dbGet(SUB_TASK_STATE_KEY_PREFIX+subTaskId, &result)
	return
}

func (this *DeviceGroups) setSubTaskState(subTaskId string, value SubTaskState) (err error) {
	err = this.dbSet(SUB_TASK_STATE_KEY_PREFIX+subTaskId, value)
	return
}
