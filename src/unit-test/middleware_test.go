package tests

import (
	"testing"

	. "github.com/pulsar-beam/src/middleware"
)

func TestSemaphore(t *testing.T) {
	var sema = NewSema(2)
	err := sema.Release()
	equals(t, "all semaphore buffer empty", err.Error())

	err = sema.Acquire()
	errNil(t, err)
	sema.Acquire()
	err = sema.Acquire()
	assertErr(t, "all semaphore buffer full", err)

	sema.Release()
	errNil(t, sema.Acquire())

	errNil(t, sema.Release())
	errNil(t, sema.Release())

	err = sema.Release()
	assertErr(t, "all semaphore buffer empty", err)
}
