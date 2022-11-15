// Copyright 2022.
// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

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
	CommitTemplate *CommitTemplate `json:"commitTemplate"`
	SubPath        string          `json:"subPath"`
	Prune          bool            `json:"prune,omitempty"`
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
