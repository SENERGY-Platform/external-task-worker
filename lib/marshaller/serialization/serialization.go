package serialization

import (
	"github.com/SENERGY-Platform/external-task-worker/lib/marshaller/serialization/base"
	_ "github.com/SENERGY-Platform/external-task-worker/lib/marshaller/serialization/json"
	_ "github.com/SENERGY-Platform/external-task-worker/lib/marshaller/serialization/xml"
)

func Get(key string) (marshaller base.Marshaller, ok bool) {
	return base.Get(key)
}
