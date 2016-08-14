package main

/*
#include "postgres.h"
#include "fmgr.h"
*/
import "C"
import (
	"log"
	"math"
	"time"
)

//export plgo_test
func plgo_test(fcinfo *FuncInfo) Datum {
	elog := &ELog{level: NOTICE}
	t := log.New(elog, "", log.Lshortfile|log.Ltime)

	TestConnection(t)
	TestQueryOutputText(t)
	TestQueryOutputInt(t)
	TestQueryOutputTime(t)
	TestQueryOutputBool(t)
	TestQueryOutputFloat32(t)
	TestQueryOutputFloat64(t)

	db, _ := Open()
	defer db.Close()

	update, _ := db.Prepare("update test set txt='abc'", []string{})
	update.Exec()

	elog.level = NOTICE
	t.Println("TEST end")
	return ToDatum(nil)
}

//export plgo_concat
func plgo_concat(fcinfo *FuncInfo) Datum {
	t := log.New(&ELog{level: NOTICE}, "", log.Lshortfile|log.Ltime)
	var a string
	var b string
	t.Print("SCAAAAAN")
	err := fcinfo.Scan(&a, &b)
	if err != nil {
		t.Print("fcinfo.Scan", err)
	}
	t.Printf("args: '%s' and '%s'", a, b)
	return ToDatum(a + b)
}

//export plgo_trigger
func plgo_trigger(fcinfo *FuncInfo) Datum {
	t := log.New(&ELog{level: NOTICE}, "", log.Lshortfile|log.Ltime)

	if !fcinfo.CalledAsTrigger() {
		t.Fatal("Not called as trigger")
	}

	triggerData := fcinfo.TriggerData()
	if !triggerData.FiredBefore() && !triggerData.FiredByUpdate() {
		t.Fatal("function not called BEFORE UPDATE :-O")
	}

	triggerData.NewRow.Set(4, time.Now().Add(-time.Hour*time.Duration(24)))

	//return ToDatum(nil) //nothing changed in the row
	//return ToDatum(triggerData.OldRow) //nothing changed in the row
	return ToDatum(triggerData.NewRow) //the new row will be changed
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
		{"select 'meh'", nil, "meh"},
		{"select '+ľščťžýáíé'::text", nil, "+ľščťžýáíé"},
		{"select lower('MEH')", nil, "meh"},
		{"select concat('foo', $1, 'bar')", []interface{}{"meh"}, "foomehbar"},
	}

	db, err := Open()
	if err != nil {
		t.Fatal("error opening", err)
	}
	defer db.Close()

	for _, test := range tests {
		var args []string = nil
		if len(test.args) > 0 {
			args = make([]string, len(test.args))
			for i := range test.args {
				args[i] = "text"
			}
		}
		stmt, err := db.Prepare(test.query, args)
		if err != nil {
			t.Fatal("prepare", err)
		}
		if stmt == nil {
			t.Fatal("plan is nil!")
		}

		rows, err := stmt.Query(test.args...)
		if err != nil {
			t.Fatal("Query ", err)
		}
		for rows.Next() {
			var res string
			err = rows.Scan(&res)
			if err != nil {
				t.Print(test, err)
			}
			if res != test.result {
				t.Print(test, "result not equal ", res, "!=", test.result)
			}
		}
	}
}

func TestQueryOutputInt(t *log.Logger) {
	var tests = []struct {
		query  string
		args   []interface{}
		result int
	}{
		{"select 1", nil, 1},
		{"select 1+1", nil, 2},
		{"select '12'::integer", nil, 12},
		{"select -123", nil, -123},
		{"select -1234567890", nil, -1234567890},
		{"select 2 * 3", nil, 2 * 3},
		{"select abs($1)", []interface{}{-100}, 100},
		{"select $1 + 200", []interface{}{-100}, 100},
	}

	db, err := Open()
	if err != nil {
		t.Fatal("error opening", err)
	}
	defer db.Close()

	for _, test := range tests {
		var args []string = nil
		if len(test.args) > 0 {
			args = make([]string, len(test.args))
			for i := range test.args {
				args[i] = "integer"
			}
		}
		stmt, err := db.Prepare(test.query, args)
		if err != nil {
			t.Fatal("prepare", err)
		}
		if stmt == nil {
			t.Fatal("plan is nil!")
		}

		rows, err := stmt.Query(test.args...)
		if err != nil {
			t.Fatal("Query ", err)
		}
		for rows.Next() {
			var res int
			err = rows.Scan(&res)
			if err != nil {
				t.Print("Scan ", test, err)
			}
			if res != test.result {
				t.Print(test, " result not equal ", res, "!=", test.result)
			}
		}
	}
}

func TestQueryOutputTime(t *log.Logger) {
	n := time.Now()
	n = n.Add(time.Nanosecond * time.Duration(-n.Nanosecond()))
	var tests = []struct {
		query  string
		args   []interface{}
		result time.Time
	}{
		{"select '2016-01-01'::date", nil, time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"select $1", []interface{}{n}, n},
		{"select '2016-01-01'::timestamp with time zone - interval '1 year'", nil, time.Date(2015, 1, 1, 0, 0, 0, 0, time.Local)},
	}

	db, err := Open()
	if err != nil {
		t.Fatal("error opening", err)
	}
	defer db.Close()

	for _, test := range tests {
		var args []string = nil
		if len(test.args) > 0 {
			args = make([]string, len(test.args))
			for i := range test.args {
				args[i] = "timestamp with time zone"
			}
		}
		stmt, err := db.Prepare(test.query, args)
		if err != nil {
			t.Fatal("prepare", err)
		}
		if stmt == nil {
			t.Fatal("plan is nil!")
		}

		rows, err := stmt.Query(test.args...)
		if err != nil {
			t.Fatal("Query ", err)
		}
		for rows.Next() {
			var res time.Time
			err = rows.Scan(&res)
			if err != nil {
				t.Print(test, err)
			}
			if res != test.result {
				t.Print(test, "result not equal ", res, "!=", test.result)
			}
		}
	}
}

func TestQueryOutputBool(t *log.Logger) {
	var tests = []struct {
		query  string
		args   []interface{}
		result bool
	}{
		{"select false", nil, false},
		{"select true", nil, true},
		{"select $1", []interface{}{true}, true},
		{"select $1", []interface{}{false}, false},
		{"select $1=true", []interface{}{false}, false},
	}

	db, err := Open()
	if err != nil {
		t.Fatal("error opening", err)
	}
	defer db.Close()

	for _, test := range tests {
		var args []string = nil
		if len(test.args) > 0 {
			args = make([]string, len(test.args))
			for i := range test.args {
				args[i] = "boolean"
			}
		}
		stmt, err := db.Prepare(test.query, args)
		if err != nil {
			t.Fatal("prepare", err)
		}
		if stmt == nil {
			t.Fatal("plan is nil!")
		}

		rows, err := stmt.Query(test.args...)
		if err != nil {
			t.Fatal("Query ", err)
		}
		for rows.Next() {
			var res bool
			err = rows.Scan(&res)
			if err != nil {
				t.Print(test, err)
			}
			if res != test.result {
				t.Print(test, "result not equal ", res, "!=", test.result)
			}
		}
	}
}

func TestQueryOutputFloat32(t *log.Logger) {
	var tests = []struct {
		query  string
		args   []interface{}
		result float32
	}{
		{"select 3.14::real", nil, 3.14},
		{"select (2 ^ 10)::real", nil, float32(math.Pow(2, 10))},
		{"select $1", []interface{}{float32(math.E)}, math.E},
		{"select $1", []interface{}{float32(math.Pi)}, math.Pi},
		{"select ($1 - 2)::real", []interface{}{float32(math.Phi)}, float32(math.Phi) - 2},
	}

	db, err := Open()
	if err != nil {
		t.Fatal("error opening", err)
	}
	defer db.Close()

	for _, test := range tests {
		var args []string = nil
		if len(test.args) > 0 {
			args = make([]string, len(test.args))
			for i := range test.args {
				args[i] = "real"
			}
		}
		stmt, err := db.Prepare(test.query, args)
		if err != nil {
			t.Fatal("prepare", err)
		}
		if stmt == nil {
			t.Fatal("plan is nil!")
		}

		rows, err := stmt.Query(test.args...)
		if err != nil {
			t.Fatal("Query ", err)
		}
		for rows.Next() {
			var res float32
			err = rows.Scan(&res)
			if err != nil {
				t.Print(test, err)
			}
			if res != test.result {
				t.Print(test, "result not equal ", res, "!=", test.result)
			}
		}
	}
}

func TestQueryOutputFloat64(t *log.Logger) {
	var tests = []struct {
		query  string
		args   []interface{}
		result float64
	}{
		{"select 3.14::double precision", nil, 3.14},
		{"select 2 ^ 10", nil, math.Pow(2, 10)},
		{"select $1", []interface{}{math.E}, math.E},
		{"select $1", []interface{}{math.Pi}, math.Pi},
		{"select pow($1,2)", []interface{}{math.Phi}, math.Pow(math.Phi, 2)},
	}

	db, err := Open()
	if err != nil {
		t.Fatal("error opening", err)
	}
	defer db.Close()

	for _, test := range tests {
		var args []string = nil
		if len(test.args) > 0 {
			args = make([]string, len(test.args))
			for i := range test.args {
				args[i] = "double precision"
			}
		}
		stmt, err := db.Prepare(test.query, args)
		if err != nil {
			t.Fatal("prepare", err)
		}
		if stmt == nil {
			t.Fatal("plan is nil!")
		}

		rows, err := stmt.Query(test.args...)
		if err != nil {
			t.Fatal("Query ", err)
		}
		for rows.Next() {
			var res float64
			err = rows.Scan(&res)
			if err != nil {
				t.Print(test, err)
			}
			if res != test.result {
				t.Print(test, "result not equal ", res, "!=", test.result)
			}
		}
	}
}
