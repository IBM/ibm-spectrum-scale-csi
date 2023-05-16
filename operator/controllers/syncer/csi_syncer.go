/**
 * Copyright 2022 IBM Corp.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package syncer

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"

	"github.com/imdario/mergo"
	"github.com/presslabs/controller-util/pkg/mergo/transformers"
	"github.com/presslabs/controller-util/pkg/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/config"
	"github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/internal/csiscaleoperator"
	"github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/util/boolptr"
	"github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/util/k8sutil"
)

const (
	socketVolumeName                     = "socket-dir"
	controllerContainerName              = "ibm-spectrum-scale-csi-operator"
	provisionerContainerName             = "ibm-spectrum-scale-csi-provisioner"
	attacherContainerName                = "ibm-spectrum-scale-csi-attacher"
	snapshotterContainerName             = "ibm-spectrum-scale-csi-snapshotter"
	resizerContainerName                 = "ibm-spectrum-scale-csi-resizer"
	controllerLivenessProbeContainerName = "liveness-probe"

	EnvVarForCSIAttacherImage    = "CSI_ATTACHER_IMAGE"
	EnvVarForCSIProvisionerImage = "CSI_PROVISIONER_IMAGE"
	EnvVarForCSISnapshotterImage = "CSI_SNAPSHOTTER_IMAGE"
	EnvVarForCSIResizerImage     = "CSI_RESIZER_IMAGE"
)

var csiLog = log.Log.WithName("csiscaleoperator_syncer")

type csiControllerSyncer struct {
	driver *csiscaleoperator.CSIScaleOperator
	obj    runtime.Object
}

// CSIConfigmapSyncer returns a new kubernetes.Object syncer for k8s configmap object.
func CSIConfigmapSyncer(c client.Client, scheme *runtime.Scheme, driver *csiscaleoperator.CSIScaleOperator) syncer.Interface {

	logger := csiLog.WithName("CSIConfigmapSyncer")
	logger.Info("Creating a syncer object for the configMap.")

	obj := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.CSIConfigMap,
			Namespace: driver.Namespace,
		},
	}

	sync := &csiControllerSyncer{
		driver: driver,
		obj:    obj,
	}

	return syncer.NewObjectSyncer(config.CSIController.String(), driver.Unwrap(), obj, c, func() error {
		return sync.SyncConfigMapFn()
	})
}

// GetAttacherSyncer returns a new kubernetes.Object syncer for k8s deployment object for CSI attacher service.
func GetAttacherSyncer(c client.Client, scheme *runtime.Scheme, driver *csiscaleoperator.CSIScaleOperator,
	restartedAtKey string, restartedAtValue string) syncer.Interface {

	logger := csiLog.WithName("GetAttacherSyncer")
	logger.Info("Creating a syncer object for the attacher deployment.")

	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.CSIControllerAttacher, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations("", ""),
			Labels:      driver.GetLabels(),
		},
	}

	sync := &csiControllerSyncer{
		driver: driver,
		obj:    obj,
	}

	return syncer.NewObjectSyncer(config.CSIController.String(), driver.Unwrap(), obj, c, func() error {
		return sync.SyncAttacherFn(restartedAtKey, restartedAtValue)
	})
}

// GetProvisionerSyncer returns a new kubernetes.Object syncer for k8s deployment object for CSI provisioner service.
func GetProvisionerSyncer(c client.Client, scheme *runtime.Scheme, driver *csiscaleoperator.CSIScaleOperator,
	restartedAtKey string, restartedAtValue string) syncer.Interface {

	logger := csiLog.WithName("GetProvisionerSyncer")
	logger.Info("Creating a syncer object for the provisioner deployment.")

	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.CSIControllerProvisioner, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations("", ""),
			Labels:      driver.GetLabels(),
		},
	}

	sync := &csiControllerSyncer{
		driver: driver,
		obj:    obj,
	}

	return syncer.NewObjectSyncer(config.CSIController.String(), driver.Unwrap(), obj, c, func() error {
		return sync.SyncProvisionerFn(restartedAtKey, restartedAtValue)
	})
}

// GetSnapshotterSyncer returns a new kubernetes.Object syncer for k8s deployment object for CSI snapshotter service.
func GetSnapshotterSyncer(c client.Client, scheme *runtime.Scheme, driver *csiscaleoperator.CSIScaleOperator,
	restartedAtKey string, restartedAtValue string) syncer.Interface {

	logger := csiLog.WithName("GetSnapshotterSyncer")
	logger.Info("Creating a syncer object for the snapshotter deployment.")

	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.CSIControllerSnapshotter, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations("", ""),
			Labels:      driver.GetLabels(),
		},
	}

	sync := &csiControllerSyncer{
		driver: driver,
		obj:    obj,
	}

	return syncer.NewObjectSyncer(config.CSIController.String(), driver.Unwrap(), obj, c, func() error {
		return sync.SyncSnapshotterFn(restartedAtKey, restartedAtValue)
	})
}

// GetResizerSyncer returns a new kubernetes.Object syncer for k8s deployment object for CSI resizer service.
func GetResizerSyncer(c client.Client, scheme *runtime.Scheme, driver *csiscaleoperator.CSIScaleOperator,
	restartedAtKey string, restartedAtValue string) syncer.Interface {

	logger := csiLog.WithName("GetResizerSyncer")
	logger.Info("Creating a syncer object for the resizer deployment.")

	obj := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.CSIControllerResizer, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations("", ""),
			Labels:      driver.GetLabels(),
		},
	}

	sync := &csiControllerSyncer{
		driver: driver,
		obj:    obj,
	}

	return syncer.NewObjectSyncer(config.CSIController.String(), driver.Unwrap(), obj, c, func() error {
		return sync.SyncResizerFn(restartedAtKey, restartedAtValue)
	})
}

// SyncConfigMapFn is a function which mutates the existing configMap object into it's desired state.
func (s *csiControllerSyncer) SyncConfigMapFn() error {

	logger := csiLog.WithName("SyncConfigMapFn")
	logger.Info("Mutating the configMap object into it's desired state.")

	out := s.obj.(*corev1.ConfigMap)
	out.ObjectMeta = metav1.ObjectMeta{Name: config.CSIConfigMap, Namespace: s.driver.Namespace, Labels: s.driver.GetLabels()}
	clustersData, err := json.Marshal(&s.driver.Spec.Clusters)
	if err != nil {
		return err
	}

	clustersDataWithKey := "{ \"clusters\": " + string(clustersData) + " }"
	out.Data = map[string]string{
		config.CSIConfigMap + ".json": clustersDataWithKey,
	}

	return nil
}

// SyncAttacherFn is a function which mutates the existing attacher deployment object into it's desired state.
func (s *csiControllerSyncer) SyncAttacherFn(restartedAtKey string, restartedAtValue string) error {

	logger := csiLog.WithName("SyncAttacherFn")
	logger.Info("Mutating the attacher deployment object into it's desired state.")

	out := s.obj.(*appsv1.Deployment)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIControllerSelectorLabels(config.GetNameForResource(config.CSIControllerAttacher, s.driver.Name)))
	out.Spec.Strategy = s.driver.GetDeploymentStrategy()
	replicas := config.ReplicaCount
	out.Spec.Replicas = &replicas

	secrets := []corev1.LocalObjectReference{}
	if len(s.driver.Spec.ImagePullSecrets) > 0 {
		for _, s := range s.driver.Spec.ImagePullSecrets {
			logger.Info("SyncAttacherFn: Got ", "ImagePullSecret:", s)
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	}
	secrets = append(secrets, corev1.LocalObjectReference{Name: config.ImagePullSecretRegistryKey},
		corev1.LocalObjectReference{Name: config.ImagePullSecretEntitlementKey})

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSIControllerPodLabels(config.GetNameForResource(config.CSIControllerAttacher, s.driver.Name))
	out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations(restartedAtKey, restartedAtValue)
	out.Spec.Template.Spec.NodeSelector = s.driver.GetNodeSelectors(s.driver.Spec.AttacherNodeSelector)
	//out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations()
	out.Spec.Template.Spec.Tolerations = []corev1.Toleration{}
	out.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{}

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensureAttacherPodSpec(secrets), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

// SyncProvisionerFn is a function which mutates the existing provisioner deployment object into it's desired state.
func (s *csiControllerSyncer) SyncProvisionerFn(restartedAtKey string, restartedAtValue string) error {

	logger := csiLog.WithName("SyncProvisionerFn")
	logger.Info("Mutating the provisioner deployment object into it's desired state.")

	out := s.obj.(*appsv1.Deployment)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIControllerSelectorLabels(config.GetNameForResource(config.CSIControllerProvisioner, s.driver.Name)))
	out.Spec.Strategy = s.driver.GetDeploymentStrategy()

	secrets := []corev1.LocalObjectReference{}
	if len(s.driver.Spec.ImagePullSecrets) > 0 {
		for _, s := range s.driver.Spec.ImagePullSecrets {
			logger.Info("SyncProvisionerFn: Got ", "ImagePullSecret:", s)
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	}
	secrets = append(secrets, corev1.LocalObjectReference{Name: config.ImagePullSecretRegistryKey},
		corev1.LocalObjectReference{Name: config.ImagePullSecretEntitlementKey})

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSIControllerPodLabels(config.GetNameForResource(config.CSIControllerProvisioner, s.driver.Name))
	out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations(restartedAtKey, restartedAtValue)
	out.Spec.Template.Spec.NodeSelector = s.driver.GetNodeSelectors(s.driver.Spec.ProvisionerNodeSelector)
	//out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations()
	out.Spec.Template.Spec.Tolerations = []corev1.Toleration{}
	out.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{}

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensureProvisionerPodSpec(secrets), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

// SyncSnapshotterFn is a function which mutates the existing snapshotter deployment object into it's desired state.
func (s *csiControllerSyncer) SyncSnapshotterFn(restartedAtKey string, restartedAtValue string) error {

	logger := csiLog.WithName("SyncSnapshotterFn")
	logger.Info("Mutating the snapshotter deployment object into it's desired state.")

	out := s.obj.(*appsv1.Deployment)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIControllerSelectorLabels(config.GetNameForResource(config.CSIControllerSnapshotter, s.driver.Name)))
	out.Spec.Strategy = s.driver.GetDeploymentStrategy()

	secrets := []corev1.LocalObjectReference{}
	if len(s.driver.Spec.ImagePullSecrets) > 0 {
		for _, s := range s.driver.Spec.ImagePullSecrets {
			logger.Info("SyncSnapshotterFn: Got ", "ImagePullSecret:", s)
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	}
	secrets = append(secrets, corev1.LocalObjectReference{Name: config.ImagePullSecretRegistryKey},
		corev1.LocalObjectReference{Name: config.ImagePullSecretEntitlementKey})

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSIControllerPodLabels(config.GetNameForResource(config.CSIControllerSnapshotter, s.driver.Name))
	out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations(restartedAtKey, restartedAtValue)
	out.Spec.Template.Spec.NodeSelector = s.driver.GetNodeSelectors(s.driver.Spec.SnapshotterNodeSelector)
	//out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations()
	out.Spec.Template.Spec.Tolerations = []corev1.Toleration{}
	out.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{}

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensureSnapshotterPodSpec(secrets), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

// SyncResizerFn is a function which mutates the existing resizer deployment object into it's desired state.
func (s *csiControllerSyncer) SyncResizerFn(restartedAtKey string, restartedAtValue string) error {

	logger := csiLog.WithName("SyncResizerFn")
	logger.Info("Mutating the resizer deployment object into it's desired state.")

	out := s.obj.(*appsv1.Deployment)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIControllerSelectorLabels(config.GetNameForResource(config.CSIControllerResizer, s.driver.Name)))
	out.Spec.Strategy = s.driver.GetDeploymentStrategy()

	secrets := []corev1.LocalObjectReference{}
	if len(s.driver.Spec.ImagePullSecrets) > 0 {
		for _, s := range s.driver.Spec.ImagePullSecrets {
			logger.Info("SyncResizerFn: Got ", "ImagePullSecret:", s)
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	}
	secrets = append(secrets, corev1.LocalObjectReference{Name: config.ImagePullSecretRegistryKey},
		corev1.LocalObjectReference{Name: config.ImagePullSecretEntitlementKey})

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSIControllerPodLabels(config.GetNameForResource(config.CSIControllerResizer, s.driver.Name))
	out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations(restartedAtKey, restartedAtValue)
	out.Spec.Template.Spec.NodeSelector = s.driver.GetNodeSelectors(s.driver.Spec.ResizerNodeSelector)
	//out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations()
	out.Spec.Template.Spec.Tolerations = []corev1.Toleration{}
	out.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{}

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensureResizerPodSpec(secrets), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

/*
TODO: Unused code. Remove if not required.
func (s *csiControllerSyncer) SyncFn() error {
	logger := csiLog.WithName("SyncFn")
	logger.Info("in SyncFn")

	out := s.obj.(*appsv1.Deployment)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIControllerSelectorLabels())
	out.Spec.ServiceName = config.GetNameForResource(config.CSIController, s.driver.Name)

	secrets := []corev1.LocalObjectReference{}
	if len(s.driver.Spec.ImagePullSecrets) > 0 {
		for _, s := range s.driver.Spec.ImagePullSecrets {
			logger.Info("SyncFn: Got ", "ImagePullSecret:", s)
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	} else {
		// Use default ImagePullSecret
		secrets = append(secrets, corev1.LocalObjectReference{Name: config.DefaultImagePullSecret})
	}

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSIControllerPodLabels()
	out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations("", "")
	out.Spec.Template.Spec.ImagePullSecrets = secrets
	out.Spec.Template.Spec.Tolerations = s.driver.Spec.Tolerations
	//out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations()

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensurePodSpec(), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}
*/

// ensureAttacherPodSpec returns an object of type corev1.PodSpec.
// PodSpec contains description of the attacher pod.
func (s *csiControllerSyncer) ensureAttacherPodSpec(secrets []corev1.LocalObjectReference) corev1.PodSpec {

	logger := csiLog.WithName("ensureAttacherPodSpec")
	logger.Info("Generating pod description for the attacher pod.")

	tolerations := s.driver.Spec.Tolerations
	pod := corev1.PodSpec{
		Containers:         s.ensureAttacherContainersSpec(),
		Volumes:            s.ensureVolumes(),
		Tolerations:        s.ensurePodTolerations(tolerations),
		Affinity:           s.driver.GetAffinity(config.Attacher.String()),
		ServiceAccountName: config.GetNameForResource(config.CSIAttacherServiceAccount, s.driver.Name),
		ImagePullSecrets:   secrets,
		PriorityClassName:  "system-node-critical",
		SecurityContext:    ensurePodSecurityContext(config.RunAsUser, config.RunAsGroup, true),
	}

	pod.Tolerations = append(pod.Tolerations, s.driver.GetNodeTolerations()...)
	return pod
}

// ensureProvisionerPodSpec returns an object of type corev1.PodSpec.
// PodSpec contains description of the provisioner pod.
func (s *csiControllerSyncer) ensureProvisionerPodSpec(secrets []corev1.LocalObjectReference) corev1.PodSpec {

	logger := csiLog.WithName("ensureProvisionerPodSpec")
	logger.Info("Generating pod description for the provisioner pod.")

	tolerations := s.driver.Spec.Tolerations
	// fsGroup := config.ControllerUserID
	pod := corev1.PodSpec{
		Containers:         s.ensureProvisionerContainersSpec(),
		Volumes:            s.ensureVolumes(),
		Tolerations:        s.ensurePodTolerations(tolerations),
		Affinity:           s.driver.GetAffinity(config.Provisioner.String()),
		ServiceAccountName: config.GetNameForResource(config.CSIProvisionerServiceAccount, s.driver.Name),
		ImagePullSecrets:   secrets,
		SecurityContext:    ensurePodSecurityContext(config.RunAsUser, config.RunAsGroup, true),
	}

	pod.Tolerations = append(pod.Tolerations, s.driver.GetNodeTolerations()...)
	return pod
}

// ensureSnapshotterPodSpec returns an object of type corev1.PodSpec.
// PodSpec contains description of the provisioner pod.
func (s *csiControllerSyncer) ensureSnapshotterPodSpec(secrets []corev1.LocalObjectReference) corev1.PodSpec {

	logger := csiLog.WithName("ensureSnapshotterPodSpec")
	logger.Info("Generating pod description for the snapshotter pod.")

	tolerations := s.driver.Spec.Tolerations
	// fsGroup := config.ControllerUserID
	pod := corev1.PodSpec{
		Containers:         s.ensureSnapshotterContainersSpec(),
		Volumes:            s.ensureVolumes(),
		Tolerations:        s.ensurePodTolerations(tolerations),
		Affinity:           s.driver.GetAffinity(config.Snapshotter.String()),
		ServiceAccountName: config.GetNameForResource(config.CSISnapshotterServiceAccount, s.driver.Name),
		ImagePullSecrets:   secrets,
		SecurityContext:    ensurePodSecurityContext(config.RunAsUser, config.RunAsGroup, true),
	}

	pod.Tolerations = append(pod.Tolerations, s.driver.GetNodeTolerations()...)
	return pod
}

// ensureResizerPodSpec returns an object of type corev1.PodSpec.
// PodSpec contains description of the provisioner pod.
func (s *csiControllerSyncer) ensureResizerPodSpec(secrets []corev1.LocalObjectReference) corev1.PodSpec {

	logger := csiLog.WithName("ensureResizerPodSpec")
	logger.Info("Generating pod description for the resizer pod.")

	tolerations := s.driver.Spec.Tolerations
	// fsGroup := config.ControllerUserID
	pod := corev1.PodSpec{
		Containers:         s.ensureResizerContainersSpec(),
		Volumes:            s.ensureVolumes(),
		Tolerations:        s.ensurePodTolerations(tolerations),
		Affinity:           s.driver.GetAffinity(config.Resizer.String()),
		ServiceAccountName: config.GetNameForResource(config.CSIResizerServiceAccount, s.driver.Name),
		ImagePullSecrets:   secrets,
		SecurityContext:    ensurePodSecurityContext(config.RunAsUser, config.RunAsGroup, true),
	}

	pod.Tolerations = append(pod.Tolerations, s.driver.GetNodeTolerations()...)
	return pod
}

/*
TODO: Unused code. Remove if not required.
// ensurePodSpec returns an object of type corev1.PodSpec.
// PodSpec contains description of the CSI node pod.
func (s *csiControllerSyncer) ensurePodSpec() corev1.PodSpec {

	logger := csiLog.WithName("ensurePodSpec")
	logger.Info("in ensurePodSpec")

	// fsGroup := config.ControllerUserID
	return corev1.PodSpec{
		Containers: s.ensureContainersSpec(),
		Volumes:    s.ensureVolumes(),
		//		SecurityContext: &corev1.PodSecurityContext{
		//			FSGroup:   &fsGroup,
		//			RunAsUser: &fsGroup,
		//		},
		Affinity:           s.driver.Spec.Affinity,
		Tolerations:        s.driver.Spec.Tolerations,
		ServiceAccountName: config.GetNameForResource(config.CSINodeServiceAccount, s.driver.Name),
	}
}
*/

// ensureAttacherContainersSpec returns an object of type corev1.Container.
// Container object contains description for the container within the attacher pod.
func (s *csiControllerSyncer) ensureAttacherContainersSpec() []corev1.Container {

	logger := csiLog.WithName("ensureAttacherContainersSpec")
	logger.Info("Generating container description for the attacher pod.", "attacherContainerName", attacherContainerName)

	attacher := s.ensureContainer(attacherContainerName,
		s.getSidecarImage(config.CSIAttacher),
		// TODO: make timeout configurable
		[]string{"--v=5", "--csi-address=$(ADDRESS)", "--resync=10m", "--timeout=2m", "--default-fstype=gpfs",
			"--leader-election=true", "--leader-election-lease-duration=$(LEADER_ELECTION_LEASE_DURATION)",
			"--leader-election-renew-deadline=$(LEADER_ELECTION_RENEW_DEADLINE)",
			"--leader-election-retry-period=$(LEADER_ELECTION_RETRY_PERIOD)",
			"--http-endpoint=:" + fmt.Sprint(config.LeaderLivenessPort)},
	)
	attacher.ImagePullPolicy = config.CSIAttacherImagePullPolicy

	return []corev1.Container{
		attacher,
	}
}

// ensureProvisionerContainersSpec returns an object of type corev1.Container.
// Container object contains description for the container within the provisioner pod.
func (s *csiControllerSyncer) ensureProvisionerContainersSpec() []corev1.Container {

	logger := csiLog.WithName("ensureProvisionerContainersSpec")
	logger.Info("Generating container description for the provisioner pod.", "provisionerContainerName", provisionerContainerName)

	provisioner := s.ensureContainer(provisionerContainerName,
		s.getSidecarImage(config.CSIProvisioner),
		// TODO: make timeout configurable
		[]string{"--csi-address=$(ADDRESS)", "--timeout=3m", "--worker-threads=10",
			"--extra-create-metadata", "--v=5", "--default-fstype=gpfs",
			"--leader-election=true", "--leader-election-lease-duration=$(LEADER_ELECTION_LEASE_DURATION)",
			"--leader-election-renew-deadline=$(LEADER_ELECTION_RENEW_DEADLINE)",
			"--leader-election-retry-period=$(LEADER_ELECTION_RETRY_PERIOD)",
			"--http-endpoint=:" + fmt.Sprint(config.LeaderLivenessPort)},
	)
	provisioner.ImagePullPolicy = config.CSIProvisionerImagePullPolicy
	return []corev1.Container{
		provisioner,
	}
}

// ensureSnapshotterContainersSpec returns an object of type corev1.Container.
// Container object contains description for the container within the snapshotter pod.
func (s *csiControllerSyncer) ensureSnapshotterContainersSpec() []corev1.Container {

	logger := csiLog.WithName("ensureSnapshotterContainersSpec")
	logger.Info("Generating container description for the snapshotter pod.", "snapshotterContainerName", snapshotterContainerName)

	snapshotter := s.ensureContainer(snapshotterContainerName,
		s.getSidecarImage(config.CSISnapshotter),
		// TODO: make timeout configurable
		[]string{"--csi-address=$(ADDRESS)", "--v=5", "--worker-threads=1",
			"--leader-election=true", "--leader-election-lease-duration=$(LEADER_ELECTION_LEASE_DURATION)",
			"--leader-election-renew-deadline=$(LEADER_ELECTION_RENEW_DEADLINE)",
			"--leader-election-retry-period=$(LEADER_ELECTION_RETRY_PERIOD)",
			"--http-endpoint=:" + fmt.Sprint(config.LeaderLivenessPort)},
	)
	snapshotter.ImagePullPolicy = config.CSISnapshotterImagePullPolicy
	return []corev1.Container{
		snapshotter,
	}
}

// ensureResizerContainersSpec returns an object of type corev1.Container.
// Container object contains description for the container within the resizer pod.
func (s *csiControllerSyncer) ensureResizerContainersSpec() []corev1.Container {

	logger := csiLog.WithName("ensureResizerContainersSpec")
	logger.Info("Generating container description for the resizer pod.", "resizerContainerName", resizerContainerName)

	resizer := s.ensureContainer(resizerContainerName,
		s.getSidecarImage(config.CSIResizer),
		[]string{"--csi-address=$(ADDRESS)", "--v=5", "--timeout=2m", "--handle-volume-inuse-error=false", "--workers=10",
			"--leader-election=true", "--leader-election-lease-duration=$(LEADER_ELECTION_LEASE_DURATION)",
			"--leader-election-renew-deadline=$(LEADER_ELECTION_RENEW_DEADLINE)",
			"--leader-election-retry-period=$(LEADER_ELECTION_RETRY_PERIOD)",
			"--http-endpoint=:" + fmt.Sprint(config.LeaderLivenessPort)},
	)
	resizer.ImagePullPolicy = config.CSIResizerImagePullPolicy
	return []corev1.Container{
		resizer,
	}
}

/*
TODO: Unused code. Remove if not required.
func (s *csiControllerSyncer) ensureContainersSpec() []corev1.Container {

	logger := csiLog.WithName("ensureContainersSpec")
	logger.Info("in ensureContainersSpec", "controllerContainerName", controllerContainerName)

	// csi provisioner sidecar
	provisioner := s.ensureContainer(provisionerContainerName,
		s.getSidecarImage(config.CSIProvisioner),
		// TODO: make timeout configurable
		[]string{"--csi-address=$(ADDRESS)", "--timeout=30s", "--default-fstype=ext4"},
	)
	provisioner.ImagePullPolicy = config.CSIProvisionerImagePullPolicy

	// csi attacher sidecar
	attacher := s.ensureContainer(attacherContainerName,
		s.getSidecarImage(config.CSIAttacher),
		// TODO: make timeout configurable
		[]string{"--csi-address=$(ADDRESS)", "--timeout=180s"},
	)
	attacher.ImagePullPolicy = config.CSIAttacherImagePullPolicy

	// csi snapshotter sidecar
	snapshotter := s.ensureContainer(snapshotterContainerName,
		s.getSidecarImage(config.CSISnapshotter),
		// TODO: make timeout configurable
		[]string{"--csi-address=$(ADDRESS)", "--timeout=30s"},
	)
	snapshotter.ImagePullPolicy = config.CSISnapshotterImagePullPolicy

	// csi resizer sidecar
	resizer := s.ensureContainer(resizerContainerName,
		s.getSidecarImage(config.CSIResizer),
		// TODO: make timeout configurable
		[]string{"--csi-address=$(ADDRESS)", "--timeout=30s"},
	)
	resizer.ImagePullPolicy = config.CSIResizerImagePullPolicy

	return []corev1.Container{
		provisioner,
		attacher,
		snapshotter,
		resizer,
	}
}
*/

// Helper function that calls ensureResources method with the resources needed for sidecar containers.
func ensureSidecarResources() corev1.ResourceRequirements {
	return ensureResources("20m", "300m", "20Mi", "300Mi", "1Gi", "5Gi")
}

// Helper function that calls ensureResources method with the resources needed for driver containers.
func ensureDriverResources() corev1.ResourceRequirements {
	return ensureResources("20m", "600m", "20Mi", "600Mi", "1Gi", "10Gi")
}

// ensureResources generates k8s resourceRequirements object.
func ensureResources(cpuRequests, cpuLimits, memoryRequests, memoryLimits, ephemeralStorageRequests, ephemeralStorageLimits string) corev1.ResourceRequirements {
	requests := corev1.ResourceList{
		corev1.ResourceCPU:              resource.MustParse(cpuRequests),
		corev1.ResourceMemory:           resource.MustParse(memoryRequests),
		corev1.ResourceEphemeralStorage: resource.MustParse(ephemeralStorageRequests),
	}
	limits := corev1.ResourceList{
		corev1.ResourceCPU:              resource.MustParse(cpuLimits),
		corev1.ResourceMemory:           resource.MustParse(memoryLimits),
		corev1.ResourceEphemeralStorage: resource.MustParse(ephemeralStorageLimits),
	}

	return corev1.ResourceRequirements{
		Limits:   limits,
		Requests: requests,
	}
}

/*
TODO: Unused code. Remove if not required.
func ensureNodeAffinity() *corev1.NodeAffinity {
	return &corev1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
			NodeSelectorTerms: []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						{
							Key:      "kubernetes.io/arch",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"amd64"},
						},
					},
				},
			},
		},
	}
}
*/

// ensureContainer generates k8s container object.
func (s *csiControllerSyncer) ensureContainer(name, image string, args []string) corev1.Container {

	logger := csiLog.WithName("ensureContainer")
	logger.Info("Container information: ", "Name", name, "Image", image)

	container := corev1.Container{
		Name:  name,
		Image: image,
		Args:  args,
		//EnvFrom:         s.getEnvSourcesFor(name),
		Env:           s.getEnvFor(name),
		VolumeMounts:  s.getVolumeMountsFor(name),
		Ports:         s.driver.GetContainerPort(),
		LivenessProbe: s.driver.GetLivenessProbe(),
		Resources:     ensureSidecarResources(),
	}
	container.SecurityContext = ensureContainerSecurityContext(true, true, true)
	fillSecurityContextCapabilities(container.SecurityContext)
	return container
}

/*
// TODO: Unused code. Remove if not required.
func (s *csiControllerSyncer) envVarFromSecret(sctName, name, key string, opt bool) corev1.EnvVar {
	env := corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: sctName,
				},
				Key:      key,
				Optional: &opt,
			},
		},
	}
	return env
}
*/

// getEnvFor returns a k8s envVar object for the CSI sidecar containers.
func (s *csiControllerSyncer) getEnvFor(name string) []corev1.EnvVar {

	switch name {
	case provisionerContainerName, attacherContainerName, snapshotterContainerName, resizerContainerName:
		return []corev1.EnvVar{
			{
				Name:  "ADDRESS",
				Value: s.driver.GetSocketPath(),
			},
			{
				Name:  "LEADER_ELECTION_LEASE_DURATION",
				Value: "137s",
			},
			{
				Name:  "LEADER_ELECTION_RENEW_DEADLINE",
				Value: "107s",
			},
			{
				Name:  "LEADER_ELECTION_RETRY_PERIOD",
				Value: "26s",
			},
		}
	}
	return nil
}

// getVolumeMountsFor returns a k8s volumeMount object for CSI sidecar containers.
func (s *csiControllerSyncer) getVolumeMountsFor(name string) []corev1.VolumeMount {
	switch name {
	case controllerContainerName, provisionerContainerName, attacherContainerName, snapshotterContainerName, resizerContainerName:
		return []corev1.VolumeMount{
			{
				Name:      socketVolumeName,
				MountPath: s.driver.GetSocketDir(),
			},
		}

	case controllerLivenessProbeContainerName:
		return []corev1.VolumeMount{
			{
				Name:      socketVolumeName,
				MountPath: s.driver.GetSocketPath(),
			},
		}
	}
	return nil
}

// ensureVolumes returns a k8s volume object for sidecar pods.
func (s *csiControllerSyncer) ensureVolumes() []corev1.Volume {
	return []corev1.Volume{
		k8sutil.EnsureVolume(socketVolumeName, k8sutil.EnsureHostPathVolumeSource(
			s.driver.GetSocketDir(), "DirectoryOrCreate")),
	}
}

// getSidecarImage gets and returns the images for sidecars from CR
// if defined in CR, otherwise returns the default images.
func (s *csiControllerSyncer) getSidecarImage(name string) string {
	logger := csiLog.WithName("getSidecarImage")
	logger.Info("Fetching image for sidecar container.", "ContainerName", name)

	image := ""
	switch name {
	case config.CSIProvisioner:
		envImage, found := os.LookupEnv(EnvVarForCSIProvisionerImage)
		if len(s.driver.Spec.Provisioner) != 0 {
			image = s.driver.Spec.Provisioner
		} else if found {
			image = envImage
		} else {
			image = s.driver.GetDefaultImage(name)
		}
		logger.Info("Got image for", " provisioner: ", image)
	case config.CSIAttacher:
		envImage, found := os.LookupEnv(EnvVarForCSIAttacherImage)
		if len(s.driver.Spec.Attacher) != 0 {
			image = s.driver.Spec.Attacher
		} else if found {
			image = envImage
		} else {
			image = s.driver.GetDefaultImage(name)
		}
		logger.Info("Got image for", " attacher: ", image)
	case config.CSISnapshotter:
		envImage, found := os.LookupEnv(EnvVarForCSISnapshotterImage)
		if len(s.driver.Spec.Snapshotter) != 0 {
			image = s.driver.Spec.Snapshotter
		} else if found {
			image = envImage
		} else {
			image = s.driver.GetDefaultImage(name)
		}
		logger.Info("Got image for", " snapshotter: ", image)
	case config.CSIResizer:
		envImage, found := os.LookupEnv(EnvVarForCSIResizerImage)
		if len(s.driver.Spec.Resizer) != 0 {
			image = s.driver.Spec.Resizer
		} else if found {
			image = envImage
		} else {
			image = s.driver.GetDefaultImage(name)
		}
		logger.Info("Got image for", " resizer: ", image)
	}
	return image
}

// ensurePodTolerations method removes the `NoExecute` & `NoSchedule` toleration for all taints
// from existing list of tolerations.
func (s *csiControllerSyncer) ensurePodTolerations(tolerations []corev1.Toleration) []corev1.Toleration {
	logger := csiLog.WithName("ensurePodTolerations")
	logger.Info("Fetching tolerations for sidecar controller pods.")

	podTolerations := []corev1.Toleration{}

	noScheduleToleration := corev1.Toleration{
		Effect:   corev1.TaintEffectNoSchedule,
		Operator: corev1.TolerationOpExists,
	}

	noExecuteToleration := corev1.Toleration{
		Effect:   corev1.TaintEffectNoExecute,
		Operator: corev1.TolerationOpExists,
	}

	for _, toleration := range tolerations {
		if !(reflect.DeepEqual(toleration, noScheduleToleration)) && !(reflect.DeepEqual(toleration, noExecuteToleration)) {
			podTolerations = append(podTolerations, toleration)
		}
	}

	return podTolerations
}

/*func ensurePorts(ports ...corev1.ContainerPort) []corev1.ContainerPort {
	return ports
}*/

func ensureProbe(delay, timeout, period int32, handler corev1.ProbeHandler) *corev1.Probe {
	return &corev1.Probe{
		InitialDelaySeconds: delay,
		TimeoutSeconds:      timeout,
		PeriodSeconds:       period,
		ProbeHandler:        handler,
		SuccessThreshold:    1,
		FailureThreshold:    30,
	}
}

// ensurePodSecurityContext set pod security with runAsUser, runAsGroup and runAsNonRoot.
func ensurePodSecurityContext(runAsUser int64, runAsGroup int64, runAsNonRoot bool) *corev1.PodSecurityContext {
	var localRunAsNonRoot bool
	if runAsNonRoot {
		localRunAsNonRoot = *boolptr.True()
	} else {
		localRunAsNonRoot = *boolptr.False()
	}
	return &corev1.PodSecurityContext{
		RunAsNonRoot: &localRunAsNonRoot,
		RunAsUser:    &runAsUser,
		RunAsGroup:   &runAsGroup,
	}
}

// ensureContainerSecurityContext configure AllowPrivilegeEscalation, Privileged, ReadOnlyRootFilesystem for the container.
func ensureContainerSecurityContext(allowPrivilegeEscalation bool, privileged bool, readOnlyRootFilesystem bool) *corev1.SecurityContext {
	var (
		localAllowPrivilegeEscalation *bool = boolptr.False()
		localPrivileged               *bool = boolptr.False()
		localReadOnlyRootFilesystem   *bool = boolptr.True()
	)
	if allowPrivilegeEscalation {
		localAllowPrivilegeEscalation = boolptr.True()
	}
	if privileged {
		localPrivileged = boolptr.True()
	}
	if !readOnlyRootFilesystem {
		localReadOnlyRootFilesystem = boolptr.False()
	}
	return &corev1.SecurityContext{
		AllowPrivilegeEscalation: localAllowPrivilegeEscalation,
		Privileged:               localPrivileged,
		ReadOnlyRootFilesystem:   localReadOnlyRootFilesystem}
}
