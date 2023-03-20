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
	Interval       metav1.Duration `json:"interval"`
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

	// ObservedGeneration is the last reconciled generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// +optional
	// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
	// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// GetConditions returns the conditions of the ComponentVersion.
func (in *GitSync) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the conditions of the ComponentVersion.
func (in *GitSync) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

// GetRequeueAfter returns the duration after which the ComponentVersion must be
// reconciled again.
func (in GitSync) GetRequeueAfter() time.Duration {
	return in.Spec.Interval.Duration
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
