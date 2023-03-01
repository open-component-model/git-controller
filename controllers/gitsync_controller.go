// Copyright 2022.
// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/fluxcd/pkg/runtime/patch"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ocmv1 "github.com/open-component-model/ocm-controller/api/v1alpha1"

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
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=snapshots,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=snapshots/status,verbs=get;update;patch

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

	if gitSync.Status.Digest != "" {
		log.Info("GitSync object already synced; status contains digest information", "digest", gitSync.Status.Digest)
		return ctrl.Result{}, nil
	}

	snapshot := &ocmv1.Snapshot{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: gitSync.Spec.SnapshotRef.Namespace,
		Name:      gitSync.Spec.SnapshotRef.Name,
	}, snapshot); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to find snapshot: %w", err)
	}
	authSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: gitSync.Spec.AuthRef.Namespace,
		Name:      gitSync.Spec.AuthRef.Name,
	}, authSecret); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to find authentication secret: %w", err)
	}

	// trim any trailing `/` and then just add.
	log.V(4).Info("crafting artifact URL to download from", "url", snapshot.Status.RepositoryURL)
	opts := &providers.PushOptions{
		URL:      gitSync.Spec.URL,
		Message:  gitSync.Spec.CommitTemplate.Message,
		Name:     gitSync.Spec.CommitTemplate.Name,
		Email:    gitSync.Spec.CommitTemplate.Email,
		Snapshot: snapshot,
		Branch:   gitSync.Spec.Branch,
		SubPath:  gitSync.Spec.SubPath,
	}
	r.parseAuthSecret(authSecret, opts)

	digest, err := r.Git.Push(ctx, opts)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to push to git repository: %w", err)
	}
	// Initialize the patch helper.
	patchHelper, err := patch.NewHelper(gitSync, r.Client)
	if err != nil {
		return ctrl.Result{
			RequeueAfter: 1 * time.Minute,
		}, fmt.Errorf("failed to create patch helper: %w", err)
	}

	gitSync.Status.Digest = digest
	if err := patchHelper.Patch(ctx, gitSync); err != nil {
		return ctrl.Result{
			RequeueAfter: 1 * time.Minute,
		}, fmt.Errorf("failed to patch git sync object: %w", err)
	}
	log.V(4).Info("patch successful")

	return ctrl.Result{}, nil
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
