package mocks

import (
	context "context"
	reflect "reflect"

	uuid "github.com/google/uuid"
	rideDomain "github.com/hextec/yolyolakay-backend/internal/ride/domain"
	userGrpc "github.com/hextec/yolyolakay-backend/pkg/api/grpc/user"
	gomock "go.uber.org/mock/gomock"
)

type MockGRPCClient struct {
	ctrl     *gomock.Controller
	recorder *MockGRPCClientMockRecorder
}

type MockGRPCClientMockRecorder struct {
	mock *MockGRPCClient
}

func NewMockGRPCClient(ctrl *gomock.Controller) *MockGRPCClient {
	mock := &MockGRPCClient{ctrl: ctrl}
	mock.recorder = &MockGRPCClientMockRecorder{mock}
	return mock
}

func (m *MockGRPCClient) EXPECT() *MockGRPCClientMockRecorder {
	return m.recorder
}

func (m *MockGRPCClient) CheckCarStatus(ctx context.Context, driverID uuid.UUID) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckCarStatus", ctx, driverID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockGRPCClientMockRecorder) CheckCarStatus(ctx, driverID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckCarStatus", reflect.TypeOf((*MockGRPCClient)(nil).CheckCarStatus), ctx, driverID)
}

func (m *MockGRPCClient) CheckRideLimit(ctx context.Context, driverID uuid.UUID) (bool, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckRideLimit", ctx, driverID)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

func (mr *MockGRPCClientMockRecorder) CheckRideLimit(ctx, driverID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckRideLimit", reflect.TypeOf((*MockGRPCClient)(nil).CheckRideLimit), ctx, driverID)
}

func (m *MockGRPCClient) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

func (mr *MockGRPCClientMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockGRPCClient)(nil).Close))
}

func (m *MockGRPCClient) GetDriverProfile(ctx context.Context, driverID uuid.UUID) (*rideDomain.DriverProfile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDriverProfile", ctx, driverID)
	ret0, _ := ret[0].(*rideDomain.DriverProfile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockGRPCClientMockRecorder) GetDriverProfile(ctx, driverID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDriverProfile", reflect.TypeOf((*MockGRPCClient)(nil).GetDriverProfile), ctx, driverID)
}

func (m *MockGRPCClient) GetDriverVehicle(ctx context.Context, driverID uuid.UUID) (*userGrpc.GetDriverVehicleResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDriverVehicle", ctx, driverID)
	ret0, _ := ret[0].(*userGrpc.GetDriverVehicleResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockGRPCClientMockRecorder) GetDriverVehicle(ctx, driverID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDriverVehicle", reflect.TypeOf((*MockGRPCClient)(nil).GetDriverVehicle), ctx, driverID)
}

func (m *MockGRPCClient) GetPassengerProfile(ctx context.Context, passengerID uuid.UUID) (*rideDomain.PassengerProfile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPassengerProfile", ctx, passengerID)
	ret0, _ := ret[0].(*rideDomain.PassengerProfile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockGRPCClientMockRecorder) GetPassengerProfile(ctx, passengerID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPassengerProfile", reflect.TypeOf((*MockGRPCClient)(nil).GetPassengerProfile), ctx, passengerID)
}
