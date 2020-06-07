/**
 * Copyright 2019 IBM Corp.
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

package connectors

type GetClusterResponse struct {
	Cluster Cluster `json:"cluster,omitempty"`
	Status  Status  `json:"status,omitempty"`
}
type Cluster struct {
	CesSummary     CesSummary        `json:"cesSummary,omitempty"`
	ClusterSummary ClusterSummary    `json:"clusterSummary"`
	CnfsSummary    CnfsSummary       `json:"cnfsSummary,omitempty"`
	Links          map[string]string `json:"links,omitempty"`
	Nodes          []ClusterNode     `json:"nodes,omitempty"`
}

type CesSummary struct {
	EnabledServices string `json:"enabledServices"`
	AddressPolicy   string `json:"addressPolicy,omitempty"`
	CesSharedRoot   string `json:"cesSharedRoot,omitempty"`
	LogLevel        int    `json:"logLevel,omitempty"`
}
type ClusterSummary struct {
	ClusterID       uint64 `json:"clusterId"`
	ClusterName     string `json:"clusterName,omitempty"`
	PrimaryServer   string `json:"primaryServer,omitempty"`
	RcpPath         string `json:"rcpPath,omitempty"`
	RepositoryType  string `json:"repositoryType,omitempty"`
	RshPath         string `json:"rshPath,omitempty"`
	SecondaryServer string `json:"secondaryServer,omitempty"`
	UIDDomain       string `json:"uidDomain,omitempty"`
}

type CnfsSummary struct {
	CNFSGanesha        string `json:"cnfsGanesha,omitempty"`
	CNFSMonitorEnabled string `json:"cnfsMonitorEnabled,omitempty"`
	CNFSMountdPort     string `json:"cnfsMountdPort,omitempty"`
	CNFSNFSDprocs      string `json:"cnfsNFSDprocs,omitempty"`
	CNFSReboot         string `json:"cnfsReboot,omitempty"`
	CNFSSharedReboot   string `json:"cnfsSharedReboot,omitempty"`
}

type ClusterNode struct {
	AdminLoginName string            `json:"adminLoginName,omitempty"`
	AdminNodeName  string            `json:"adminNodeName,omitempty"`
	CESNode        CESNode           `json:"cesNode,omitempty"`
	CNFSNode       CNFSNode          `json:"cnfsNode,omitempty"`
	DaemonNodeName string            `json:"daemonNodeName,omitempty"`
	Designation    string            `json:"designation,omitempty"`
	IPAddress      string            `json:"ipAddress,omitempty"`
	Links          map[string]string `json:"links,omitempty"`
	NodeNumber     int               `json:"nodeNumber,omitempty"`
	OtherNodeRoles string            `json:"otherNodeRoles,omitempty"`
}

type CESNode struct {
	CESGroup  string `json:"cesGroup,omitempty"`
	CESIPList string `json:"cesIpList,omitempty"`
	CESState  string `json:"cesState,omitempty"`
}
type CNFSNode struct {
	CNFSGroupID string `json:"cnfsGroupID,omitempty"`
	CNFSIPList  string `json:"cnfsIpList,omitempty"`
	CNFSState   string `json:"cnfsState,omitempty"`
}

type Status struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Job struct {
	Result    Respresult  `json:"result,omitempty"`
	Request   Resprequest `json:"request,omitempty"`
	JobID     uint64      `json:"jobid,omitempty"`
	Submitted string      `json:"submitted,omitempty"`
	Completed string      `json:"completed,omitempty"`
	Status    string      `json:"status,omitempty"`
}

type Respresult struct {
	Commands []string `json:"commands,omitempty"`
	Progress []string `json:"progress,omitempty"`
	ExitCode int      `json:"exitCode,omitempty"`
	Stderr   []string `json:"stderr,omitempty"`
	Stdout   []string `json:"stdout,omitempty"`
}

type Resprequest struct {
	Type string                 `json:"type,omitempty"`
	Url  string                 `json:"url,omitempty"`
	Data map[string]interface{} `json:"data,omitempty"`
}

type GetFilesystemResponse struct {
	FileSystems []FileSystem `json:"filesystems,omitempty"`
	Status      Status       `json:"status,omitempty"`
}

type FileSystem struct {
	ACLSemantics               string            `json:"ACLSemantics,omitempty"`
	DMAPIEnabled               string            `json:"DMAPIEnabled,omitempty"`
	UID                        string            `json:"UID,omitempty"`
	AdditionalMountOptions     string            `json:"additionalMountOptions,omitempty"`
	AutomaticMountOption       string            `json:"automaticMountOption,omitempty"`
	BlockAllocationType        string            `json:"blockAllocationType,omitempty"`
	BlockSize                  string            `json:"blockSize,omitempty"`
	CreateTime                 string            `json:"createTime,omitempty"`
	DefaultMetadataReplicas    string            `json:"defaultMetadataReplicas,omitempty"`
	DefaultMountPoint          string            `json:"defaultMountPoint,omitempty"`
	DefaultQuotasEnabled       string            `json:"defaultQuotasEnabled,omitempty"`
	Disks                      string            `json:"disks,omitempty"`
	Encryption                 string            `json:"encryption,omitempty"`
	ExactMTime                 string            `json:"exactMTime,omitempty"`
	FastEAEnabled              string            `json:"FastEAEnabled,omitempty"`
	FileLockingSemantics       string            `json:"fileLockingSemantics,omitempty"`
	FilesetdfEnabled           string            `json:"filesetdfEnabled,omitempty"`
	FilesystemHighestSupported string            `json:"filesystemHighestSupported,omitempty"`
	FilesystemName             string            `json:"filesystemName,omitempty"`
	FilesystemVersion          string            `json:"filesystemVersion,omitempty"`
	FilesystemVersionLocal     string            `json:"filesystemVersionLocal,omitempty"`
	FilesystemVersionManager   string            `json:"filesystemVersionManager,omitempty"`
	FilesystemVersionOriginal  string            `json:"filesystemVersionOriginal,omitempty"`
	IndirectBlockSize          string            `json:"indirectBlockSize,omitempty"`
	InodeSize                  string            `json:"inodeSize,omitempty"`
	Is4KAligned                string            `json:"is4KAligned,omitempty"`
	Links                      map[string]string `json:"links,omitempty"`
	LogReplicas                string            `json:"logReplicas,omitempty"`
	LogfileSize                string            `json:"logfileSize,omitempty"`
	MaxDataReplicas            string            `json:"maxDataReplicas,omitempty"`
	MaxMetadataReplicas        string            `json:"maxMetadataReplicas,omitempty"`
	MaxNumberOfInodes          string            `json:"maxNumberOfInodes,omitempty"`
	MaxSnapshotId              string            `json:"maxSnapshotId,omitempty"`
	MinFragmentSize            string            `json:"minFragmentSize,omitempty"`
	MountPriority              string            `json:"mountPriority,omitempty"`
	NumNodes                   string            `json:"numNodes,omitempty"`
	OtherPools                 []Pool            `json:"otherPools,omitempty"`
	PerfilesetQuotas           string            `json:"perfilesetQuotas,omitempty"`
	QuotasAccountingEnable     string            `json:"quotasAccountingEnable,omitempty"`
	QuotasEnforced             string            `json:"quotasEnforced,omitempty"`
	RapidRepairEnabled         string            `json:"rapidRepairEnabled,omitempty"`
	StoragePools               string            `json:"storagePools,omitempty"`
	StrictReplication          string            `json:"strictReplication,omitempty"`
	SuppressATime              string            `json:"suppressATime,omitempty"`
	WriteCacheThreshold        string            `json:"writeCacheThreshold,omitempty"`
}
type Pool struct {
	BlockSize       string `json:"blockSize,omitempty"`
	MinFragmentSize string `json:"minFragmentSize,omitempty"`
}

type GetFilesetResponse struct {
	Filesets []Fileset `json:"filesets,omitempty"`
	Status   Status    `json:"status,omitempty"`
}

type Fileset struct {
	AFM    AFM               `json:"afm,omitempty"`
	Config FilesetConfig     `json:"config,omitempty"`
	Links  map[string]string `json:"links,omitempty"`
	State  FilesetState      `json:"state,omitempty"`
}

type AFM struct {
	AFMPrimaryID                 string `json:"afmPrimaryID"`
	AFMMode                      string `json:"afmMode"`
	AFMTarget                    string `json:"afmTarget"`
	AFMAsyncDelay                int    `json:"afmAsyncDelay"`
	AFMDirLookupRefreshInterval  int    `json:"afmDirLookupRefreshInterval"`
	AFMDirOpenRefreshInterval    int    `json:"afmDirOpenRefreshInterval"`
	AFMExpirationTimeout         int    `json:"afmExpirationTimeout"`
	AFMFileLookupRefreshInterval int    `json:"afmFileLookupRefreshInterval"`
	AFMNumFlushThreads           int    `json:"afmNumFlushThreads"`
	AFMParallelReadChunkSize     int    `json:"afmParallelReadChunkSize"`
	AFMParallelReadThreshold     int    `json:"afmParallelReadThreshold"`
	AFMParallelWriteChunkSize    int    `json:"afmParallelWriteChunkSize"`
	AFMParallelWriteThreshold    int    `json:"afmParallelWriteThreshold"`
	AFMPrefetchThreshold         int    `json:"afmPrefetchThreshold"`
	AFMRPO                       int    `json:"afmRPO"`
	AFMEnableAutoEviction        bool   `json:"afmEnableAutoEviction"`
	AFMShowHomeSnapshots         bool   `json:"afmShowHomeSnapshots"`
}

type FilesetConfig struct {
	Comment              string `json:"comment,omitempty"`
	FilesetName          string `json:"filesetName,omitempty"`
	FilesystemName       string `json:"filesystemName,omitempty"`
	IAMMode              string `json:"iamMode,omitempty"`
	INodeSpace           string `json:"inodeSpace,omitempty"`
	MaxNumInodes         string `json:"maxNumInodes,omitempty"`
	AllocInodes          string `json:"allocInodes,omitempty"`
	Owner                string `json:"owner,omitempty"`
	Path                 string `json:"path,omitempty"`
	PermissionChangeMode string `json:"permissionChangeMode,omitempty"`
	Permissions          string `json:"permissions,omitempty"`
}
type FilesetState struct {
	AFMState          string `json:"afmState,omitempty"`
	Created           string `json:"created,omitempty"`
	DataInKB          string `json:"dataInKB,omitempty"`
	FreeInodes        string `json:"freeInodes,omitempty"`
	ID                string `json:"id,omitempty"`
	InodeSpaceMAsk    string `json:"inodeSpaceMask,omitempty"`
	Inodes            string `json:"inodes,omitempty"`
	IsInodeSpaceOwner string `json:"isInodeSpaceOwner,omitempty"`
	ParentID          string `json:"parentId,omitempty"`
	RootInode         string `json:"rootInode,omitempty"`
	SnapID            string `json:"snapID,omitempty"`
	Status            string `json:"status,omitempty"`
}

type CreateFilesetResquest struct {
	AFM    AFM           `json:"afm,omitempty"`
	Config FilesetConfig `json:"config,omitempty"`
}
type GenericResponse struct {
	Status Status `json:"status,omitempty"`
	Jobs   []Job  `json:"jobs,omitempty"`
}

type NfsExportRequest struct {
	Path         string   `json:"path,omitempty"`
	ClientDetail []string `json:"nfsClients,omitempty"`
}

type GetQuotaResponse struct {
	Links  map[string]string `json:"links,omitempty"`
	Quotas []Quota           `json:"quotas,omitempty"`
	Status Status            `json:"status,omitempty"`
}
type SetQuotaRequest struct {
	BlockGracePeriod string `json:"blockGracePeriod,omitempty"`
	BlockHardLimit   string `json:"blockHardLimit,omitempty"`
	BlockSoftLimit   string `json:"blockSoftLimit,omitempty"`
	FilesGracePeriod string `json:"filesGracePeriod,omitempty"`
	FilesHardLimit   string `json:"filesHardLimit,omitempty"`
	FilesSoftLimit   string `json:"filesSoftLimit,omitempty"`
	FilesetName      string `json:"filesetName,omitempty"`
	FilesystemName   string `json:"filesystemName,omitempty"`
	ObjectName       string `json:"objectName,omitempty"`
	OperationType    string `json:"operationType,omitempty"`
	QuotaType        string `json:"quotaType,omitempty"`
}

type Quota struct {
	BlockGrace     string `json:"blockGrace,omitempty"`
	BlockInDoubt   string `json:"blockInDoubt,omitempty"`
	BlockLimit     string `json:"blockLimit,omitempty"`
	BlockQuota     string `json:"blockQuota,omitempty"`
	BlockUsage     string `json:"blockUsage,omitempty"`
	DefQuota       string `json:"defQuota,omitempty"`
	FilesGrace     string `json:"filesGrace,omitempty"`
	FilesInDoubt   string `json:"filesInDoubt,omitempty"`
	FilesLimit     string `json:"filesLimit,omitempty"`
	FilesQuota     string `json:"filesQuota,omitempty"`
	FilesUsage     string `json:"filesUsage,omitempty"`
	FilesetId      string `json:"filesetId,omitempty"`
	FilesetName    string `json:"filesetName,omitempty"`
	FilesystemName string `json:"filesystemName,omitempty"`
	ObjectID       string `json:"objectID,omitempty"`
	ObjectName     string `json:"objectName,omitempty"`
	QuotaType      string `json:"quotaType,omitempty"`
}

type GetNodesResponse struct {
	Nodes  []Node `json:"nodes,omitempty"`
	Status Status `json:"status,omitempty"`
}
type Node struct {
	AdminLogin        string            `json:"adminLogin,omitempty"`
	CESNode           string            `json:"cesNode,omitempty"`
	ClientNode        string            `json:"clientNode,omitempty"`
	CNFSInterface     string            `json:"cnfsInterface,omitempty"`
	CNFSState         string            `json:"cnfsState,omitempty"`
	DaemonIPAddress   string            `json:"daemonIPAddress,omitempty"`
	DaemonInterface   string            `json:"daemonInterface,omitempty"`
	DesignatedLicence string            `json:"designatedLicense,omitempty"`
	GatewayNode       string            `json:"gatewayNode,omitempty"`
	GPFSState         string            `json:"gpfsState,omitempty"`
	Links             map[string]string `json:"links,omitempty"`
	ManagerNode       string            `json:"managerNode,omitempty"`
	NodeName          string            `json:"nodeName,omitempty"`
	NodeNumber        string            `json:"nodeNumber,omitempty"`
	OSName            string            `json:"osName,omitempty"`
	ProductVersion    string            `json:"productVersion,omitempty"`
	QuorumNode        string            `json:"quorumNode,omitempty"`
	SNMPNode          string            `json:"snmpNode,omitempty"`
}

/* Version 2*/

type GetFilesystemResponse_v2 struct {
	FileSystems []FileSystem_v2 `json:"filesystems,omitempty"`
	Status      Status          `json:"status,omitempty"`
	Paging      Pages           `json:"paging,omitempty"`
}

type OwnerInfo struct {
	User  string `json:"user,omitempty"`
	UID   int    `json:"uid,omitempty"`
	Group string `json:"group,omitempty"`
	GID   int    `json:"gid,omitempty"`
}

type OwnerResp_v2 struct {
	Status Status    `json:"status,omitempty"`
	Owner  OwnerInfo `json:"owner,omitempty"`
}

type BlockInfo struct {
	Pools               string `json:"pools,omitempty"`
	Disks               string `json:"disks,omitempty"`
	BlockSize           int    `json:"blockSize,omitempty"`
	MetaDataBlockSize   int    `json:"metaDataBlockSize,omitempty"`
	IndirectBlockSize   int    `json:"indirectBlockSize,omitempty"`
	MinFragmentSize     int    `json:"minFragmentSize,omitempty"`
	InodeSize           int    `json:"inodeSize,omitempty"`
	LogfileSize         int    `json:"logfileSize,omitempty"`
	WriteCacheThreshold int    `json:"writeCacheThreshold,omitempty"`
}

type MountInfo struct {
	MountPoint             string `json:"mountPoint,omitempty"`
	AutomaticMountOption   string `json:"automaticMountOption,omitempty"`
	AdditionalMountOptions string `json:"additionalMountOptions,omitempty"`
	MountPriority          int    `json:"mountPriority,omitempty"`
	//	DriveLetter            string `json:"driveLetter,omitempty"`
	RemoteDeviceName string   `json:"remoteDeviceName,omitempty"`
	NodesMounted     []string `json:"nodesMountedReadWrite,omitempty"`
	ReadOnly         bool     `json:"readOnly,omitempty"`
	Status           string   `json:"status,omitempty"`
}

type QuotaInfo struct {
	QuotasAccountingEnabled string `json:"quotasAccountingEnabled,omitempty"`
	QuotasEnforced          string `json:"quotasEnforced,omitempty"`
	DefaultQuotasEnabled    string `json:"defaultQuotasEnabled,omitempty"`
	PerfilesetQuotas        bool   `json:"perfilesetQuotas,omitempty"`
	FilesetdfEnabled        bool   `json:"filesetdfEnabled,omitempty"`
}

type SettingInfo struct {
	BlockAllocationType  string `json:"blockAllocationType,omitempty"`
	FileLockingSemantics string `json:"fileLockingSemantics,omitempty"`
	SuppressATime        string `json:"suppressATime,omitempty"`
	StripeMethod         string `json:"stripeMethod"`
	ACLSemantics         string `json:"ACLSemantics,omitempty"`
	NumNodes             int    `json:"numNodes,omitempty"`
	MaxNumberOfInodes    int    `json:"maxNumberOfInodes,omitempty"`
	ExactMTime           bool   `json:"exactMTime,omitempty"`
	FastEAEnabled        bool   `json:"fastEAEnabled,omitempty"`
	Encryption           bool   `json:"encryption,omitempty"`
	Is4KAligned          bool   `json:"is4KAligned,omitempty"`
	RapidRepairEnabled   bool   `json:"rapidRepairEnabled,omitempty"`
	StripedLogs          bool   `json:"stripedLogs,omitempty"`
	DMAPIEnabled         bool   `json:"DMAPIEnabled,omitempty"`
}

type ReplicationInfo struct {
	DefaultMetadataReplicas int    `json:"defaultMetadataReplicas,omitempty"`
	MaxMetadataReplicas     int    `json:"maxMetadataReplicas,omitempty"`
	DefaultDataReplicas     int    `json:"defaultDataReplicas,omitempty"`
	MaxDataReplicas         int    `json:"maxDataReplicas,omitempty"`
	StrictReplication       string `json:"strictReplication,omitempty"`
	LogReplicas             int    `json:"logReplicas,omitempty"`
}

type FileSystem_v2 struct {
	Oid         int             `json:"oid,omitempty"`
	UUID        string          `json:"uuid,omitempty"`
	Name        string          `json:"name,omitempty"`
	Version     string          `json:"version,omitempty"`
	Type        string          `json:"type,omitempty"`
	CreateTime  string          `json:"createTime,omitempty"`
	Block       BlockInfo       `json:"block,omitempty"`
	Mount       MountInfo       `json:"mount,omitempty"`
	Replication ReplicationInfo `json:"replication,omitempty"`
	Quota       QuotaInfo       `json:"quota,omitempty"`
	Settings    SettingInfo     `json:"settings,omitempty"`
}

type Pages struct {
	Next string `json:"next,omitempty"`
}

type GetNodesResponse_v2 struct {
	Nodes  []Node_v2 `json:"nodes,omitempty"`
	Status Status    `json:"status,omitempty"`
	Paging Pages     `json:"paging,omitempty"`
}

type NodeConfig struct {
	AdminLoginName    string `json:"adminLoginName,omitempty"`
	DesignatedLicense string `json:"designatedLicense,omitempty"`
	RequiredLicense   string `json:"requiredLicense,omitempty"`
}

type NodeStatus struct {
	OSName         string `json:"osName,omitempty"`
	NodeState      string `json:"nodeState,omitempty"`
	GPFSState      string `json:"gpfsState,omitempty"`
	ProductVersion string `json:"productVersion,omitempty"`
}

type Network struct {
	AdminIPAddress  string `json:"adminIPAddress,omitempty"`
	DaemonNodeName  string `json:"daemonNodeName,omitempty"`
	DaemonIPAddress string `json:"daemonIPAddress,omitempty"`
	GetcnfsNodeName string `json:"getcnfsNodeName,omitempty"`
}

type NodeRoles struct {
	SNMPNode         string `json:"snmpNode,omitempty"`
	ManagerNode      string `json:"managerNode,omitempty"`
	GatewayNode      string `json:"gatewayNode,omitempty"`
	CNFSNode         string `json:"cnfsNode,omitempty"`
	ClientNode       string `json:"clientNode,omitempty"`
	CESNode          string `json:"cesNode,omitempty"`
	QuorumNode       string `json:"quorumNode,omitempty"`
	CloudGatewayNode string `json:"cloudGatewayNode,omitempty"`
	OtherNodeRoles   string `json:"otherNodeRoles,omitempty"`
	Designation      string `json:"designation,omitempty"`
}

type Node_v2 struct {
	AdminNodename string     `json:"adminNodename,omitempty"`
	NodeNumber    int        `json:"nodeNumber,omitempty"`
	Config        NodeConfig `json:"config,omitempty"`
	Status        NodeStatus `json:"status,omitempty"`
	Network       Network    `json:"network,omitempty"`
	Roles         NodeRoles  `json:"roles,omitempty"`
	CNFSInfo      CNFSNode   `json:"cnfsInfo,omitempty"`
	CESInfo       CESNode    `json:"cesInfo,omitempty"`
}

type Fileset_v2 struct {
	AFM         AFM              `json:"afm,omitempty"`
	Config      FilesetConfig_v2 `json:"config,omitempty"`
	FilesetName string           `json:"filesetName,omitempty"`
}

type GetFilesetResponse_v2 struct {
	Filesets []Fileset_v2 `json:"filesets,omitempty"`
	Status   Status       `json:"status,omitempty"`
	Paging   Pages        `json:"paging,omitempty"`
}

type FilesetConfig_v2 struct {
	FilesetName          string `json:"filesetName,omitempty"`
	FilesystemName       string `json:"filesystemName,omitempty"`
	Path                 string `json:"path,omitempty"`
	InodeSpace           int    `json:"inodeSpace,omitempty"`
	MaxNumInodes         int    `json:"maxNumInodes,omitempty"`
	PermissionChangeMode string `json:"permissionChangeMode,omitempty"`
	Comment              string `json:"comment,omitempty"`
	IamMode              string `json:"iamMode,omitempty"`
	Oid                  int    `json:"oid,omitempty"`
	Id                   int    `json:"id,omitempty"`
	Status               string `json:"status,omitempty"`
	ParentId             int    `json:"parentId,omitempty"`
	Created              string `json:"created,omitempty"`
	IsInodeSpaceOwner    bool   `json:"isInodeSpaceOwner,omitempty"`
	InodeSpaceMask       int    `json:"inodeSpaceMask,omitempty"`
	SnapID               int    `json:"snapID,omitempty"`
	RootInode            int    `json:"rootInode,omitempty"`
}

type GetQuotaResponse_v2 struct {
	Quotas []Quota_v2 `json:"quotas,omitempty"`
	Status Status     `json:"status,omitempty"`
	Paging Pages      `json:"paging,omitempty"`
}

type Quota_v2 struct {
	QuotaID        int    `json:"quotaID,omitempty"`
	FilesystemName string `json:"filesystemName,omitempty"`
	FilesetName    string `json:"filesetName,omitempty"`
	QuotaType      string `json:"quotaType,omitempty"`
	ObjectName     string `json:"objectName,omitempty"`
	ObjectId       int    `json:"objectId,omitempty"`
	BlockUsage     int    `json:"blockUsage,omitempty"`
	BlockLimit     int    `json:"blockLimit,omitempty"`
	BlockQuota     int    `json:"blockQuota,omitempty"`
	BlockInDoubt   int    `json:"blockInDoubt,omitempty"`
	BlockGrace     string `json:"blockGrace,omitempty"`
	FilesUsage     int    `json:"filesUsage,omitempty"`
	FilesQuota     int    `json:"filesQuota,omitempty"`
	FilesLimit     int    `json:"filesLimit,omitempty"`
	FilesInDoubt   int    `json:"filesInDoubt,omitempty"`
	FilesGrace     string `json:"filesGrace,omitempty"`
	DefaultQuota   string `json:"defaultQuota,omitempty"`
}

type SetQuotaRequest_v2 struct {
	BlockGracePeriod string `json:"blockGracePeriod,omitempty"`
	BlockHardLimit   string `json:"blockHardLimit,omitempty"`
	BlockSoftLimit   string `json:"blockSoftLimit,omitempty"`
	FilesGracePeriod string `json:"filesGracePeriod,omitempty"`
	FilesHardLimit   string `json:"filesHardLimit,omitempty"`
	FilesSoftLimit   string `json:"filesSoftLimit,omitempty"`
	OperationType    string `json:"operationType,omitempty"`
	QuotaType        string `json:"quotaType,omitempty"`
	ObjectName       string `json:"objectName,omitempty"`
}

type LinkFilesetRequest struct {
	Path string `json:"path,omitempty"`
}

type UnlinkFilesetRequest struct {
	Force bool `json:"force,omitempty"`
}

type CreateFilesetRequest struct {
	FilesetName                  string `json:"filesetName,omitempty"`
	Path                         string `json:"path,omitempty"`
	Owner                        string `json:"owner,omitempty"`
	Permissions                  string `json:"permissions,omitempty"`
	InodeSpace                   string `json:"inodeSpace,omitempty"`
	MaxNumInodes                 string `json:"maxNumInodes,omitempty"`
	AllocInodes                  string `json:"allocInodes,omitempty"`
	PermissionChangeMode         string `json:"permissionChangeMode,omitempty"`
	Comment                      string `json:"comment,omitempty"`
	IamMode                      string `json:"iamMode,omitempty"`
	AfmTarget                    string `json:"afmTarget,omitempty"`
	AfmAsyncDelay                string `json:"afmAsyncDelay,omitempty"`
	AfmDirLookupRefreshInterval  string `json:"afmDirLookupRefreshInterval,omitempty"`
	AfmDirOpenRefreshInterval    string `json:"afmDirOpenRefreshInterval,omitempty"`
	AfmEnableAutoEviction        string `json:"afmEnableAutoEviction,omitempty"`
	AfmExpirationTimeout         string `json:"afmExpirationTimeout,omitempty"`
	AfmFileLookupRefreshInterval string `json:"afmFileLookupRefreshInterval,omitempty"`
	AfmMode                      string `json:"afmMode,omitempty"`
	AfmNumFlushThreads           int    `json:"afmNumFlushThreads,omitempty"`
	AfmParallelReadChunkSize     int    `json:"afmParallelReadChunkSize,omitempty"`
	AfmParallelReadThreshold     int    `json:"afmParallelReadThreshold,omitempty"`
	AfmParallelWriteChunkSize    int    `json:"afmParallelWriteChunkSize,omitempty"`
	AfmParallelWriteThreshold    int    `json:"afmParallelWriteThreshold,omitempty"`
	AfmPrefetchThreshold         int    `json:"afmPrefetchThreshold,omitempty"`
	AfmPrimaryID                 string `json:"afmPrimaryID,omitempty"`
	AfmRPO                       int    `json:"afmRPO,omitempty"`
	AfmShowHomeSnapshots         string `json:"afmShowHomeSnapshots,omitempty"`
}

type CreateMakeDirRequest struct {
	UID   string `json:"uid,omitempty"`   //uidnumber
	GID   string `json:"gid,omitempty"`   //gidnumber
	USER  string `json:"user,omitempty"`  //username
	GROUP string `json:"group,omitempty"` //groupname
}

type SymLnkRequest struct {
	FilesystemName string `json:"filesystemName"`
	RelativePath   string `json:"relativePath"`
}

type MountFilesystemRequest struct {
	Nodes        []string `json:"nodes,omitempty"`
	MountOptions string   `json:"mountOptions,omitempty"`
}

type UnmountFilesystemRequest struct {
	Nodes []string `json:"nodes,omitempty"`
	Force bool     `json:"force,omitempty"`
}

const (
	UserSpecifiedUID string = "uid"
	UserSpecifiedGID string = "gid"
)
