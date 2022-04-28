// Code generated by mockery v2.12.0. DO NOT EDIT.

package classservice

import (
	context "context"

	primitive "github.com/angusgmorrison/hexagonal/internal/primitive"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// MockAtomicClassScribe is an autogenerated mock type for the AtomicClassScribe type
type MockAtomicClassScribe struct {
	mock.Mock
}

// Begin provides a mock function with given fields: ctx
func (_m *MockAtomicClassScribe) Begin(ctx context.Context) error {
	ret := _m.Called(ctx)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Commit provides a mock function with given fields:
func (_m *MockAtomicClassScribe) Commit() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EnrollStudents provides a mock function with given fields: ctx, c, s
func (_m *MockAtomicClassScribe) EnrollStudents(ctx context.Context, c Course, s Students) (Class, error) {
	ret := _m.Called(ctx, c, s)

	var r0 Class
	if rf, ok := ret.Get(0).(func(context.Context, Course, Students) Class); ok {
		r0 = rf(ctx, c, s)
	} else {
		r0 = ret.Get(0).(Class)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, Course, Students) error); ok {
		r1 = rf(ctx, c, s)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetClassByCourseCode provides a mock function with given fields: ctx, courseCode
func (_m *MockAtomicClassScribe) GetClassByCourseCode(ctx context.Context, courseCode string) (Class, error) {
	ret := _m.Called(ctx, courseCode)

	var r0 Class
	if rf, ok := ret.Get(0).(func(context.Context, string) Class); ok {
		r0 = rf(ctx, courseCode)
	} else {
		r0 = ret.Get(0).(Class)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, courseCode)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetStudentsByEmail provides a mock function with given fields: ctx, emails
func (_m *MockAtomicClassScribe) GetStudentsByEmail(ctx context.Context, emails []primitive.EmailAddress) (Students, error) {
	ret := _m.Called(ctx, emails)

	var r0 Students
	if rf, ok := ret.Get(0).(func(context.Context, []primitive.EmailAddress) Students); ok {
		r0 = rf(ctx, emails)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Students)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []primitive.EmailAddress) error); ok {
		r1 = rf(ctx, emails)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Rollback provides a mock function with given fields:
func (_m *MockAtomicClassScribe) Rollback() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewMockAtomicClassScribe creates a new instance of MockAtomicClassScribe. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockAtomicClassScribe(t testing.TB) *MockAtomicClassScribe {
	mock := &MockAtomicClassScribe{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
