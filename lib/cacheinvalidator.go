/*
 * Copyright 2023 InfAI (CC SES)
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
	"context"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"github.com/SENERGY-Platform/service-commons/pkg/cache/invalidator"
	"github.com/SENERGY-Platform/service-commons/pkg/kafka"
	"log"
	"runtime/debug"
	"time"
)

func StartCacheInvalidator(ctx context.Context, conf util.Config) (err error) {
	if conf.KafkaUrl == "" || conf.KafkaUrl == "-" {
		return nil
	}
	kafkaConf := kafka.Config{
		KafkaUrl:               conf.KafkaUrl,
		StartOffset:            kafka.LastOffset,
		Debug:                  conf.Debug,
		PartitionWatchInterval: time.Minute,
		InitTopic:              conf.InitTopics,
		OnError: func(err error) {
			log.Println("ERROR:", err)
			debug.PrintStack()
		},
	}
	if len(conf.CacheInvalidationAllKafkaTopics) > 0 {
		err = invalidator.StartCacheInvalidatorAll(ctx, kafkaConf, conf.CacheInvalidationAllKafkaTopics, nil)
		if err != nil {
			return err
		}
	}

	err = invalidator.StartKnownCacheInvalidators(ctx, kafkaConf, invalidator.KnownTopics{
		DeviceTopic:      conf.DeviceKafkaTopic,
		DeviceGroupTopic: conf.DeviceGroupKafkaTopic,
	}, nil)
	if err != nil {
		return err
	}

	return nil
}
