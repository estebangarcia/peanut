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

// PeanutStageSpec defines the desired state of PeanutStage
type PeanutStageSpec struct {
	Image   string   `json:"image,omitempty"`
	Command []string `json:"command,omitempty"`
	Args    []string `json:"args,omitempty"`
	Timeout int      `json:"timeout,omitempty"`
}

// PeanutStageStatus defines the observed state of PeanutStage
type PeanutStageStatus struct {
	// TBD
	Ack string `json:"acknowledged,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PeanutStage is the Schema for the peanutstages API
type PeanutStage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PeanutStageSpec   `json:"spec,omitempty"`
	Status PeanutStageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PeanutStageList contains a list of PeanutStage
type PeanutStageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PeanutStage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PeanutStage{}, &PeanutStageList{})
}
