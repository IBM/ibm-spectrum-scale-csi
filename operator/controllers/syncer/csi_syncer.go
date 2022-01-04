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
	"os"

	"github.com/imdario/mergo"
	"github.com/presslabs/controller-util/mergo/transformers"
	"github.com/presslabs/controller-util/syncer"
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

// GetAttacherSyncer returns a new kubernetes.Object syncer for k8s statefulset object for CSI attacher service.
func GetAttacherSyncer(c client.Client, scheme *runtime.Scheme, driver *csiscaleoperator.CSIScaleOperator) syncer.Interface {

	logger := csiLog.WithName("GetAttacherSyncer")
	logger.Info("Creating a syncer object for the attacher statefulset.")

	obj := &appsv1.StatefulSet{
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
		return sync.SyncAttacherFn()
	})
}

// GetProvisionerSyncer returns a new kubernetes.Object syncer for k8s statefulset object for CSI provisioner service.
func GetProvisionerSyncer(c client.Client, scheme *runtime.Scheme, driver *csiscaleoperator.CSIScaleOperator) syncer.Interface {

	logger := csiLog.WithName("GetProvisionerSyncer")
	logger.Info("Creating a syncer object for the provisioner statefulset.")

	obj := &appsv1.StatefulSet{
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
		return sync.SyncProvisionerFn()
	})
}

// GetSnapshotterSyncer returns a new kubernetes.Object syncer for k8s statefulset object for CSI snapshotter service.
func GetSnapshotterSyncer(c client.Client, scheme *runtime.Scheme, driver *csiscaleoperator.CSIScaleOperator) syncer.Interface {

	logger := csiLog.WithName("GetSnapshotterSyncer")
	logger.Info("Creating a syncer object for the snapshotter statefulset.")

	obj := &appsv1.StatefulSet{
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
		return sync.SyncSnapshotterFn()
	})
}

// GetResizerSyncer returns a new kubernetes.Object syncer for k8s statefulset object for CSI resizer service.
func GetResizerSyncer(c client.Client, scheme *runtime.Scheme, driver *csiscaleoperator.CSIScaleOperator) syncer.Interface {

	logger := csiLog.WithName("GetResizerSyncer")
	logger.Info("Creating a syncer object for the resizer statefulset.")

	obj := &appsv1.StatefulSet{
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
		return sync.SyncResizerFn()
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

// SyncAttacherFn is a function which mutates the existing attacher statefulset object into it's desired state.
func (s *csiControllerSyncer) SyncAttacherFn() error {

	logger := csiLog.WithName("SyncAttacherFn")
	logger.Info("Mutating the attacher statefulset object into it's desired state.")

	out := s.obj.(*appsv1.StatefulSet)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIControllerSelectorLabels(config.GetNameForResource(config.CSIControllerAttacher, s.driver.Name)))
	out.Spec.ServiceName = config.GetNameForResource(config.CSIControllerAttacher, s.driver.Name)

	secrets := []corev1.LocalObjectReference{}
	if len(s.driver.Spec.ImagePullSecrets) > 0 {
		for _, s := range s.driver.Spec.ImagePullSecrets {
			logger.Info("SyncAttacherFn: Got ", "ImagePullSecret:", s)
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	}

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSIControllerPodLabels(config.GetNameForResource(config.CSIControllerAttacher, s.driver.Name))
	out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations("", "")
	if len(secrets) != 0 {
		out.Spec.Template.Spec.ImagePullSecrets = secrets
	}
	out.Spec.Template.Spec.Tolerations = s.driver.Spec.Tolerations
	out.Spec.Template.Spec.NodeSelector = s.driver.GetNodeSelectors(s.driver.Spec.AttacherNodeSelector)
	//out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations()

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensureAttacherPodSpec(secrets), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

// SyncProvisionerFn is a function which mutates the existing provisioner statefulset object into it's desired state.
func (s *csiControllerSyncer) SyncProvisionerFn() error {

	logger := csiLog.WithName("SyncProvisionerFn")
	logger.Info("Mutating the provisioner statefulset object into it's desired state.")

	out := s.obj.(*appsv1.StatefulSet)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIControllerSelectorLabels(config.GetNameForResource(config.CSIControllerProvisioner, s.driver.Name)))
	out.Spec.ServiceName = config.GetNameForResource(config.CSIControllerProvisioner, s.driver.Name)

	secrets := []corev1.LocalObjectReference{}
	if len(s.driver.Spec.ImagePullSecrets) > 0 {
		for _, s := range s.driver.Spec.ImagePullSecrets {
			logger.Info("SyncProvisionerFn: Got ", "ImagePullSecret:", s)
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	}

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSIControllerPodLabels(config.GetNameForResource(config.CSIControllerProvisioner, s.driver.Name))
	out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations("", "")
	if len(secrets) != 0 {
		out.Spec.Template.Spec.ImagePullSecrets = secrets
	}
	out.Spec.Template.Spec.Tolerations = s.driver.Spec.Tolerations
	out.Spec.Template.Spec.NodeSelector = s.driver.GetNodeSelectors(s.driver.Spec.ProvisionerNodeSelector)
	//out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations()

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensureProvisionerPodSpec(secrets), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

// SyncSnapshotterFn is a function which mutates the existing snapshotter statefulset object into it's desired state.
func (s *csiControllerSyncer) SyncSnapshotterFn() error {

	logger := csiLog.WithName("SyncSnapshotterFn")
	logger.Info("Mutating the snapshotter statefulset object into it's desired state.")

	out := s.obj.(*appsv1.StatefulSet)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIControllerSelectorLabels(config.GetNameForResource(config.CSIControllerSnapshotter, s.driver.Name)))
	out.Spec.ServiceName = config.GetNameForResource(config.CSIControllerSnapshotter, s.driver.Name)

	secrets := []corev1.LocalObjectReference{}
	if len(s.driver.Spec.ImagePullSecrets) > 0 {
		for _, s := range s.driver.Spec.ImagePullSecrets {
			logger.Info("SyncSnapshotterFn: Got ", "ImagePullSecret:", s)
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	}

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSIControllerPodLabels(config.GetNameForResource(config.CSIControllerSnapshotter, s.driver.Name))
	out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations("", "")
	if len(secrets) != 0 {
		out.Spec.Template.Spec.ImagePullSecrets = secrets
	}
	out.Spec.Template.Spec.Tolerations = s.driver.Spec.Tolerations
	out.Spec.Template.Spec.NodeSelector = s.driver.GetNodeSelectors(s.driver.Spec.SnapshotterNodeSelector)
	//out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations()

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensureSnapshotterPodSpec(secrets), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	return nil
}

// SyncResizerFn is a function which mutates the existing resizer statefulset object into it's desired state.
func (s *csiControllerSyncer) SyncResizerFn() error {

	logger := csiLog.WithName("SyncResizerFn")
	logger.Info("Mutating the resizer statefulset object into it's desired state.")

	out := s.obj.(*appsv1.StatefulSet)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIControllerSelectorLabels(config.GetNameForResource(config.CSIControllerResizer, s.driver.Name)))
	out.Spec.ServiceName = config.GetNameForResource(config.CSIControllerResizer, s.driver.Name)

	secrets := []corev1.LocalObjectReference{}
	if len(s.driver.Spec.ImagePullSecrets) > 0 {
		for _, s := range s.driver.Spec.ImagePullSecrets {
			logger.Info("SyncResizerFn: Got ", "ImagePullSecret:", s)
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	}

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSIControllerPodLabels(config.GetNameForResource(config.CSIControllerResizer, s.driver.Name))
	out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations("", "")
	if len(secrets) != 0 {
		out.Spec.Template.Spec.ImagePullSecrets = secrets
	}
	out.Spec.Template.Spec.Tolerations = s.driver.Spec.Tolerations
	out.Spec.Template.Spec.NodeSelector = s.driver.GetNodeSelectors(s.driver.Spec.ResizerNodeSelector)
	//out.Spec.Template.ObjectMeta.Annotations = s.driver.GetAnnotations()

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

	out := s.obj.(*appsv1.StatefulSet)

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSIControllerSelectorLabels())
	out.Spec.ServiceName = config.GetNameForResource(config.CSIController, s.driver.Name)

	secrets := []corev1.LocalObjectReference{}
	if len(s.driver.Spec.ImagePullSecrets) > 0 {
		for _, s := range s.driver.Spec.ImagePullSecrets {
			logger.Info("SyncFn: Got ", "ImagePullSecret:", s)
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
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

	// fsGroup := config.ControllerUserID
	pod := corev1.PodSpec{
		Containers: s.ensureAttacherContainersSpec(),
		Volumes:    s.ensureVolumes(),
		//		SecurityContext: &corev1.PodSecurityContext{
		//			FSGroup:   &fsGroup,
		//			RunAsUser: &fsGroup,
		//		},
		Affinity:           s.driver.Spec.Affinity,
		Tolerations:        s.driver.Spec.Tolerations,
		ServiceAccountName: config.GetNameForResource(config.CSIAttacherServiceAccount, s.driver.Name),
	}
	if len(secrets) != 0 {
		pod.ImagePullSecrets = secrets
	}
	return pod
}

// ensureProvisionerPodSpec returns an object of type corev1.PodSpec.
// PodSpec contains description of the provisioner pod.
func (s *csiControllerSyncer) ensureProvisionerPodSpec(secrets []corev1.LocalObjectReference) corev1.PodSpec {

	logger := csiLog.WithName("ensureProvisionerPodSpec")
	logger.Info("Generating pod description for the provisioner pod.")

	// fsGroup := config.ControllerUserID
	pod := corev1.PodSpec{
		Containers: s.ensureProvisionerContainersSpec(),
		Volumes:    s.ensureVolumes(),
		//		SecurityContext: &corev1.PodSecurityContext{
		//			FSGroup:   &fsGroup,
		//			RunAsUser: &fsGroup,
		//		},
		Affinity:           s.driver.Spec.Affinity,
		Tolerations:        s.driver.Spec.Tolerations,
		ServiceAccountName: config.GetNameForResource(config.CSIProvisionerServiceAccount, s.driver.Name),
	}
	if len(secrets) != 0 {
		pod.ImagePullSecrets = secrets
	}
	return pod
}

// ensureSnapshotterPodSpec returns an object of type corev1.PodSpec.
// PodSpec contains description of the provisioner pod.
func (s *csiControllerSyncer) ensureSnapshotterPodSpec(secrets []corev1.LocalObjectReference) corev1.PodSpec {

	logger := csiLog.WithName("ensureSnapshotterPodSpec")
	logger.Info("Generating pod description for the snapshotter pod.")

	// fsGroup := config.ControllerUserID
	pod := corev1.PodSpec{
		Containers: s.ensureSnapshotterContainersSpec(),
		Volumes:    s.ensureVolumes(),
		//		SecurityContext: &corev1.PodSecurityContext{
		//			FSGroup:   &fsGroup,
		//			RunAsUser: &fsGroup,
		//		},
		Affinity:           s.driver.Spec.Affinity,
		Tolerations:        s.driver.Spec.Tolerations,
		ServiceAccountName: config.GetNameForResource(config.CSISnapshotterServiceAccount, s.driver.Name),
	}
	if len(secrets) != 0 {
		pod.ImagePullSecrets = secrets
	}
	return pod
}

// ensureResizerPodSpec returns an object of type corev1.PodSpec.
// PodSpec contains description of the provisioner pod.
func (s *csiControllerSyncer) ensureResizerPodSpec(secrets []corev1.LocalObjectReference) corev1.PodSpec {

	logger := csiLog.WithName("ensureResizerPodSpec")
	logger.Info("Generating pod description for the resizer pod.")

	// fsGroup := config.ControllerUserID
	pod := corev1.PodSpec{
		Containers: s.ensureResizerContainersSpec(),
		Volumes:    s.ensureVolumes(),
		//		SecurityContext: &corev1.PodSecurityContext{
		//			FSGroup:   &fsGroup,
		//			RunAsUser: &fsGroup,
		//		},
		Affinity:           s.driver.Spec.Affinity,
		Tolerations:        s.driver.Spec.Tolerations,
		ServiceAccountName: config.GetNameForResource(config.CSIResizerServiceAccount, s.driver.Name),
	}
	if len(secrets) != 0 {
		pod.ImagePullSecrets = secrets
	}
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
		[]string{"--v=5", "--csi-address=$(ADDRESS)", "--resync=10m", "--timeout=2m"},
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
		[]string{"--csi-address=$(ADDRESS)", "--timeout=2m", "--worker-threads=10", "--extra-create-metadata", "--v=5"},
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
		[]string{"--csi-address=$(ADDRESS)", "--v=5", "--leader-election=false"},
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
		[]string{"--csi-address=$(ADDRESS)", "--v=5", "--timeout=2m", "--handle-volume-inuse-error=false", "--workers=10"},
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
		[]string{"--csi-address=$(ADDRESS)", "--v=5", "--timeout=30s", "--default-fstype=ext4"},
	)
	provisioner.ImagePullPolicy = config.CSIProvisionerImagePullPolicy

	// csi attacher sidecar
	attacher := s.ensureContainer(attacherContainerName,
		s.getSidecarImage(config.CSIAttacher),
		// TODO: make timeout configurable
		[]string{"--csi-address=$(ADDRESS)", "--v=5", "--timeout=180s"},
	)
	attacher.ImagePullPolicy = config.CSIAttacherImagePullPolicy

	// csi snapshotter sidecar
	snapshotter := s.ensureContainer(snapshotterContainerName,
		s.getSidecarImage(config.CSISnapshotter),
		// TODO: make timeout configurable
		[]string{"--csi-address=$(ADDRESS)", "--v=5", "--timeout=30s"},
	)
	snapshotter.ImagePullPolicy = config.CSISnapshotterImagePullPolicy

	// csi resizer sidecar
	resizer := s.ensureContainer(resizerContainerName,
		s.getSidecarImage(config.CSIResizer),
		// TODO: make timeout configurable
		[]string{"--csi-address=$(ADDRESS)", "--v=5", "--timeout=30s"},
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

// Helper function that calls ensureResources method.
func ensureDefaultResources() corev1.ResourceRequirements {
	return ensureResources("20m", "200m", "20Mi", "200Mi")
}

// ensureResources generates k8s resourceRequirements object.
func ensureResources(cpuRequests, cpuLimits, memoryRequests, memoryLimits string) corev1.ResourceRequirements {
	requests := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(cpuRequests),
		corev1.ResourceMemory: resource.MustParse(memoryRequests),
	}
	limits := corev1.ResourceList{
		corev1.ResourceCPU:    resource.MustParse(cpuLimits),
		corev1.ResourceMemory: resource.MustParse(memoryLimits),
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

	sc := &corev1.SecurityContext{
		//		AllowPrivilegeEscalation: boolptr.False(),
		Privileged: boolptr.True(),
	}
	fillSecurityContextCapabilities(sc)
	container := corev1.Container{
		Name:  name,
		Image: image,
		Args:  args,
		//EnvFrom:         s.getEnvSourcesFor(name),
		Env:          s.getEnvFor(name),
		VolumeMounts: s.getVolumeMountsFor(name),
	}
	_, isOpenShift := os.LookupEnv(config.ENVIsOpenShift)
	if isOpenShift {
		container.SecurityContext = sc
	}
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
	case controllerContainerName:
		return []corev1.EnvVar{
			{
				Name:  "CSI_ENDPOINT",
				Value: s.driver.GetCSIEndpoint(),
			},
			{
				Name:  "CSI_LOGLEVEL",
				Value: config.DefaultLogLevel,
			},
		}

	case provisionerContainerName, attacherContainerName, snapshotterContainerName, resizerContainerName:
		return []corev1.EnvVar{
			{
				Name:  "ADDRESS",
				Value: s.driver.GetSocketPath(),
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
		logger.Info("got image for", " provisioner: ", image)
	case config.CSIAttacher:
		envImage, found := os.LookupEnv(EnvVarForCSIAttacherImage)
		if len(s.driver.Spec.Attacher) != 0 {
			image = s.driver.Spec.Attacher
		} else if found {
			image = envImage
		} else {
			image = s.driver.GetDefaultImage(name)
		}
		logger.Info("got image for", " attacher: ", image)
	case config.CSISnapshotter:
		envImage, found := os.LookupEnv(EnvVarForCSISnapshotterImage)
		if len(s.driver.Spec.Snapshotter) != 0 {
			image = s.driver.Spec.Snapshotter
		} else if found {
			image = envImage
		} else {
			image = s.driver.GetDefaultImage(name)
		}
		logger.Info("got image for", " snapshotter: ", image)
	case config.CSIResizer:
		envImage, found := os.LookupEnv(EnvVarForCSIResizerImage)
		if len(s.driver.Spec.Resizer) != 0 {
			image = s.driver.Spec.Resizer
		} else if found {
			image = envImage
		} else {
			image = s.driver.GetDefaultImage(name)
		}
		logger.Info("got image for", " resizer: ", image)
	}
	return image
}

func ensurePorts(ports ...corev1.ContainerPort) []corev1.ContainerPort {
	return ports
}

func ensureProbe(delay, timeout, period int32, handler corev1.Handler) *corev1.Probe {
	return &corev1.Probe{
		InitialDelaySeconds: delay,
		TimeoutSeconds:      timeout,
		PeriodSeconds:       period,
		Handler:             handler,
		SuccessThreshold:    1,
		FailureThreshold:    30,
	}
}

/*
// TODO: Unused code. Remove if not required.
// Sidecars as a separate list of fields:
// Sidecar images are already present in current CSI CR as different
// fields under 'spec'. In future, if it is decided to have sidecars
// as a separate list of fields under a field 'spec.sidecars' in CR,
// uncomment following helper functions and use these as needed.

func getSidecarByName(driver *csiscaleoperator.CSIScaleOperator, name string) *csiv1.CSISidecar {
	for _, sidecar := range driver.Spec.Sidecars {
	if sidecar.Name == name {
			return &sidecar
		}
	}
	return nil
}

func (s *csiControllerSyncer) getSidecarPullPolicy(sidecarName string) corev1.PullPolicy {
	sidecar := s.getSidecarByName(sidecarName)
	if sidecar != nil && sidecar.ImagePullPolicy != "" {
		return sidecar.ImagePullPolicy
	}
	return corev1.PullIfNotPresent
}

func (s *csiControllerSyncer) getCSIAttacherPullPolicy() corev1.PullPolicy {
	return s.getSidecarPullPolicy(config.CSIAttacher)
}

func (s *csiControllerSyncer) getCSIProvisionerPullPolicy() corev1.PullPolicy {
	return s.getSidecarPullPolicy(config.CSIProvisioner)
}

func (s *csiControllerSyncer) getSidecarByName(name string) *csiv1.CSISidecar {

 	logger := csiLog.WithName("getSidecarByName")
 	logger.Info("in getSidecarByName")

 	return getSidecarByName(s.driver, name)
}

func (s *csiControllerSyncer) getSidecarImageByName(name string) string {

 	logger := csiLog.WithName("getSidecarImageByName")
 	logger.Info("in getSidecarImageByName", "name", name)

 	sidecar := s.getSidecarByName(name)
 	if sidecar != nil {
 		return fmt.Sprintf("%s:%s", sidecar.Repository, sidecar.Tag)
 	}
 	logger.Info("didn't find sidecar image")
 	return s.driver.GetDefaultSidecarImageByName(name)
}

func (s *csiControllerSyncer) getCSISnapshotterPullPolicy() corev1.PullPolicy {
 	return s.getSidecarPullPolicy(config.CSISnapshotter)
}

func (s *csiControllerSyncer) getCSIResizerPullPolicy() corev1.PullPolicy {
 	return s.getSidecarPullPolicy(config.CSIResizer)
}

func (s *csiControllerSyncer) getCSIProvisionerImage() string {
 	logger := csiLog.WithName("getCSIProvisionerImage")
 	logger.Info("in getCSIProvisionerImage")
 	return s.getSidecarImage(config.CSIProvisioner)
}

func (s *csiControllerSyncer) getCSISnapshotterImage() string {
 	logger := csiLog.WithName("getCSISnapshotterImage")
 	logger.Info("in getCSISnapshotterImage")
 	return s.getSidecarImage(config.CSISnapshotter)
}

func (s *csiControllerSyncer) getCSIResizerImage() string {
 	logger := csiLog.WithName("getCSIResizerImage")
 	logger.Info("in getCSIResizerImage")
 	return config.CSIResizerImage
}

func (s *csiControllerSyncer) getCSIAttacherImage() string {
 	logger := csiLog.WithName("getCSIAttacherImage")
 	logger.Info("in getCSIAttacherImage")
 	return s.getSidecarImage(config.CSIAttacher)
}
*/
