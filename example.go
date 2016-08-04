package main

/*
#include "postgres.h"
#include "fmgr.h"
*/
import "C"
import "log"

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
	plan, err := db.Prepare("select * from test where id=$1", []string{"integer"})
	if err != nil {
		logger.Fatal(err)
	}
	insert, err := db.Prepare("insert into test (txt) values ($1)", []string{"text"})
	if err != nil {
		logger.Fatal(err)
	}

	//running statements
	err = insert.Exec("hello")
	if err != nil {
		logger.Fatal(err)
	}
	row, err := plan.QueryRow(1)
	if err != nil {
		logger.Fatal(err)
	}

	//scanning result row
	var id int
	var txt string
	err = row.Scan(&id, &txt)
	if err != nil {
		logger.Fatal(err)
	}
	logger.Printf("id: %d txt: %s", id, txt)

	//some magic with return value :)
	var ret string
	for i := 0; i < x; i++ {
		ret += t + txt
	}

	//return value must be converted to Datum
	return ToDatum(ret)
}
