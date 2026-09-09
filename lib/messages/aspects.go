/*
 * Copyright 2026 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package messages

import (
	"slices"
	"strings"

	"github.com/SENERGY-Platform/external-task-worker/lib/devicerepository/model"
	marshallermodel "github.com/SENERGY-Platform/marshaller/lib/marshaller/model"
)

//The single aspect of a command and of a protocol message is deprecated in favor of an
//aspect list, following ContentVariable.AspectIds of the device-repository. It stays an
//alias for a list with one element: it is folded into the list where a message is read and
//it is filled from the list where a message is written, so that a task or a response of a
//deployment that predates the lists is still handled, and one written here is still
//understood by a reader that only knows the single field.

// GetAspects returns the aspects the command asks for. A command names more than one aspect
// if the deployed criteria did; a service then has to serve all of them, the way the
// device-repository reads a filter-criteria with several aspects.
func (this Command) GetAspects() []model.AspectNode {
	if this.Aspect == nil {
		return this.Aspects
	}
	return marshallermodel.AspectNodesAlias(*this.Aspect, this.Aspects)
}

// SetAspects folds the deprecated Aspect into Aspects, so that everything behind the camunda
// boundary evaluates Aspects only.
func (this *Command) SetAspects() {
	this.Aspects = this.GetAspects()
}

// GetOutputAspectNodes returns the aspects of the response the metadata describes. Messages
// that were sent before this worker knew the list carry only the deprecated single node.
func (this Metadata) GetOutputAspectNodes() []model.AspectNode {
	if this.OutputAspectNode == nil {
		return this.OutputAspectNodes
	}
	return marshallermodel.AspectNodesAlias(*this.OutputAspectNode, this.OutputAspectNodes)
}

// SetOutputAspectNodes sets the aspects of the expected response and keeps the deprecated
// single node in sync, for a reader of the message that does not know the list yet.
func (this *Metadata) SetOutputAspectNodes(aspectNodes []model.AspectNode) {
	this.OutputAspectNodes = aspectNodes
	this.OutputAspectNode = deprecatedAspectNode(aspectNodes)
}

// deprecatedAspectNode picks the node a written message reports in its deprecated single
// aspect field. It is the node with the alphabetically first id, the way the
// device-repository fills the deprecated field of a path option.
func deprecatedAspectNode(aspectNodes []model.AspectNode) *model.AspectNode {
	if len(aspectNodes) == 0 {
		return nil
	}
	first := slices.MinFunc(aspectNodes, func(a model.AspectNode, b model.AspectNode) int {
		return strings.Compare(a.Id, b.Id)
	})
	return &first
}
