// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/driscollco-cluster/go-service-rest/internal/interfaces (interfaces: Log)
//
// Generated by this command:
//
//	mockgen -destination=../mocks/mock-log.go -package=mocks . Log
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	log "github.com/driscollco-core/log"
	gomock "go.uber.org/mock/gomock"
)

// MockLog is a mock of Log interface.
type MockLog struct {
	ctrl     *gomock.Controller
	recorder *MockLogMockRecorder
	isgomock struct{}
}

// MockLogMockRecorder is the mock recorder for MockLog.
type MockLogMockRecorder struct {
	mock *MockLog
}

// NewMockLog creates a new mock instance.
func NewMockLog(ctrl *gomock.Controller) *MockLog {
	mock := &MockLog{ctrl: ctrl}
	mock.recorder = &MockLogMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLog) EXPECT() *MockLogMockRecorder {
	return m.recorder
}

// Alert mocks base method.
func (m *MockLog) Alert(msg string, keyvals ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range keyvals {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Alert", varargs...)
}

// Alert indicates an expected call of Alert.
func (mr *MockLogMockRecorder) Alert(msg any, keyvals ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, keyvals...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Alert", reflect.TypeOf((*MockLog)(nil).Alert), varargs...)
}

// Child mocks base method.
func (m *MockLog) Child(keyvals ...any) log.Log {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range keyvals {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Child", varargs...)
	ret0, _ := ret[0].(log.Log)
	return ret0
}

// Child indicates an expected call of Child.
func (mr *MockLogMockRecorder) Child(keyvals ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Child", reflect.TypeOf((*MockLog)(nil).Child), keyvals...)
}

// Constants mocks base method.
func (m *MockLog) Constants(keyvals ...any) {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range keyvals {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Constants", varargs...)
}

// Constants indicates an expected call of Constants.
func (mr *MockLogMockRecorder) Constants(keyvals ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Constants", reflect.TypeOf((*MockLog)(nil).Constants), keyvals...)
}

// Critical mocks base method.
func (m *MockLog) Critical(msg string, keyvals ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range keyvals {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Critical", varargs...)
}

// Critical indicates an expected call of Critical.
func (mr *MockLogMockRecorder) Critical(msg any, keyvals ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, keyvals...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Critical", reflect.TypeOf((*MockLog)(nil).Critical), varargs...)
}

// Debug mocks base method.
func (m *MockLog) Debug(msg string, keyvals ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range keyvals {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Debug", varargs...)
}

// Debug indicates an expected call of Debug.
func (mr *MockLogMockRecorder) Debug(msg any, keyvals ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, keyvals...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockLog)(nil).Debug), varargs...)
}

// Emergency mocks base method.
func (m *MockLog) Emergency(msg string, keyvals ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range keyvals {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Emergency", varargs...)
}

// Emergency indicates an expected call of Emergency.
func (mr *MockLogMockRecorder) Emergency(msg any, keyvals ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, keyvals...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Emergency", reflect.TypeOf((*MockLog)(nil).Emergency), varargs...)
}

// Error mocks base method.
func (m *MockLog) Error(msg string, keyvals ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range keyvals {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Error", varargs...)
}

// Error indicates an expected call of Error.
func (mr *MockLogMockRecorder) Error(msg any, keyvals ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, keyvals...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockLog)(nil).Error), varargs...)
}

// Info mocks base method.
func (m *MockLog) Info(msg string, keyvals ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range keyvals {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Info", varargs...)
}

// Info indicates an expected call of Info.
func (mr *MockLogMockRecorder) Info(msg any, keyvals ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, keyvals...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockLog)(nil).Info), varargs...)
}

// Notice mocks base method.
func (m *MockLog) Notice(msg string, keyvals ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range keyvals {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Notice", varargs...)
}

// Notice indicates an expected call of Notice.
func (mr *MockLogMockRecorder) Notice(msg any, keyvals ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, keyvals...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Notice", reflect.TypeOf((*MockLog)(nil).Notice), varargs...)
}

// Warn mocks base method.
func (m *MockLog) Warn(msg string, keyvals ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range keyvals {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Warn", varargs...)
}

// Warn indicates an expected call of Warn.
func (mr *MockLogMockRecorder) Warn(msg any, keyvals ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, keyvals...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockLog)(nil).Warn), varargs...)
}
