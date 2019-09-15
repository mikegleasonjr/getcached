package mocks

import (
	"github.com/stretchr/testify/mock"
)

// HashFn mocks a hash function.
type HashFn struct {
	mock.Mock
}

// Fn is the actual mocked method.
func (m *HashFn) Fn(data []byte) uint32 {
	ret := m.Called(data)

	var r0 uint32
	if rf, ok := ret.Get(0).(func([]byte) uint32); ok {
		r0 = rf(data)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(uint32)
		}
	}

	return r0
}
