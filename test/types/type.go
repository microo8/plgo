package main

/*
#include "postgres.h"
#include "fmgr.h"
*/
import "C"

import (
	"fmt"
	"log"
	"unsafe"
)

//export Complex
type Complex struct {
	X, Y float64
}

//export ComplexIn
func ComplexIn(fcinfo *funcInfo) Datum {
	logger := NewNoticeLogger("", log.Lshortfile|log.Ltime)
	var targ string
	fcinfo.Scan(&targ)
	logger.Println("meh")
	logger.Println(targ)
	var c Complex
	fmt.Sscanf(targ, "(%lf,%lf)", &c.X, &c.Y)
	return toDatum(unsafe.Pointer(&c))
}

//export ComplexOut
func ComplexOut(fcinfo *funcInfo) Datum {
	var c *Complex
	fcinfo.Scan(unsafe.Pointer(c))
	return toDatum(fmt.Sprintf("(%g,%g)", c.X, c.Y))
}
