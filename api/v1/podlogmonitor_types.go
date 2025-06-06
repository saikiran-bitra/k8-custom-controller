/*
Copyright 2025 Saikiran Bitra.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PodLogMonitorSpec defines the desired state of PodLogMonitor.
type PodLogMonitorSpec struct {
        // INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
        // Important: Run "make" to regenerate code after modifying this file

        // Foo is an example field of PodLogMonitor. Edit podlogmonitor_types.go to remove/update
        //Foo string `json:"foo,omitempty"`
        Namespace string `json:"namespace"`
        LogMessage string `json:"logMessage"`
}

// PodLogMonitorStatus defines the observed state of PodLogMonitor.
type PodLogMonitorStatus struct {
        // INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
        // Important: Run "make" to regenerate code after modifying this file
        LastRestartedPodName string `json:"lastRestartedPodName,omitempty"`
        LastRestartTime metav1.Time `json:"lastRestartTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// PodLogMonitor is the Schema for the podlogmonitors API.
type PodLogMonitor struct {
        metav1.TypeMeta   `json:",inline"`
        metav1.ObjectMeta `json:"metadata,omitempty"`

        Spec   PodLogMonitorSpec   `json:"spec,omitempty"`
        Status PodLogMonitorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PodLogMonitorList contains a list of PodLogMonitor.
type PodLogMonitorList struct {
        metav1.TypeMeta `json:",inline"`
        metav1.ListMeta `json:"metadata,omitempty"`
        Items           []PodLogMonitor `json:"items"`
}

func init() {
        SchemeBuilder.Register(&PodLogMonitor{}, &PodLogMonitorList{})
}
