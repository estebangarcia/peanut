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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PeanutPipelineExecutionSpec defines the desired state of PeanutPipelineExecution
type PeanutPipelineExecutionSpec struct {
	// +kubebuilder:validation:Required
	PipelineName string `json:"pipelineName,omitempty"`
}

type PhaseState string

const (
	Running PhaseState = "Running"
	Pending PhaseState = "Pending"
	Failed  PhaseState = "Failed"
)

// PeanutPipelineExecutionStatus defines the observed state of PeanutPipelineExecution
type PeanutPipelineExecutionStatus struct {
	Phase PhaseState `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PeanutPipelineExecution is the Schema for the peanutpipelineexecutions API
type PeanutPipelineExecution struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PeanutPipelineExecutionSpec   `json:"spec,omitempty"`
	Status PeanutPipelineExecutionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PeanutPipelineExecutionList contains a list of PeanutPipelineExecution
type PeanutPipelineExecutionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PeanutPipelineExecution `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PeanutPipelineExecution{}, &PeanutPipelineExecutionList{})
}
