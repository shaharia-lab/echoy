// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// MockCommander is an autogenerated mock type for the Commander type
type MockCommander struct {
	mock.Mock
}

type MockCommander_Expecter struct {
	mock *mock.Mock
}

func (_m *MockCommander) EXPECT() *MockCommander_Expecter {
	return &MockCommander_Expecter{mock: &_m.Mock}
}

// Execute provides a mock function with given fields: ctx, cmd, args
func (_m *MockCommander) Execute(ctx context.Context, cmd string, args []string) (string, error) {
	ret := _m.Called(ctx, cmd, args)

	if len(ret) == 0 {
		panic("no return value specified for Execute")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) (string, error)); ok {
		return rf(ctx, cmd, args)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) string); ok {
		r0 = rf(ctx, cmd, args)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []string) error); ok {
		r1 = rf(ctx, cmd, args)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockCommander_Execute_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Execute'
type MockCommander_Execute_Call struct {
	*mock.Call
}

// Execute is a helper method to define mock.On call
//   - ctx context.Context
//   - cmd string
//   - args []string
func (_e *MockCommander_Expecter) Execute(ctx interface{}, cmd interface{}, args interface{}) *MockCommander_Execute_Call {
	return &MockCommander_Execute_Call{Call: _e.mock.On("Execute", ctx, cmd, args)}
}

func (_c *MockCommander_Execute_Call) Run(run func(ctx context.Context, cmd string, args []string)) *MockCommander_Execute_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].([]string))
	})
	return _c
}

func (_c *MockCommander_Execute_Call) Return(_a0 string, _a1 error) *MockCommander_Execute_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCommander_Execute_Call) RunAndReturn(run func(context.Context, string, []string) (string, error)) *MockCommander_Execute_Call {
	_c.Call.Return(run)
	return _c
}

// IsRunning provides a mock function with given fields: ctx
func (_m *MockCommander) IsRunning(ctx context.Context) (bool, string) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for IsRunning")
	}

	var r0 bool
	var r1 string
	if rf, ok := ret.Get(0).(func(context.Context) (bool, string)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) bool); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context) string); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Get(1).(string)
	}

	return r0, r1
}

// MockCommander_IsRunning_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsRunning'
type MockCommander_IsRunning_Call struct {
	*mock.Call
}

// IsRunning is a helper method to define mock.On call
//   - ctx context.Context
func (_e *MockCommander_Expecter) IsRunning(ctx interface{}) *MockCommander_IsRunning_Call {
	return &MockCommander_IsRunning_Call{Call: _e.mock.On("IsRunning", ctx)}
}

func (_c *MockCommander_IsRunning_Call) Run(run func(ctx context.Context)) *MockCommander_IsRunning_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *MockCommander_IsRunning_Call) Return(_a0 bool, _a1 string) *MockCommander_IsRunning_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockCommander_IsRunning_Call) RunAndReturn(run func(context.Context) (bool, string)) *MockCommander_IsRunning_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockCommander creates a new instance of MockCommander. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockCommander(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockCommander {
	mock := &MockCommander{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
