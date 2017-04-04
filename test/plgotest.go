package main

/*
#cgo CFLAGS: -I/usr/include/postgresql/server
#cgo LDFLAGS: -shared

#include "postgres.h"
#include "fmgr.h"
*/
import "C"
import (
	"log"
	"math"
	"sync"
	"time"
)

//PLGoTest testing function
//export PLGoTest
func PLGoTest(fcinfo *FuncInfo) Datum {
	elog := &ELog{Level: NOTICE}
	t := log.New(elog, "", log.Lshortfile|log.Ltime)

	testConnection(t)
	testQueryOutputText(t)
	testQueryOutputInt(t)
	testQueryOutputTime(t)
	testQueryOutputBool(t)
	testQueryOutputFloat32(t)
	testQueryOutputFloat64(t)
	testQueryOutputArrayText(t)
	testQueryOutputArrayInt(t)
	testQueryOutputArrayFloat(t)
	testGoroutines(t)

	db, _ := Open()
	defer db.Close()

	update, _ := db.Prepare("update test set txt='abc'", []string{})
	update.Exec()

	elog.Level = NOTICE
	t.Println("TEST end")
	return ToDatum(nil)
}

//PLGoConcat concatenates two strings
//export PLGoConcat
func PLGoConcat(fcinfo *FuncInfo) Datum {
	t := log.New(&ELog{Level: NOTICE}, "", log.Lshortfile|log.Ltime)
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

//PLGoTrigger is an trigger test function
//export PLGoTrigger
func PLGoTrigger(fcinfo *FuncInfo) Datum {
	t := log.New(&ELog{Level: NOTICE}, "", log.Lshortfile|log.Ltime)

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

func testConnection(t *log.Logger) {
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

func testQueryOutputText(t *log.Logger) {
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
		var args []string
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

func testQueryOutputInt(t *log.Logger) {
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
		var args []string
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

func testQueryOutputTime(t *log.Logger) {
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
		var args []string
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

func testQueryOutputBool(t *log.Logger) {
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
		var args []string
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

func testQueryOutputFloat32(t *log.Logger) {
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
		var args []string
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

func testQueryOutputFloat64(t *log.Logger) {
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
		var args []string
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

func testQueryOutputArrayText(t *log.Logger) {
	var tests = []struct {
		query  string
		args   []interface{}
		result []string
	}{
		{"select array['1','2']", nil, []string{"1", "2"}},
		{"select string_to_array('meh#foo#bar','#')", nil, []string{"meh", "foo", "bar"}},
		{"select $1", []interface{}{[]string{"meh", "foo"}}, []string{"meh", "foo"}},
		{"select array_append($1,'bar')", []interface{}{[]string{"meh", "foo"}}, []string{"meh", "foo", "bar"}},
		{"select array_remove($1,'meh')", []interface{}{[]string{"meh", "foo"}}, []string{"foo"}},
	}

	db, err := Open()
	if err != nil {
		t.Fatal("error opening", err)
	}
	defer db.Close()

	for _, test := range tests {
		var args []string
		if len(test.args) > 0 {
			args = make([]string, len(test.args))
			for i := range test.args {
				args[i] = "text[]"
			}
		}
		stmt, err := db.Prepare(test.query, args)
		if err != nil {
			t.Print("prepare", err)
		}
		if stmt == nil {
			t.Print("plan is nil!")
		}

		rows, err := stmt.Query(test.args...)
		if err != nil {
			t.Print("Query ", err)
		}
		for rows.Next() {
			var res []string
			err = rows.Scan(&res)
			if err != nil {
				t.Print(test, err)
			}
			eq := len(res) == len(test.result)
			for i := 0; eq && i < len(res); i++ {
				eq = res[i] == test.result[i]
			}
			if !eq {
				t.Print(test, "result not equal ", res, "!=", test.result)
			}
		}
	}
}

func testQueryOutputArrayInt(t *log.Logger) {
	var tests = []struct {
		query  string
		args   []interface{}
		result []int
	}{
		{"select array[1,2]", nil, []int{1, 2}},
		{"select $1", []interface{}{[]int{123, 456}}, []int{123, 456}},
		{"select array_append($1,100)", []interface{}{[]int{1234, 5678}}, []int{1234, 5678, 100}},
		{"select array_remove($1,100)", []interface{}{[]int{12345, 100, 67890}}, []int{12345, 67890}},
	}

	db, err := Open()
	if err != nil {
		t.Fatal("error opening", err)
	}
	defer db.Close()

	for _, test := range tests {
		var args []string
		if len(test.args) > 0 {
			args = make([]string, len(test.args))
			for i := range test.args {
				args[i] = "int[]"
			}
		}
		stmt, err := db.Prepare(test.query, args)
		if err != nil {
			t.Print("prepare", err)
		}
		if stmt == nil {
			t.Print("plan is nil!")
		}

		rows, err := stmt.Query(test.args...)
		if err != nil {
			t.Print("Query ", err)
		}
		for rows.Next() {
			var res []int
			err = rows.Scan(&res)
			if err != nil {
				t.Print(test, err)
			}
			eq := len(res) == len(test.result)
			for i := 0; eq && i < len(res); i++ {
				eq = res[i] == test.result[i]
			}
			if !eq {
				t.Print(test, "result not equal ", res, "!=", test.result)
			}
		}
	}
}

func testQueryOutputArrayFloat(t *log.Logger) {
	var tests = []struct {
		query  string
		args   []interface{}
		result []float64
	}{
		{"select array[1.2::double precision,2.3::double precision]", nil, []float64{1.2, 2.3}},
		{"select $1", []interface{}{[]float64{123.2, 456.34}}, []float64{123.2, 456.34}},
		{"select $1", []interface{}{[]float64{1e-23, 2.3e-45}}, []float64{1e-23, 2.3e-45}},
		{"select array_append($1,100.001::double precision)", []interface{}{[]float64{1234.123123, 5678.456456}}, []float64{1234.123123, 5678.456456, 100.001}},
		{"select array_remove($1,100.001::double precision)", []interface{}{[]float64{12345.123123, 100.001, 67890.456456}}, []float64{12345.123123, 67890.456456}},
	}

	db, err := Open()
	if err != nil {
		t.Fatal("error opening", err)
	}
	defer db.Close()

	for _, test := range tests {
		var args []string
		if len(test.args) > 0 {
			args = make([]string, len(test.args))
			for i := range test.args {
				args[i] = "double precision[]"
			}
		}
		stmt, err := db.Prepare(test.query, args)
		if err != nil {
			t.Print("prepare", err)
		}
		if stmt == nil {
			t.Print("plan is nil!")
		}

		rows, err := stmt.Query(test.args...)
		if err != nil {
			t.Print("Query ", err)
		}
		for rows.Next() {
			var res []float64
			err = rows.Scan(&res)
			if err != nil {
				t.Print(test, err)
			}
			eq := len(res) == len(test.result)
			for i := 0; eq && i < len(res); i++ {
				eq = res[i] == test.result[i]
			}
			if !eq {
				t.Print(test, "result not equal ", res, "!=", test.result)
			}
		}
	}
}

func testGoroutines(t *log.Logger) {
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

	var wg sync.WaitGroup

	for i, test := range tests {

		var args []string
		if len(test.args) > 0 {
			args = make([]string, len(test.args))
			for i := range test.args {
				args[i] = "text"
			}
		}
		t.Println(i, "Prepare1", test)
		stmt, err := db.Prepare(test.query, args)
		t.Println(i, "Prepare2", test)
		if err != nil {
			t.Fatal("prepare", err)
		}
		t.Println(i, "Prepare3", test)
		if stmt == nil {
			t.Fatal("plan is nil!")
		}
		t.Println(i, "Prepare4", test)

		wg.Add(1)
		i := i
		test := test
		go func() {
			t.Println("goroutine", i)
			defer wg.Done()
			time.Sleep(time.Second)
			t.Println("query", i)
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
		}()
	}

	wg.Wait()
}
