package repo

import "github.com/SENERGY-Platform/external-task-worker/util"

type FactoryType struct {}

var Factory FactoryType

func (FactoryType) Get(config util.Config) RepoInterface {
	return NewIot(config)
}