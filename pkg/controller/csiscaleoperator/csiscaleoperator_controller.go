package csiscaleoperator

import (
	"context"
	"github.com/operator-framework/operator-sdk/pkg/ansible/runner"
	//"github.com/operator-framework/operator-sdk/pkg/ansible/watches"

	"sigs.k8s.io/controller-runtime/pkg/predicate"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	ibmv1alpha1 "github.ibm.com/jdunham/ibm-spectrum-scale-csi-operator/pkg/apis/ibm/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	//	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Track that the secret was added to the controller.
var SecretAddedCOS = false
var log = logf.Log.WithName("controller_csiscaleoperator")

// Add creates a new CSIScaleOperator Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	r :=   &ReconcileCSIScaleOperator{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}

	// Create a new controller
	c, err := controller.New("csiscaleoperator-controller", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CSIScaleOperator
	err = c.Watch(&source.Kind{Type: &ibmv1alpha1.CSIScaleOperator{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	log.Info("Add the Secrets to be watched.")

	if !SecretAddedCOS {
		SecretAddedCOS = true

		// Add the secret to the controller.
		// Define the label to look for and the contant for the operator.
		const LabelName  = "app.kubernetes.io/name"
		const LabelConst = "ibm-spectrum-scale-csi-operator"

		// Set the source to monitor secrets.
		src := &source.Kind{Type: &v1.Secret{}}

		// Setup a handler mapping to access all custom resources.
		hdl := &handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(a handler.MapObject) []reconcile.Request {
				// Query for all Operator resources in the namespace.
				cso := &ibmv1alpha1.CSIScaleOperatorList{}
				opts := &client.ListOptions{ Namespace: a.Meta.GetNamespace() }
				_ = mgr.GetClient().List(context.TODO(), opts, cso)

				// Compose the Requests.
				reqs := make([]reconcile.Request, len(cso.Items))
				for  i, _ := range reqs {
					reqs[i].NamespacedName.Name = cso.Items[i].Name
					reqs[i].Namespace = a.Meta.GetNamespace()
				}

				return reqs
			}),
		}

		// Setup the predicate filter for secret updates and creates.
		prd := predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				labels :=  e.MetaNew.GetLabels()
				if labels != nil {
					return  labels[LabelName] == LabelConst
				}
				return  false
			},
			CreateFunc: func(e event.CreateEvent) bool {
				labels := e.Meta.GetLabels()
				if labels != nil {
					value := labels[LabelName]
					return value == LabelConst
				}
				return false
			},
		}

		err = c.Watch(src, hdl, prd)
		if err != nil {
			return err
		}
	}

	return nil
}



// blank assignment to verify that ReconcileCSIScaleOperator implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCSIScaleOperator{}

// ReconcileCSIScaleOperator reconciles a CSIScaleOperator object
type ReconcileCSIScaleOperator struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	Runner  runner.Runner
}

// Reconcile reads that state of the cluster for a CSIScaleOperator object and makes changes based on the state read
// and what is in the CSIScaleOperator.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCSIScaleOperator) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues(
		"Request.Namespace", request.Namespace, 
		"Request.Name", request.Name,
		"Request.NamespacedName", request.NamespacedName)
	reqLogger.Info("Reconciling CSIScaleOperator")


	// Fetch the CSIScaleOperator instance
	instance := &ibmv1alpha1.CSIScaleOperator{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//reconcileResult := reconcile.Result{}





	return reconcile.Result{}, nil
}

