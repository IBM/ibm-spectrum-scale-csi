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

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"go.uber.org/zap/zapcore"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	csiv1 "github.com/IBM/ibm-spectrum-scale-csi/operator/api/v1"
	"github.com/IBM/ibm-spectrum-scale-csi/operator/controllers"
	"github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/config"
	"github.com/robfig/cron/v3"

	configv1 "github.com/openshift/api/config/v1"
	securityv1 "github.com/openshift/api/security/v1"
	//+kubebuilder:scaffold:imports
)

const OCPControllerNamespace = "openshift-controller-manager"

// gitCommit that is injected via go build -ldflags "-X main.gitCommit=$(git rev-parse HEAD)"
var (
	gitCommit string
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(csiv1.AddToScheme(scheme))
	utilruntime.Must(securityv1.AddToScheme(scheme))
	utilruntime.Must(configv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

// getMetricsBindAddress returns the metrics bind address for the operator
func getMetricsBindAddress() string {
	var metricsBindAddrEnvVar = "METRICS_BIND_ADDRESS"

	defaultBindAddr := ":8383"

	bindAddr, found := os.LookupEnv(metricsBindAddrEnvVar)
	if found {
		_, err := strconv.Atoi(bindAddr)
		if err != nil {
			msg := fmt.Errorf("%s %s: %s", "supplied METRICS_BIND_ADDRESS is not a number", "METRICS_BIND_ADDRESS", bindAddr)
			setupLog.Error(msg, "Using default METRICS_BIND_ADDRESS: 8383")
			return defaultBindAddr
		} else {
			return ":" + bindAddr
		}
	}
	return defaultBindAddr
}

// getWatchNamespace returns the Namespace the operator should be watching for changes
func getWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s %s", "did not find WATCH_NAMESPACE", "Environment variable WATCH_NAMESPACE must be set")
	}
	return ns, nil
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	//	var probeAddr string

	bindAddr := getMetricsBindAddress()

	flag.StringVar(&metricsAddr, "metrics-bind-address", bindAddr, "The address the metric endpoint binds to.")
	//	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leaderElection", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.ISO8601TimeEncoder,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog.Info("Version Info", "commit", gitCommit)

	watchNamespace, err := getWatchNamespace()
	if err != nil {
		setupLog.Error(err, "unable to get WatchNamespace, "+
			"the manager will watch and manage resources in all namespaces")
	}

	namespaces := []string{watchNamespace, OCPControllerNamespace}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		//		HealthProbeBindAddress: probeAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionID:        "ibm-spectrum-scale-csi-operator",
		LeaderElectionNamespace: watchNamespace, // TODO: Flag should be set to select the namespace where operator is running. Needed for running operator locally.
		NewCache:                cache.MultiNamespacedCacheBuilder(namespaces),
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	//eventRec := mgr.GetEventRecorderFor("CSIScaleOperator")
	csiScaleOperatorReconciler := &controllers.CSIScaleOperatorReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		Recorder: mgr.GetEventRecorderFor("CSIScaleOperator"),
	}
	if err = csiScaleOperatorReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CSIScaleOperator")
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

	cron := cron.New()
	cron.AddFunc("@every 2m", func() {
		controllers.MonitorPodsAndTriggerEvent(
			csiScaleOperatorReconciler,
			config.Product,
			watchNamespace,
		)
	})

	cron.AddFunc("@every 3m", func() {
		controllers.MonitorPodsAndTriggerEvent(
			csiScaleOperatorReconciler,
			config.ProvisionerLabel,
			watchNamespace,
		)
	})
	cron.Start()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
