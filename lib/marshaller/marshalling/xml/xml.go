package xml

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/marshalling/base"
)

type Marshaller struct {
}

const Format = "xml"

func init() {
	base.Register(Format, Marshaller{})
}
