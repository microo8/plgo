package main

/*
#include "postgres.h"
#include "fmgr.h"
*/
import "C"
import (
	"log"
)

//export plgo_test
func plgo_test(fcinfo *FuncInfo) Datum {
	elog := &ELog{level: ERROR}
	t := log.New(elog, "", log.Lshortfile|log.Ltime)
	TestConnection(t)
	TestQueryOutputText(t)
	elog.level = NOTICE
	t.Println("TEST end")
	return ToDatum(nil)
}

func TestConnection(t *log.Logger) {
	db, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	_, err = Open()
	if err == nil {
		t.Fatal("Double openned")
	}
	err = db.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = db.Close()
	if err == nil {
		t.Fatal("Double closed")
	}
}

func TestQueryOutputText(t *log.Logger) {
	var tests = []struct {
		query  string
		args   []interface{}
		result string
	}{
		{"select '1'::text", nil, "1"},
		{"select 1::text", nil, "1"},
		{"select 'meh'::text", nil, "meh"},
		{"select '+ľščťžýáíé'::text", nil, "+ľščťžýáíé"},
		{"select lower('MEH')", nil, "meh"},
		{"select concat('foo', $1, 'bar')", []interface{}{"meh"}, "foomehbar"},
	}

	db, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for _, test := range tests {
		t.Println("Running", test)
		args := make([]string, len(test.args))
		for range test.args {
			args = append(args, "text")
		}
		stmt, err := db.Prepare(test.query, args)
		if err != nil {
			t.Fatal(err)
		}
		if stmt == nil {
			t.Fatal("plan is nil!")
		}
		rows, err := stmt.Query(test.args...)
		for rows.Next() {
			var res string
			err = rows.Scan(&res)
			if err != nil {
				t.Fatal(test, err)
			}
			if res != test.result {
				t.Fatal("result not equal", res, test.result)
			}
		}
		t.Println("End", test)
	}
}
