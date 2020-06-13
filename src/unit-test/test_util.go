package tests

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

// assert fails the test if the condition is false.
func assert(tb testing.TB, condition bool, msg string, v ...interface{}) {
	if !condition {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("[%s:%d: "+msg+"]\n\n", append([]interface{}{filepath.Base(file), line}, v...)...)
		tb.FailNow()
	}
}

// test if an err is not nil.
func errNil(tb testing.TB, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("[%s:%d: unexpected error: %s]\n\n", filepath.Base(file), line, err.Error())
		tb.FailNow()
	}
}

// test if an err is not nil.
func assertErr(tb testing.TB, exp string, err error) {
	if err == nil {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("[%s:%d: error is expected.]\n\n", filepath.Base(file), line)
		tb.FailNow()
	} else {
		equals(tb, exp, err.Error())
	}
}

// equals fails the test if exp is not equal to act.
func equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}
