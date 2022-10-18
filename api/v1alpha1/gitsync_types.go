/*
Copyright 2022.

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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Ref defines a name and namespace ref to any object.
type Ref struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// CommitTemplate defines the details of the commit to the external repository.
type CommitTemplate struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// GitSyncSpec defines the desired state of GitSync
type GitSyncSpec struct {
	ComponentRef   Ref             `json:"componentRef"`
	SnapshotRef    Ref             `json:"snapshotRef"`
	Interval       time.Duration   `json:"interval"`
	URL            string          `json:"url"`
	Branch         string          `json:"branch"`
	AuthRef        Ref             `json:"authRef"`
	Destination    string          `json:"destination"`
	CommitTemplate *CommitTemplate `json:"commitTemplate"`
}

// GitSyncStatus defines the observed state of GitSync
type GitSyncStatus struct {
	Digest string `json:"digest,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GitSync is the Schema for the gitsyncs API
type GitSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitSyncSpec   `json:"spec,omitempty"`
	Status GitSyncStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GitSyncList contains a list of GitSync
type GitSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GitSync{}, &GitSyncList{})
}
