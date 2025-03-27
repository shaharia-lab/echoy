// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	theme "github.com/shaharia-lab/echoy/internal/theme"
	mock "github.com/stretchr/testify/mock"
)

// MockTheme is an autogenerated mock type for the Theme type
type MockTheme struct {
	mock.Mock
}

type MockTheme_Expecter struct {
	mock *mock.Mock
}

func (_m *MockTheme) EXPECT() *MockTheme_Expecter {
	return &MockTheme_Expecter{mock: &_m.Mock}
}

// Custom provides a mock function with given fields: name
func (_m *MockTheme) Custom(name string) *theme.Style {
	ret := _m.Called(name)

	if len(ret) == 0 {
		panic("no return value specified for Custom")
	}

	var r0 *theme.Style
	if rf, ok := ret.Get(0).(func(string) *theme.Style); ok {
		r0 = rf(name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*theme.Style)
		}
	}

	return r0
}

// MockTheme_Custom_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Custom'
type MockTheme_Custom_Call struct {
	*mock.Call
}

// Custom is a helper method to define mock.On call
//   - name string
func (_e *MockTheme_Expecter) Custom(name interface{}) *MockTheme_Custom_Call {
	return &MockTheme_Custom_Call{Call: _e.mock.On("Custom", name)}
}

func (_c *MockTheme_Custom_Call) Run(run func(name string)) *MockTheme_Custom_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockTheme_Custom_Call) Return(_a0 *theme.Style) *MockTheme_Custom_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTheme_Custom_Call) RunAndReturn(run func(string) *theme.Style) *MockTheme_Custom_Call {
	_c.Call.Return(run)
	return _c
}

// Disabled provides a mock function with no fields
func (_m *MockTheme) Disabled() *theme.Style {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Disabled")
	}

	var r0 *theme.Style
	if rf, ok := ret.Get(0).(func() *theme.Style); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*theme.Style)
		}
	}

	return r0
}

// MockTheme_Disabled_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Disabled'
type MockTheme_Disabled_Call struct {
	*mock.Call
}

// Disabled is a helper method to define mock.On call
func (_e *MockTheme_Expecter) Disabled() *MockTheme_Disabled_Call {
	return &MockTheme_Disabled_Call{Call: _e.mock.On("Disabled")}
}

func (_c *MockTheme_Disabled_Call) Run(run func()) *MockTheme_Disabled_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTheme_Disabled_Call) Return(_a0 *theme.Style) *MockTheme_Disabled_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTheme_Disabled_Call) RunAndReturn(run func() *theme.Style) *MockTheme_Disabled_Call {
	_c.Call.Return(run)
	return _c
}

// Error provides a mock function with no fields
func (_m *MockTheme) Error() *theme.Style {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Error")
	}

	var r0 *theme.Style
	if rf, ok := ret.Get(0).(func() *theme.Style); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*theme.Style)
		}
	}

	return r0
}

// MockTheme_Error_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Error'
type MockTheme_Error_Call struct {
	*mock.Call
}

// Error is a helper method to define mock.On call
func (_e *MockTheme_Expecter) Error() *MockTheme_Error_Call {
	return &MockTheme_Error_Call{Call: _e.mock.On("Error")}
}

func (_c *MockTheme_Error_Call) Run(run func()) *MockTheme_Error_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTheme_Error_Call) Return(_a0 *theme.Style) *MockTheme_Error_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTheme_Error_Call) RunAndReturn(run func() *theme.Style) *MockTheme_Error_Call {
	_c.Call.Return(run)
	return _c
}

// Info provides a mock function with no fields
func (_m *MockTheme) Info() *theme.Style {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Info")
	}

	var r0 *theme.Style
	if rf, ok := ret.Get(0).(func() *theme.Style); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*theme.Style)
		}
	}

	return r0
}

// MockTheme_Info_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Info'
type MockTheme_Info_Call struct {
	*mock.Call
}

// Info is a helper method to define mock.On call
func (_e *MockTheme_Expecter) Info() *MockTheme_Info_Call {
	return &MockTheme_Info_Call{Call: _e.mock.On("Info")}
}

func (_c *MockTheme_Info_Call) Run(run func()) *MockTheme_Info_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTheme_Info_Call) Return(_a0 *theme.Style) *MockTheme_Info_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTheme_Info_Call) RunAndReturn(run func() *theme.Style) *MockTheme_Info_Call {
	_c.Call.Return(run)
	return _c
}

// IsEnabled provides a mock function with no fields
func (_m *MockTheme) IsEnabled() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for IsEnabled")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockTheme_IsEnabled_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'IsEnabled'
type MockTheme_IsEnabled_Call struct {
	*mock.Call
}

// IsEnabled is a helper method to define mock.On call
func (_e *MockTheme_Expecter) IsEnabled() *MockTheme_IsEnabled_Call {
	return &MockTheme_IsEnabled_Call{Call: _e.mock.On("IsEnabled")}
}

func (_c *MockTheme_IsEnabled_Call) Run(run func()) *MockTheme_IsEnabled_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTheme_IsEnabled_Call) Return(_a0 bool) *MockTheme_IsEnabled_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTheme_IsEnabled_Call) RunAndReturn(run func() bool) *MockTheme_IsEnabled_Call {
	_c.Call.Return(run)
	return _c
}

// Primary provides a mock function with no fields
func (_m *MockTheme) Primary() *theme.Style {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Primary")
	}

	var r0 *theme.Style
	if rf, ok := ret.Get(0).(func() *theme.Style); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*theme.Style)
		}
	}

	return r0
}

// MockTheme_Primary_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Primary'
type MockTheme_Primary_Call struct {
	*mock.Call
}

// Primary is a helper method to define mock.On call
func (_e *MockTheme_Expecter) Primary() *MockTheme_Primary_Call {
	return &MockTheme_Primary_Call{Call: _e.mock.On("Primary")}
}

func (_c *MockTheme_Primary_Call) Run(run func()) *MockTheme_Primary_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTheme_Primary_Call) Return(_a0 *theme.Style) *MockTheme_Primary_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTheme_Primary_Call) RunAndReturn(run func() *theme.Style) *MockTheme_Primary_Call {
	_c.Call.Return(run)
	return _c
}

// Secondary provides a mock function with no fields
func (_m *MockTheme) Secondary() *theme.Style {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Secondary")
	}

	var r0 *theme.Style
	if rf, ok := ret.Get(0).(func() *theme.Style); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*theme.Style)
		}
	}

	return r0
}

// MockTheme_Secondary_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Secondary'
type MockTheme_Secondary_Call struct {
	*mock.Call
}

// Secondary is a helper method to define mock.On call
func (_e *MockTheme_Expecter) Secondary() *MockTheme_Secondary_Call {
	return &MockTheme_Secondary_Call{Call: _e.mock.On("Secondary")}
}

func (_c *MockTheme_Secondary_Call) Run(run func()) *MockTheme_Secondary_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTheme_Secondary_Call) Return(_a0 *theme.Style) *MockTheme_Secondary_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTheme_Secondary_Call) RunAndReturn(run func() *theme.Style) *MockTheme_Secondary_Call {
	_c.Call.Return(run)
	return _c
}

// SetEnabled provides a mock function with given fields: enabled
func (_m *MockTheme) SetEnabled(enabled bool) {
	_m.Called(enabled)
}

// MockTheme_SetEnabled_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SetEnabled'
type MockTheme_SetEnabled_Call struct {
	*mock.Call
}

// SetEnabled is a helper method to define mock.On call
//   - enabled bool
func (_e *MockTheme_Expecter) SetEnabled(enabled interface{}) *MockTheme_SetEnabled_Call {
	return &MockTheme_SetEnabled_Call{Call: _e.mock.On("SetEnabled", enabled)}
}

func (_c *MockTheme_SetEnabled_Call) Run(run func(enabled bool)) *MockTheme_SetEnabled_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(bool))
	})
	return _c
}

func (_c *MockTheme_SetEnabled_Call) Return() *MockTheme_SetEnabled_Call {
	_c.Call.Return()
	return _c
}

func (_c *MockTheme_SetEnabled_Call) RunAndReturn(run func(bool)) *MockTheme_SetEnabled_Call {
	_c.Run(run)
	return _c
}

// Subtle provides a mock function with no fields
func (_m *MockTheme) Subtle() *theme.Style {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Subtle")
	}

	var r0 *theme.Style
	if rf, ok := ret.Get(0).(func() *theme.Style); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*theme.Style)
		}
	}

	return r0
}

// MockTheme_Subtle_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Subtle'
type MockTheme_Subtle_Call struct {
	*mock.Call
}

// Subtle is a helper method to define mock.On call
func (_e *MockTheme_Expecter) Subtle() *MockTheme_Subtle_Call {
	return &MockTheme_Subtle_Call{Call: _e.mock.On("Subtle")}
}

func (_c *MockTheme_Subtle_Call) Run(run func()) *MockTheme_Subtle_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTheme_Subtle_Call) Return(_a0 *theme.Style) *MockTheme_Subtle_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTheme_Subtle_Call) RunAndReturn(run func() *theme.Style) *MockTheme_Subtle_Call {
	_c.Call.Return(run)
	return _c
}

// Success provides a mock function with no fields
func (_m *MockTheme) Success() *theme.Style {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Success")
	}

	var r0 *theme.Style
	if rf, ok := ret.Get(0).(func() *theme.Style); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*theme.Style)
		}
	}

	return r0
}

// MockTheme_Success_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Success'
type MockTheme_Success_Call struct {
	*mock.Call
}

// Success is a helper method to define mock.On call
func (_e *MockTheme_Expecter) Success() *MockTheme_Success_Call {
	return &MockTheme_Success_Call{Call: _e.mock.On("Success")}
}

func (_c *MockTheme_Success_Call) Run(run func()) *MockTheme_Success_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTheme_Success_Call) Return(_a0 *theme.Style) *MockTheme_Success_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTheme_Success_Call) RunAndReturn(run func() *theme.Style) *MockTheme_Success_Call {
	_c.Call.Return(run)
	return _c
}

// Warning provides a mock function with no fields
func (_m *MockTheme) Warning() *theme.Style {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Warning")
	}

	var r0 *theme.Style
	if rf, ok := ret.Get(0).(func() *theme.Style); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*theme.Style)
		}
	}

	return r0
}

// MockTheme_Warning_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Warning'
type MockTheme_Warning_Call struct {
	*mock.Call
}

// Warning is a helper method to define mock.On call
func (_e *MockTheme_Expecter) Warning() *MockTheme_Warning_Call {
	return &MockTheme_Warning_Call{Call: _e.mock.On("Warning")}
}

func (_c *MockTheme_Warning_Call) Run(run func()) *MockTheme_Warning_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockTheme_Warning_Call) Return(_a0 *theme.Style) *MockTheme_Warning_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockTheme_Warning_Call) RunAndReturn(run func() *theme.Style) *MockTheme_Warning_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockTheme creates a new instance of MockTheme. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockTheme(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockTheme {
	mock := &MockTheme{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
