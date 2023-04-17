// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package delivery

import (
	"context"
	"testing"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/conditions"
	"github.com/open-component-model/git-controller/pkg/providers/fakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	ocmv1 "github.com/open-component-model/ocm-controller/api/v1alpha1"

	"github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg"
)

func TestSyncReconciler(t *testing.T) {
	snapshot := DefaultSnapshot.DeepCopy()
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
	repository := &mpasv1alpha1.Repository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repository",
			Namespace: "default",
		},
		Spec: mpasv1alpha1.RepositorySpec{
			Provider:       "github",
			Owner:          "Skarlso",
			RepositoryName: "test",
			Credentials: mpasv1alpha1.Credentials{
				SecretRef: v1.LocalObjectReference{
					Name: secret.Name,
				},
			},
			Interval:                 metav1.Duration{Duration: 10 * time.Second},
			Visibility:               "public",
			ExistingRepositoryPolicy: mpasv1alpha1.ExistingRepositoryPolicyAdopt,
		},
	}
	sync := &v1alpha1.Sync{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "git-test",
			Namespace: "default",
		},
		Spec: v1alpha1.SyncSpec{
			SnapshotRef: v1.LocalObjectReference{
				Name: snapshot.Name,
			},
			RepositoryRef: v1.LocalObjectReference{
				Name: repository.Name,
			},
			Branch: "main",
			CommitTemplate: &v1alpha1.CommitTemplate{
				Name:    "Skarlso",
				Email:   "email@mail.com",
				Message: "This is my message",
			},
			SubPath: "./subpath",
			Prune:   true,
		},
	}

	client := env.FakeKubeClient(WithObjets(sync, snapshot, secret, repository), WithAddToScheme(ocmv1.AddToScheme), WithAddToScheme(mpasv1alpha1.AddToScheme))
	m := &mockGit{
		digest: "test-digest",
	}
	gsr := SyncReconciler{
		Client: client,
		Scheme: env.scheme,
		Git:    m,
	}

	_, err := gsr.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: sync.Namespace,
			Name:      sync.Name,
		},
	})
	require.NoError(t, err)

	err = client.Get(context.Background(), types.NamespacedName{
		Name:      sync.Name,
		Namespace: sync.Namespace,
	}, sync)
	require.NoError(t, err)

	assert.Equal(t, "test-digest", sync.Status.Digest)
	assert.True(t, conditions.IsTrue(sync, meta.ReadyCondition))
}

func TestSyncReconcilerWithAutomaticPullRequest(t *testing.T) {
	snapshot := DefaultSnapshot.DeepCopy()
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
	repository := &mpasv1alpha1.Repository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-repository",
			Namespace: "default",
		},
		Spec: mpasv1alpha1.RepositorySpec{
			Provider:       "github",
			Owner:          "Skarlso",
			RepositoryName: "test",
			Credentials: mpasv1alpha1.Credentials{
				SecretRef: v1.LocalObjectReference{
					Name: secret.Name,
				},
			},
			Interval:                 metav1.Duration{Duration: 10 * time.Second},
			Visibility:               "public",
			ExistingRepositoryPolicy: mpasv1alpha1.ExistingRepositoryPolicyAdopt,
		},
	}
	sync := &v1alpha1.Sync{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "git-test",
			Namespace: "default",
		},
		Spec: v1alpha1.SyncSpec{
			SnapshotRef: v1.LocalObjectReference{
				Name: snapshot.Name,
			},
			RepositoryRef: v1.LocalObjectReference{
				Name: repository.Name,
			},
			Branch: "main",
			CommitTemplate: &v1alpha1.CommitTemplate{
				Name:    "Skarlso",
				Email:   "email@mail.com",
				Message: "This is my message",
			},
			SubPath:                      "./subpath",
			Prune:                        true,
			AutomaticPullRequestCreation: true,
		},
	}

	client := env.FakeKubeClient(WithObjets(sync, snapshot, secret, repository), WithAddToScheme(ocmv1.AddToScheme), WithAddToScheme(mpasv1alpha1.AddToScheme))
	m := &mockGit{
		digest: "test-digest",
	}
	fakeProvider := fakes.NewProvider()

	gsr := SyncReconciler{
		Client:   client,
		Scheme:   env.scheme,
		Git:      m,
		Provider: fakeProvider,
	}

	_, err := gsr.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: sync.Namespace,
			Name:      sync.Name,
		},
	})
	require.NoError(t, err)

	err = client.Get(context.Background(), types.NamespacedName{
		Name:      sync.Name,
		Namespace: sync.Namespace,
	}, sync)
	require.NoError(t, err)

	assert.Equal(t, "test-digest", sync.Status.Digest)
	assert.True(t, conditions.IsTrue(sync, meta.ReadyCondition))

	args, err := fakeProvider.CreatePullRequestCallArgsForNumber(0)
	require.NoError(t, err)

	branch := args[0]
	assert.NotEmpty(t, branch.(string))
}

type mockGit struct {
	digest string
	err    error
}

func (g *mockGit) Push(ctx context.Context, opts *pkg.PushOptions) (string, error) {
	return g.digest, g.err
}
