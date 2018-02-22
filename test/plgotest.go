package main

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/microo8/plgo"
)

//PLGoTest testing function
func PLGoTest() {
	t := plgo.NewNoticeLogger("", log.Ltime|log.Lshortfile)
	defer t.Println("TEST end")

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
	testJSON(t)
	//testGoroutines(t)
}

func testConnection(t *log.Logger) {
	db, err := plgo.Open()
	if err != nil {
		t.Fatal(err)
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

	db, err := plgo.Open()
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

	db, err := plgo.Open()
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
	n := time.Now().Round(time.Second)
	var tests = []struct {
		query  string
		args   []interface{}
		result time.Time
	}{
		{"select '2016-01-01'::date", nil, time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"select $1", []interface{}{n}, n},
		{"select '2016-01-01'::timestamp with time zone - interval '1 year'", nil, time.Date(2015, 1, 1, 0, 0, 0, 0, time.Local)},
	}

	db, err := plgo.Open()
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

	db, err := plgo.Open()
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

	db, err := plgo.Open()
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

	db, err := plgo.Open()
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

	db, err := plgo.Open()
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

	db, err := plgo.Open()
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

	db, err := plgo.Open()
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

	db, err := plgo.Open()
	if err != nil {
		t.Fatal("error opening", err)
	}
	defer db.Close()

	var wg sync.WaitGroup

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

		wg.Add(1)
		test := test
		go func() {
			defer wg.Done()
			time.Sleep(time.Second)
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

type exampleStruct struct {
	Val1 int    `json:"val1"`
	Val2 string `json:"val2"`
}

func initJSONTable(db *plgo.DB) error {
	drop, err := db.Prepare("drop table if exists example", nil)
	if err != nil {
		return fmt.Errorf("prepare %s", err)
	}
	if err = drop.Exec(); err != nil {
		return fmt.Errorf("cannot drop table %s", err)
	}
	create, err := db.Prepare("create table example (id serial primary key, jsonval json, jsonbval jsonb)", nil)
	if err != nil {
		return fmt.Errorf("prepare %s", err)
	}
	if err = create.Exec(); err != nil {
		return fmt.Errorf("cannot create table %s", err)
	}
	insert, err := db.Prepare(`insert into example (jsonval, jsonbval) values ('{"val1":1,"val2":"foo"}','{"val1":1,"val2":"foo"}')`, nil)
	if err != nil {
		return fmt.Errorf("prepare %s", err)
	}
	if err = insert.Exec(); err != nil {
		return fmt.Errorf("cannot insert into table %s", err)
	}
	return nil
}

func testJSON(t *log.Logger) {
	db, err := plgo.Open()
	if err != nil {
		t.Fatal("error opening", err)
	}
	defer db.Close()
	if err = initJSONTable(db); err != nil {
		t.Fatal(err)
	}

	stmt, err := db.Prepare("select jsonval from example limit 1", nil)
	if err != nil {
		t.Fatal("prepare", err)
	}
	row, err := stmt.QueryRow()
	if err != nil {
		t.Fatal("query ", err)
	}
	var e exampleStruct
	if err = row.Scan(&e); err != nil {
		t.Fatal("json scan", err)
	}
	if e.Val1 != 1 || e.Val2 != "foo" {
		t.Fatalln("not correctly loaded json val", e)
	}

	stmt, err = db.Prepare("select jsonbval from example limit 1", nil)
	if err != nil {
		t.Fatal("prepare", err)
	}
	row, err = stmt.QueryRow()
	if err != nil {
		t.Fatal("query ", err)
	}
	var eb exampleStruct
	if err = row.Scan(&eb); err != nil {
		t.Fatal("json scan", err)
	}
	if eb.Val1 != 1 || eb.Val2 != "foo" {
		t.Fatalln("not correctly loaded json val", e)
	}

	insert, err := db.Prepare(`insert into example (id, jsonval, jsonbval) values (100, $1,$2)`, []string{"json", "jsonb"})
	if err != nil {
		t.Fatal("prepare", err)
	}
	e.Val1 = 2
	e.Val2 = "bar"
	eb.Val1 = 2
	eb.Val2 = "bar"
	if err = insert.Exec(e, eb); err != nil {
		t.Fatal("cannot insert into table", err)
	}

	stmt, err = db.Prepare("select jsonval, jsonbval from example where id=100", nil)
	if err != nil {
		t.Fatal("prepare", err)
	}
	row, err = stmt.QueryRow()
	if err != nil {
		t.Fatal("query ", err)
	}
	var es, esb exampleStruct
	err = row.Scan(&es, &esb)
	if err != nil {
		t.Fatal("json scan", err)
	}
	if es.Val1 != 2 || es.Val2 != "bar" {
		t.Fatalln("not correctly loaded json val", e)
	}
	if esb.Val1 != 2 || esb.Val2 != "bar" {
		t.Fatalln("not correctly loaded json val", e)
	}
}
