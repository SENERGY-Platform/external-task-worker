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
	"errors"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/coocood/freecache"
)

var LocalDbSize = 100 * 1024 * 1024 //100MB

type LocalDb struct {
	db *freecache.Cache
}

func NewLocalDb() *LocalDb {
	return &LocalDb{db: freecache.NewCache(LocalDbSize)}
}

func (this *LocalDb) Get(key string) (result *memcache.Item, err error) {
	value, err := this.db.Get([]byte(key))
	if err == freecache.ErrNotFound {
		err = ErrNotFount
	}
	if err != nil {
		return nil, err
	}
	return &memcache.Item{
		Key:   key,
		Value: value,
	}, nil
}

func (this *LocalDb) Set(item *memcache.Item) error {
	if item == nil {
		return errors.New("missing item")
	}
	return this.db.Set([]byte(item.Key), item.Value, int(item.Expiration))
}
