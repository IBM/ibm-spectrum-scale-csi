// Code generated by MockGen. DO NOT EDIT.
// Source: gpfs.go

// Package mock_scale is a generated GoMock package.
package mock_scale

import (
	reflect "reflect"
	sync "sync"

	scale "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin"
	connectors "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/connectors"
	settings "github.com/IBM/ibm-spectrum-scale-csi/driver/csiplugin/settings"
	csi "github.com/container-storage-interface/spec/lib/go/csi"
	gomock "github.com/golang/mock/gomock"
)

// MockScaleDriverInterface is a mock of ScaleDriverInterface interface.
type MockScaleDriverInterface struct {
	ctrl     *gomock.Controller
	recorder *MockScaleDriverInterfaceMockRecorder
}

// MockScaleDriverInterfaceMockRecorder is the mock recorder for MockScaleDriverInterface.
type MockScaleDriverInterfaceMockRecorder struct {
	mock *MockScaleDriverInterface
}

// NewMockScaleDriverInterface creates a new mock instance.
func NewMockScaleDriverInterface(ctrl *gomock.Controller) *MockScaleDriverInterface {
	mock := &MockScaleDriverInterface{ctrl: ctrl}
	mock.recorder = &MockScaleDriverInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockScaleDriverInterface) EXPECT() *MockScaleDriverInterfaceMockRecorder {
	return m.recorder
}

// AddControllerServiceCapabilities mocks base method.
func (m *MockScaleDriverInterface) AddControllerServiceCapabilities(cl []csi.ControllerServiceCapability_RPC_Type) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddControllerServiceCapabilities", cl)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddControllerServiceCapabilities indicates an expected call of AddControllerServiceCapabilities.
func (mr *MockScaleDriverInterfaceMockRecorder) AddControllerServiceCapabilities(cl interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddControllerServiceCapabilities", reflect.TypeOf((*MockScaleDriverInterface)(nil).AddControllerServiceCapabilities), cl)
}

// AddNodeServiceCapabilities mocks base method.
func (m *MockScaleDriverInterface) AddNodeServiceCapabilities(nl []csi.NodeServiceCapability_RPC_Type) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddNodeServiceCapabilities", nl)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddNodeServiceCapabilities indicates an expected call of AddNodeServiceCapabilities.
func (mr *MockScaleDriverInterfaceMockRecorder) AddNodeServiceCapabilities(nl interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddNodeServiceCapabilities", reflect.TypeOf((*MockScaleDriverInterface)(nil).AddNodeServiceCapabilities), nl)
}

// AddVolumeCapabilityAccessModes mocks base method.
func (m *MockScaleDriverInterface) AddVolumeCapabilityAccessModes(vc []csi.VolumeCapability_AccessMode_Mode) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddVolumeCapabilityAccessModes", vc)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddVolumeCapabilityAccessModes indicates an expected call of AddVolumeCapabilityAccessModes.
func (mr *MockScaleDriverInterfaceMockRecorder) AddVolumeCapabilityAccessModes(vc interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddVolumeCapabilityAccessModes", reflect.TypeOf((*MockScaleDriverInterface)(nil).AddVolumeCapabilityAccessModes), vc)
}

// CreatePrimaryFileset mocks base method.
func (m *MockScaleDriverInterface) CreatePrimaryFileset(sc connectors.SpectrumScaleConnector, primaryFS, fsmount, filesetName, inodeLimit string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePrimaryFileset", sc, primaryFS, fsmount, filesetName, inodeLimit)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreatePrimaryFileset indicates an expected call of CreatePrimaryFileset.
func (mr *MockScaleDriverInterfaceMockRecorder) CreatePrimaryFileset(sc, primaryFS, fsmount, filesetName, inodeLimit interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePrimaryFileset", reflect.TypeOf((*MockScaleDriverInterface)(nil).CreatePrimaryFileset), sc, primaryFS, fsmount, filesetName, inodeLimit)
}

// CreateSymlinkPath mocks base method.
func (m *MockScaleDriverInterface) CreateSymlinkPath(sc connectors.SpectrumScaleConnector, fs, fsmount, fsetlinkpath string) (string, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSymlinkPath", sc, fs, fsmount, fsetlinkpath)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// CreateSymlinkPath indicates an expected call of CreateSymlinkPath.
func (mr *MockScaleDriverInterfaceMockRecorder) CreateSymlinkPath(sc, fs, fsmount, fsetlinkpath interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSymlinkPath", reflect.TypeOf((*MockScaleDriverInterface)(nil).CreateSymlinkPath), sc, fs, fsmount, fsetlinkpath)
}

// DeferDeleteReqMap mocks base method.
func (m *MockScaleDriverInterface) DeferDeleteReqMap(name string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeferDeleteReqMap", name)
}

// DeferDeleteReqMap indicates an expected call of DeferDeleteReqMap.
func (mr *MockScaleDriverInterfaceMockRecorder) DeferDeleteReqMap(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeferDeleteReqMap", reflect.TypeOf((*MockScaleDriverInterface)(nil).DeferDeleteReqMap), name)
}

// GetClusterMap mocks base method.
func (m *MockScaleDriverInterface) GetClusterMap() *sync.Map {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClusterMap")
	ret0, _ := ret[0].(*sync.Map)
	return ret0
}

// GetClusterMap indicates an expected call of GetClusterMap.
func (mr *MockScaleDriverInterfaceMockRecorder) GetClusterMap() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClusterMap", reflect.TypeOf((*MockScaleDriverInterface)(nil).GetClusterMap))
}

// GetConnMap mocks base method.
func (m *MockScaleDriverInterface) GetConnMap(conntype string) (connectors.SpectrumScaleConnector, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConnMap", conntype)
	ret0, _ := ret[0].(connectors.SpectrumScaleConnector)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetConnMap indicates an expected call of GetConnMap.
func (mr *MockScaleDriverInterfaceMockRecorder) GetConnMap(conntype interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConnMap", reflect.TypeOf((*MockScaleDriverInterface)(nil).GetConnMap), conntype)
}

// GetPrimary mocks base method.
func (m *MockScaleDriverInterface) GetPrimary() *settings.Primary {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPrimary")
	ret0, _ := ret[0].(*settings.Primary)
	return ret0
}

// GetPrimary indicates an expected call of GetPrimary.
func (mr *MockScaleDriverInterfaceMockRecorder) GetPrimary() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPrimary", reflect.TypeOf((*MockScaleDriverInterface)(nil).GetPrimary))
}

// GetReqMap mocks base method.
func (m *MockScaleDriverInterface) GetReqMap(name string) (int64, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetReqMap", name)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetReqMap indicates an expected call of GetReqMap.
func (mr *MockScaleDriverInterfaceMockRecorder) GetReqMap(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetReqMap", reflect.TypeOf((*MockScaleDriverInterface)(nil).GetReqMap), name)
}

// GetScaleDriver mocks base method.
func (m *MockScaleDriverInterface) GetScaleDriver() *scale.ScaleDriver {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetScaleDriver")
	ret0, _ := ret[0].(*scale.ScaleDriver)
	return ret0
}

// GetScaleDriver indicates an expected call of GetScaleDriver.
func (mr *MockScaleDriverInterfaceMockRecorder) GetScaleDriver() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetScaleDriver", reflect.TypeOf((*MockScaleDriverInterface)(nil).GetScaleDriver))
}

// GetSnapJobStatusMap mocks base method.
func (m *MockScaleDriverInterface) GetSnapJobStatusMap() sync.Map {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSnapJobStatusMap")
	ret0, _ := ret[0].(sync.Map)
	return ret0
}

// GetSnapJobStatusMap indicates an expected call of GetSnapJobStatusMap.
func (mr *MockScaleDriverInterfaceMockRecorder) GetSnapJobStatusMap() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSnapJobStatusMap", reflect.TypeOf((*MockScaleDriverInterface)(nil).GetSnapJobStatusMap))
}

// GetVolCopyJobStatusMap mocks base method.
func (m *MockScaleDriverInterface) GetVolCopyJobStatusMap() sync.Map {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetVolCopyJobStatusMap")
	ret0, _ := ret[0].(sync.Map)
	return ret0
}

// GetVolCopyJobStatusMap indicates an expected call of GetVolCopyJobStatusMap.
func (mr *MockScaleDriverInterfaceMockRecorder) GetVolCopyJobStatusMap() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetVolCopyJobStatusMap", reflect.TypeOf((*MockScaleDriverInterface)(nil).GetVolCopyJobStatusMap))
}

// Getcscap mocks base method.
func (m *MockScaleDriverInterface) Getcscap() []*csi.ControllerServiceCapability {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Getcscap")
	ret0, _ := ret[0].([]*csi.ControllerServiceCapability)
	return ret0
}

// Getcscap indicates an expected call of Getcscap.
func (mr *MockScaleDriverInterfaceMockRecorder) Getcscap() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Getcscap", reflect.TypeOf((*MockScaleDriverInterface)(nil).Getcscap))
}

// NewControllerServer mocks base method.
func (m *MockScaleDriverInterface) NewControllerServer(connMap map[string]connectors.SpectrumScaleConnector, cmap settings.ScaleSettingsConfigMap, primary settings.Primary) *scale.ScaleControllerServer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewControllerServer", connMap, cmap, primary)
	ret0, _ := ret[0].(*scale.ScaleControllerServer)
	return ret0
}

// NewControllerServer indicates an expected call of NewControllerServer.
func (mr *MockScaleDriverInterfaceMockRecorder) NewControllerServer(connMap, cmap, primary interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewControllerServer", reflect.TypeOf((*MockScaleDriverInterface)(nil).NewControllerServer), connMap, cmap, primary)
}

// NewIdentityServer mocks base method.
func (m *MockScaleDriverInterface) NewIdentityServer() *scale.ScaleIdentityServer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewIdentityServer")
	ret0, _ := ret[0].(*scale.ScaleIdentityServer)
	return ret0
}

// NewIdentityServer indicates an expected call of NewIdentityServer.
func (mr *MockScaleDriverInterfaceMockRecorder) NewIdentityServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewIdentityServer", reflect.TypeOf((*MockScaleDriverInterface)(nil).NewIdentityServer))
}

// NewNodeServer mocks base method.
func (m *MockScaleDriverInterface) NewNodeServer() *scale.ScaleNodeServer {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewNodeServer")
	ret0, _ := ret[0].(*scale.ScaleNodeServer)
	return ret0
}

// NewNodeServer indicates an expected call of NewNodeServer.
func (mr *MockScaleDriverInterfaceMockRecorder) NewNodeServer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewNodeServer", reflect.TypeOf((*MockScaleDriverInterface)(nil).NewNodeServer))
}

// PluginInitialize mocks base method.
func (m *MockScaleDriverInterface) PluginInitialize() (map[string]connectors.SpectrumScaleConnector, settings.ScaleSettingsConfigMap, settings.Primary, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PluginInitialize")
	ret0, _ := ret[0].(map[string]connectors.SpectrumScaleConnector)
	ret1, _ := ret[1].(settings.ScaleSettingsConfigMap)
	ret2, _ := ret[2].(settings.Primary)
	ret3, _ := ret[3].(error)
	return ret0, ret1, ret2, ret3
}

// PluginInitialize indicates an expected call of PluginInitialize.
func (mr *MockScaleDriverInterfaceMockRecorder) PluginInitialize() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PluginInitialize", reflect.TypeOf((*MockScaleDriverInterface)(nil).PluginInitialize))
}

// Run mocks base method.
func (m *MockScaleDriverInterface) Run(endpoint string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Run", endpoint)
}

// Run indicates an expected call of Run.
func (mr *MockScaleDriverInterfaceMockRecorder) Run(endpoint interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Run", reflect.TypeOf((*MockScaleDriverInterface)(nil).Run), endpoint)
}

// SetReqMap mocks base method.
func (m *MockScaleDriverInterface) SetReqMap(name string, value int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetReqMap", name, value)
}

// SetReqMap indicates an expected call of SetReqMap.
func (mr *MockScaleDriverInterfaceMockRecorder) SetReqMap(name, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetReqMap", reflect.TypeOf((*MockScaleDriverInterface)(nil).SetReqMap), name, value)
}

// SetupScaleDriver mocks base method.
func (m *MockScaleDriverInterface) SetupScaleDriver(name, vendorVersion, nodeID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetupScaleDriver", name, vendorVersion, nodeID)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetupScaleDriver indicates an expected call of SetupScaleDriver.
func (mr *MockScaleDriverInterfaceMockRecorder) SetupScaleDriver(name, vendorVersion, nodeID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetupScaleDriver", reflect.TypeOf((*MockScaleDriverInterface)(nil).SetupScaleDriver), name, vendorVersion, nodeID)
}

// ValidateControllerServiceRequest mocks base method.
func (m *MockScaleDriverInterface) ValidateControllerServiceRequest(c csi.ControllerServiceCapability_RPC_Type) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateControllerServiceRequest", c)
	ret0, _ := ret[0].(error)
	return ret0
}

// ValidateControllerServiceRequest indicates an expected call of ValidateControllerServiceRequest.
func (mr *MockScaleDriverInterfaceMockRecorder) ValidateControllerServiceRequest(c interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateControllerServiceRequest", reflect.TypeOf((*MockScaleDriverInterface)(nil).ValidateControllerServiceRequest), c)
}

// ValidateScaleConfigParameters mocks base method.
func (m *MockScaleDriverInterface) ValidateScaleConfigParameters(scaleConfig settings.ScaleSettingsConfigMap) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateScaleConfigParameters", scaleConfig)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateScaleConfigParameters indicates an expected call of ValidateScaleConfigParameters.
func (mr *MockScaleDriverInterfaceMockRecorder) ValidateScaleConfigParameters(scaleConfig interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateScaleConfigParameters", reflect.TypeOf((*MockScaleDriverInterface)(nil).ValidateScaleConfigParameters), scaleConfig)
}
