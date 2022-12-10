package scale_test

import (
	"fmt"
	"strconv"

	scale "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin"
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	mock_scale "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/mock_scale"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"golang.org/x/net/context"

	mock_connectors "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/mock_scale/connectors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scale", func() {
	var (
		mockCtrl       *gomock.Controller
		mockConnectors *mock_connectors.MockSpectrumScaleConnector
		mockScale      *mock_scale.MockScaleDriverInterface
		// driver       *scale.ScaleDriver
		fileset *connectors.Fileset_v2
	)

	Context("CTest", func() {

		BeforeEach(func() {
			// driver := scale.GetScaleDriver()
			// conn := make(map[string]settings.ScaleSettingsConfigMap)
			// driver.PluginInitialize() = conn
			mockCtrl = gomock.NewController(GinkgoT())
			// mockObj = mock_main.NewMockStudentInt(mockCtrl)
			mockScale = mock_scale.NewMockScaleDriverInterface(mockCtrl)
			mockConnectors = mock_connectors.NewMockSpectrumScaleConnector(mockCtrl)
			mockScale.EXPECT().GetScaleDriver().Return(&scale.ScaleDriver{}).AnyTimes()
			mockScale.EXPECT().ValidateControllerServiceRequest(gomock.Any()).Return(nil).AnyTimes()
			/* newscaleConfig := settings.ScaleSettingsConfigMap{
				Clusters: []settings.Clusters{
					{
						ID: "18359298820404492091",
						Primary: settings.Primary{
							PrimaryFs: "fs1",
						},
						SecureSslMode: false,
						Cacert:        "",
						Secrets:       "guisecret",
						RestAPI: []settings.RestAPI{
							{GuiHost: "10.11.105.138"},
						},

						MgmtUsername: "csiadmin",
						MgmtPassword: "adminuser",
					}},
			} */
			/* connectionObj, _ := connectors.NewSpectrumRestV2(settings.Clusters{

				ID: "18359298820404492091",
				Primary: settings.Primary{
					PrimaryFs: "fs1",
				},
				SecureSslMode: false,
				Cacert:        "",
				Secrets:       "guisecret",
				RestAPI: []settings.RestAPI{
					{GuiHost: "10.11.105.138"},
				},

				MgmtUsername: "csiadmin",
				MgmtPassword: "adminuser",
			}) */

			mockScale.EXPECT().GetConnMap("18359298820404492091").Return(mockConnectors, true).AnyTimes()
			mockConnectors.EXPECT().GetFilesystemName(gomock.Any()).Return("fs1", nil).AnyTimes()
			mockConnectors.EXPECT().CheckIfFilesetExist("fs1", gomock.Any()).Return(true, nil).AnyTimes()
			mockConnectors.EXPECT().ListFilesetQuota("fs1", gomock.Any()).Return("2097152K", nil).AnyTimes()

			fileset = &connectors.Fileset_v2{
				AFM: connectors.AFM{AFMPrimaryID: "",
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

				Config: connectors.FilesetConfig_v2{FilesetName: "", FilesystemName: "",
					Path: "/ibm/fs1/pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7", InodeSpace: 2, MaxNumInodes: 100352, PermissionChangeMode: "chmodAndSetacl", Comment: "Fileset created by IBM Container Storage Interface driver", IamMode: "off", Oid: 4, Id: 2, Status: "Linked", ParentId: 0, Created: "2022-11-22 11:10:45,000", IsInodeSpaceOwner: true, InodeSpaceMask: 1536, SnapID: 0, RootInode: 1048579},
				FilesetName: "pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7"}

			mockConnectors.EXPECT().ListFileset("fs1", gomock.Any()).Return(*fileset, nil).AnyTimes()

			opt := make(map[string]interface{})
			opt[connectors.UserSpecifiedInodeLimit] = strconv.FormatUint(200000, 10)
			mockConnectors.EXPECT().UpdateFileset("fs1", gomock.Any(), opt).Return(nil).AnyTimes()
			/* err := mockScale.SetupScaleDriver("name", "vendorVersion", "nodeID")
			if err != nil {
				fmt.Errorf("Failed to initialize Scale CSI Driver: %v", err)
			} */

			/* 	newscaleConfig := settings.ScaleSettingsConfigMap{
				Clusters: []settings.Clusters{
					{
						ID: "18359298820404492091",
						Primary: settings.Primary{
							PrimaryFs: "fs1",
						},
						SecureSslMode: false,
						Cacert:        "",
						Secrets:       "guisecret",
						RestAPI: []settings.RestAPI{
							{GuiHost: "10.11.105.138"},
						},

						MgmtUsername: "csiadmin",
						MgmtPassword: "adminuser",
					}},
			} */

			// fmt.Printf("______newscaleConfig_______ %+v \n\n", newscaleConfig)
			/*
			   			______scmap_______ map[18359298820404492091:0xc0000ba640 primary:0xc0000ba640]
			   ______cmap_______ {Clusters:[{ID:18359298820404492091 Primary:{PrimaryFSDep: PrimaryFs:fs1 PrimaryFset:spectrum-scale-csi-volume-store PrimaryCid:18359298820404492091 InodeLimitDep: InodeLimits: RemoteCluster: PrimaryFSMount:/ibm/fs1 PrimaryFsetLink: SymlinkAbsolutePath: SymlinkRelativePath:} SecureSslMode:false Cacert: Secrets:guisecret RestAPI:[{GuiHost:10.11.105.138 GuiPort:0}] MgmtUsername:csiadmin MgmtPassword:adminuser CacertValue:[]}]}
			   ______primary_______ {PrimaryFSDep: PrimaryFs:fs1 PrimaryFset:spectrum-scale-csi-volume-store PrimaryCid:18359298820404492091 InodeLimitDep: InodeLimits: RemoteCluster: PrimaryFSMount:/ibm/fs1 PrimaryFsetLink:/ibm/fs1/spectrum-scale-csi-volume-store SymlinkAbsolutePath:/ibm/fs1/spectrum-scale-csi-volume-store/.volumes SymlinkRelativePath:spectrum-scale-csi-volume-store/.volumes} */
			/* 	mockCtrl = gomock.NewController(GinkgoT())
			mockSettings = mock_connectors.NewMockSpectrumScaleConnector(mockCtrl)
			mockScale = mock_scale.NewMockScaleDriverInterface(mockCtrl)
			// mockScale.EXPECT().GetScaleDriver().AnyTimes()
			scmap := make(map[string]connectors.SpectrumScaleConnector)
			for i := 0; i < len(newscaleConfig.Clusters); i++ {
				cluster := newscaleConfig.Clusters[0]
				sc, err := connectors.NewSpectrumRestV2(cluster)
				fmt.Printf("err : %+v", err)
				fmt.Printf("sc : %+v", sc)
				scmap["18359298820404492091"] = sc
				scmap["primary"] = sc
			}

			cmap := newscaleConfig
			primarySettings := settings.Primary{
				PrimaryFSDep:        "",
				PrimaryFs:           "fs1",
				PrimaryFset:         "spectrum-scale-csi-volume-store",
				PrimaryCid:          "18359298820404492091",
				InodeLimitDep:       "",
				InodeLimits:         "",
				RemoteCluster:       "",
				PrimaryFSMount:      "/ibm/fs1",
				PrimaryFsetLink:     "/ibm/fs1/spectrum-scale-csi-volume-store",
				SymlinkAbsolutePath: "/ibm/fs1/spectrum-scale-csi-volume-store/.volumes",
				SymlinkRelativePath: "spectrum-scale-csi-volume-store/.volumes",
			}
			mockScale.EXPECT().PluginInitialize().Return(scmap, cmap, primarySettings, nil).AnyTimes()
			mockSettings.EXPECT().GetClusterId().Return("18359298820404492091", nil).AnyTimes()
			// mockScale.EXPECT().PluginInitialize().Return().AnyTimes()
			mockScale.EXPECT().ValidateControllerServiceRequest(gomock.Any()).AnyTimes()
			mockScale.EXPECT().GetConnMap(gomock.Any()).Return(cmap, true).AnyTimes()
			*/
		})
		AfterEach(func() {
			defer mockCtrl.Finish()
		})
		It("should give nil", func() {
			out := mockScale.GetScaleDriver()
			fmt.Printf("mockscale %+v", out)
			// n1, n2 := mockSettings.GetClusterId()
			// fmt.Printf("n1111111%+v\n", n1)
			// fmt.Printf("n2222222222222%+v", n2)
			req := &csi.ControllerExpandVolumeRequest{
				VolumeId: "0;2;18359298820404492091;17680B0A:6375380F;;pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7;/ibm/fs1/spectrum-scale-csi-volume-store/.volumes/pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7",
				CapacityRange: &csi.CapacityRange{
					RequiredBytes: 80000,
					LimitBytes:    100000,
				},
			}
			// local_cscap := csi.ControllerServiceCapability{
			// 	Type:
			// }
			// driver := mockScale
			// mockScale.SetupScaleDriver("name", "vendorVersion", "nodeID")
			// mockScale.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{9})
			scaleControllerServer := scale.ScaleControllerServer{
				Driver: mockScale,
			}

			resp, erro := scaleControllerServer.ControllerExpandVolume(context.Background(), req)
			fmt.Println("erroroooooo :====", erro)
			fmt.Println("resp :====", resp)
			Expect(erro).NotTo(HaveOccurred())
		})
	})
})
