package main

import (
	"bytes"
	"compress/gzip"
	"log"
	"strings"

	"github.com/paulhatch/plgo"
)

//Meh prints out message to error elog
func Meh() {
	logger := plgo.NewErrorLogger("", log.Ltime|log.Lshortfile)
	logger.Println("meh")
}

//ConcatAll concatenates all values of an column in a given table
func ConcatAll(tableName, colName string) string {
	logger := plgo.NewErrorLogger("", log.Ltime|log.Lshortfile)
	db, err := plgo.Open()
	if err != nil {
		logger.Fatalf("Cannot open DB: %s", err)
	}
	defer db.Close()
	query := "select " + colName + " from " + tableName
	stmt, err := db.Prepare(query, nil)
	if err != nil {
		logger.Fatalf("Cannot prepare query statement (%s): %s", query, err)
	}
	rows, err := stmt.Query()
	if err != nil {
		logger.Fatalf("Query (%s) error: %s", query, err)
	}
	var ret string
	for rows.Next() {
		var val string
		cols, err := rows.Columns()
		if err != nil {
			logger.Fatalln("Cannot get columns", err)
		}
		logger.Println(cols)
		err = rows.Scan(&val)
		if err != nil {
			logger.Fatalln("Cannot scan value", err)
		}
		ret += val
	}
	return ret
}

//CreatedTimeTrigger example trigger
func CreatedTimeTrigger(td *plgo.TriggerData) *plgo.TriggerRow {
	var id int
	var value string
	td.NewRow.Scan(&id, &value)
	td.NewRow.Set(0, id+10)
	td.NewRow.Set(1, value+value)
	return td.NewRow
}

//ConcatArray concatenates an array of strings
func ConcatArray(strs []string) string {
	return strings.Join(strs, "")
}

func GzipCompress(data []byte) []byte {
	logger := plgo.NewErrorLogger("", log.Ltime|log.Lshortfile)
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(data)
	if err != nil {
		logger.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		logger.Fatal(err)
	}
	return buf.Bytes()
}
