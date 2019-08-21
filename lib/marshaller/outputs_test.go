package marshaller

import (
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/example"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
)

func ExampleMarshalOutput1() {
	protocol := model.Protocol{
		Id:      "p1",
		Name:    "p1",
		Handler: "p1",
		ProtocolSegments: []model.ProtocolSegment{
			{Id: "p1.1", Name: "body"},
			{Id: "p1.2", Name: "head"},
		},
	}
	service := model.Service{
		Id:          "s1",
		LocalId:     "s1l",
		Name:        "s1n",
		Description: "s1d",
		ProtocolId:  "p1",
		Outputs: []model.Content{
			{
				Id: "c1",
				ContentVariable: model.ContentVariable{
					Id:   "c1.1",
					Name: "payload",
					Type: model.Structure,
					SubContentVariables: []model.ContentVariable{
						{
							Id:   "c1.1.1",
							Name: "color",
							Type: model.Structure,
							SubContentVariables: []model.ContentVariable{
								{
									Id:               "sr",
									Name:             "red",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".r",
								},
								{
									Id:               "sg",
									Name:             "green",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".g",
								},
								{
									Id:               "sb",
									Name:             "blue",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".b",
								},
							},
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "p1.1",
			},
		},
	}
	output := map[string]string{"body": "{\"color\":{\"blue\":100,\"green\":0,\"red\":255}}"}
	outputCharacteristic := example.Hex
	result, err := UnmarshalOutputs(protocol, service, output, outputCharacteristic)
	fmt.Println(result, err)

	//output:
	//#ff0064 <nil>
}
