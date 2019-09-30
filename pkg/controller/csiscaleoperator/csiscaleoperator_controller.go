package csiscaleoperator

import (
	"context"
	//"github.com/operator-framework/operator-sdk/pkg/ansible/runner"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"

	ibmv1alpha1 "github.ibm.com/jdunham/ibm-spectrum-scale-csi-operator/pkg/apis/ibm/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_csiscaleoperator")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new CSIScaleOperator Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCSIScaleOperator{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("csiscaleoperator-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CSIScaleOperator
	err = c.Watch(&source.Kind{Type: &ibmv1alpha1.CSIScaleOperator{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	//// TODO(user): Modify this to be the types you create that are owned by the primary resource
	//// Watch for changes to secondary resource Pods and requeue the owner CSIScaleOperator
	//err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
	//	IsController: true,
	//	OwnerType:    &ibmv1alpha1.CSIScaleOperator{},
	//})
	//if err != nil {
	//	return err
	//}

	// Set up the watches for the Secret events:
	// ----------------------------------------------------------------
	const LabelName  = "app.kubernetes.io/name"
	const LabelConst = "ibm-spectrum-scale-csi-operator"
	src := &source.Kind{Type: &v1.Secret{}}
	hdl := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ibmv1alpha1.CSIScaleOperator{},
	}
	prd := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			labels :=  e.MetaNew.GetLabels()
			if labels != nil {
				value, ok := labels[LabelName]
				return ok && value == LabelConst
			}
			return  false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			labels := e.Meta.GetLabels()
			if labels != nil {
				value, ok := labels[LabelName]
				return ok && value == LabelConst
			}
			return false
		},
		// TODO create?
	}

	err = c.Watch(src, hdl, prd)
	if err != nil {
		return err
	}
	// ----------------------------------------------------------------





	//err = c.Watch(&source.Kind{Type: &v1.Secret{}}, 
	//run, err := runner.New(nil)

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
}

// Reconcile reads that state of the cluster for a CSIScaleOperator object and makes changes based on the state read
// and what is in the CSIScaleOperator.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCSIScaleOperator) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
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

	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set CSIScaleOperator instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists
	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *ibmv1alpha1.CSIScaleOperator) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}
