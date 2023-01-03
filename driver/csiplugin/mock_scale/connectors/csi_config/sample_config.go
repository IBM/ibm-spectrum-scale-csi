package csi_config

import (
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
)

var ScaleConfig = settings.ScaleSettingsConfigMap{
	Clusters: []settings.Clusters{
		{
			ID: "18359298820404492091",
			Primary: settings.Primary{
				PrimaryFSDep:        "",
				PrimaryFs:           "fs1",
				PrimaryFset:         "",
				PrimaryCid:          "",
				InodeLimitDep:       "",
				InodeLimits:         "",
				RemoteCluster:       "",
				PrimaryFSMount:      "",
				PrimaryFsetLink:     "",
				SymlinkAbsolutePath: "",
				SymlinkRelativePath: "",
			},
			SecureSslMode: false,
			Cacert:        "",
			Secrets:       "guisecret",
			RestAPI: []settings.RestAPI{
				{
					GuiHost: "10.11.105.138",
					GuiPort: 0,
				},
			},
			MgmtUsername: "csiadmin",
			MgmtPassword: "adminuser",
		},
	},
}

var Fileset = connectors.Fileset_v2{
	AFM: connectors.AFM{
		AFMPrimaryID:                 "",
		AFMMode:                      "",
		AFMTarget:                    "",
		AFMAsyncDelay:                0,
		AFMDirLookupRefreshInterval:  0,
		AFMDirOpenRefreshInterval:    0,
		AFMExpirationTimeout:         0,
		AFMFileLookupRefreshInterval: 0,
		AFMNumFlushThreads:           0,
		AFMParallelReadChunkSize:     0,
		AFMParallelReadThreshold:     0,
		AFMParallelWriteChunkSize:    0,
		AFMParallelWriteThreshold:    0,
		AFMPrefetchThreshold:         0,
		AFMRPO:                       0,
		AFMEnableAutoEviction:        false,
		AFMShowHomeSnapshots:         false,
	},
	Config: connectors.FilesetConfig_v2{
		FilesetName:          "",
		FilesystemName:       "",
		Path:                 "/ibm/fs1/pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7",
		InodeSpace:           2,
		MaxNumInodes:         100352,
		PermissionChangeMode: "chmodAndSetacl",
		Comment:              "Fileset created by IBM Container Storage Interface driver",
		IamMode:              "off",
		Oid:                  4,
		Id:                   2,
		Status:               "Linked",
		ParentId:             0,
		Created:              "2022-11-22 11:10:45,000",
		IsInodeSpaceOwner:    true,
		InodeSpaceMask:       1536,
		SnapID:               0,
		RootInode:            1048579,
	},
	FilesetName: "pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7",
}

var Fsmount = connectors.MountInfo{
	MountPoint:             "/ibm/fs1",
	AutomaticMountOption:   "yes",
	AdditionalMountOptions: "none",
	MountPriority:          0,
	RemoteDeviceName:       "mspectrumscale.ibm.com:fs1",
	NodesMounted:           []string{"bnp2-scalegui.fyre.ibm.com", "bnp2-worker-1.fyre.ibm.com", "bnp2-worker-2.fyre.ibm.com"},
	ReadOnly:               false,
	Status:                 "mounted",
}
