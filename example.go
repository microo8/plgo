package main

/*
#include "postgres.h"
#include "fmgr.h"
*/
import "C"

//export plgo_example
func plgo_example(fcinfo *FuncInfo) Datum {
	t := fcinfo.Text(0)
	x := fcinfo.Int(1)
	var ret string = ""
	for i := 0; i < x; i++ {
		ret += t
	}
	return PGVal(ret)
}
