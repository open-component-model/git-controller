// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package mpas

import (
	"context"
	"errors"
	"fmt"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/patch"
	rreconcile "github.com/fluxcd/pkg/runtime/reconcile"
	"github.com/open-component-model/ocm-controller/pkg/status"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	kuberecorder "k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg/providers"
)

// RepositoryReconciler reconciles a Repository object.
type RepositoryReconciler struct {
	client.Client
	kuberecorder.EventRecorder
	Scheme   *runtime.Scheme
	Provider providers.Provider
}

//+kubebuilder:rbac:groups=mpas.ocm.software,resources=repositories,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mpas.ocm.software,resources=repositories/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mpas.ocm.software,resources=repositories/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *RepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	obj := &mpasv1alpha1.Repository{}

	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get component object: %w", err)
	}

	patchHelper := patch.NewSerialPatcher(obj, r.Client)

	// Always attempt to patch the object and status after each reconciliation.
	defer func() {
		// Patching has not been set up, or the controller errored earlier.
		if patchHelper == nil {
			return
		}

		if derr := status.UpdateStatus(ctx, patchHelper, obj, r.EventRecorder, obj.GetRequeueAfter()); derr != nil {
			err = errors.Join(err, derr)
		}
	}()

	// Starts the progression by setting ReconcilingCondition.
	// This will be checked in defer.
	// Should only be deleted on a success.
	rreconcile.ProgressiveStatus(false, obj, meta.ProgressingReason, "reconciliation in progress for resource: %s", obj.Name)

	if err := r.reconcile(ctx, obj); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mpasv1alpha1.Repository{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}

func (r *RepositoryReconciler) reconcile(ctx context.Context, obj *mpasv1alpha1.Repository) error {
	if obj.Generation != obj.Status.ObservedGeneration {
		rreconcile.ProgressiveStatus(
			false,
			obj,
			meta.ProgressingReason,
			"processing object: new generation %d -> %d",
			obj.Status.ObservedGeneration,
			obj.Generation,
		)
	}

	rreconcile.ProgressiveStatus(false, obj, meta.ProgressingReason, "creating repository: %s", obj.Name)

	if err := r.Provider.CreateRepository(ctx, *obj); err != nil {
		err := fmt.Errorf("failed to create repository: %w", err)
		status.MarkNotReady(r.EventRecorder, obj, mpasv1alpha1.RepositoryCreateFailedReason, err.Error())

		return err
	}

	rreconcile.ProgressiveStatus(false, obj, meta.ProgressingReason, "setting up branch protection rules: %s", obj.Name)

	if err := r.Provider.CreateBranchProtection(ctx, *obj); err != nil {
		if errors.Is(err, providers.ErrNotSupported) {
			status.MarkReady(r.EventRecorder, obj, "Successful reconciliation")

			// ignore and return without branch protection rules.
			return nil
		}

		err := fmt.Errorf("failed to update branch protection rules: %w", err)
		status.MarkNotReady(r.EventRecorder, obj, mpasv1alpha1.UpdatingBranchProtectionFailedReason, err.Error())

		return err
	}

	status.MarkReady(r.EventRecorder, obj, "Successful reconciliation")

	return nil
}
