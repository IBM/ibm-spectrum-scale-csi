package scale_test

import (
	"fmt"

	scale "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin"
	mock_scale "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/mock_scale"

	// mock_connectors "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/mock_scale/connectors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	// . "github.com/onsi/gomega"
)

var _ = Describe("Scale", func() {
	var (
		mockCtrl *gomock.Controller
		// mockSettings *mock_connectors.MockSpectrumScaleConnector
		mockScale *mock_scale.MockScaleDriverInterface
		// driver       *scale.ScaleDriver
	)

	Context("CTest", func() {

		BeforeEach(func() {
			// driver := scale.GetScaleDriver()
			// conn := make(map[string]settings.ScaleSettingsConfigMap)
			// driver.PluginInitialize() = conn
			mockCtrl = gomock.NewController(GinkgoT())
			// mockObj = mock_main.NewMockStudentInt(mockCtrl)
			mockScale = mock_scale.NewMockScaleDriverInterface(mockCtrl)
			mockScale.EXPECT().GetScaleDriver().Return(&scale.ScaleDriver{}).AnyTimes()
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
			/* 			req := &csi.ControllerExpandVolumeRequest{
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
			   			Expect(erro).NotTo(HaveOccurred()) */
		})
	})
})
