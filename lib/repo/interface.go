package repo

import (
	"github.com/SENERGY-Platform/external-task-worker/util"
	"github.com/SENERGY-Platform/iot-device-repository/lib/model"
)

type FactoryInterface interface {
	Get(configType util.Config)RepoInterface
}

type RepoInterface interface{
	GetDeviceInfo(instanceId string, serviceId string, user string) (instance model.DeviceInstance, service model.Service, err error)
}

