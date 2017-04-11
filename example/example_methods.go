package main

import (
	"log"
	"strings"

	"github.com/microo8/plgo"
)

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
		rows.Scan(&val)
		ret += val
	}
	return ret
}

func CreatedTimeTrigger(td *plgo.TriggerData) *plgo.TriggerRow {
	return nil
}

//ConcatArray concatenates an array of strings
func ConcatArray(strs []string) string {
	return strings.Join(strs, "")
}
