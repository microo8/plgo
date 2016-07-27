package main

/*
#include "postgres.h"
#include "fmgr.h"
*/
import "C"

//export plgo_func
func plgo_func(fcinfo *FuncInfo) Datum {
	t := fcinfo.Text(0)
	x := fcinfo.Int(1)
	var ret string = ""
	for i := 0; i < x; i++ {
		ret += t
	}
	return PGVal(ret)
}
