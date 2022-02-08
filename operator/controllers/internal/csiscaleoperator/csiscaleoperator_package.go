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

package csiscaleoperator

import (
	securityv1 "github.com/openshift/api/security/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	v1 "github.com/IBM/ibm-spectrum-scale-csi/operator/api/v1"
	"github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/config"
	"github.com/IBM/ibm-spectrum-scale-csi/operator/controllers/util/boolptr"
)

const (
	snapshotStorageApiGroup              string = "snapshot.storage.k8s.io"
	securityOpenshiftApiGroup            string = "security.openshift.io"
	storageApiGroup                      string = "storage.k8s.io"
	rbacAuthorizationApiGroup            string = "rbac.authorization.k8s.io"
	podSecurityPolicyApiGroup            string = "extensions"
	coordinationAPIGroup                 string = "coordination.k8s.io"
	storageClassesResource               string = "storageclasses"
	persistentVolumesResource            string = "persistentvolumes"
	persistentVolumeClaimsResource       string = "persistentvolumeclaims"
	persistentVolumeClaimsStatusResource string = "persistentvolumeclaims/status"
	podsResource                         string = "pods"
	volumeAttachmentsResource            string = "volumeattachments"
	volumeAttachmentsStatusResource      string = "volumeattachments/status"
	volumeSnapshotClassesResource        string = "volumesnapshotclasses"
	volumeSnapshotsResource              string = "volumesnapshots"
	volumeSnapshotContentsResource       string = "volumesnapshotcontents"
	volumeSnapshotContentsStatusResource string = "volumesnapshotcontents/status"
	eventsResource                       string = "events"
	nodesResource                        string = "nodes"
	csiNodesResource                     string = "csinodes"
	namespacesResource                   string = "namespaces"
	securityContextConstraintsResource   string = "securitycontextconstraints"
	podSecurityPolicyResource            string = "podsecuritypolicies"
	leaseResource                        string = "leases"
	verbGet                              string = "get"
	verbList                             string = "list"
	verbWatch                            string = "watch"
	verbCreate                           string = "create"
	verbUpdate                           string = "update"
	verbPatch                            string = "patch"
	verbDelete                           string = "delete"
	verbUse                              string = "use"
)

// GenerateCSIDriver returns a non-namespaced CSIDriver object.
func (c *CSIScaleOperator) GenerateCSIDriver() *storagev1.CSIDriver {
	return &storagev1.CSIDriver{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.DriverName,
			Labels: c.GetLabels(),
		},
		Spec: storagev1.CSIDriverSpec{
			AttachRequired: boolptr.True(),
			PodInfoOnMount: boolptr.True(),
		},
	}
}

/*
// GenerateControllerServiceAccount creates a kubernetes service account for the operator controllers
// and modify the service account to use secret as an imagePullSecret.
// It returns an object of type *corev1.ServiceAccount.
func (c *CSIScaleOperator) GenerateControllerServiceAccount() *corev1.ServiceAccount {
	logger := csiLog.WithName("GenerateControllerServiceAccount")
	logger.Info("Inside GenerateControllerServiceAccount method")

	secrets := []corev1.LocalObjectReference{}
	if len(c.Spec.ImagePullSecrets) > 0 {
		for _, s := range c.Spec.ImagePullSecrets {
			logger.Info("GenerateControllerServiceAccount: Got ", "ImagePullSecret:", s)
			secrets = append(secrets, corev1.LocalObjectReference{Name: s})
		}
	}

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
		ImagePullSecrets: secrets,
	}
}
*/

// GenerateNodeServiceAccount creates a kubernetes service account for the node/driver service
// and modify the service account to use secret as an imagePullSecret.
// It returns an object of type *corev1.ServiceAccount.
func (c *CSIScaleOperator) GenerateNodeServiceAccount() *corev1.ServiceAccount {

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.GetNameForResource(config.CSINodeServiceAccount, c.Name),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
	}
}

// GenerateControllerServiceAccount creates a kubernetes service account for the CSI sidecar services
// and modify the service account to use secret as an imagePullSecret.
// It returns an object of type *corev1.ServiceAccount.
func (c *CSIScaleOperator) GenerateControllerServiceAccount() *corev1.ServiceAccount {

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
	}
}

// GenerateProvisionerClusterRole returns a kubernetes clusterrole object for the provisioner service.
func (c *CSIScaleOperator) GenerateProvisionerClusterRole() *rbacv1.ClusterRole {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.Provisioner, c.Name),
			Labels: c.GetLabels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumesResource},
				Verbs:     []string{verbGet, verbList, verbWatch, verbCreate, verbDelete},
			},
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumeClaimsResource},
				Verbs:     []string{verbGet, verbList, verbWatch, verbUpdate},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{storageClassesResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{eventsResource},
				Verbs:     []string{verbList, verbWatch, verbCreate, verbUpdate, verbPatch},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotsResource},
				Verbs:     []string{verbGet, verbList},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotContentsResource},
				Verbs:     []string{verbGet, verbList},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{csiNodesResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{nodesResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{volumeAttachmentsResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{coordinationAPIGroup},
				Resources: []string{leaseResource},
				Verbs:     []string{verbCreate, verbGet, verbList, verbPatch, verbUpdate, verbDelete},
			},
		},
	}
	if len(c.Spec.CSIpspname) != 0 {
		clusterRole.Rules = append(clusterRole.Rules, rbacv1.PolicyRule{
			APIGroups:     []string{podSecurityPolicyApiGroup},
			Resources:     []string{podSecurityPolicyResource},
			ResourceNames: []string{c.Spec.CSIpspname},
			Verbs:         []string{verbUse},
		})
	}
	return clusterRole
}

// GenerateProvisionerClusterRoleBinding returns a kubernetes clusterrolebinding object for the provisioner service.
func (c *CSIScaleOperator) GenerateProvisionerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.Provisioner, c.Name),
			Labels: c.GetLabels(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.Provisioner, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}

// GenerateAttacherClusterRole returns a kubernetes clusterrole object for the attacher service.
func (c *CSIScaleOperator) GenerateAttacherClusterRole() *rbacv1.ClusterRole {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.Attacher, c.Name),
			Labels: c.GetLabels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{eventsResource},
				Verbs:     []string{verbGet, verbList, verbWatch, verbUpdate},
			},
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumesResource},
				Verbs:     []string{verbGet, verbList, verbWatch, verbPatch},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{csiNodesResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{volumeAttachmentsResource},
				Verbs:     []string{verbGet, verbList, verbWatch, verbPatch},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{volumeAttachmentsStatusResource},
				Verbs:     []string{verbPatch},
			},
			{
				APIGroups: []string{coordinationAPIGroup},
				Resources: []string{leaseResource},
				Verbs:     []string{verbCreate, verbGet, verbList, verbPatch, verbUpdate, verbDelete},
			},
		},
	}
	if len(c.Spec.CSIpspname) != 0 {
		clusterRole.Rules = append(clusterRole.Rules, rbacv1.PolicyRule{
			APIGroups:     []string{podSecurityPolicyApiGroup},
			Resources:     []string{podSecurityPolicyResource},
			ResourceNames: []string{c.Spec.CSIpspname},
			Verbs:         []string{verbUse},
		})
	}
	return clusterRole
}

// GenerateAttacherClusterRoleBinding returns a kubernetes clusterrolebinding object for the attacher service.
func (c *CSIScaleOperator) GenerateAttacherClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.Attacher, c.Name),
			Labels: c.GetLabels(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.Attacher, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}

// GenerateSnapshotterClusterRole returns a kubernetes clusterrole object for the snapshotter service.
func (c *CSIScaleOperator) GenerateSnapshotterClusterRole() *rbacv1.ClusterRole {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.Snapshotter, c.Name),
			Labels: c.GetLabels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{eventsResource},
				Verbs:     []string{verbList, verbWatch, verbCreate, verbUpdate, verbPatch},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotClassesResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotContentsResource},
				Verbs:     []string{verbCreate, verbGet, verbList, verbWatch, verbUpdate, verbDelete, verbPatch},
			},
			{
				APIGroups: []string{snapshotStorageApiGroup},
				Resources: []string{volumeSnapshotContentsStatusResource},
				Verbs:     []string{verbUpdate, verbPatch},
			},
			{
				APIGroups: []string{coordinationAPIGroup},
				Resources: []string{leaseResource},
				Verbs:     []string{verbCreate, verbGet, verbList, verbPatch, verbUpdate, verbDelete},
			},
		},
	}
	if len(c.Spec.CSIpspname) != 0 {
		clusterRole.Rules = append(clusterRole.Rules, rbacv1.PolicyRule{
			APIGroups:     []string{podSecurityPolicyApiGroup},
			Resources:     []string{podSecurityPolicyResource},
			ResourceNames: []string{c.Spec.CSIpspname},
			Verbs:         []string{verbUse},
		})
	}
	return clusterRole
}

// GenerateSnapshotterClusterRoleBinding returns a kubernetes clusterrolebinding object for the snapshotter service.
func (c *CSIScaleOperator) GenerateSnapshotterClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.Snapshotter, c.Name),
			Labels: c.GetLabels(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.Snapshotter, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}

// GenerateResizerClusterRoleBinding returns a kubernetes clusterrolebinding object for the resizer service.
func (c *CSIScaleOperator) GenerateResizerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.Resizer, c.Name),
			Labels: c.GetLabels(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.Resizer, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}

// GenerateResizerClusterRole returns a kubernetes clusterrole object for the resizer service.
func (c *CSIScaleOperator) GenerateResizerClusterRole() *rbacv1.ClusterRole {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.Resizer, c.Name),
			Labels: c.GetLabels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumesResource},
				Verbs:     []string{verbGet, verbList, verbWatch, verbPatch, verbUpdate},
			},
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumeClaimsResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{podsResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumeClaimsStatusResource},
				Verbs:     []string{verbPatch, verbUpdate},
			},
			{
				APIGroups: []string{""},
				Resources: []string{eventsResource},
				Verbs:     []string{verbList, verbWatch, verbCreate, verbUpdate, verbPatch},
			},
			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{storageClassesResource},
				Verbs:     []string{verbGet, verbList, verbWatch},
			},
			{
				APIGroups: []string{coordinationAPIGroup},
				Resources: []string{leaseResource},
				Verbs:     []string{verbCreate, verbGet, verbList, verbPatch, verbUpdate, verbDelete},
			},
		},
	}
	if len(c.Spec.CSIpspname) != 0 {
		clusterRole.Rules = append(clusterRole.Rules, rbacv1.PolicyRule{
			APIGroups:     []string{podSecurityPolicyApiGroup},
			Resources:     []string{podSecurityPolicyResource},
			ResourceNames: []string{c.Spec.CSIpspname},
			Verbs:         []string{verbUse},
		})
	}
	return clusterRole
}

// GenerateNodePluginClusterRole returns a kubernetes clusterrole object for the
// CSI driver node plugin.
func (c *CSIScaleOperator) GenerateNodePluginClusterRole() *rbacv1.ClusterRole {
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.NodePlugin, c.Name),
			Labels: c.GetLabels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{nodesResource},
				Verbs:     []string{verbGet, verbList, verbUpdate},
			},

			{
				APIGroups: []string{""},
				Resources: []string{persistentVolumesResource},
				Verbs:     []string{verbGet, verbList, verbWatch, verbUpdate},
			},

			{
				APIGroups: []string{storageApiGroup},
				Resources: []string{volumeAttachmentsResource},
				Verbs:     []string{verbGet, verbList, verbWatch, verbUpdate},
			},

			{
				APIGroups: []string{""},
				Resources: []string{namespacesResource},
				Verbs:     []string{verbGet, verbList},
			},
		},
	}
	if len(c.Spec.CSIpspname) != 0 {
		clusterRole.Rules = append(clusterRole.Rules, rbacv1.PolicyRule{
			APIGroups:     []string{podSecurityPolicyApiGroup},
			Resources:     []string{podSecurityPolicyResource},
			ResourceNames: []string{c.Spec.CSIpspname},
			Verbs:         []string{verbUse},
		})
	}
	return clusterRole
}

// GenerateNodePluginClusterRoleBinding returns a kubernetes clusterrolebinding object for the
// CSI driver node plugin.
func (c *CSIScaleOperator) GenerateNodePluginClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.NodePlugin, c.Name),
			Labels: c.GetLabels(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSINodeServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.NodePlugin, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}

/*
func (c *CSIScaleOperator) GenerateSCCForControllerClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.CSIControllerSCCClusterRole, c.Name),
			Labels: c.GetLabels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{securityOpenshiftApiGroup},
				Resources:     []string{securityContextConstraintsResource},
				ResourceNames: []string{"anyuid"},
				Verbs:         []string{"use"},
			},
		},
	}
}
*/
/*func (c *CSIScaleOperator) GenerateSCCForControllerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.CSIControllerSCCClusterRoleBinding, c.Name),
			Labels: c.GetLabels(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIControllerServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.CSIControllerSCCClusterRole, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}
*/
/*
func (c *CSIScaleOperator) GenerateSCCForNodeClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.CSINodeSCCClusterRole, c.Name),
			Labels: c.GetLabels(),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{securityOpenshiftApiGroup},
				Resources:     []string{securityContextConstraintsResource},
				ResourceNames: []string{"privileged"},
				Verbs:         []string{"use"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{nodesResource},
				Verbs:     []string{verbGet},
			},
		},
	}
}
*/
/*
func (c *CSIScaleOperator) GenerateSCCForNodeClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.CSINodeSCCClusterRoleBinding, c.Name),
			Labels: c.GetLabels(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSINodeServiceAccount, c.Name),
				Namespace: c.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     config.GetNameForResource(config.CSINodeSCCClusterRole, c.Name),
			APIGroup: rbacAuthorizationApiGroup,
		},
	}
}
*/

// GenerateSecurityContextConstraint returns an openshift securitycontextconstraints object.
func (c *CSIScaleOperator) GenerateSecurityContextConstraint(users []string) *securityv1.SecurityContextConstraints {

	var (
		FSTypeHostPath              securityv1.FSType = "hostPath"
		FSTypeEmptyDir              securityv1.FSType = "emptyDir"
		FSTypeSecret                securityv1.FSType = "secret"
		FSTypePersistentVolumeClaim securityv1.FSType = "persistentVolumeClaim"
		FSTypeDownwardAPI           securityv1.FSType = "downwardAPI"
		FSTypeConfigMap             securityv1.FSType = "configMap"
		FSProjected                 securityv1.FSType = "projected"
	)

	return &securityv1.SecurityContextConstraints{
		ObjectMeta: metav1.ObjectMeta{
			Name: config.CSISCC,
		},
		ReadOnlyRootFilesystem:   false,
		RequiredDropCapabilities: []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"},
		RunAsUser: securityv1.RunAsUserStrategyOptions{
			Type: securityv1.RunAsUserStrategyType("RunAsAny"),
		},
		SELinuxContext: securityv1.SELinuxContextStrategyOptions{
			Type: securityv1.SELinuxContextStrategyType("RunAsAny"),
		},
		SupplementalGroups: securityv1.SupplementalGroupsStrategyOptions{
			Type: securityv1.SupplementalGroupsStrategyType("RunAsAny"),
		},
		Volumes: []securityv1.FSType{
			FSTypeHostPath,
			FSTypeEmptyDir,
			FSTypeSecret,
			FSTypePersistentVolumeClaim,
			FSTypeDownwardAPI,
			FSTypeConfigMap,
			FSProjected,
		},
		AllowHostDirVolumePlugin: true,
		AllowHostIPC:             false,
		AllowHostNetwork:         true,
		AllowHostPID:             false,
		AllowHostPorts:           false,
		// AllowPrivilegedEscalation: true, // Note: Not supported by the package, If not specificed, defaults to true.
		AllowPrivilegedContainer: true,
		AllowedCapabilities:      []corev1.Capability{},
		DefaultAddCapabilities:   []corev1.Capability{},
		FSGroup: securityv1.FSGroupStrategyOptions{
			Type: securityv1.FSGroupStrategyType("MustRunAs"),
		},
		Users: users,
	}
}

// GetNodeSelectors converts the given nodeselector array into a map.
func (c *CSIScaleOperator) GetNodeSelectors(nodeSelectorObj []v1.CSINodeSelector) map[string]string {

	nodeSelectors := make(map[string]string)

	if len(nodeSelectorObj) != 0 {
		for _, item := range nodeSelectorObj {
			nodeSelectors[item.Key] = item.Value
		}
	}

	return nodeSelectors
}

// GetNodeTolerations returns an array of kubernetes object of type corev1.Tolerations
func (c *CSIScaleOperator) GetNodeTolerations() []corev1.Toleration {

	tolerationsSeconds := config.TolerationsSeconds
	tolerations := []corev1.Toleration{
		{
			Key:               "node.kubernetes.io/unreachable",
			Operator:          "Exists",
			Effect:            "NoExecute",
			TolerationSeconds: &tolerationsSeconds,
		},
		{
			Key:               "node.kubernetes.io/not-ready",
			Operator:          "Exists",
			Effect:            "NoExecute",
			TolerationSeconds: &tolerationsSeconds,
		},
	}

	return tolerations
}

// GetPodAntiAffinity returns kubernetes podAntiAffinity for the sidecar controller pod.
func (c *CSIScaleOperator) GetPodAntiAffinity() *corev1.PodAntiAffinity {
	podAntiAffinity := corev1.PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
			{
				LabelSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      config.LabelApp,
							Operator: "In",
							Values:   []string{config.GetNameForResource(config.CSIController, c.Name)},
						},
					},
				},
				TopologyKey: "kubernetes.io/hostname",
			},
		},
	}

	return &podAntiAffinity
}

// GetLivenessProbe returns liveness probe information for sidecar controller.
func (c *CSIScaleOperator) GetLivenessProbe() *corev1.Probe {
	//tolerationsSeconds := config.TolerationsSeconds
	probe := corev1.Probe{
		FailureThreshold:    int32(1),
		InitialDelaySeconds: int32(30), // TODO: With increase in sidecar containers, initial delay needs to be increased.
		TimeoutSeconds:      int32(10),
		PeriodSeconds:       int32(20),
		Handler:             c.GetHandler(),
	}
	return &probe
}

// GetAttacherContainerPort returns port details for the attacher sidecar container.
func (c *CSIScaleOperator) GetAttacherContainerPort() []corev1.ContainerPort {
	ports := []corev1.ContainerPort{
		{
			ContainerPort: config.AttacherLeaderLivenessPort,
			Name:          "http-endpoint",
			Protocol:      c.GetProtocol(),
		},
	}
	return ports
}

// GetProvisionerContainerPort returns port details for the provisioner sidecar container.
func (c *CSIScaleOperator) GetProvisionerContainerPort() []corev1.ContainerPort {
	ports := []corev1.ContainerPort{
		{
			ContainerPort: config.ProvisionerLeaderLivenessPort,
			Name:          "http-endpoint",
			Protocol:      c.GetProtocol(),
		},
	}
	return ports
}

// GetResizerContainerPort returns port details for the resizer sidecar container.
func (c *CSIScaleOperator) GetResizerContainerPort() []corev1.ContainerPort {
	ports := []corev1.ContainerPort{
		{
			ContainerPort: config.ResizerLeaderLivenessPort,
			Name:          "http-endpoint",
			Protocol:      c.GetProtocol(),
		},
	}
	return ports
}

// GetSnapshotterContainerPort returns port details for the snapshotter sidecar container.
func (c *CSIScaleOperator) GetSnapshotterContainerPort() []corev1.ContainerPort {
	ports := []corev1.ContainerPort{
		{
			ContainerPort: config.SnapshotterLeaderLivenessPort,
			Name:          "http-endpoint",
			Protocol:      c.GetProtocol(),
		},
	}
	return ports
}

// GetProtocol returns the protocol to be used by liveness probe with httpGet request.
func (c *CSIScaleOperator) GetProtocol() corev1.Protocol {
	var protocol corev1.Protocol = "TCP"
	return protocol
}

// GetHandler returns a handler with httpGet information.
func (c *CSIScaleOperator) GetHandler() corev1.Handler {
	handler := corev1.Handler{
		HTTPGet: c.GetHTTPGetAction(),
	}
	return handler
}

// GetHTTPGetAction returns httpGet information for the liveness probe.
func (c CSIScaleOperator) GetHTTPGetAction() *corev1.HTTPGetAction {
	action := corev1.HTTPGetAction{
		Path: "/healthz/leader-election",
		Port: intstr.FromString("http-endpoint"),
	}
	return &action
}

// GetDeploymentStrategy returns update strategy details for kubernetes deployment.
func (c CSIScaleOperator) GetDeploymentStrategy() appsv1.DeploymentStrategy {
	strategy := appsv1.DeploymentStrategy{
		RollingUpdate: c.GetRollingUpdateDeployment(),
		Type:          c.GetDeploymentStrategyType(),
	}
	return strategy
}

// GetRollingUpdateDeployment returns rollingUpdate details. MaxSurge as 25% and MaxUnavailable as 50%.
func (c CSIScaleOperator) GetRollingUpdateDeployment() *appsv1.RollingUpdateDeployment {
	maxSurge := intstr.FromString("25%")
	maxUnavailable := intstr.FromString("50%")
	deploy := appsv1.RollingUpdateDeployment{
		MaxSurge:       &maxSurge,
		MaxUnavailable: &maxUnavailable,
	}
	return &deploy
}

// GetDeploymentStrategyType returns deployment strategy type as `RollingUpdate` for kubernetes deployment.
func (c *CSIScaleOperator) GetDeploymentStrategyType() appsv1.DeploymentStrategyType {
	var StrategyType appsv1.DeploymentStrategyType = "RollingUpdate"
	return StrategyType
}
