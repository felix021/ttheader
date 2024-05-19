package ttheader

import (
	"testing"
)

// TODO: complete tests for Exception

func Test_exception_BytesLength(t *testing.T) {
	exc := NewException("method", 0, "message", 0)
	length, err := exc.Bytes()
	assert(t, err == nil)
	assert(t, len(length) == sizeFixed+len(exc.Message)+len(exc.MethodName))
}
