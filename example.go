package main

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

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
