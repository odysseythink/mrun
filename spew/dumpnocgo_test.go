// NOTE: Due to the following build constraints, this file will only be compiled
// when either cgo is not supported or "-tags testcgo" is not added to the go
// test command line.  This file intentionally does not setup any cgo tests in
// this scenario.
//go:build !cgo || !testcgo
// +build !cgo !testcgo

package spew_test

func addCgoDumpTests() {
	// Don't add any tests for cgo since this file is only compiled when
	// there should not be any cgo tests.
}
