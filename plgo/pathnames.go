// +build !windows

package main

import (
	"io/ioutil"
	"strings"
)

// getcorrectpath used on Windows, see file pathnames_windows.go
func getcorrectpath(p string) string {
	ret := strings.TrimRight(p, "\n")
	return ret
}

// addOtherIncludesAndLDFLAGS used on Windows, see file pathnames_windows.go
func addOtherIncludesAndLDFLAGS(plgoSource *string, postgresIncludeDir string) {
	return
}

func buildPath() (string, error) {
	return ioutil.TempDir("", plgo)
}
