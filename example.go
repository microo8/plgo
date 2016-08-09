package main

/*
#include "postgres.h"
#include "fmgr.h"
*/
import "C"
import (
	"log"
	"time"
)

//export plgo_example
func plgo_example(fcinfo *FuncInfo) Datum {
	//getting the function parameters
	t := fcinfo.Text(0)
	x := fcinfo.Int(1)

	//Creating notice logger
	logger := log.New(&elog{}, "", log.Ltime|log.Lshortfile)

	//connect to DB
	db, err := Open()
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	//preparing query statements
	plan, err := db.Prepare("select * from test", nil)
	if err != nil {
		logger.Fatal(err)
	}

	//running statements
	rows, err := plan.Query()
	if err != nil {
		logger.Fatal(err)
	}

	//scanning result rows
	for rows.Next() {
		var id int
		var txt string
		var date time.Time
		var ti time.Time
		var titz time.Time
		err = rows.Scan(&id, &txt, &date, &ti, &titz)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Printf("id: %d txt: %s date: %s time: %s timetz: %s", id, txt, date, ti, titz)
	}

	//some magic with return value :)
	var ret string
	for i := 0; i < x; i++ {
		ret += t //+ txt
	}

	//return value must be converted to Datum
	return ToDatum(ret)
}
