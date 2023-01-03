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
	"github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/mock_scale/connectors/csi_config"

	"github.com/golang/glog"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CSI Scale Unit Testing", func() {
	Context("For Controller Expand Volume", func() {
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
			fsmount           connectors.MountInfo
		)
		scaleConfig = csi_config.ScaleConfig
		fsmount = csi_config.Fsmount
		Context("When volume request is passed correctly", func() {

			BeforeEach(func() {

				mockCtrl = gomock.NewController(GinkgoT())
				mockConnectors = mock_connectors.NewMockSpectrumScaleConnector(mockCtrl)
				mockGetConnectors = mock_connectors.NewMockGetSpectrumScaleConnectorInt(mockCtrl)
				fileset = &csi_config.Fileset

				opt := make(map[string]interface{})
				opt[connectors.UserSpecifiedInodeLimit] = strconv.FormatUint(200000, 10)

				mockConnectors.EXPECT().GetFilesystemName(gomock.Any()).Return("fs1", nil).AnyTimes()
				mockConnectors.EXPECT().CheckIfFilesetExist("fs1", gomock.Any()).Return(true, nil).AnyTimes()
				mockConnectors.EXPECT().ListFilesetQuota("fs1", gomock.Any()).Return("4718592K", nil)
				mockConnectors.EXPECT().ListFileset("fs1", gomock.Any()).Return(*fileset, nil).AnyTimes()
				mockConnectors.EXPECT().UpdateFileset("fs1", gomock.Any(), opt).Return(nil).AnyTimes()
				mockConnectors.EXPECT().GetClusterId().Return("18359298820404492091", nil).AnyTimes()

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

				Expect(erro).NotTo(HaveOccurred())
				Expect(resp.CapacityBytes).Should(Equal(int64(80000)))
			})
		})
		Context("when got error at ListFilesetQuota", func() {

			BeforeEach(func() {

				mockCtrl = gomock.NewController(GinkgoT())
				mockConnectors = mock_connectors.NewMockSpectrumScaleConnector(mockCtrl)
				mockGetConnectors = mock_connectors.NewMockGetSpectrumScaleConnectorInt(mockCtrl)
				fileset = &csi_config.Fileset

				opt := make(map[string]interface{})
				opt[connectors.UserSpecifiedInodeLimit] = strconv.FormatUint(200000, 10)

				mockConnectors.EXPECT().GetFilesystemName(gomock.Any()).Return("fs1", nil).AnyTimes()
				mockConnectors.EXPECT().CheckIfFilesetExist("fs1", gomock.Any()).Return(true, nil).AnyTimes()
				mockConnectors.EXPECT().ListFilesetQuota("fs1", gomock.Any()).Return("7K", fmt.Errorf("error while list file set quota"))
				mockConnectors.EXPECT().ListFileset("fs1", gomock.Any()).Return(*fileset, nil).AnyTimes()
				mockConnectors.EXPECT().GetClusterId().Return("18359298820404492091", nil).AnyTimes()
				// fsmount := csi_config.Fsmount
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
				scaleConnMap := make(map[string]connectors.SpectrumScaleConnector)

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
				Expect(erro).To(HaveOccurred())
				Expect(resp).Should(BeNil())
			})
		})
		Context("when quota is less than required size", func() {

			BeforeEach(func() {

				mockCtrl = gomock.NewController(GinkgoT())
				mockConnectors = mock_connectors.NewMockSpectrumScaleConnector(mockCtrl)
				mockGetConnectors = mock_connectors.NewMockGetSpectrumScaleConnectorInt(mockCtrl)
				fileset = &csi_config.Fileset

				opt := make(map[string]interface{})
				opt[connectors.UserSpecifiedInodeLimit] = strconv.FormatUint(200000, 10)

				mockConnectors.EXPECT().GetFilesystemName(gomock.Any()).Return("fs1", nil).AnyTimes()
				mockConnectors.EXPECT().CheckIfFilesetExist("fs1", gomock.Any()).Return(true, nil).AnyTimes()
				mockConnectors.EXPECT().ListFilesetQuota("fs1", gomock.Any()).Return("7K", nil)
				mockConnectors.EXPECT().ListFileset("fs1", gomock.Any()).Return(*fileset, nil).AnyTimes()
				mockConnectors.EXPECT().UpdateFileset("fs1", gomock.Any(), opt).Return(nil).AnyTimes()
				mockConnectors.EXPECT().GetClusterId().Return("18359298820404492091", nil).AnyTimes()
				mockConnectors.EXPECT().SetFilesetQuota(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

				mockConnectors.EXPECT().GetFilesystemMountDetails("fs1").Return(fsmount, nil).AnyTimes()
				mockConnectors.EXPECT().IsFilesystemMountedOnGUINode("fs1").Return(true, nil).AnyTimes()
				mockConnectors.EXPECT().MakeDirectory(gomock.Any(), gomock.Any(), "0", "0").Return(nil).AnyTimes()
				mockGetConnectors.EXPECT().GetSpectrumScaleConnector(gomock.Any()).Return(mockConnectors, nil).AnyTimes()

			})
			AfterEach(func() {
				defer mockCtrl.Finish()
			})
			It("should pass when quota is less than required size in the expand volume request", func() {
				driver = scale.GetScaleDriver()
				scaleConnMap := make(map[string]connectors.SpectrumScaleConnector)
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
				Expect(erro).NotTo(HaveOccurred())
				Expect(resp.CapacityBytes).Should(Equal(int64(80000)))
			})
		})

		Context("when volumeID is not provided in req", func() {

			BeforeEach(func() {

				mockCtrl = gomock.NewController(GinkgoT())
				mockConnectors = mock_connectors.NewMockSpectrumScaleConnector(mockCtrl)
				mockGetConnectors = mock_connectors.NewMockGetSpectrumScaleConnectorInt(mockCtrl)
				fileset = &csi_config.Fileset

				opt := make(map[string]interface{})
				opt[connectors.UserSpecifiedInodeLimit] = strconv.FormatUint(200000, 10)

				mockConnectors.EXPECT().GetFilesystemName(gomock.Any()).Return("fs1", nil).AnyTimes()
				mockConnectors.EXPECT().CheckIfFilesetExist("fs1", gomock.Any()).Return(true, nil).AnyTimes()
				mockConnectors.EXPECT().ListFilesetQuota("fs1", gomock.Any()).Return("7K", nil).AnyTimes()
				mockConnectors.EXPECT().ListFileset("fs1", gomock.Any()).Return(*fileset, nil).AnyTimes()
				mockConnectors.EXPECT().UpdateFileset("fs1", gomock.Any(), opt).Return(nil).AnyTimes()
				mockConnectors.EXPECT().GetClusterId().Return("18359298820404492091", nil).AnyTimes()
				mockConnectors.EXPECT().SetFilesetQuota(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
				mockConnectors.EXPECT().GetFilesystemMountDetails("fs1").Return(fsmount, nil).AnyTimes()
				mockConnectors.EXPECT().IsFilesystemMountedOnGUINode("fs1").Return(true, nil).AnyTimes()
				mockConnectors.EXPECT().MakeDirectory(gomock.Any(), gomock.Any(), "0", "0").Return(nil).AnyTimes()
				mockGetConnectors.EXPECT().GetSpectrumScaleConnector(gomock.Any()).Return(mockConnectors, nil).AnyTimes()

			})
			AfterEach(func() {
				defer mockCtrl.Finish()
			})
			It("should pass when volumeID is not provided in the expand volume request", func() {
				driver = scale.GetScaleDriver()
				scaleConnMap := make(map[string]connectors.SpectrumScaleConnector)
				err := driver.SetupScaleDriver(*driverName, vendorVersion, *nodeID, scaleConfig, scaleConnMap, mockGetConnectors)
				if err != nil {
					glog.Fatalf("Failed to initialize Scale CSI Driver: %v", err)
				}
				req := &csi.ControllerExpandVolumeRequest{
					CapacityRange: &csi.CapacityRange{
						RequiredBytes: 80000,
						LimitBytes:    100000,
					},
				}
				scaleControllerServer := scale.ScaleControllerServer{
					Driver: driver,
				}
				resp, erro := scaleControllerServer.ControllerExpandVolume(context.Background(), req)
				Expect(erro).To(HaveOccurred())
				Expect(resp).Should(BeNil())
			})

			It("should pass when capacityRange is not provided in the expand volume request", func() {
				driver = scale.GetScaleDriver()
				scaleConnMap := make(map[string]connectors.SpectrumScaleConnector)
				err := driver.SetupScaleDriver(*driverName, vendorVersion, *nodeID, scaleConfig, scaleConnMap, mockGetConnectors)
				if err != nil {
					glog.Fatalf("Failed to initialize Scale CSI Driver: %v", err)
				}
				req := &csi.ControllerExpandVolumeRequest{
					VolumeId: "0;2;18359298820404492091;17680B0A:6375380F;;pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7;/ibm/fs1/spectrum-scale-csi-volume-store/.volumes/pvc-80a0976b-e5a8-4a10-9f27-81aaec7436b7",
				}
				scaleControllerServer := scale.ScaleControllerServer{
					Driver: driver,
				}
				resp, erro := scaleControllerServer.ControllerExpandVolume(context.Background(), req)
				Expect(erro).To(HaveOccurred())
				Expect(resp).Should(BeNil())
			})
		})
	})
})
