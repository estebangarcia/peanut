/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"github.com/goombaio/dag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PeanutPipelineStageDefinition defines the stage that the pipeline will execute when triggered
type PeanutPipelineStageDefinition struct {
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`

	// +kubebuilder:validation:Required
	Type string `json:"type,omitempty"`

	// +kubebuilder:validation:Optional
	Dependencies []string `json:"dependsOn,omitempty"`
}

// PeanutPipelineSpec defines the desired state of PeanutPipeline
type PeanutPipelineSpec struct {
	// +kubebuilder:validation:Required
	Stages []PeanutPipelineStageDefinition `json:"stages,omitempty"`
}

// PeanutPipelineStatus defines the observed state of PeanutPipeline
type PeanutPipelineStatus struct {
	// TBD
	Ack string `json:"acknowledged,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PeanutPipeline is the Schema for the peanutpipelines API
type PeanutPipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PeanutPipelineSpec   `json:"spec,omitempty"`
	Status PeanutPipelineStatus `json:"status,omitempty"`
}

func (p *PeanutPipeline) BuildStageGraph() (*dag.DAG, error) {
	stageGraph := dag.NewDAG()

	rootVertex := dag.NewVertex("root", nil)

	if err := stageGraph.AddVertex(rootVertex); err != nil {
		return nil, err
	}

	for _, stage := range p.Spec.Stages {
		if err := stageGraph.AddVertex(dag.NewVertex(stage.Name, stage)); err != nil {
			return nil, err
		}
	}

	for _, stage := range p.Spec.Stages {
		v, _ := stageGraph.GetVertex(stage.Name)
		if len(stage.Dependencies) == 0 {
			if err := stageGraph.AddEdge(rootVertex, v); err != nil {
				return nil, err
			}
		} else {
			for _, dep := range stage.Dependencies {
				d, err := stageGraph.GetVertex(dep)

				if err != nil {
					return nil, err
				}

				err = stageGraph.AddEdge(d, v)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return stageGraph, nil
}

// +kubebuilder:object:root=true

// PeanutPipelineList contains a list of PeanutPipeline
type PeanutPipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PeanutPipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PeanutPipeline{}, &PeanutPipelineList{})
}
