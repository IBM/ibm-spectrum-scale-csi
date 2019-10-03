package csiscalesecret

import (
	"context"
	//"github.com/operator-framework/operator-sdk/pkg/ansible/runner"
	//"github.com/operator-framework/operator-sdk/pkg/ansible/watches"

	"sigs.k8s.io/controller-runtime/pkg/predicate"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	ibmv1alpha1 "github.ibm.com/jdunham/ibm-spectrum-scale-csi-operator/pkg/apis/ibm/v1alpha1"
	//"k8s.io/apimachinery/pkg/api/errors"
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
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Track that the secret was added to the controller.
var SecretAddedCOS = false
var log = logf.Log.WithName("controller_csiscaleoperator")

var GVK = schema.GroupVersionKind{
		Version: "v1",
		Group: "core",
		Kind: "Secret",
	}

// Add creates a new CSIScaleOperator Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager)  *controller.Controller {
	r :=   &ReconcileCSIScaleOperator{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
	}

	// Create a new controller
	c, err := controller.New("csiscalesecret-controller", mgr, controller.Options{
		Reconciler: r,
	})
	if err != nil {
		log.Error(nil, "Unable to create csiscaleoperator-controller")
		return nil
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

				log.Info(fmt.Sprintf("In Mapping function, mapping to %v items", len(cso.Items))


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
					return labels[LabelName] == LabelConst
				}
				return false
			},
		}

		err = c.Watch(src, hdl, prd)
		if err != nil {
			log.Error(nil, "Unable to setup secret watch.")
			return nil
		}
	}

	return &c
}



// blank assignment to verify that ReconcileCSIScaleOperator implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCSIScaleOperator{}

// ReconcileCSIScaleOperator reconciles a CSIScaleOperator object
type ReconcileCSIScaleOperator struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a CSIScaleOperator object and makes changes based on the state read
// and what is in the CSIScaleOperator.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCSIScaleOperator) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

