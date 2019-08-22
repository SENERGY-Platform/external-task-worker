package devicerepository

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/model"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

type FactoryInterface interface {
	Get(configType util.Config) RepoInterface
}

type RepoInterface interface {
	GetDevice(token Impersonate, id string) (model.Device, error)
	GetService(token Impersonate, device model.Device, serviceId string) (model.Service, error)
	GetProtocol(token Impersonate, id string) (model.Protocol, error)
	GetToken(user string) (Impersonate, error)
}
