package xml

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/serialization/base"
)

type Marshaller struct {
}

const Format = "xml"

func init() {
	base.Register(Format, Marshaller{})
}
