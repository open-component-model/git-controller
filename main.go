// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Open Component Model contributors.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"flag"
	"os"

	"github.com/fluxcd/pkg/runtime/events"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/open-component-model/ocm-controller/api/v1alpha1"
	"github.com/open-component-model/ocm-controller/pkg/oci"

	deliveryv1alpha1 "github.com/open-component-model/git-controller/apis/delivery/v1alpha1"
	mpasv1alpha1 "github.com/open-component-model/git-controller/apis/mpas/v1alpha1"
	"github.com/open-component-model/git-controller/controllers/delivery"
	mpascontrollers "github.com/open-component-model/git-controller/controllers/mpas"
	"github.com/open-component-model/git-controller/pkg/gogit"
	"github.com/open-component-model/git-controller/pkg/providers/gitea"
	"github.com/open-component-model/git-controller/pkg/providers/github"
	"github.com/open-component-model/git-controller/pkg/providers/gitlab"
	//+kubebuilder:scaffold:imports
)

const controllerName = "git-controller"

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(deliveryv1alpha1.AddToScheme(scheme))
	utilruntime.Must(mpasv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var (
		metricsAddr               string
		enableLeaderElection      bool
		eventsAddr                string
		probeAddr                 string
		storagePath               string
		ociRegistryAddr           string
		ociRegistryCertSecretName string
		ociRegistryNamespace      string
	)
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.StringVar(&storagePath, "storage-path", "/data", "The location which to use for temporary storage. Should be mounted into the pod.")
	flag.StringVar(&ociRegistryAddr, "oci-registry-addr", ":5000", "The address of the OCI registry.")
	flag.StringVar(&ociRegistryCertSecretName, "certificate-secret-name", "ocm-registry-tls-certs", "")
	flag.StringVar(&ociRegistryNamespace, "oci-registry-namespace", "ocm-system", "The namespace in which the registry is running in.")
	flag.StringVar(&eventsAddr, "events-addr", "", "The address of the events receiver.")

	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	const metricsServerPort = 9443
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   metricsServerPort,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "76b5aa10.ocm.software",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	cache := oci.NewClient(ociRegistryAddr,
		oci.WithClient(mgr.GetClient()),
		oci.WithNamespace(ociRegistryNamespace),
		oci.WithCertificateSecret(ociRegistryCertSecretName),
	)
	gitClient := gogit.NewGoGit(ctrl.Log, cache)
	giteaProvider := gitea.NewClient(mgr.GetClient(), nil)
	gitlabProvider := gitlab.NewClient(mgr.GetClient(), giteaProvider)
	githubProvider := github.NewClient(mgr.GetClient(), gitlabProvider)

	var eventsRecorder *events.Recorder
	if eventsRecorder, err = events.NewRecorder(mgr, ctrl.Log, eventsAddr, controllerName); err != nil {
		setupLog.Error(err, "unable to create event recorder")
		os.Exit(1)
	}

	if err = (&delivery.SyncReconciler{
		EventRecorder: eventsRecorder,
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		Git:           gitClient,
		Provider:      githubProvider,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Sync")
		os.Exit(1)
	}

	if err = (&mpascontrollers.RepositoryReconciler{
		EventRecorder: eventsRecorder,
		Client:        mgr.GetClient(),
		Scheme:        mgr.GetScheme(),
		Provider:      githubProvider,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Repository")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
