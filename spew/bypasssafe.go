// NOTE: Due to the following build constraints, this file will only be compiled
// when the code is running on Google App Engine, compiled by GopherJS, or
// "-tags safe" is added to the go build command line.  The "disableunsafe"
// tag is deprecated and thus should not be used.
//go:build js || appengine || safe || disableunsafe || !go1.4
// +build js appengine safe disableunsafe !go1.4

package spew

import "reflect"

const (
	// UnsafeDisabled is a build-time constant which specifies whether or
	// not access to the unsafe package is available.
	UnsafeDisabled = true
)

// unsafeReflectValue typically converts the passed reflect.Value into a one
// that bypasses the typical safety restrictions preventing access to
// unaddressable and unexported data.  However, doing this relies on access to
// the unsafe package.  This is a stub version which simply returns the passed
// reflect.Value when the unsafe package is not available.
func unsafeReflectValue(v reflect.Value) reflect.Value {
	return v
}
