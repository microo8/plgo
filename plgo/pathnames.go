// +build !windows

package main

// getcorrectpath used on Windows, see file pathnames_windows.go
func getcorrectpath(p string) string {
	return p
}

// addOtherIncludesAndLDFLAGS used on Windows, see file pathnames_windows.go
func addOtherIncludesAndLDFLAGS(plgoSource *string, postgresIncludeDir string) {
	return
}
