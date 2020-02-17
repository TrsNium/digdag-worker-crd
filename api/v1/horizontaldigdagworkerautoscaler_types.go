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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// HorizontalDigdagWorkerAutoscalerSpec defines the desired state of HorizontalDigdagWorkerAutoscaler
type HorizontalDigdagWorkerAutoscalerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ScaleTargetDeployment string `json:"scaleTargetDeployment"`
	// +kubebuilder:validation:Minimum=1
	DigdagWorkerMaxTaskThreads int32  `json:"digdagMaxTaskThreads"`
	PostgresqlHost             string `json:"postgresqlHost"`
	PostgresqlPort             string `json:"postgresqlPort"`
	PostgresqlDatabase         string `json:"postgresqlDatabase"`
	PostgresqlUser             string `json:"postgresqlUser"`
	PostgresqlPassword         string `json:"postgresqlPassword"`
}

// HorizontalDigdagWorkerAutoscalerStatus defines the observed state of HorizontalDigdagWorkerAutoscaler
type HorizontalDigdagWorkerAutoscalerStatus struct{}

// +kubebuilder:object:root=true

// HorizontalDigdagWorkerAutoscaler is the Schema for the horizontaldigdagworkerautoscalers API
type HorizontalDigdagWorkerAutoscaler struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HorizontalDigdagWorkerAutoscalerSpec   `json:"spec,omitempty"`
	Status HorizontalDigdagWorkerAutoscalerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HorizontalDigdagWorkerAutoscalerList contains a list of HorizontalDigdagWorkerAutoscaler
type HorizontalDigdagWorkerAutoscalerList struct {
	metav1.TypeMeta `json:"typemeta,inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HorizontalDigdagWorkerAutoscaler `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HorizontalDigdagWorkerAutoscaler{}, &HorizontalDigdagWorkerAutoscalerList{})
}
