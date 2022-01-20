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
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

// GenerateAttacherServiceAccount creates a kubernetes service account for the attacher service
// and modify the service account to use secret as an imagePullSecret.
// It returns an object of type *corev1.ServiceAccount.
func (c *CSIScaleOperator) GenerateAttacherServiceAccount() *corev1.ServiceAccount {

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.GetNameForResource(config.CSIAttacherServiceAccount, c.Name),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
	}
}

// GenerateProvisionerServiceAccount creates a kubernetes service account for the provisioner service
// and modify the service account to use secret as an imagePullSecret.
// It returns an object of type *corev1.ServiceAccount.
func (c *CSIScaleOperator) GenerateProvisionerServiceAccount() *corev1.ServiceAccount {

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.GetNameForResource(config.CSIProvisionerServiceAccount, c.Name),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
	}
}

// GenerateSnapshotterServiceAccount creates a kubernetes service account for the snapshotter service
// and modify the service account to use secret as an imagePullSecret.
// It returns an object of type *corev1.ServiceAccount.
func (c *CSIScaleOperator) GenerateSnapshotterServiceAccount() *corev1.ServiceAccount {

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.GetNameForResource(config.CSISnapshotterServiceAccount, c.Name),
			Namespace: c.Namespace,
			Labels:    c.GetLabels(),
		},
	}
}

// GenerateResizerServiceAccount creates a kubernetes service account for the resizer service
// and modify the service account to use secret as an imagePullSecret.
// It returns an object of type *corev1.ServiceAccount.
func (c *CSIScaleOperator) GenerateResizerServiceAccount() *corev1.ServiceAccount {

	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      config.GetNameForResource(config.CSIResizerServiceAccount, c.Name),
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

// GenerateProvisionerClusterRole returns a kubernetes clusterrolebinding object for the provisioner service.
func (c *CSIScaleOperator) GenerateProvisionerClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   config.GetNameForResource(config.Provisioner, c.Name),
			Labels: c.GetLabels(),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIProvisionerServiceAccount, c.Name),
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
				Name:      config.GetNameForResource(config.CSIAttacherServiceAccount, c.Name),
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
				Name:      config.GetNameForResource(config.CSISnapshotterServiceAccount, c.Name),
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
			Name: config.GetNameForResource(config.Resizer, c.Name),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      config.GetNameForResource(config.CSIResizerServiceAccount, c.Name),
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
func (s *CSIScaleOperator) GenerateSecurityContextConstraint(users []string) *securityv1.SecurityContextConstraints {

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

// getNodeSelector converts the given nodeselector array into a map.
func (s *CSIScaleOperator) GetNodeSelectors(nodeSelectorObj []v1.CSINodeSelector) map[string]string {

	nodeSelectors := make(map[string]string)

	if len(nodeSelectorObj) != 0 {
		for _, item := range nodeSelectorObj {
			nodeSelectors[item.Key] = item.Value
		}
	}

	return nodeSelectors
}
