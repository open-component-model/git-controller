// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package delivery

import (
	"context"
	"errors"
	"fmt"
	"time"

	eventv1 "github.com/fluxcd/pkg/apis/event/v1beta1"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/fluxcd/pkg/runtime/conditions"
	"github.com/fluxcd/pkg/runtime/patch"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	kuberecorder "k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	ocmv1 "github.com/open-component-model/ocm-controller/api/v1alpha1"

	"github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/pkg"
	"github.com/open-component-model/git-controller/pkg/event"
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

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SyncReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, err error) {
	log := log.FromContext(ctx)

	obj := &v1alpha1.Sync{}
	if err = r.Get(ctx, req.NamespacedName, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		return ctrl.Result{}, fmt.Errorf("failed to get git sync object: %w", err)
	}
	log.V(4).Info("found reconciling object", "sync", obj)

	// The replication controller doesn't need a shouldReconcile, because it should always reconcile,
	// that is its purpose.
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

	// it's important that this happens here so any residual status condition can be overwritten / set.
	if obj.Status.Digest != "" {
		log.Info("Sync object already synced; status contains digest information", "digest", obj.Status.Digest)
		event.New(r.EventRecorder, obj, eventv1.EventSeverityInfo, fmt.Sprintf("sync object already synced with digest %s", obj.Status.Digest), nil)
		conditions.MarkTrue(obj, meta.ReadyCondition, meta.SucceededReason, "Reconciliation success")

		return ctrl.Result{}, nil
	}

	snapshot := &ocmv1.Snapshot{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: obj.Namespace,
		Name:      obj.Spec.SnapshotRef.Name,
	}, snapshot); err != nil {
		err = fmt.Errorf("failed to find snapshot: %w", err)
		r.markAndEmitEvent(obj, v1alpha1.SnapshotGetFailedReason, err)

		return ctrl.Result{}, err
	}

	log.V(4).Info("found target snapshot")

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
		r.markAndEmitEvent(obj, v1alpha1.RepositoryGetFailedReason, err)

		return ctrl.Result{}, err
	}

	log.V(4).Info("found target repository")

	authSecret := &corev1.Secret{}
	if err = r.Get(ctx, types.NamespacedName{
		Namespace: repository.Namespace,
		Name:      repository.Spec.Credentials.SecretRef.Name,
	}, authSecret); err != nil {
		err = fmt.Errorf("failed to find authentication secret: %w", err)
		r.markAndEmitEvent(obj, v1alpha1.CredentialsNotFoundReason, err)

		return ctrl.Result{}, err
	}

	log.V(4).Info("found authentication secret")

	baseBranch := obj.Spec.CommitTemplate.BaseBranch
	if baseBranch == "" {
		baseBranch = "main"
	}

	targetBranch := obj.Spec.CommitTemplate.TargetBranch
	if targetBranch == "" && obj.Spec.AutomaticPullRequestCreation {
		targetBranch = fmt.Sprintf("branch-%d", time.Now().Unix())
	} else if targetBranch == "" && !obj.Spec.AutomaticPullRequestCreation {
		err = fmt.Errorf("branch cannot be empty if automatic pull request creation is not enabled")
		r.markAndEmitEvent(obj, v1alpha1.GitRepositoryPushFailedReason, err)

		return ctrl.Result{}, err
	}

	log.Info("preparing to push snapshot content", "base", baseBranch, "target", targetBranch)

	// trim any trailing `/` and then just add.
	log.V(4).Info("crafting artifact URL to download from", "url", snapshot.Status.RepositoryURL)
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
		r.markAndEmitEvent(obj, v1alpha1.GitRepositoryPushFailedReason, err)

		return ctrl.Result{}, err
	}

	log.Info("target content pushed with digest", "base", baseBranch, "target", targetBranch, "digest", digest)

	obj.Status.Digest = digest

	if obj.Spec.AutomaticPullRequestCreation {
		log.Info("automatic pull-request creation is enabled, preparing to create a pull request")

		id, err := r.Provider.CreatePullRequest(ctx, targetBranch, *obj, *repository)
		if err != nil {
			err = fmt.Errorf("failed to create pull request: %w", err)
			r.markAndEmitEvent(obj, v1alpha1.CreatePullRequestFailedReason, err)

			return ctrl.Result{}, err
		}

		obj.Status.PullRequestID = id
	}

	log.Info("successfully reconciled sync object")
	conditions.MarkTrue(obj, meta.ReadyCondition, meta.SucceededReason, "Reconciliation success")
	event.New(r.EventRecorder, obj, eventv1.EventSeverityInfo, "Reconciliation success", nil)

	return ctrl.Result{}, nil
}

func (r *SyncReconciler) markAndEmitEvent(obj *v1alpha1.Sync, reason string, err error) {
	event.New(r.EventRecorder, obj, eventv1.EventSeverityError, err.Error(), nil)
	conditions.MarkFalse(obj, meta.ReadyCondition, reason, err.Error())
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
