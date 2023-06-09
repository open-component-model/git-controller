// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package mpas

import (
	"context"
	"errors"
	"fmt"

	eventv1 "github.com/fluxcd/pkg/apis/event/v1beta1"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/conditions"
	"github.com/fluxcd/pkg/runtime/patch"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	kuberecorder "k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg/event"
	"github.com/open-component-model/git-controller/pkg/providers"
)

// RepositoryReconciler reconciles a Repository object
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
	logger := log.FromContext(ctx).WithName("repository")
	logger.V(4).Info("entering repository loop...")

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

		// Set status observed generation option if the component is stalled or ready.
		if conditions.IsReady(obj) {
			obj.Status.ObservedGeneration = obj.Generation
		}

		// Update the object.
		if perr := patchHelper.Patch(ctx, obj); perr != nil {
			err = errors.Join(err, perr)
		}
	}()

	return r.reconcile(ctx, obj)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RepositoryReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mpasv1alpha1.Repository{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}

func (r *RepositoryReconciler) reconcile(ctx context.Context, obj *mpasv1alpha1.Repository) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("creating or adopting repository")
	if err := r.Provider.CreateRepository(ctx, *obj); err != nil {
		event.New(r.EventRecorder, obj, eventv1.EventSeverityError, err.Error(), nil)
		conditions.MarkFalse(obj, meta.ReadyCondition, mpasv1alpha1.RepositoryCreateFailedReason, err.Error())

		return ctrl.Result{}, fmt.Errorf("failed to create repository: %w", err)
	}

	logger.Info("updating branch protection rules")
	if err := r.Provider.CreateBranchProtection(ctx, *obj); err != nil {
		if errors.Is(err, providers.NotSupportedError) {
			// ignore and return without branch protection rules.
			logger.Error(err, fmt.Sprintf("provider %s does not support updating branch protection rules", obj.Spec.Provider))

			return ctrl.Result{}, nil
		}

		conditions.MarkFalse(obj, meta.ReadyCondition, mpasv1alpha1.UpdatingBranchProtectionFailedReason, err.Error())
		event.New(r.EventRecorder, obj, eventv1.EventSeverityError, err.Error(), nil)

		return ctrl.Result{}, fmt.Errorf("failed to update branch protection rules: %w", err)
	}

	logger.Info("done reconciling repository")
	conditions.MarkTrue(obj, meta.ReadyCondition, meta.SucceededReason, "Reconciliation success")
	event.New(r.EventRecorder, obj, eventv1.EventSeverityInfo, "Reconciliation success", nil)

	return ctrl.Result{}, nil
}
