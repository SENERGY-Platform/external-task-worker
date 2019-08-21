package marshaller

import (
	"fmt"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/casting/example"
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
)

func ExampleUnmarshalOutput1() {
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

func ExampleUnmarshalOutput2() {
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
					Id:               "c1.1",
					Name:             "color",
					Type:             model.String,
					CharacteristicId: example.Hex,
				},
				Serialization:     "json",
				ProtocolSegmentId: "p1.1",
			},
		},
	}
	output := map[string]string{"body": "\"#ff0064\""}
	outputCharacteristic := example.Rgb
	result, err := UnmarshalOutputs(protocol, service, output, outputCharacteristic)
	fmt.Println(result, err)

	//output:
	//map[b:100 g:0 r:255] <nil>
}

func ExampleUnmarshalOutputMulti() {
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
									Value:            float64(255),
								},
								{
									Id:               "sg",
									Name:             "green",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".g",
									Value:            float64(255),
								},
								{
									Id:               "sb",
									Name:             "blue",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".b",
									Value:            float64(255),
								},
							},
						},
						{
							Id:               "c1.1.2",
							Name:             "bri",
							Type:             model.Integer,
							CharacteristicId: example.Lux,
							Value:            float64(100),
						},
					},
				},
				Serialization:     "json",
				ProtocolSegmentId: "p1.1",
			},
		},
	}
	fmt.Println(UnmarshalOutputs(protocol, service, map[string]string{"body": "{\"bri\":100,\"color\":{\"blue\":100,\"green\":0,\"red\":255}}"}, example.Hex))
	fmt.Println(UnmarshalOutputs(protocol, service, map[string]string{"body": "{\"bri\":25,\"color\":{\"blue\":255,\"green\":255,\"red\":255}}"}, example.Lux))

	//output:
	//#ff0064 <nil>
	//25 <nil>
}

func ExampleMarshalOutputMultiXml() {
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
									Name:             "-red",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".r",
									Value:            float64(255),
								},
								{
									Id:               "sg",
									Name:             "-green",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".g",
									Value:            float64(255),
								},
								{
									Id:               "sb",
									Name:             "-blue",
									Type:             model.Integer,
									CharacteristicId: example.Rgb + ".b",
									Value:            float64(255),
								},
							},
						},
						{
							Id:               "c1.1.2",
							Name:             "bri",
							Type:             model.Integer,
							CharacteristicId: example.Lux,
							Value:            float64(100),
						},
					},
				},
				Serialization:     "xml",
				ProtocolSegmentId: "p1.1",
			},
		},
	}
	fmt.Println(UnmarshalOutputs(protocol, service, map[string]string{"body": "<payload><bri>100</bri><color blue=\"100\" green=\"0\" red=\"255\"/></payload>"}, example.Hex))
	fmt.Println(UnmarshalOutputs(protocol, service, map[string]string{"body": "<payload><bri>25</bri><color blue=\"255\" green=\"255\" red=\"255\"></color></payload>"}, example.Lux))

	//output:
	//#ff0064 <nil>
	//25 <nil>
}
