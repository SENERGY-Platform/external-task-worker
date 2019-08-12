package test

import (
	"context"
	"encoding/json"
	"github.com/SENERGY-Platform/external-task-worker/lib"
	"github.com/SENERGY-Platform/external-task-worker/lib/messages"
	"github.com/SENERGY-Platform/external-task-worker/lib/test/mock"
	"github.com/SENERGY-Platform/external-task-worker/util"
	formatter_lib "github.com/SENERGY-Platform/formatter-lib"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
	uuid "github.com/satori/go.uuid"
	"log"
	"reflect"
	"testing"
	"time"
)

func TestExample(t *testing.T) {
	config, err := util.LoadConfig("../../config.json")
	if err != nil {
		log.Fatal(err)
	}

	config.CompletionStrategy = util.OPTIMISTIC
	config.CamundaWorkerTimeout = 100

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go lib.Worker(ctx, config, mock.Kafka, mock.Repo, mock.Camunda)

	time.Sleep(1 * time.Second)

	//populate repository
	mock.Repo.RegisterDevice(model.DeviceInstance{
		Id:         "device_1",
		Name:       "d1",
		DeviceType: "dt1",
		Url:        "d1u",
	})

	mock.Repo.RegisterService(model.Service{
		Id:   "service_1",
		Name: "s1",
		Protocol: model.Protocol{
			ProtocolHandlerUrl: "protocol1",
			MsgStructure:       []model.MsgSegment{{Id: "ms1", Name: "body"}},
		},
		Input: []model.TypeAssignment{{
			Name:       "metrics",
			MsgSegment: model.MsgSegment{Id: "ms1", Name: "body"},
			Type: model.ValueType{
				Name:     "metrics_type",
				BaseType: model.StructBaseType,
				Fields: []model.FieldType{{
					Id:   uuid.NewV4().String(),
					Name: "level",
					Type: model.ValueType{
						Id:       uuid.NewV4().String(),
						Name:     "int",
						BaseType: model.XsdInt,
					},
				}},
			},
			Format: formatter_lib.JSON_ID,
		}},
		Url: "s1u",
	})

	//create camunda task
	cmd := messages.Command{
		InstanceId: "device_1",
		ServiceId:  "service_1",
		Inputs: map[string]interface{}{
			"metrics": map[string]interface{}{
				"level": 42,
			},
		},
	}

	cmdMsg, err := json.Marshal(cmd)
	if err != nil {
		t.Fatal(err)
	}

	mock.Camunda.AddTask(messages.CamundaTask{
		Id: "1",
		Variables: map[string]messages.CamundaVariable{
			util.CAMUNDA_VARIABLES_PAYLOAD: {
				Value: string(cmdMsg),
			},
		},
	})

	time.Sleep(1 * time.Second)

	protocolMessageStrings := mock.Kafka.GetProduced("protocol1")

	if len(protocolMessageStrings) != 1 {
		t.Fatal(protocolMessageStrings)
	}

	envelope := messages.Envelope{}
	err = json.Unmarshal([]byte(protocolMessageStrings[0]), &envelope)
	if err != nil {
		t.Fatal(err, protocolMessageStrings[0])
	}

	temp, err := json.Marshal(envelope.Value)
	if err != nil {
		t.Fatal(err, envelope.Value)
	}
	protocolMessage := messages.ProtocolMsg{}
	err = json.Unmarshal(temp, &protocolMessage)
	if err != nil {
		t.Fatal(err, envelope.Value)
	}

	if !reflect.DeepEqual(protocolMessage.ProtocolParts, []messages.ProtocolPart{{Name: "body", Value: "{\n    \"level\": 42\n}\n"}}) {
		t.Fatal(protocolMessage.ProtocolParts, protocolMessage, protocolMessageStrings[0])
	}

	fetched, finished, failed := mock.Camunda.GetStatus()
	if len(fetched) != 0 {
		t.Fatal(fetched)
	}
	if len(finished) != 1 {
		t.Fatal(finished)
	}
	if len(failed) != 0 {
		t.Fatal(failed)
	}

}
