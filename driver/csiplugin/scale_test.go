package scale_test

import (
	"flag"
	"fmt"
	"strconv"

	scale "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"

	mock_connectors "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/mock_scale/connectors"

	"github.com/golang/glog"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CSI Scale Unit Testing", func() {
	var (
		mockCtrl          *gomock.Controller
		mockConnectors    *mock_connectors.MockSpectrumScaleConnector
		mockGetConnectors *mock_connectors.MockGetSpectrumScaleConnectorInt
		fileset           *connectors.Fileset_v2
		driver            *scale.ScaleDriver
		driverName        = flag.String("drivername", "spectrumscale.csi.ibm.com", "name of the driver")
		nodeID            = flag.String("nodeid", "", "node id")
		vendorVersion     = "2.8.0"
		scaleConfig       settings.ScaleSettingsConfigMap
	)
	scaleConfig = settings.ScaleSettingsConfigMap{
		Clusters: []settings.Clusters{
			{ID: "18359298820404492091",
				Primary: settings.Primary{
					PrimaryFSDep: "",
					PrimaryFs:    "fs1", PrimaryFset: "", PrimaryCid: "", InodeLimitDep: "", InodeLimits: "", RemoteCluster: "", PrimaryFSMount: "", PrimaryFsetLink: "", SymlinkAbsolutePath: "", SymlinkRelativePath: "",
				},
				SecureSslMode: false, Cacert: "", Secrets: "guisecret",
				RestAPI:      []settings.RestAPI{{GuiHost: "10.11.105.138", GuiPort: 0}},
				MgmtUsername: "csiadmin", MgmtPassword: "adminuser"},
		}}
	Context("For Controller Expand Volume", func() {

		BeforeEach(func() {

			mockCtrl = gomock.NewController(GinkgoT())
			mockConnectors = mock_connectors.NewMockSpectrumScaleConnector(mockCtrl)
			mockGetConnectors = mock_connectors.NewMockGetSpectrumScaleConnectorInt(mockCtrl)
			// Socket Connection functionality mock
			fileset = &connectors.Fileset_v2{
				AFM: connectors.AFM{AFMPrimaryID: "", AFMMode: "", AFMTarget: "", AFMAsyncDelay: 0, AFMDirLookupRefreshInterval: 0, AFMDirOpenRefreshInterval: 0, AFMExpirationTimeout: 0, AFMFileLookupRefreshInterval: 0, AFMNumFlushThreads: 0, AFMParallelReadChunkSize: 0, AFMParallelReadThreshold: 0, AFMParallelWriteChunkSize: 0, AFMParallelWriteThreshold: 0, AFMPrefetchThreshold: 0, AFMRPO: 0, AFMEnableAutoEviction: false, AFMShowHomeSnapshots: false}, Config: connectors.FilesetConfig_v2{FilesetName: "", FilesystemName: "",
					Path: "/ibm/fs1/pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7", InodeSpace: 2, MaxNumInodes: 100352, PermissionChangeMode: "chmodAndSetacl", Comment: "Fileset created by IBM Container Storage Interface driver", IamMode: "off", Oid: 4, Id: 2, Status: "Linked", ParentId: 0, Created: "2022-11-22 11:10:45,000", IsInodeSpaceOwner: true, InodeSpaceMask: 1536, SnapID: 0, RootInode: 1048579},
				FilesetName: "pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7"}

			opt := make(map[string]interface{})
			opt[connectors.UserSpecifiedInodeLimit] = strconv.FormatUint(200000, 10)

			mockConnectors.EXPECT().GetFilesystemName(gomock.Any()).Return("fs1", nil).AnyTimes()
			mockConnectors.EXPECT().CheckIfFilesetExist("fs1", gomock.Any()).Return(true, nil).AnyTimes()
			mockConnectors.EXPECT().ListFilesetQuota("fs1", gomock.Any()).Return("4718592K", nil)
			mockConnectors.EXPECT().ListFileset("fs1", gomock.Any()).Return(*fileset, nil).AnyTimes()
			mockConnectors.EXPECT().UpdateFileset("fs1", gomock.Any(), opt).Return(nil).AnyTimes()
			mockConnectors.EXPECT().GetClusterId().Return("18359298820404492091", nil).AnyTimes()
			fsmount := connectors.MountInfo{
				MountPoint:             "/ibm/fs1",
				AutomaticMountOption:   "yes",
				AdditionalMountOptions: "none",
				MountPriority:          0,
				RemoteDeviceName:       "mspectrumscale.ibm.com:fs1",
				NodesMounted:           []string{"bnp2-scalegui.fyre.ibm.com", "bnp2-worker-1.fyre.ibm.com", "bnp2-worker-2.fyre.ibm.com"},
				ReadOnly:               false,
				Status:                 "mounted"}
			mockConnectors.EXPECT().GetFilesystemMountDetails("fs1").Return(fsmount, nil).AnyTimes()
			mockConnectors.EXPECT().IsFilesystemMountedOnGUINode("fs1").Return(true, nil).AnyTimes()
			mockConnectors.EXPECT().MakeDirectory(gomock.Any(), gomock.Any(), "0", "0").Return(nil).AnyTimes()
			mockGetConnectors.EXPECT().GetSpectrumScaleConnector(gomock.Any()).Return(mockConnectors, nil).AnyTimes()

		})
		AfterEach(func() {
			defer mockCtrl.Finish()
		})
		It("should successfully expand volume", func() {
			driver = scale.GetScaleDriver()

			scaleConnMap := make(map[string]connectors.SpectrumScaleConnector)
			// fmt.Printf("driver %+v\n", driver)
			// *driverName, vendorVersion, *nodeID, scaleConfig, scaleConnMap, &connectors.GetSpec{}
			err := driver.SetupScaleDriver(*driverName, vendorVersion, *nodeID, scaleConfig, scaleConnMap, mockGetConnectors)
			if err != nil {
				glog.Fatalf("Failed to initialize Scale CSI Driver: %v", err)
			}
			req := &csi.ControllerExpandVolumeRequest{
				VolumeId: "0;2;18359298820404492091;17680B0A:6375380F;;pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7;/ibm/fs1/spectrum-scale-csi-volume-store/.volumes/pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 80000,
					LimitBytes:    100000,
				},
			}
			scaleControllerServer := scale.ScaleControllerServer{
				Driver: driver,
			}
			resp, erro := scaleControllerServer.ControllerExpandVolume(context.Background(), req)
			fmt.Printf("erroroooooo :====%+v\n", erro)
			fmt.Printf("resp :====%+v\n", resp)
			Expect(erro).NotTo(HaveOccurred())
			Expect(resp.CapacityBytes).Should(Equal(int64(80000)))
		})
	})
	Context("For Controller Expand Volume", func() {

		BeforeEach(func() {

			mockCtrl = gomock.NewController(GinkgoT())
			mockConnectors = mock_connectors.NewMockSpectrumScaleConnector(mockCtrl)
			mockGetConnectors = mock_connectors.NewMockGetSpectrumScaleConnectorInt(mockCtrl)
			// Socket Connection functionality mock
			fileset = &connectors.Fileset_v2{
				AFM: connectors.AFM{AFMPrimaryID: "", AFMMode: "", AFMTarget: "", AFMAsyncDelay: 0, AFMDirLookupRefreshInterval: 0, AFMDirOpenRefreshInterval: 0, AFMExpirationTimeout: 0, AFMFileLookupRefreshInterval: 0, AFMNumFlushThreads: 0, AFMParallelReadChunkSize: 0, AFMParallelReadThreshold: 0, AFMParallelWriteChunkSize: 0, AFMParallelWriteThreshold: 0, AFMPrefetchThreshold: 0, AFMRPO: 0, AFMEnableAutoEviction: false, AFMShowHomeSnapshots: false}, Config: connectors.FilesetConfig_v2{FilesetName: "", FilesystemName: "",
					Path: "/ibm/fs1/pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7", InodeSpace: 2, MaxNumInodes: 100352, PermissionChangeMode: "chmodAndSetacl", Comment: "Fileset created by IBM Container Storage Interface driver", IamMode: "off", Oid: 4, Id: 2, Status: "Linked", ParentId: 0, Created: "2022-11-22 11:10:45,000", IsInodeSpaceOwner: true, InodeSpaceMask: 1536, SnapID: 0, RootInode: 1048579},
				FilesetName: "pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7"}

			opt := make(map[string]interface{})
			opt[connectors.UserSpecifiedInodeLimit] = strconv.FormatUint(200000, 10)

			mockConnectors.EXPECT().GetFilesystemName(gomock.Any()).Return("fs1", nil).AnyTimes()
			mockConnectors.EXPECT().CheckIfFilesetExist("fs1", gomock.Any()).Return(true, nil).AnyTimes()
			mockConnectors.EXPECT().ListFilesetQuota("fs1", gomock.Any()).Return("10K", fmt.Errorf("error while list file set quota"))
			mockConnectors.EXPECT().ListFileset("fs1", gomock.Any()).Return(*fileset, nil).AnyTimes()
			mockConnectors.EXPECT().UpdateFileset("fs1", gomock.Any(), opt).Return(nil).AnyTimes()
			mockConnectors.EXPECT().GetClusterId().Return("18359298820404492091", nil).AnyTimes()
			fsmount := connectors.MountInfo{
				MountPoint:             "/ibm/fs1",
				AutomaticMountOption:   "yes",
				AdditionalMountOptions: "none",
				MountPriority:          0,
				RemoteDeviceName:       "mspectrumscale.ibm.com:fs1",
				NodesMounted:           []string{"bnp2-scalegui.fyre.ibm.com", "bnp2-worker-1.fyre.ibm.com", "bnp2-worker-2.fyre.ibm.com"},
				ReadOnly:               false,
				Status:                 "mounted"}
			mockConnectors.EXPECT().GetFilesystemMountDetails("fs1").Return(fsmount, nil).AnyTimes()
			mockConnectors.EXPECT().IsFilesystemMountedOnGUINode("fs1").Return(true, nil).AnyTimes()
			mockConnectors.EXPECT().MakeDirectory(gomock.Any(), gomock.Any(), "0", "0").Return(nil).AnyTimes()
			mockGetConnectors.EXPECT().GetSpectrumScaleConnector(gomock.Any()).Return(mockConnectors, nil).AnyTimes()

		})
		AfterEach(func() {
			defer mockCtrl.Finish()
		})
		It("should fail in the expand volume request", func() {
			driver = scale.GetScaleDriver()
			// scaleConfig := settings.LoadScaleConfigSettings()
			scaleConnMap := make(map[string]connectors.SpectrumScaleConnector)
			// fmt.Printf("driver %+v\n", driver)
			// *driverName, vendorVersion, *nodeID, scaleConfig, scaleConnMap, &connectors.GetSpec{}
			err := driver.SetupScaleDriver(*driverName, vendorVersion, *nodeID, scaleConfig, scaleConnMap, mockGetConnectors)
			if err != nil {
				glog.Fatalf("Failed to initialize Scale CSI Driver: %v", err)
			}
			req := &csi.ControllerExpandVolumeRequest{
				VolumeId: "0;2;18359298820404492091;17680B0A:6375380F;;pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7;/ibm/fs1/spectrum-scale-csi-volume-store/.volumes/pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 80000,
					LimitBytes:    100000,
				},
			}
			scaleControllerServer := scale.ScaleControllerServer{
				Driver: driver,
			}
			resp, erro := scaleControllerServer.ControllerExpandVolume(context.Background(), req)
			fmt.Printf("erroroooooo :====%+v\n", erro)
			fmt.Printf("resp :====%+v\n", resp)
			Expect(erro).To(HaveOccurred())
			// Expect(resp.CapacityBytes).Should(Equal(int64(80000)))
		})
	})
})
