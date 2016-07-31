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
	plan, err := PLGoPrepare("select * from test where id=$1", []string{"integer"})
	if err != nil {
		return PGVal(err)
	}
	defer plan.Close()
	rows, err := plan.Query(1)
	if err != nil {
		return PGVal(err)
	}
	var ret string
	for rows.Next() {
		//return PGVal("meh")
		var id int
		var txt string
		err = rows.Scan(&id, &txt)
		if err != nil {
			return PGVal(err)
		}
		for i := 0; i < x; i++ {
			ret += t + txt
		}
	}
	return PGVal(ret)
}
