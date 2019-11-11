package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/example"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"
	"log"
	"time"
)

func ExampleWorkerRetries() {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.PESSIMISTIC
	config.CamundaWorkerTimeout = 100
	config.CamundaFetchLockDuration = 100

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mock.Camunda)

	time.Sleep(1 * time.Second)

	//populate repository
	mock.Repo.RegisterDevice(model.Device{
		Id:           "device_1",
		Name:         "d1",
		DeviceTypeId: "dt1",
		LocalId:      "d1u",
	})

	mock.Repo.RegisterProtocol(model.Protocol{
		Id:               "p1",
		Name:             "protocol1",
		Handler:          "protocol1",
		ProtocolSegments: []model.ProtocolSegment{{Id: "ms1", Name: "body"}},
	})

	mock.Repo.RegisterService(model.Service{
		Id:         "service_1",
		Name:       "s1",
		LocalId:    "s1u",
		ProtocolId: "p1",
		Outputs: []model.Content{
			{
				Id: "metrics",
				ContentVariable: model.ContentVariable{
					Id:   "metrics",
					Name: "metrics",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:               "level",
							Name:             "level",
							Type:             model.Integer,
							CharacteristicId: example.Hex,
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "ms1",
			},
		},
	})

	cmd1 := messages.Command{
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
		CharacteristicId: example.Rgb,
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
		Retries:          2,
	}

	cmdMsg1, err := json.Marshal(cmd1)
	if err != nil {
		log.Fatal(err)
	}

	cmd2 := messages.Command{
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
		CharacteristicId: example.Hex,
		DeviceId:         "device_1",
		ServiceId:        "service_1",
		ProtocolId:       "p1",
	}

	cmdMsg2, err := json.Marshal(cmd2)
	if err != nil {
		log.Fatal(err)
	}

	mock.Camunda.AddTask(messages.CamundaExternalTask{
		Id: "1",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg1),
			},
		},
	})

	time.Sleep(1 * time.Second)

	mock.Camunda.AddTask(messages.CamundaExternalTask{
		Id: "2",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg2),
			},
		},
	})

	time.Sleep(time.Duration(config.CamundaFetchLockDuration) * time.Millisecond * 10) //wait for all fetches and timeouts

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	if len(protocolMessageStrings) != 4 {
		fmt.Println(len(protocolMessageStrings))
		fmt.Println(protocolMessageStrings)
		return
	}

	fetched, completed, failed := mock.Camunda.GetStatus()

	if len(fetched) != 0 || len(failed) != 2 || len(completed) != 0 {
		fmt.Println("fetched:", fetched)
		fmt.Println("failed:", failed)
		fmt.Println("completed:", completed)
		return
	}

	//output:
	//
}
