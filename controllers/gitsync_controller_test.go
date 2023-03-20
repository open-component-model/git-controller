package controllers

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

	"github.com/open-component-model/git-sync-controller/api/v1alpha1"
	"github.com/open-component-model/git-sync-controller/pkg"
)

func TestGitSyncReconciler(t *testing.T) {
	cv := DefaultComponent.DeepCopy()
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
	gitSync := &v1alpha1.GitSync{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "git-sync-test",
			Namespace: "default",
		},
		Spec: v1alpha1.GitSyncSpec{
			ComponentRef: v1alpha1.Ref{
				Name:      cv.Name,
				Namespace: cv.Namespace,
			},
			SnapshotRef: v1alpha1.Ref{
				Name:      snapshot.Name,
				Namespace: snapshot.Namespace,
			},
			AuthRef: v1alpha1.Ref{
				Name:      secret.Name,
				Namespace: secret.Namespace,
			},
			URL:    "https://github.com/Skarlso/test",
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

	client := env.FakeKubeClient(WithObjets(gitSync, snapshot, cv, secret), WithAddToScheme(ocmv1.AddToScheme))
	m := &mockGit{
		digest: "test-digest",
	}
	gsr := GitSyncReconciler{
		Client: client,
		Scheme: env.scheme,
		Git:    m,
	}

	_, err := gsr.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{
			Namespace: gitSync.Namespace,
			Name:      gitSync.Name,
		},
	})
	require.NoError(t, err)

	err = client.Get(context.Background(), types.NamespacedName{
		Name:      gitSync.Name,
		Namespace: gitSync.Namespace,
	}, gitSync)
	require.NoError(t, err)

	assert.Equal(t, "test-digest", gitSync.Status.Digest)
	assert.True(t, conditions.IsTrue(gitSync, meta.ReadyCondition))
}

type mockGit struct {
	digest string
	err    error
}

func (g *mockGit) Push(ctx context.Context, opts *pkg.PushOptions) (string, error) {
	return g.digest, g.err
}
