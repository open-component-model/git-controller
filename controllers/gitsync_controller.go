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

package controllers

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/open-component-model/git-sync-controller/api/v1alpha1"
	providers "github.com/open-component-model/git-sync-controller/pkg"
)

// GitSyncReconciler reconciles a GitSync object
type GitSyncReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Git providers.Git
}

//+kubebuilder:rbac:groups=delivery.ocm.software,resources=gitsyncs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=gitsyncs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=gitsyncs/finalizers,verbs=update
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=ocmresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *GitSyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.V(4).Info("starting reconcile loop for snapshot")
	gitSync := &v1alpha1.GitSync{}
	if err := r.Get(ctx, req.NamespacedName, gitSync); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get git sync object: %w", err)
	}
	log.V(4).Info("found reconciling object", "gitSync", gitSync)

	snapshot := &v1alpha1.OCMSnapshot{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: gitSync.Spec.SnapshotRef.Namespace,
		Name:      gitSync.Spec.SnapshotRef.Name,
	}, snapshot); err != nil {
		return requeue(gitSync.Spec.Interval), fmt.Errorf("failed to find snapshot: %w", err)
	}
	authSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: gitSync.Spec.AuthRef.Namespace,
		Name:      gitSync.Spec.AuthRef.Name,
	}, snapshot); err != nil {
		return requeue(gitSync.Spec.Interval), fmt.Errorf("failed to find authentication secret: %w", err)
	}
	opts := &providers.PushOptions{
		URL:         gitSync.Spec.URL,
		Message:     gitSync.Spec.CommitTemplate.Message,
		Name:        gitSync.Spec.CommitTemplate.Name,
		Email:       gitSync.Spec.CommitTemplate.Email,
		SnapshotURL: snapshot.Spec.URL,
		Branch:      gitSync.Spec.Branch,
	}
	r.parseAuthSecret(authSecret, opts)

	if err := r.Git.Push(ctx, opts); err != nil {
		return requeue(gitSync.Spec.Interval), fmt.Errorf("failed to push to git repository: %w", err)
	}

	// TODO: update with status so we know that GitSync has been processed.

	return ctrl.Result{}, nil
}

func requeue(seconds time.Duration) ctrl.Result {
	return ctrl.Result{
		RequeueAfter: seconds * time.Second,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *GitSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.GitSync{}).
		Complete(r)
}

func (r *GitSyncReconciler) parseAuthSecret(secret *corev1.Secret, opts *providers.PushOptions) {
	if _, ok := secret.Data["identity"]; ok {
		opts.Auth = &providers.Auth{
			SSH: &providers.SSH{
				PemBytes: secret.Data["identity"],
				User:     string(secret.Data["username"]),
				Password: string(secret.Data["password"]),
			},
		}
		return
	}
	// default to basic auth.
	opts.Auth = &providers.Auth{
		BasicAuth: &providers.BasicAuth{
			Username: string(secret.Data["username"]),
			Password: string(secret.Data["password"]),
		},
	}
}
