package main

import "log"

func PLGoConcat(a, b string) string {
	t := log.New(&ELog{level: NOTICE}, "", log.Lshortfile|log.Ltime)
	err := log.Scan(&a, &b)
	t.Print("SCAAAAAN")
	if err != nil {
		panic(err)
	}
	t.Printf("args: '%s' and '%s'", a, b)
	return a + b
}
