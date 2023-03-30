// Copyright 2022.
// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"errors"
	"fmt"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/conditions"
	"github.com/fluxcd/pkg/runtime/patch"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ocmv1 "github.com/open-component-model/ocm-controller/api/v1alpha1"

	"github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
	providers "github.com/open-component-model/git-controller/pkg"
)

// SyncReconciler reconciles a Sync object
type SyncReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	Git providers.Git
}

//+kubebuilder:rbac:groups=delivery.ocm.software,resources=syncs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=syncs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=syncs/finalizers,verbs=update
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=ocmresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=snapshots,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=snapshots/status,verbs=get;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var (
		result ctrl.Result
		retErr error
	)

	log := log.FromContext(ctx)
	log.V(4).Info("starting reconcile loop for snapshot")
	obj := &v1alpha1.Sync{}
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get git sync object: %w", err)
	}
	log.V(4).Info("found reconciling object", "sync", obj)

	// The replication controller doesn't need a shouldReconcile, because it should always reconcile,
	// that is its purpose.
	patchHelper, err := patch.NewHelper(obj, r.Client)
	if err != nil {
		retErr = errors.Join(retErr, err)
		conditions.MarkFalse(obj, meta.ReadyCondition, v1alpha1.PatchFailedReason, err.Error())

		return ctrl.Result{}, retErr
	}

	// Always attempt to patch the object and status after each reconciliation.
	defer func() {
		// Patching has not been set up, or the controller errored earlier.
		if patchHelper == nil {
			return
		}

		if condition := conditions.Get(obj, meta.StalledCondition); condition != nil && condition.Status == metav1.ConditionTrue {
			conditions.Delete(obj, meta.ReconcilingCondition)
		}

		// Check if it's a successful reconciliation.
		// We don't set Requeue in case of error, so we can safely check for Requeue.
		if result.RequeueAfter == obj.GetRequeueAfter() && !result.Requeue && retErr == nil {
			// Remove the reconciling condition if it's set.
			conditions.Delete(obj, meta.ReconcilingCondition)

			// Set the return err as the ready failure message if the resource is not ready, but also not reconciling or stalled.
			if ready := conditions.Get(obj, meta.ReadyCondition); ready != nil && ready.Status == metav1.ConditionFalse && !conditions.IsStalled(obj) {
				retErr = errors.New(conditions.GetMessage(obj, meta.ReadyCondition))
			}
		}

		// If still reconciling then reconciliation did not succeed, set to ProgressingWithRetry to
		// indicate that reconciliation will be retried.
		if conditions.IsReconciling(obj) {
			reconciling := conditions.Get(obj, meta.ReconcilingCondition)
			reconciling.Reason = meta.ProgressingWithRetryReason
			conditions.Set(obj, reconciling)
		}

		// If not reconciling or stalled than mark Ready=True
		if !conditions.IsReconciling(obj) &&
			!conditions.IsStalled(obj) &&
			retErr == nil {
			conditions.MarkTrue(obj, meta.ReadyCondition, meta.SucceededReason, "Reconciliation success")
		}

		// Set status observed generation option if the component is stalled or ready.
		if conditions.IsStalled(obj) || conditions.IsReady(obj) {
			obj.Status.ObservedGeneration = obj.Generation
		}

		// Update the object.
		if err := patchHelper.Patch(ctx, obj); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}()

	// it's important that this happens here so any residual status condition can be overwritten / set.
	if obj.Status.Digest != "" {
		log.Info("Sync object already synced; status contains digest information", "digest", obj.Status.Digest)
		return ctrl.Result{}, nil
	}

	snapshot := &ocmv1.Snapshot{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: obj.Spec.SnapshotRef.Namespace,
		Name:      obj.Spec.SnapshotRef.Name,
	}, snapshot); err != nil {
		retErr = fmt.Errorf("failed to find snapshot: %w", err)
		conditions.MarkFalse(obj, meta.ReadyCondition, v1alpha1.SnapshotGetFailedReason, retErr.Error())

		return ctrl.Result{}, retErr
	}

	authSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{
		Namespace: obj.Spec.AuthRef.Namespace,
		Name:      obj.Spec.AuthRef.Name,
	}, authSecret); err != nil {
		retErr = fmt.Errorf("failed to find authentication secret: %w", err)
		conditions.MarkFalse(obj, meta.ReadyCondition, v1alpha1.CredentialsNotFoundReason, retErr.Error())

		return ctrl.Result{}, retErr
	}

	// trim any trailing `/` and then just add.
	log.V(4).Info("crafting artifact URL to download from", "url", snapshot.Status.RepositoryURL)
	opts := &providers.PushOptions{
		URL:      obj.Spec.URL,
		Message:  obj.Spec.CommitTemplate.Message,
		Name:     obj.Spec.CommitTemplate.Name,
		Email:    obj.Spec.CommitTemplate.Email,
		Snapshot: snapshot,
		Branch:   obj.Spec.Branch,
		SubPath:  obj.Spec.SubPath,
	}

	r.parseAuthSecret(authSecret, opts)

	digest, err := r.Git.Push(ctx, opts)
	if err != nil {
		retErr = fmt.Errorf("failed to push to git repository: %w", err)
		conditions.MarkFalse(obj, meta.ReadyCondition, v1alpha1.GitRepositoryPushFailedReason, retErr.Error())

		return ctrl.Result{}, retErr
	}

	obj.Status.Digest = digest

	// Remove any stale Ready condition, most likely False, set above. Its value
	// is derived from the overall result of the reconciliation in the deferred
	// block at the very end.
	conditions.Delete(obj, meta.ReadyCondition)

	return result, retErr
}

// SetupWithManager sets up the controller with the Manager.
func (r *SyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Sync{}).
		Complete(r)
}

func (r *SyncReconciler) parseAuthSecret(secret *corev1.Secret, opts *providers.PushOptions) {
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
