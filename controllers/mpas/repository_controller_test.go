// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package mpas

import (
	"context"
	"testing"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/conditions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	ocmv1 "github.com/open-component-model/ocm-controller/api/v1alpha1"

	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg/providers/fakes"
)

func TestRepositoryReconciler(t *testing.T) {
	repository := DefaultRepository.DeepCopy()
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "auth-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"username": []byte("username"),
			"password": []byte("password"),
		},
	}

	client := env.FakeKubeClient(WithAddToScheme(mpasv1alpha1.AddToScheme), WithObjets(repository, secret), WithAddToScheme(ocmv1.AddToScheme))
	fakeProvider := fakes.NewProvider()
	controller := &RepositoryReconciler{
		Client:   client,
		Scheme:   env.scheme,
		Provider: fakeProvider,
	}

	_, err := controller.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: repository.Namespace,
			Name:      repository.Name,
		},
	})
	require.NoError(t, err)

	err = client.Get(context.Background(), types.NamespacedName{
		Namespace: repository.Namespace,
		Name:      repository.Name,
	}, repository)
	require.NoError(t, err)

	assert.True(t, conditions.IsTrue(repository, meta.ReadyCondition))
}
