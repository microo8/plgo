package main

import "log"
import "github.com/microo8/plgo/lib"

func PLGoConcat(a, b string) string {
	t := log.New(&plgo.ELog{level: NOTICE}, "", log.Lshortfile|log.Ltime)
	err := log.Scan(&a, &b)
	t.Print("SCAAAAAN")
	if err != nil {
		panic(err)
	}
	t.Printf("args: '%s' and '%s'", a, b)
	return a + b
}
