package main

/*
#include "postgres.h"
#include "fmgr.h"
*/
import "C"

//export plgo_example
func plgo_example(fcinfo *FuncInfo) Datum {
	//getting the function parameters
	t := fcinfo.Text(0)
	x := fcinfo.Int(1)

	//preparing query statements
	plan, err := Prepare("select * from test where id=$1", []string{"integer"})
	if err != nil {
		return ToDatum(err)
	}
	defer plan.Close()
	insert, err := Prepare("insert into test (txt) values ($1)", []string{"text"})
	if err != nil {
		return ToDatum(err)
	}
	defer insert.Close()

	//running statements
	err = insert.Exec("hello")
	if err != nil {
		return ToDatum(err)
	}
	row, err := plan.QueryRow(1)
	if err != nil {
		return ToDatum(err)
	}

	//scanning result row
	var id int
	var txt string
	err = row.Scan(&id, &txt)
	if err != nil {
		return ToDatum(err)
	}

	//some magic with return value :)
	var ret string
	for i := 0; i < x; i++ {
		ret += t + txt
	}
	return ToDatum(ret)
}
