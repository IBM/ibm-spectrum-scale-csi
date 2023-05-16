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
	"errors"
	"os"
	"sort"
	"strconv"

	"github.com/imdario/mergo"
	"github.com/presslabs/controller-util/pkg/mergo/transformers"
	"github.com/presslabs/controller-util/pkg/syncer"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/config"
	"github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/internal/csiscaleoperator"
	"github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/util/boolptr"
	"github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/util/k8sutil"
)

const (
	nodeContainerName                = "ibm-spectrum-scale-csi"
	nodeDriverRegistrarContainerName = "driver-registrar"
	nodeLivenessProbeContainerName   = "liveness-probe"
	nodeContainerHealthPortName      = "healthz"
	nodeContainerHealthPortNumber    = 9821
	podMountDir                      = "pods-mount-dir"
	hostDev                          = "host-dev"
	hostDevPath                      = "/dev"
	pluginDir                        = "plugin-dir"
	registrationDir                  = "registration-dir"
	registrationDirPath              = "/registration"

	// FS-Group requirement.
	// Mount `/` filesystem from host machine to driver container on path `/host`
	hostDir          = "host-dir"
	hostDirPath      = "/"
	hostDirMountPath = "/host"

	//EnvVarForDriverImage is the name of environment variable for
	//CSI driver image name, passed by operator.
	EnvVarForDriverImage           = "CSI_DRIVER_IMAGE"
	EnvVarForCSINodeRegistrarImage = "CSI_NODE_REGISTRAR_IMAGE"
	EnvVarForCSILivenessProbeImage = "CSI_LIVENESSPROBE_IMAGE"
	EnvVarForLivenessHealthPort    = "LIVENESS_HEALTH_PORT"
	EnvVarForShortNodeNameMapping  = "SHORTNAME_NODE_MAPPING"
)

var (
	// UUID is a unique cluster ID assigned to the kubernetes/ OCP platform.
	UUID                       string
	nodeContainerHealthPort    = intstr.FromInt(nodeContainerHealthPortNumber)
	cmEnvVars                  []corev1.EnvVar
	daemonMaxUnavailableGlobal string
)

type csiNodeSyncer struct {
	driver *csiscaleoperator.CSIScaleOperator
	obj    runtime.Object
}

// GetCSIDaemonsetSyncer creates and returns a syncer for CSI driver daemonset.
func GetCSIDaemonsetSyncer(c client.Client, scheme *runtime.Scheme, driver *csiscaleoperator.CSIScaleOperator,
	daemonSetRestartedKey string, daemonSetRestartedValue string, CGPrefix string, envVars map[string]string, daemonSetMaxUnavailable string) syncer.Interface {
	obj := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        config.GetNameForResource(config.CSINode, driver.Name),
			Namespace:   driver.Namespace,
			Annotations: driver.GetAnnotations(daemonSetRestartedKey, daemonSetRestartedValue),
			Labels:      driver.GetLabels(),
		},
	}

	sync := &csiNodeSyncer{
		driver: driver,
		obj:    obj,
	}

	UUID = CGPrefix

	cmEnvVars = []corev1.EnvVar{}
	var keys []string
	for k := range envVars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		cmEnvVars = append(cmEnvVars, corev1.EnvVar{
			Name:  k,
			Value: envVars[k],
		})
	}

	daemonMaxUnavailableGlobal = daemonSetMaxUnavailable

	return syncer.NewObjectSyncer(config.CSINode.String(), driver.Unwrap(), obj, c, func() error {
		return sync.SyncCSIDaemonsetFn(daemonSetRestartedKey, daemonSetRestartedValue)
	})
}

// SyncCSIDaemonsetFn handles reconciliation of CSI driver daemonset.
func (s *csiNodeSyncer) SyncCSIDaemonsetFn(daemonSetRestartedKey string, daemonSetRestartedValue string) error {
	logger := csiLog.WithName("SyncCSIDaemonsetFn")

	out := s.obj.(*appsv1.DaemonSet)

	secrets := []corev1.LocalObjectReference{}
	if len(s.driver.Spec.ImagePullSecrets) > 0 {
		for _, s := range s.driver.Spec.ImagePullSecrets {
			logger.Info("Got ", "ImagePullSecret:", s)
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	}
	secrets = append(secrets, corev1.LocalObjectReference{Name: config.ImagePullSecretRegistryKey},
		corev1.LocalObjectReference{Name: config.ImagePullSecretEntitlementKey})

	annotations := s.driver.GetAnnotations(daemonSetRestartedKey, daemonSetRestartedValue)
	annotations["kubectl.kubernetes.io/default-container"] = config.Product

	out.Spec.Selector = metav1.SetAsLabelSelector(s.driver.GetCSINodeSelectorLabels(config.GetNameForResource(config.CSINode, s.driver.Name)))

	// ensure template
	out.Spec.Template.ObjectMeta.Labels = s.driver.GetCSINodePodLabels(config.GetNameForResource(config.CSINode, s.driver.Name))
	out.Spec.Template.ObjectMeta.Annotations = annotations
	out.Spec.Template.Spec.NodeSelector = s.driver.GetNodeSelectors(s.driver.Spec.PluginNodeSelector)
	if out.Spec.Template.Spec.Containers != nil {
		for i := range out.Spec.Template.Spec.Containers {
			out.Spec.Template.Spec.Containers[i].Env = []corev1.EnvVar{}
		}
	}
	out.Spec.Template.Spec.Tolerations = []corev1.Toleration{}
	out.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{}

	// set daemonSetUpgradeStrategy
	var maxUnavailableLocal intstr.IntOrString
	if len(daemonMaxUnavailableGlobal) > 0 {
		maxUnavailableLocal = intstr.FromString(daemonMaxUnavailableGlobal)
	} else {
		maxUnavailableLocal = intstr.FromInt(1)
	}
	logger.Info("UpdateStrategy for RollingUpdate set for ", "MaxUnavailable", maxUnavailableLocal)
	deploy := appsv1.RollingUpdateDaemonSet{
		MaxUnavailable: &maxUnavailableLocal,
	}
	var strategyType appsv1.DaemonSetUpdateStrategyType = config.CSIDaemonSetUpgradeUpdateStrateyType
	strategy := appsv1.DaemonSetUpdateStrategy{
		RollingUpdate: &deploy,
		Type:          strategyType,
	}
	out.Spec.UpdateStrategy = strategy
	//reset variable after setting daemonSet spec.
	daemonMaxUnavailableGlobal = ""

	err := mergo.Merge(&out.Spec.Template.Spec, s.ensurePodSpec(secrets), mergo.WithTransformers(transformers.PodSpec))
	if err != nil {
		return err
	}

	logger.Info("Synchronization of node DaemonSet is successful")
	return nil
}

// ensurePodSpec creates and returns pod specs for CSI driver pod.
func (s *csiNodeSyncer) ensurePodSpec(secrets []corev1.LocalObjectReference) corev1.PodSpec {

	pod := corev1.PodSpec{
		Containers:         s.ensureContainersSpec(),
		Volumes:            s.ensureVolumes(),
		HostIPC:            false,
		HostNetwork:        true,
		DNSPolicy:          config.ClusterFirstWithHostNet,
		ServiceAccountName: config.GetNameForResource(config.CSINodeServiceAccount, s.driver.Name),
		Tolerations:        s.driver.Spec.Tolerations,
		ImagePullSecrets:   secrets,
		Affinity:           s.driver.GetAffinity(config.NodePlugin.String()),
		PriorityClassName:  "system-node-critical",
	}
	return pod
}

// ensureContainersSpec returns array of containers which has the desired
// fields for all 3 containers driver plugin, driver registrar and
// liveness probe.
func (s *csiNodeSyncer) ensureContainersSpec() []corev1.Container {

	logger := csiLog.WithName("ensureContainersSpec")

	// node plugin container
	nodePlugin := s.ensureContainer(nodeContainerName,
		s.getImage(config.GetNameForResource(config.CSINode, s.driver.Name)),
		[]string{
			"--nodeid=$(NODE_ID)",
			"--endpoint=$(CSI_ENDPOINT)",
			"--kubeletRootDirPath=$(KUBELET_ROOT_DIR_PATH)",
		},
	)

	nodePlugin.Resources = ensureDriverResources()

	//nodePlugin.Ports = ensurePorts(corev1.ContainerPort{
	//	Name:          nodeContainerHealthPortName,
	//	ContainerPort: nodeContainerHealthPortNumber,
	//})

	nodePlugin.ImagePullPolicy = config.CSIDriverImagePullPolicy

	// Check if there is any environment variable passing liveness
	// health port number
	healthPort := nodeContainerHealthPort
	healthPortStr, found := os.LookupEnv(EnvVarForLivenessHealthPort)
	if found {
		port, err := strconv.Atoi(healthPortStr)
		if err == nil {
			logger.Info("Got liveness probe", " port: ", port)
			healthPort = intstr.FromInt(port)
		} else {
			logger.Info("Invalid liveness probe port number", "received port: ", healthPortStr)
		}
	}
	nodePlugin.LivenessProbe = ensureProbe(10, 30, 120, corev1.ProbeHandler{
		HTTPGet: &corev1.HTTPGetAction{
			Path:   "/healthz",
			Port:   healthPort,
			Scheme: corev1.URISchemeHTTP,
		},
	})

	nodePlugin.SecurityContext = ensureDriverContainersSecurityContext(true, true, true, false)
	fillSecurityContextCapabilities(nodePlugin.SecurityContext)

	nodePlugin.Lifecycle = &corev1.Lifecycle{
		PreStop: &corev1.LifecycleHandler{
			Exec: &corev1.ExecAction{
				Command: []string{"/bin/sh", "-c", "rm -f " + s.driver.GetSocketPath()},
			},
		},
	}

	// node driver registrar sidecar
	registrar := s.ensureContainer(nodeDriverRegistrarContainerName,
		s.getImage(config.CSINodeDriverRegistrar),
		[]string{
			"--csi-address=$(ADDRESS)",
			"--kubelet-registration-path=$(DRIVER_REG_SOCK_PATH)",
			"--v=5",
		},
	)

	registrar.SecurityContext = ensureDriverContainersSecurityContext(true, true, true, true)
	fillSecurityContextCapabilities(registrar.SecurityContext)
	registrar.ImagePullPolicy = config.CSINodeDriverRegistrarImagePullPolicy
	registrar.Resources = ensureSidecarResources()

	// liveness probe sidecar
	livenessProbe := s.ensureContainer(nodeLivenessProbeContainerName,
		s.getImage(config.LivenessProbe),
		[]string{
			"--health-port=" + healthPort.String(),
			"--csi-address=$(ADDRESS)",
			"--v=5",
		},
	)
	livenessProbe.SecurityContext = ensureDriverContainersSecurityContext(false, false, true, false)
	fillSecurityContextCapabilities(livenessProbe.SecurityContext)
	livenessProbe.ImagePullPolicy = config.LivenessProbeImagePullPolicy
	livenessProbe.Resources = ensureSidecarResources()

	return []corev1.Container{
		nodePlugin,
		registrar,
		livenessProbe,
	}
}

// ensureContainer returns a container with given name, image and
// some other fields.
func (s *csiNodeSyncer) ensureContainer(name, image string, args []string) corev1.Container {
	return corev1.Container{
		Name:         name,
		Image:        image,
		Args:         args,
		Env:          s.getEnvFor(name),
		VolumeMounts: s.getVolumeMountsFor(name),
	}
}

// envVarFromField returns an environment variable with given name
// and path.
func envVarFromField(name, fieldPath string) corev1.EnvVar {
	env := corev1.EnvVar{
		Name: name,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				APIVersion: config.APIVersion,
				FieldPath:  fieldPath,
			},
		},
	}
	return env
}

// getEnvFor returns list of environment variables for given container name.
func (s *csiNodeSyncer) getEnvFor(name string) []corev1.EnvVar {

	switch name {
	case nodeContainerName:
		EnvVars := []corev1.EnvVar{}
		if len(s.driver.Spec.NodeMapping) != 0 {
			for _, item := range s.driver.Spec.NodeMapping {
				obj := corev1.EnvVar{}
				obj.Name = item.K8sNode
				obj.Value = item.SpectrumscaleNode
				EnvVars = append(EnvVars, obj)
			}
		}

		shortNodeNameMappingObj := corev1.EnvVar{}
		shortNodeNameMappingObj.Name = "SHORTNAME_NODE_MAPPING"
		shortNodeNameMappingObj.Value = "no"
		shortNodeNameMapping, found := os.LookupEnv(EnvVarForShortNodeNameMapping)
		if found {
			shortNodeNameMappingObj.Value = shortNodeNameMapping
		}
		EnvVars = append(EnvVars, shortNodeNameMappingObj)

		CGPrefixObj := corev1.EnvVar{}
		CGPrefixObj.Name = config.ENVCGPrefix
		CGPrefixObj.Value = UUID
		EnvVars = append(EnvVars, CGPrefixObj)

		/*for _, cmEnv := range cmEnvVars {
			EnvVars = append(EnvVars, cmEnv)
		}*/
		EnvVars = append(EnvVars, cmEnvVars...)
		return append(EnvVars, []corev1.EnvVar{
			{
				Name:  "CSI_ENDPOINT",
				Value: s.driver.GetCSIEndpoint(),
			},
			{
				Name:  "KUBELET_ROOT_DIR_PATH",
				Value: config.CSIKubeletRootDirPath,
			},
			{
				Name:  "SKIP_MOUNT_UNMOUNT",
				Value: "yes",
			},
			envVarFromField("NODE_ID", "spec.nodeName"),
			// envVarFromField("KUBE_NODE_NAME", "spec.nodeName"),
		}...)

	case nodeDriverRegistrarContainerName:
		return []corev1.EnvVar{
			{
				Name:  "ADDRESS",
				Value: s.driver.GetSocketPath(),
			},
			{
				Name:  "DRIVER_REG_SOCK_PATH",
				Value: s.driver.GetSocketPath(),
			},
			envVarFromField("KUBE_NODE_NAME", "spec.nodeName"),
		}

	case nodeLivenessProbeContainerName:
		return []corev1.EnvVar{
			{
				Name:  "ADDRESS",
				Value: s.driver.GetSocketPath(),
			},
		}
	}
	return nil
}

// getVolumeMountsFor returns volume mounts for given container name.
func (s *csiNodeSyncer) getVolumeMountsFor(name string) []corev1.VolumeMount {
	// TODO: Keeping mount propogation as Bidirectional for now, investigate and
	// use an appropriate mount propogation.
	// More information here: https://kubernetes.io/docs/concepts/storage/volumes/
	mountPropagationB := corev1.MountPropagationBidirectional
	//mountPropagationH := corev1.MountPropagationHostToContainer
	switch name {
	case nodeContainerName:
		volumeMounts := []corev1.VolumeMount{
			{
				Name:             hostDir,
				MountPath:        hostDirMountPath,
				MountPropagation: &mountPropagationB,
			},
			{
				Name:      pluginDir,
				MountPath: s.driver.GetSocketDir(),
			},
			{
				Name:             podMountDir,
				MountPath:        s.driver.GetKubeletPodsDir(),
				MountPropagation: &mountPropagationB,
			},
			{
				Name:      hostDev,
				MountPath: hostDevPath,
			},
			{
				Name:      config.CSIConfigMap,
				MountPath: config.ConfigMapPath,
			},
		}

		for _, cluster := range s.driver.Spec.Clusters {
			secret := cluster.Secrets
			secretVolumeMount := corev1.VolumeMount{
				Name:      secret,
				MountPath: config.SecretsMountPath + secret}
			volumeMounts = append(volumeMounts, secretVolumeMount)

			//There is already an error mesaage, logged by operator if the cacert
			//is missing in case of secureSslMode is set to true. The error is
			//logged when k8s tries to create a volume for cacert.
			isSecureSslMode := cluster.SecureSslMode
			cacert := cluster.Cacert
			if isSecureSslMode && len(cacert) != 0 {
				cacertVolumeMount := corev1.VolumeMount{
					Name:      cacert,
					MountPath: config.CAcertMountPath + cacert}
				volumeMounts = append(volumeMounts, cacertVolumeMount)
			}
		}
		return volumeMounts

	case nodeDriverRegistrarContainerName:
		return []corev1.VolumeMount{
			{
				Name:      pluginDir,
				MountPath: s.driver.GetSocketDir(),
			},
			{
				Name:      registrationDir,
				MountPath: registrationDirPath,
			},
		}

	case nodeLivenessProbeContainerName:
		return []corev1.VolumeMount{
			{
				Name:      pluginDir,
				MountPath: s.driver.GetSocketDir(),
			},
		}
	}
	return nil
}

// ensureVolumes returns volumes for CSI driver pods.
func (s *csiNodeSyncer) ensureVolumes() []corev1.Volume {
	logger := csiLog.WithName("ensureVolumes")
	volumes := []corev1.Volume{
		k8sutil.EnsureVolume(pluginDir, k8sutil.EnsureHostPathVolumeSource(s.driver.GetSocketDir(), "DirectoryOrCreate")),
		k8sutil.EnsureVolume(registrationDir, k8sutil.EnsureHostPathVolumeSource(s.driver.GetKubeletRootDirPath()+config.PluginsRegistry, "Directory")),
		k8sutil.EnsureVolume(podMountDir, k8sutil.EnsureHostPathVolumeSource(s.driver.GetKubeletPodsDir(), "Directory")),
		k8sutil.EnsureVolume(hostDev, k8sutil.EnsureHostPathVolumeSource(hostDevPath, "Directory")),
		k8sutil.EnsureVolume(config.CSIConfigMap, k8sutil.EnsureConfigMapVolumeSource(config.CSIConfigMap)),
		k8sutil.EnsureVolume(hostDir, k8sutil.EnsureHostPathVolumeSource(hostDirPath, "Directory")),
	}

	for _, cluster := range s.driver.Spec.Clusters {
		volume := k8sutil.EnsureVolume(cluster.Secrets,
			ensureSecretVolumeSource(cluster.Secrets))
		volumes = append(volumes, volume)

		// To enable SSL, both secureSslMode and cacert have to be passed in
		// CSI CR manifest file.
		// TODO: See if this validation can be done by CRD itself.
		isSecureSslMode := cluster.SecureSslMode
		if isSecureSslMode {
			cacert := cluster.Cacert
			if len(cacert) == 0 {
				logger.Error(errors.New("SecureSslMode Error"),
					"CA certificate is not specified in secure SSL mode for cluster with ID "+cluster.Id)
			} else {
				cacertVolume := k8sutil.EnsureVolume(cacert,
					k8sutil.EnsureConfigMapVolumeSource(cacert))
				volumes = append(volumes, cacertVolume)
				logger.Info("SSL communication with GPFS cluster with ID " +
					cluster.Id + " is enabled!")
			}
		}
	}
	return volumes
}

// getImage gets and returns the images for CSI driver from CR
// if defined in CR, otherwise returns the default images.
func (s *csiNodeSyncer) getImage(name string) string {
	logger := csiLog.WithName("getImage")
	logger.Info("Getting image for: ", "name", name)

	image := ""
	csiNodeDriverPlugin := config.GetNameForResource(config.CSINode, s.driver.Name)
	switch name {
	case config.CSINodeDriverRegistrar:
		nodeRegistrarImage, found := os.LookupEnv(EnvVarForCSINodeRegistrarImage)
		if len(s.driver.Spec.DriverRegistrar) != 0 {
			image = s.driver.Spec.DriverRegistrar
		} else if found {
			image = nodeRegistrarImage
		} else {
			image = s.driver.GetDefaultImage(name)
		}
		logger.Info("Got image for", " node driver registrar: ", image)
	case csiNodeDriverPlugin:
		driverImage, found := os.LookupEnv(EnvVarForDriverImage)
		if len(s.driver.Spec.SpectrumScale) != 0 {
			image = s.driver.Spec.SpectrumScale
		} else if found {
			image = driverImage
		} else {
			image = s.driver.GetDefaultImage(config.CSINodeDriverPlugin)
		}
		logger.Info("Got image for", " node plugin: ", image)
	case config.LivenessProbe:
		livenessProbeImage, found := os.LookupEnv(EnvVarForCSILivenessProbeImage)
		if len(s.driver.Spec.LivenessProbe) != 0 {
			image = s.driver.Spec.LivenessProbe
		} else if found {
			image = livenessProbeImage
		} else {
			image = s.driver.GetDefaultImage(name)
		}
		logger.Info("Got image for", " liveness probe: ", image)
	}
	logger.Info("Exiting getImage", " got image:", image)
	return image
}

// ensureSecretVolumeSource returns SecretVolumeSource with given name
// with items username and password.
func ensureSecretVolumeSource(name string) corev1.VolumeSource {
	return corev1.VolumeSource{
		Secret: &corev1.SecretVolumeSource{
			SecretName: name,
			Items: []corev1.KeyToPath{
				{
					Key:  config.SecretUsername,
					Path: config.SecretUsername,
				},
				{
					Key:  config.SecretPassword,
					Path: config.SecretPassword,
				},
			},
		},
	}
}

// fillSecurityContextCapabilities adds POSIX capabilities to given SCC.
func fillSecurityContextCapabilities(sc *corev1.SecurityContext, add ...string) {
	sc.Capabilities = &corev1.Capabilities{
		Drop: []corev1.Capability{"ALL"},
	}
	if len(add) > 0 {
		adds := []corev1.Capability{}
		for _, a := range add {
			adds = append(adds, corev1.Capability(a))
		}
		sc.Capabilities.Add = adds
	}
}

// ensureDriverContainersSecurityContext sets AllowPrivilegeEscalation, Privileged, ReadOnlyRootFilesystem
// and runAsNonRoot for the containers of driver pod, if runAsNonRoot is true, default values are set for
// runAsUser and runAsGroup.
func ensureDriverContainersSecurityContext(allowPrivilegeEscalation bool, privileged bool, readOnlyRootFilesystem bool, runAsNonRoot bool) *corev1.SecurityContext {
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
	securityContext := &corev1.SecurityContext{
		AllowPrivilegeEscalation: localAllowPrivilegeEscalation,
		Privileged:               localPrivileged,
		ReadOnlyRootFilesystem:   localReadOnlyRootFilesystem}

	if runAsNonRoot {
		runAsUser := int64(config.RunAsUser)
		runAsGroup := int64(config.RunAsGroup)

		securityContext.RunAsNonRoot = boolptr.True()
		securityContext.RunAsUser = &runAsUser
		securityContext.RunAsGroup = &runAsGroup
	} else {
		securityContext.RunAsNonRoot = boolptr.False()
	}
	return securityContext
}
