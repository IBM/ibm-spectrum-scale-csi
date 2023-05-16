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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CSIScaleOperatorSpec specifies the desired state of CSI
type CSIScaleOperatorSpec struct {

	// Note: Sidecar images are currently fetched by spec.attacher, spec.provisioner, spec.resizer, spec.snapshotter separately.
	// sidecars is a list of sidecar images.
	// // +listType=set
	// // +kubebuilder:validation:Optional
	// Sidecars []CSISidecar `json:"sidecars"`

	// attacher is the attacher sidecar image for CSI (actually attaches to the storage).
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Attacher Image",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	Attacher string `json:"attacher,omitempty"`

	// attacherNodeSelector is the node selector for attacher sidecar.
	// +kubebuilder:default:={{key:scale,value:`true`}}
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Attacher Node Selector",xDescriptors="urn:alm:descriptor:com.tectonic.ui:advanced"
	AttacherNodeSelector []CSINodeSelector `json:"attacherNodeSelector,omitempty"`

	// clusters is a collection of IBM Storage Scale cluster properties for the CSI driver to mount.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Clusters"
	Clusters []CSICluster `json:"clusters"`

	// driverRegistrar is the Sidecar container image for the IBM Storage Scale CSI plugin pods.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Driver Registrar",xDescriptors="urn:alm:descriptor:com.tectonic.ui:advanced"
	DriverRegistrar string `json:"driverRegistrar,omitempty"`

	// nodeMapping specifies mapping of K8s node with IBM Storage Scale node.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Node Mapping",xDescriptors="urn:alm:descriptor:com.tectonic.ui:advanced"
	NodeMapping []NodeMapping `json:"nodeMapping,omitempty"`

	// pluginNodeSelector is the node selector for IBM Storage Scale CSI plugin.
	// +kubebuilder:default:={{key:scale,value:`true`}}
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Plugin Node Selector",xDescriptors="urn:alm:descriptor:com.tectonic.ui:advanced"
	PluginNodeSelector []CSINodeSelector `json:"pluginNodeSelector,omitempty"`

	// provisioner is the provisioner sidecar image for CSI (actually issues provision requests).
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Provisioner Image",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	Provisioner string `json:"provisioner,omitempty"`

	// provisionerNodeSelector is the node selector for provisioner sidecar.
	// +kubebuilder:default:={{key:scale,value:`true`}}
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Provisioner Node Selector",xDescriptors="urn:alm:descriptor:com.tectonic.ui:advanced"
	ProvisionerNodeSelector []CSINodeSelector `json:"provisionerNodeSelector,omitempty"`

	// snapshotter is the snapshotter sidecar image for CSI (issues volume snapshot requests).
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Snapshotter Image",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	Snapshotter string `json:"snapshotter,omitempty"`

	// snapshotterNodeSelector is the snapshotter node selector for snapshotter sidecar.
	// +kubebuilder:default:={{key:scale,value:`true`}}
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Snapshotter Node Selector",xDescriptors="urn:alm:descriptor:com.tectonic.ui:advanced"
	SnapshotterNodeSelector []CSINodeSelector `json:"snapshotterNodeSelector,omitempty"`

	// resizer is the resizer sidecar image for CSI (issues volume expansion requests).
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Resizer Image",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	Resizer string `json:"resizer,omitempty"`

	// resizerNodeSelector is the node selector for resizer sidecar.
	// +kubebuilder:default:={{key:scale,value:`true`}}
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Resizer Node Selector",xDescriptors="urn:alm:descriptor:com.tectonic.ui:advanced"
	ResizerNodeSelector []CSINodeSelector `json:"resizerNodeSelector,omitempty"`

	// livenessprobe is the image for livenessProbe container (liveness probe is used to know when to restart a container).
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="LivenessProbe",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	LivenessProbe string `json:"livenessprobe,omitempty"`

	// spectrumScale is the image name for the IBM Storage Scale CSI node driver plugin container.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="IBM Storage Scale Image",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	SpectrumScale string `json:"spectrumScale,omitempty"`

	// A passthrough option that distributes an imagePullSecrets array to the
	// containers generated by the CSI scale operator. Please refer to official
	// k8s documentation for your environment for more details.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Image Pull Secrets",xDescriptors={"urn:alm:descriptor:com.tectonic.ui:label","urn:alm:descriptor:com.tectonic.ui:advanced"}
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	// Array of tolerations that will be distributed to CSI pods. Please refer to
	// official k8s documentation for your environment for more details.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Tolerations",xDescriptors="urn:alm:descriptor:com.tectonic.ui:advanced"
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// ControllerRepository string `json:"repository,omitempty"`

	// ControllerTag string `json:"tag,omitempty"`

	//	NodeRepository string `json:"repository"`
	//	NodeTag        string `json:"tag"`

	// node is a group of CSIScaleOperatorNodeSpec properties.
	// Node CSIScaleOperatorNodeSpec `json:"node,omitempty"`

	// affinity is a group of affinity scheduling rules.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Affinity",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// status defines the observed state of CSIScaleOperator
	// Status CSIScaleOperatorStatus `json:"status,omitempty"`

	// kubeletRootDirPath is the path for kubelet root directory.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Kubelet Root Directory Path",xDescriptors="urn:alm:descriptor:com.tectonic.ui:advanced"
	KubeletRootDirPath string `json:"kubeletRootDirPath,omitempty"`

	// PodSecurityPolicy name for CSI driver and sidecar pods.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="CSI Pod Security Policy Name",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	CSIpspname string `json:"csipspname,omitempty"`

	// consistencyGroupPrefix is a prefix of consistency group of an application.
	// This is expected to be an RFC4122 UUID value (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx in hexadecimal values)
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Consistency Group Prefix",xDescriptors="urn:alm:descriptor:com.tectonic.ui:hidden"
	CGPrefix string `json:"consistencyGroupPrefix,omitempty"`
}

// CSIScaleOperatorStatus defines the observed state of CSIScaleOperator
type CSIScaleOperatorStatus struct {

	/* TODO: Status should display driver state.
	// Phase is the driver running phase
	Phase           DriverPhase `json:"phase,omitempty"`
	ControllerReady bool        `json:"controllerReady,omitempty"`
	NodeReady       bool        `json:"nodeReady,omitempty"`
	Conditions []CSICondition `json:"conditions,omitempty"`
	*/

	// version is the current CSIDriver version installed by the operator.
	Versions []Version `json:"versions,omitempty"`

	// conditions contains the details for one aspect of the current state of this custom resource.
	//+operator-sdk:csv:customresourcedefinitions:type=status,displayName="Conditions",xDescriptors="urn:alm:descriptor:io.kubernetes.conditions"
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type Version struct {

	//name is the name of the particular operand this version is for.
	Name string `json:"name,omitempty"`

	// version of a particular operand that is currently being managed.
	Version string `json:"version,omitempty"`
}

/*
TODO: Unused code. Remove if not required.
// CSIScaleOperatorNodeSpec defines the desired state of CSIScaleOperatorNode
// +k8s:openapi-gen=true
type CSIScaleOperatorNodeSpec struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`

	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// // +listType=set

	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}
*/

/* Note: Uncomment when status.Phase is in use.
type DriverPhase string

const (
	DriverPhaseNone     DriverPhase = ""
	DriverPhaseCreating DriverPhase = "Creating"
	DriverPhaseRunning  DriverPhase = "Running"
	DriverPhaseFailed   DriverPhase = "Failed"
)
*/

/*
// Note: Uncomment this when spec.sidecars field is in use.
type CSISidecar struct {
	// The name of the CSI sidecar image
	Name string `json:"name"`

	// The repository of the CSI sidecar image
	Repository string `json:"repository"`

	// The tag of the CSI sidecar image
	Tag string `json:"tag"`

	// The pullPolicy of the CSI sidecar image
	// +kubebuilder:default:=IfNotPresent
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`
}
*/

/*
//  Note: Uncomment this when CSICondition is in use.
type CSICondition struct {
	// +optional
	// Indicates that the plugin is running
	Ready bool `json:"Ready"`
}
*/

// CSINodeSelector defines the fields of Node Selector
type CSINodeSelector struct {

	// Key for node selector
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Key",xDescriptors="urn:alm:descriptor:com.tectonic.ui:label"
	Key string `json:"key"`

	// Value for key
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Value",xDescriptors="urn:alm:descriptor:com.tectonic.ui:label"
	Value string `json:"value"`
}

/*
TODO: Unused code. Remove if not required.
type Toleration struct {

	// +optional

	// Node taint key name
	Key string `json:"key"`

	// +optional

	// Valid values are "Exists" and "Equal"
	Operator Operator `json:"operator"`

	// +optional

	// Required if operator is "Equal"
	Value string `json:"value"`

	// +optional

	// Valid values are "NoSchedule", "PreferNoSchedule" and "NoExecute".
	// An empty effect matches all effects with given key.
	// // +kubebuilder:validation:Enum=NoSchedule;PreferNoSchedule;NoExecute
	Effect string `json:"effect"`
}
*/

/*
// Note: Uncomment this when Toleration structure is in use
type Effect string

const (

	// TODO: add doc
	NoSchedule Effect = "NoSchedule"

	// TODO: add doc
	PreferNoSchedule Effect = "PreferNoSchedule"

	// TODO: add doc
	NoExecute Effect = "NoExecute"

	// TODO: add doc
	None Effect = ""
)


// +kubebuilder:validation:Enum=Exists;Equal
type Operator string

const (

	// TODO: add doc
	Exists Operator = "Exists"

	// TODO: add doc
	Equal Operator = "Equal"
)
*/

// Defines mapping between kubernetes node and IBM Storage Scale nodes
type NodeMapping struct {

	// k8sNode is the name of the kubernetes node
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Kubernetes Node",xDescriptors="urn:alm:descriptor:com.tectonic.ui:label"
	K8sNode string `json:"k8sNode"`

	// spectrumscaleNode is the name of the IBM Storage Scale node
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="IBM Storage Scale Node",xDescriptors="urn:alm:descriptor:com.tectonic.ui:label"
	SpectrumscaleNode string `json:"spectrumscaleNode"`
}

// Defines the fields of a IBM Storage Scale cluster specification
type CSICluster struct {

	// cacert is the name of the configMap storing GUI certificates. Mandatory if secureSslMode is true.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="CA Certificate Resource Name",xDescriptors="urn:alm:descriptor:io.kubernetes:ConfigMap"
	Cacert string `json:"cacert,omitempty"` // TODO: Rename to CACert or caCert

	// id is the cluster ID of the IBM Storage Scale cluster.
	// +kubebuilder:validation:MaxLength:=20
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Cluster ID",xDescriptors="urn:alm:descriptor:com.tectonic.ui:label"
	Id string `json:"id"` // TODO: Rename to ID or id

	// primary is the primary file system for the IBM Storage Scale cluster.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Primary",xDescriptors="urn:alm:descriptor:com.tectonic.ui:label"
	Primary *CSIFilesystem `json:"primary,omitempty"`

	// restApi is a collection of targets for REST calls
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="REST API",xDescriptors="urn:alm:descriptor:com.tectonic.ui:label"
	RestApi []RestApi `json:"restApi"` // TODO: Rename to RESTApi or restApi

	// secret is the name of the basic-auth secret containing credentials to connect to IBM Storage Scale REST API server.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secrets",xDescriptors="urn:alm:descriptor:io.kubernetes:Secret"
	Secrets string `json:"secrets"` // TODO: Secrets should be Singular

	// secureSslMode specifies if a secure SSL connection to connect to IBM Storage Scale cluster is required.
	// +kubebuilder:default:=false
	// +kubebuilder:validation:Enum:=true;false
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="Secure SSL Mode",xDescriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	SecureSslMode bool `json:"secureSslMode"`
}

// Defines the fields for CSI for IBM Storage Scale file system
type CSIFilesystem struct {

	// Inode limit for Primary Fileset
	InodeLimit string `json:"inodeLimit,omitempty"`

	// The name of the primary CSIFilesystem
	PrimaryFs string `json:"primaryFs,omitempty"`

	// The name of the primary fileset, created in primaryFs
	PrimaryFset string `json:"primaryFset,omitempty"`

	// Remote IBM Storage Scale cluster ID
	RemoteCluster string `json:"remoteCluster,omitempty"`
}

// Defines the fields for REST API server information.
type RestApi struct {

	// guiHost is the hostname/IP of the IBM Storage Scale GUI node.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="GUI Host",xDescriptors="urn:alm:descriptor:com.tectonic.ui:label"
	GuiHost string `json:"guiHost"`

	// guiPort is the port number of the IBM Storage Scale GUI node.
	//+operator-sdk:csv:customresourcedefinitions:type=spec,displayName="GUI Port",xDescriptors="urn:alm:descriptor:com.tectonic.ui:number"
	GuiPort int `json:"guiPort,omitempty"`
}

// // +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="TODO: Add description."

// +kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.versions[0].version`,description="CSIDriver version."
// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="Success",type=string,JSONPath=`.status.conditions[?(@ "status")].status`,description="CSI driver resource creation status."
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=cso, categories=scale, scope=Namespaced

// CSIScaleOperator is the Schema for the csiscaleoperators API
// +operator-sdk:csv:customresourcedefinitions:displayName="IBM Storage Scale CSI Driver",resources={{Deployment,v1beta2},{DaemonSet,v1beta2},{Pod,v1},{ConfigMap,v1}}
type CSIScaleOperator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CSIScaleOperatorSpec   `json:"spec,omitempty"`
	Status CSIScaleOperatorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CSIScaleOperatorList contains a list of CSIScaleOperator
type CSIScaleOperatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CSIScaleOperator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CSIScaleOperator{}, &CSIScaleOperatorList{})
}

type CSIReason string

const (
	CSIConfigured CSIReason = "CSIConfigured"
	Unknown       CSIReason = "Unknown"

	GetFileSystemFailed          CSIReason = "GetFileSystemFailed"
	FilesetRefreshFailed         CSIReason = "FilesetRefreshFailed"
	GetFilesetFailed             CSIReason = "GetFilesetFailed"
	CreateDirFailed              CSIReason = "CreateDirFailed"
	CreateFilesetFailed          CSIReason = "CreateFilesetFailed"
	LinkFilesetFailed            CSIReason = "LinkFilesetFailed"
	ValidationFailed             CSIReason = "ValidationFailed"
	GUIConnFailed                CSIReason = "GUIConnFailed"
	ClusterIDMismatch            CSIReason = "ClusterIDMismatch"
	PrimaryClusterUndefined      CSIReason = "PrimaryClusterUndefined"
	GetRemoteFileSystemFailed    CSIReason = "GetRemoteFileSystemFailed"
	PrimaryClusterStanzaModified CSIReason = "PrimaryClusterStanzaModified"
	UnmarshalFailed              CSIReason = "UnmarshalFailed"

	//for create/update/delete/get operations on k8s resources
	GetFailed    CSIReason = "GetFailed"
	CreateFailed CSIReason = "CreateFailed"
	UpdateFailed CSIReason = "UpdateFailed"
	DeleteFailed CSIReason = "DeleteFailed"
)
