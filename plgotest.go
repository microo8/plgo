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
	elog := &ELog{level: NOTICE}
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
		args   []string
		result string
	}{
		{"select '1'::text", nil, "1"},
		{"select 1::text", nil, "1"},
		{"select 'meh'", nil, "meh"},
		{"select '+ľščťžýáíé'::text", nil, "+ľščťžýáíé"},
		{"select lower('MEH')", nil, "meh"},
		{"select concat('foo', $1, 'bar')", []string{"meh"}, "foomehbar"},
	}

	db, err := Open()
	if err != nil {
		t.Fatal("error opening", err)
	}
	defer db.Close()

	for _, test := range tests {
		t.Print("running: ", test)
		var args []string = nil
		if len(test.args) > 0 {
			args = make([]string, len(test.args))
			for i := range test.args {
				args[i] = "text"
			}
		}
		t.Print(1)
		stmt, err := db.Prepare(test.query, args)
		t.Print(2)
		if err != nil {
			t.Fatal("prepare", err)
		}
		if stmt == nil {
			t.Fatal("plan is nil!")
		}
		t.Print(3, test.args)
		var rows *Rows
		if len(test.args) > 0 {
			rows, err = stmt.Query(test.args[0])
		} else {
			rows, err = stmt.Query()
		}
		t.Print(4)
		for rows.Next() {
			t.Print(5)
			var res string
			err = rows.Scan(&res)
			if err != nil {
				t.Fatal(test, err)
			}
			if res != test.result {
				t.Fatal("result not equal ", res, "!=", test.result)
			}
		}
	}
}
