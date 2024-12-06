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
	"github.com/patrickmn/go-cache"
	"time"
)

type LocalDb struct {
	db *cache.Cache
}

func NewLocalDb() *LocalDb {
	return &LocalDb{db: cache.New(time.Hour, time.Minute)}
}

func (this *LocalDb) Get(key string) (result *memcache.Item, err error) {
	value, found := this.db.Get(key)
	if !found {
		return nil, ErrNotFount
	}
	valueAsBytest, ok := value.([]byte)
	if !ok {
		return nil, errors.New("value is not bytes")
	}
	return &memcache.Item{
		Key:   key,
		Value: valueAsBytest,
	}, nil
}

func (this *LocalDb) Set(item *memcache.Item) error {
	if item == nil {
		return errors.New("missing item")
	}
	this.db.Set(item.Key, item.Value, time.Second*time.Duration(item.Expiration))
	return nil
}

func (this *LocalDb) Delete(key string) error {
	this.db.Delete(key)
	return nil
}
