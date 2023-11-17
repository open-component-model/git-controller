// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package delivery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/patch"
	rreconcile "github.com/fluxcd/pkg/runtime/reconcile"
	"github.com/open-component-model/ocm-controller/pkg/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kuberecorder "k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	ocmv1 "github.com/open-component-model/ocm-controller/api/v1alpha1"

	"github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg"
	"github.com/open-component-model/git-controller/pkg/providers"
)

// SyncReconciler reconciles a Sync object
type SyncReconciler struct {
	client.Client
	kuberecorder.EventRecorder
	Scheme *runtime.Scheme

	Git      pkg.Git
	Provider providers.Provider
}

//+kubebuilder:rbac:groups=delivery.ocm.software,resources=syncs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=syncs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=syncs/finalizers,verbs=update
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=ocmresources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=snapshots,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=delivery.ocm.software,resources=snapshots/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, err error) {
	obj := &v1alpha1.Sync{}
	if err = r.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get git sync object: %w", err)
	}

	// The replication controller doesn't need a shouldReconcile, because it should always reconcile,
	// that is its purpose.
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

	// it's important that this happens here so any residual status condition can be overwritten / set.
	if obj.Status.Digest != "" {
		status.MarkReady(r.EventRecorder, obj, "Digest already reconciled")

		return ctrl.Result{}, nil
	}

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

	snapshot := &ocmv1.Snapshot{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: obj.Namespace,
		Name:      obj.Spec.SnapshotRef.Name,
	}, snapshot); err != nil {
		err = fmt.Errorf("failed to find snapshot: %w", err)
		status.MarkNotReady(r.EventRecorder, obj, v1alpha1.SnapshotGetFailedReason, err.Error())

		return ctrl.Result{}, err
	}

	namespace := obj.Spec.RepositoryRef.Namespace
	if namespace == "" {
		namespace = obj.Namespace
	}

	repository := &mpasv1alpha1.Repository{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      obj.Spec.RepositoryRef.Name,
	}, repository); err != nil {
		err = fmt.Errorf("failed to find repository: %w", err)
		status.MarkNotReady(r.EventRecorder, obj, v1alpha1.RepositoryGetFailedReason, err.Error())

		return ctrl.Result{}, err
	}

	authSecret := &corev1.Secret{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: repository.Namespace,
		Name:      repository.Spec.Credentials.SecretRef.Name,
	}, authSecret); err != nil {
		err = fmt.Errorf("failed to find authentication secret: %w", err)
		status.MarkNotReady(r.EventRecorder, obj, v1alpha1.CredentialsNotFoundReason, err.Error())

		return ctrl.Result{}, err
	}

	baseBranch := obj.Spec.CommitTemplate.BaseBranch
	if baseBranch == "" {
		baseBranch = "main"
	}

	targetBranch := obj.Spec.CommitTemplate.TargetBranch
	if targetBranch == "" && obj.Spec.AutomaticPullRequestCreation {
		targetBranch = fmt.Sprintf("branch-%d", time.Now().Unix())
	} else if targetBranch == "" && !obj.Spec.AutomaticPullRequestCreation {
		err = fmt.Errorf("branch cannot be empty if automatic pull request creation is not enabled")
		status.MarkNotReady(r.EventRecorder, obj, v1alpha1.GitRepositoryPushFailedReason, err.Error())

		return ctrl.Result{}, err
	}

	rreconcile.ProgressiveStatus(false, obj, meta.ProgressingReason, "preparing to push snapshot content with base branch %s and target %s", baseBranch, targetBranch)

	opts := &pkg.PushOptions{
		URL:          repository.GetRepositoryURL(),
		Message:      obj.Spec.CommitTemplate.Message,
		Name:         obj.Spec.CommitTemplate.Name,
		Email:        obj.Spec.CommitTemplate.Email,
		Snapshot:     snapshot,
		BaseBranch:   baseBranch,
		TargetBranch: targetBranch,
		SubPath:      obj.Spec.SubPath,
	}

	r.parseAuthSecret(authSecret, opts)

	var digest string
	digest, err = r.Git.Push(ctx, opts)
	if err != nil {
		err = fmt.Errorf("failed to push to git repository: %w", err)
		status.MarkNotReady(r.EventRecorder, obj, v1alpha1.GitRepositoryPushFailedReason, err.Error())

		return ctrl.Result{}, err
	}

	obj.Status.Digest = digest

	if obj.Spec.AutomaticPullRequestCreation {
		rreconcile.ProgressiveStatus(false, obj, meta.ProgressingReason, "creating pull request")

		id, err := r.Provider.CreatePullRequest(ctx, targetBranch, *obj, *repository)
		if err != nil {
			err = fmt.Errorf("failed to create pull request: %w", err)
			status.MarkNotReady(r.EventRecorder, obj, v1alpha1.CreatePullRequestFailedReason, err.Error())

			return ctrl.Result{}, err
		}

		obj.Status.PullRequestID = id
	}

	status.MarkReady(r.EventRecorder, obj, "Reconciliation success")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Sync{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}

func (r *SyncReconciler) parseAuthSecret(secret *corev1.Secret, opts *pkg.PushOptions) {
	if _, ok := secret.Data["identity"]; ok {
		opts.Auth = &pkg.Auth{
			SSH: &pkg.SSH{
				PemBytes: secret.Data["identity"],
				User:     string(secret.Data["username"]),
				Password: string(secret.Data["password"]),
			},
		}
		return
	}
	// default to basic auth.
	opts.Auth = &pkg.Auth{
		BasicAuth: &pkg.BasicAuth{
			Username: string(secret.Data["username"]),
			Password: string(secret.Data["password"]),
		},
	}
}
