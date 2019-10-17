package main

import (
	"context"
	//"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/IBM/ibm-spectrum-scale-csi-operator/stable/ibm-spectrum-scale-csi-operator-bundle/operators/ibm-spectrum-scale-csi-operator/pkg/apis"
	"github.com/IBM/ibm-spectrum-scale-csi-operator/stable/ibm-spectrum-scale-csi-operator-bundle/operators/ibm-spectrum-scale-csi-operator/pkg/controller/csiscalesecret"

	aocontroller "github.com/operator-framework/operator-sdk/pkg/ansible/controller"
	aoflags "github.com/operator-framework/operator-sdk/pkg/ansible/flags"
	proxy "github.com/operator-framework/operator-sdk/pkg/ansible/proxy"
	"github.com/operator-framework/operator-sdk/pkg/ansible/proxy/controllermap"
	"github.com/operator-framework/operator-sdk/pkg/ansible/runner"
	"github.com/operator-framework/operator-sdk/pkg/ansible/watches"

	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/operator-framework/operator-sdk/pkg/leader"
	"github.com/operator-framework/operator-sdk/pkg/log/zap"
	"github.com/operator-framework/operator-sdk/pkg/metrics"
	"github.com/operator-framework/operator-sdk/pkg/restmapper"
	sdkVersion "github.com/operator-framework/operator-sdk/version"
	"github.com/spf13/pflag"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	kubemetrics "github.com/operator-framework/operator-sdk/pkg/kube-metrics"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Change below variables to serve metrics on different host or port.
var (
	log                       = logf.Log.WithName("cmd")
	metricsHost               = "0.0.0.0"
	metricsPort         int32 = 8383
	operatorMetricsPort int32 = 8686
)

func printVersion() {
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
	log.Info(fmt.Sprintf("Version of operator-sdk: %v", sdkVersion.Version))
}

func main() {
	// Use the ansible flag mechanism.
	flags := aoflags.AddTo(pflag.CommandLine)
	pflag.Parse()
	logf.SetLogger(zap.Logger())

	printVersion()

	namespace, found := os.LookupEnv(k8sutil.WatchNamespaceEnvVar)
	//log = log.WithValues("Namespace", namespace)
	if found {
		log.Info("Watching namespace.")
	} else {
		log.Info(fmt.Sprintf("%v environment variable not set. This operator is watching all namespaces.",
			k8sutil.WatchNamespaceEnvVar))
		namespace = metav1.NamespaceAll
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		Namespace:          namespace,
		MapperProvider:     restmapper.NewDynamicRESTMapper,
		MetricsBindAddress: fmt.Sprintf("%s:%d", metricsHost, metricsPort),
	})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// This reimplements the Ansible run code.
	// Effectively this patches in the runners.
	// ---------------------------------------------------------
	var gvks []schema.GroupVersionKind
	cMap := controllermap.NewControllerMap()
	watches, err := watches.Load(flags.WatchesFile)
	if err != nil {
		log.Error(err, "Failed to load watches.")
		os.Exit(1)
	}

	for _, w := range watches {
		runner, err := runner.New(w)
		if err != nil {
			log.Error(err, "Failed to create runner")
			os.Exit(1)
		}
		log.Info(fmt.Sprintf("Role %s", w.Role))
		ctr := aocontroller.Add(mgr, aocontroller.Options{
			GVK:             w.GroupVersionKind,
			Runner:          runner,
			ManageStatus:    w.ManageStatus,
			MaxWorkers:      getMaxWorkers(w.GroupVersionKind, flags.MaxWorkers),
			ReconcilePeriod: w.ReconcilePeriod,
		})
		if ctr == nil {
			log.Error(nil, "failed to add controller for GVK %v", w.GroupVersionKind.String())
			os.Exit(1)
		}

		cMap.Store(w.GroupVersionKind, &controllermap.Contents{Controller: *ctr,
			WatchDependentResources:     w.WatchDependentResources,
			WatchClusterScopedResources: w.WatchClusterScopedResources,
			OwnerWatchMap:               controllermap.NewWatchMap(),
			AnnotationWatchMap:          controllermap.NewWatchMap(),
		})
		gvks = append(gvks, w.GroupVersionKind)
	}
	// ---------------------------------------------------------

	// This is what we needed to inject for secret monitoring.
	ctr := csiscalesecret.Add(mgr)
	if ctr == nil {
		log.Error(nil, "failed to add controller for secrets")
		os.Exit(1)
	}
	//cMap.Store(csiscalesecret.GVK, &controllermap.Contents{Controller: *ctr,
	//	WatchDependentResources:     true,
	//	WatchClusterScopedResources: true,
	//	OwnerWatchMap:               controllermap.NewWatchMap(),
	//	AnnotationWatchMap:          controllermap.NewWatchMap(),
	//})
	gvks = append(gvks, csiscalesecret.GVK)

	operatorName, err := k8sutil.GetOperatorName()
	if err != nil {
		log.Error(err, "Failed to get the operator name")
		os.Exit(1)
	}

	// Become the leader before proceeding
	err = leader.Become(context.TODO(), operatorName+"-lock")
	if err != nil {
		log.Error(err, "Failed to become leader.")
		os.Exit(1)
	}

	// -------------------------------------

	err = kubemetrics.GenerateAndServeCRMetrics(cfg, []string{namespace}, gvks, metricsHost, operatorMetricsPort)
	if err != nil {
		log.Info("Could not generate and serve custom resource metrics", "error", err.Error())
	}
	servicePorts := []v1.ServicePort{
		{Port: metricsPort, Name: metrics.OperatorPortName, Protocol: v1.ProtocolTCP, TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: metricsPort}},
	}

	// Create Service object to expose the metrics port(s).
	// TODO: probably should expose the port as an environment variable
	_, err = metrics.CreateMetricsService(context.TODO(), cfg, servicePorts)
	if err != nil {
		log.Error(err, "Exposing metrics port failed.")
		os.Exit(1)
	}

	done := make(chan error)

	// start the proxy
	err = proxy.Run(done, proxy.Options{
		Address:           "localhost",
		Port:              8888,
		KubeConfig:        mgr.GetConfig(),
		Cache:             mgr.GetCache(),
		RESTMapper:        mgr.GetRESTMapper(),
		ControllerMap:     cMap,
		OwnerInjection:    flags.InjectOwnerRef,
		WatchedNamespaces: []string{namespace},
	})
	if err != nil {
		log.Error(err, "Error starting proxy.")
		os.Exit(1)
	}

	go func() {
		done <- mgr.Start(signals.SetupSignalHandler())
	}()
	// -------------------------------------

	// Wait for proxy or cmd to finish
	err = <-done
	if err != nil {
		log.Error(err, "Proxy or operator exited with error.")
		os.Exit(1)
	}
	log.Info("Exiting.")
}

// The following function is borrowed from the operator sdk run.
func getMaxWorkers(gvk schema.GroupVersionKind, defValue int) int {
	envVar := strings.ToUpper(strings.Replace(
		fmt.Sprintf("WORKER_%s_%s", gvk.Kind, gvk.Group),
		".",
		"_",
		-1,
	))
	switch maxWorkers, err := strconv.Atoi(os.Getenv(envVar)); {
	case maxWorkers <= 1:
		return defValue
	case err != nil:
		// we don't care why we couldn't parse it just use default
		log.Info("Failed to parse %v from environment. Using default %v", envVar, defValue)
		return defValue
	default:
		return maxWorkers
	}
}
