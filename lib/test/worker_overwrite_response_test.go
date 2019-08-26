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
	"sort"

	"log"
	"time"
)

func ExampleWorkerOverwriteResponse() {
	util.TimeNow = func() time.Time {
		return time.Time{}
	}
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.PESSIMISTIC
	config.CamundaWorkerTimeout = 100

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

	cmdOverwrite := messages.Overwrite{
		DeviceId:   "device_1",
		ServiceId:  "service_1",
		ProtocolId: "p1",
	}

	cmdOverwriteStr, err := json.Marshal(cmdOverwrite)
	if err != nil {
		log.Fatal(err)
	}

	cmd1 := messages.Command{
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
		CharacteristicId: example.Rgb,
	}

	cmdMsg1, err := json.Marshal(cmd1)
	if err != nil {
		log.Fatal(err)
	}

	cmd2 := messages.Command{
		Function:         model.Function{RdfType: model.SES_ONTOLOGY_MEASURING_FUNCTION},
		CharacteristicId: example.Hex,
	}

	cmdMsg2, err := json.Marshal(cmd2)
	if err != nil {
		log.Fatal(err)
	}

	mock.Camunda.AddTask(messages.CamundaTask{
		Id: "1",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg1),
			},
			util.CAMUNDA_VARIABLES_OVERWRITE: {
				Value: string(cmdOverwriteStr),
			},
		},
	})

	time.Sleep(1 * time.Second)

	mock.Camunda.AddTask(messages.CamundaTask{
		Id: "2",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg2),
			},
			util.CAMUNDA_VARIABLES_OVERWRITE: {
				Value: string(cmdOverwriteStr),
			},
		},
	})

	time.Sleep(1 * time.Second)

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	if len(protocolMessageStrings) != 2 {
		log.Fatal(protocolMessageStrings)
	}

	for _, message := range protocolMessageStrings {
		msg := messages.ProtocolMsg{}
		err = json.Unmarshal([]byte(message), &msg)
		if err != nil {
			log.Fatal(err)
		}
		msg.Response.Output = map[string]string{
			"body": "{\"level\":\"#c83200\"}",
		}
		resp, err := json.Marshal(msg)
		if err != nil {
			log.Fatal(err)
		}
		mock.Kafka.Produce(config.ResponseTopic, string(resp))
		time.Sleep(1 * time.Second)
	}

	fetched, completed, failed := mock.Camunda.GetStatus()

	if len(fetched) != 0 || len(failed) != 0 || len(completed) != 2 {
		log.Println("fetched:", fetched)
		log.Println("failed:", failed)
		log.Println("completed:", completed)
		log.Fatal(len(fetched), len(failed), len(completed))
	}

	list := []string{}

	for _, cmd := range completed {
		temp, err := json.Marshal(cmd)
		if err != nil {
			log.Fatal(err)
		}
		list = append(list, string(temp))
	}
	sort.Strings(list)
	for _, cmd := range list {
		fmt.Println(cmd)
	}

	//output:
	//{"function":{"id":"","name":"","concept_ids":null,"rdf_type":""},"characteristic_id":"example_hex","device_id":"device_1","device":{"id":"device_1","local_id":"d1u","name":"d1","device_type_id":"dt1"},"service_id":"service_1","service":{"id":"service_1","local_id":"s1u","name":"s1","description":"","aspects":null,"protocol_id":"p1","inputs":null,"outputs":[{"id":"metrics","content_variable":{"id":"metrics","name":"metrics","type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"level","name":"level","type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_hex","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"ms1"}],"functions":null,"rdf_type":""},"protocol":{"id":"p1","name":"protocol1","handler":"protocol1","protocol_segments":[{"id":"ms1","name":"body"}]},"protocol_id":"p1","output":"#c83200"}
	//{"function":{"id":"","name":"","concept_ids":null,"rdf_type":""},"characteristic_id":"example_rgb","device_id":"device_1","device":{"id":"device_1","local_id":"d1u","name":"d1","device_type_id":"dt1"},"service_id":"service_1","service":{"id":"service_1","local_id":"s1u","name":"s1","description":"","aspects":null,"protocol_id":"p1","inputs":null,"outputs":[{"id":"metrics","content_variable":{"id":"metrics","name":"metrics","type":"https://schema.org/StructuredValue","sub_content_variables":[{"id":"level","name":"level","type":"https://schema.org/Integer","sub_content_variables":null,"characteristic_id":"example_hex","value":null,"serialization_options":null}],"characteristic_id":"","value":null,"serialization_options":null},"serialization":"json","protocol_segment_id":"ms1"}],"functions":null,"rdf_type":""},"protocol":{"id":"p1","name":"protocol1","handler":"protocol1","protocol_segments":[{"id":"ms1","name":"body"}]},"protocol_id":"p1","output":{"b":0,"g":50,"r":200}}
}
