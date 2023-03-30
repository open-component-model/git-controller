// Copyright 2022.
// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	ocmv1alpha1 "github.com/open-component-model/ocm-controller/api/v1alpha1"

	"github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
)

type testEnv struct {
	scheme *runtime.Scheme
	obj    []client.Object
}

// FakeKubeClientOption defines options to construct a fake kube client. There are some defaults involved.
// Scheme gets corev1 and v1alpha1 schemes by default. Anything that is passed in will override current
// defaults.
type FakeKubeClientOption func(testEnv *testEnv)

// WithAddToScheme adds the scheme
func WithAddToScheme(addToScheme func(s *runtime.Scheme) error) FakeKubeClientOption {
	return func(testEnv *testEnv) {
		if err := addToScheme(testEnv.scheme); err != nil {
			panic(err)
		}
	}
}

// WithObjects provides an option to set objects for the fake client.
func WithObjets(obj ...client.Object) FakeKubeClientOption {
	return func(testEnv *testEnv) {
		testEnv.obj = obj
	}
}

// FakeKubeClient creates a fake kube client with some defaults and optional arguments.
func (t *testEnv) FakeKubeClient(opts ...FakeKubeClientOption) client.Client {
	for _, o := range opts {
		o(t)
	}
	return fake.NewClientBuilder().WithScheme(t.scheme).WithObjects(t.obj...).Build()
}

var (
	DefaultComponent = &ocmv1alpha1.ComponentVersion{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-component",
			Namespace: "default",
		},
		Spec: ocmv1alpha1.ComponentVersionSpec{
			Interval:  metav1.Duration{Duration: 10 * time.Minute},
			Component: "github.com/open-component-model/test-component",
			Version: ocmv1alpha1.Version{
				Semver: "v0.0.1",
			},
			Repository: ocmv1alpha1.Repository{
				URL: "github.com/open-component-model/test",
			},
			Verify: []ocmv1alpha1.Signature{},
			References: ocmv1alpha1.ReferencesConfig{
				Expand: true,
			},
		},
	}
	DefaultSnapshot = &ocmv1alpha1.Snapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-snapshot",
			Namespace: "default",
		},
		Spec: ocmv1alpha1.SnapshotSpec{
			Identity: ocmv1alpha1.Identity{
				ocmv1alpha1.ComponentNameKey:    DefaultComponent.Name,
				ocmv1alpha1.ComponentVersionKey: "v0.0.1",
				ocmv1alpha1.ResourceNameKey:     "test-resource",
				ocmv1alpha1.ResourceVersionKey:  "v0.0.1",
			},
			Digest: "test-digest",
			Tag:    "v0.0.1",
		},
		Status: ocmv1alpha1.SnapshotStatus{},
	}
)

var env *testEnv

func TestMain(m *testing.M) {
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	env = &testEnv{
		scheme: scheme,
	}
	m.Run()
}
