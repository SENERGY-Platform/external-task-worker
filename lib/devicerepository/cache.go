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

package devicerepository

import (
	"encoding/json"
	"errors"
	"github.com/coocood/freecache"
	"log"
)

var L1Expiration = 60         // 60sec
var L1Size = 40 * 1024 * 1024 //40MB
var Debug = false

type Cache struct {
	l1 *freecache.Cache
}

type Item struct {
	Key   string
	Value []byte
}

var ErrNotFound = errors.New("key not found in cache")

func NewCache() *Cache {
	return &Cache{l1: freecache.NewCache(L1Size)}
}

func (this *Cache) Get(key string) (item Item, err error) {
	item.Value, err = this.l1.Get([]byte(key))
	if err != nil && err != freecache.ErrNotFound {
		log.Println("ERROR: in Cache::l1.Get()", err)
	}
	return
}

func (this *Cache) Set(key string, value []byte) {
	err := this.l1.Set([]byte(key), value, L1Expiration)
	if err != nil {
		log.Println("ERROR: in Cache::l1.Set()", err)
	}
	return
}

func (this *Cache) Use(key string, getter func() (interface{}, error), result interface{}) (err error) {
	item, err := this.Get(key)
	if err == nil {
		err = json.Unmarshal(item.Value, result)
		return
	}
	temp, err := getter()
	if err != nil {
		return err
	}
	value, err := json.Marshal(temp)
	if err != nil {
		return err
	}
	this.Set(key, value)
	return json.Unmarshal(value, &result)
}
