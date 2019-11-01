// +build windows

// also need to include
// include\server\port\win32_msvc
// include\server\port\win32
// include\server
// include

package main

import (
	"strings"

	"golang.org/x/sys/windows"
)

// getcorrectpath works around 8.3 name returned by pg_config --includedir-server.
func getcorrectpath(p string) string {
	repl := p
	if p[len(p)-2:] == "\r\n" {
		repl = p[:len(p)-2]
	}

	ret, err := shortToLongPath(repl)
	if err != nil {
		panic("in getcorrectpath: " + err.Error() + "\"" + repl + "\"")
	}
	return ret
}

// shortToLongPath makes syscall to transform 8.3 filaname form to 'long' filename
func shortToLongPath(short83path string) (string, error) {
	ptrshort, err := windows.UTF16PtrFromString(short83path)
	if err != nil {
		return "", err
	}
	n := uint32(400) // "enough" space
morechars:
	b := make([]uint16, n)
	n, err = windows.GetLongPathName(ptrshort, &b[0], uint32(len(b)))
	if err != nil {
		if n != 0 {
			// here n is already set by kernel to right count of tchars
			goto morechars
		}
		return "", err
	}
	return windows.UTF16ToString(b[:n]), nil
}

// addOtherIncludesAndLDFLAGS used on Windows.
// adds to CFLAGS -I .../server/port/win32
// adds minGw gcc peculiarities workarounds:
// adds #cgo CFLAGS:  -DHAVE_LONG_LONG_INT_64 -I"{{postgresinclude}}/port/win32"
// adds #cgo LDFLAGS: -L../ -L "{{postgresinclude}}/../../lib/"
// adds #cgo LDFLAGS: -lpostgresInterfaceLib
func addOtherIncludesAndLDFLAGS(plgoSource *string, postgresIncludeDir string) {
	windowsCFLAGS := `
#cgo CFLAGS:  -DHAVE_LONG_LONG_INT_64 -I"{{postgresinclude}}/port/win32"
// our import(interface) library libpostgresInterfaceLib.a build by mingw's dlltool.exe is in -L../
#cgo LDFLAGS: -L../ -L "{{postgresinclude}}/../../lib/"
// import library postgres.lib build by msvc CAN'T be used by mingw gcc on windows, silent erronous usage.
#cgo LDFLAGS: -lpostgresInterfaceLib  
`
	adds := strings.Replace(windowsCFLAGS, "{{postgresinclude}}", postgresIncludeDir, -1)

	*plgoSource = strings.Replace(*plgoSource, "//{windowsCFLAGS}", adds, 1)

	return
}
