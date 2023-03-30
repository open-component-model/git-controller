// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SecretRef is a reference to a secret in the same namespace as the referencing object.
type SecretRef struct {
	Name string `json:"name"`
}

// Credentials contains ways of authenticating the creation of a repository.
type Credentials struct {
	SecretRef SecretRef `json:"secretRef"`
}

// ExistingRepositoryPolicy defines what to do in case a requested repository already exists.
type ExistingRepositoryPolicy string

var (
	// ExistingRepositoryPolicyAdopt will use the repository if it exists.
	ExistingRepositoryPolicyAdopt ExistingRepositoryPolicy = "adopt"
	// ExistingRepositoryPolicyFail will fail if the requested repository already exists.
	ExistingRepositoryPolicyFail ExistingRepositoryPolicy = "fail"
)

// RepositorySpec defines the desired state of Repository
type RepositorySpec struct {
	//+required
	Provider string `json:"provider"`
	//+required
	Owner string `json:"owner"`
	//+required
	RepositoryName string `json:"repositoryName"`
	//+required
	Credentials Credentials `json:"credentials"`

	//+optional
	Interval metav1.Duration `json:"interval,omitempty"`
	//+optional
	//+kubebuilder:default:=private
	Visibility string `json:"visibility,omitempty"`
	//+kubebuilder:default:=true
	IsOrganization bool `json:"isOrganization,omitempty"`
	//+optional
	Domain string `json:"domain,omitempty"`
	//+optional
	Maintainers []string `json:"maintainers,omitempty"`
	//+optional
	//+kubebuilder:default:=true
	AutomaticPullRequestCreation bool `json:"automaticPullRequestCreation,omitempty"`
	//+optional
	//+kubebuilder:default:=adopt
	//+kubebuilder:validation:Enum=adopt;fail
	ExistingRepositoryPolicy ExistingRepositoryPolicy `json:"existingRepositoryPolicy,omitempty"`
}

// RepositoryStatus defines the observed state of Repository
type RepositoryStatus struct {
	// ObservedGeneration is the last reconciled generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status",description=""
	// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].message",description=""
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// GetConditions returns the conditions of the ComponentVersion.
func (in *Repository) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the conditions of the ComponentVersion.
func (in *Repository) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

// GetRequeueAfter returns the duration after which the ComponentVersion must be
// reconciled again.
func (in Repository) GetRequeueAfter() time.Duration {
	return in.Spec.Interval.Duration
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Repository is the Schema for the repositories API
type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RepositorySpec   `json:"spec,omitempty"`
	Status RepositoryStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RepositoryList contains a list of Repository
type RepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Repository `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Repository{}, &RepositoryList{})
}
