// Code generated by mockery v2.42.0. DO NOT EDIT.

package mocks

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"

	types "github.com/yuyang0/vmimage/types"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// ListLocalImages provides a mock function with given fields: ctx, user
func (_m *Manager) ListLocalImages(ctx context.Context, user string) ([]*types.Image, error) {
	ret := _m.Called(ctx, user)

	if len(ret) == 0 {
		panic("no return value specified for ListLocalImages")
	}

	var r0 []*types.Image
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*types.Image, error)); ok {
		return rf(ctx, user)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []*types.Image); ok {
		r0 = rf(ctx, user)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*types.Image)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, user)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LoadImage provides a mock function with given fields: ctx, imgName
func (_m *Manager) LoadImage(ctx context.Context, imgName string) (*types.Image, error) {
	ret := _m.Called(ctx, imgName)

	if len(ret) == 0 {
		panic("no return value specified for LoadImage")
	}

	var r0 *types.Image
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*types.Image, error)); ok {
		return rf(ctx, imgName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *types.Image); ok {
		r0 = rf(ctx, imgName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Image)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, imgName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Prepare provides a mock function with given fields: fname, img
func (_m *Manager) Prepare(fname string, img *types.Image) (io.ReadCloser, error) {
	ret := _m.Called(fname, img)

	if len(ret) == 0 {
		panic("no return value specified for Prepare")
	}

	var r0 io.ReadCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(string, *types.Image) (io.ReadCloser, error)); ok {
		return rf(fname, img)
	}
	if rf, ok := ret.Get(0).(func(string, *types.Image) io.ReadCloser); ok {
		r0 = rf(fname, img)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(string, *types.Image) error); ok {
		r1 = rf(fname, img)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Pull provides a mock function with given fields: ctx, img, pullPolicy
func (_m *Manager) Pull(ctx context.Context, img *types.Image, pullPolicy types.PullPolicy) (io.ReadCloser, error) {
	ret := _m.Called(ctx, img, pullPolicy)

	if len(ret) == 0 {
		panic("no return value specified for Pull")
	}

	var r0 io.ReadCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.Image, types.PullPolicy) (io.ReadCloser, error)); ok {
		return rf(ctx, img, pullPolicy)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *types.Image, types.PullPolicy) io.ReadCloser); ok {
		r0 = rf(ctx, img, pullPolicy)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *types.Image, types.PullPolicy) error); ok {
		r1 = rf(ctx, img, pullPolicy)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Push provides a mock function with given fields: ctx, img, force
func (_m *Manager) Push(ctx context.Context, img *types.Image, force bool) (io.ReadCloser, error) {
	ret := _m.Called(ctx, img, force)

	if len(ret) == 0 {
		panic("no return value specified for Push")
	}

	var r0 io.ReadCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.Image, bool) (io.ReadCloser, error)); ok {
		return rf(ctx, img, force)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *types.Image, bool) io.ReadCloser); ok {
		r0 = rf(ctx, img, force)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.ReadCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *types.Image, bool) error); ok {
		r1 = rf(ctx, img, force)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RemoveLocal provides a mock function with given fields: ctx, img
func (_m *Manager) RemoveLocal(ctx context.Context, img *types.Image) error {
	ret := _m.Called(ctx, img)

	if len(ret) == 0 {
		panic("no return value specified for RemoveLocal")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *types.Image) error); ok {
		r0 = rf(ctx, img)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewManager creates a new instance of Manager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewManager(t interface {
	mock.TestingT
	Cleanup(func())
}) *Manager {
	mock := &Manager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
