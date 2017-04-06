package main

const (
	plgo        = "plgo"
	triggerData = "TriggerData"
	triggerRow  = "TriggerRow"

	//TODO triggers (will have TriggerData arg, so must be extracted from fcinfo)
	methods = `package main

/*
#include "postgres.h"
#include "fmgr.h"
*/
import "C"

{{range $funcName, $func := .}}
//export {{$funcName}}
func {{$funcName}}(fcinfo *funcInfo) Datum {
    {{if $func.IsTrigger }}{{end}}
	{{range $func.Params}}var {{.Name}} {{.Type}}
	{{end}}
	fcinfo.Scan(
		{{range $func.Params}}&{{.Name}},
		{{end}})
	{{ if ne (len $func.ReturnType) 0 }}ret :={{end}} {{$funcName | ToLower }}(
		{{range $func.Params}}{{.Name}},
		{{end}})
	{{ if ne (len $func.ReturnType) 0 }}return toDatum(ret){{else}}return toDatum(nil){{end}}
}
{{end}}
`
	//TODO triggers
	sql = `{{range . }}
CREATE OR REPLACE FUNCTION {{.Schema}}.{{.Name}}({{range $funcParams}}{{.Name}} {{.Type}}, {{end}})
RETURNS {{.ReturnType}} AS
'$libdir/{{..Package}}', '{{.Name}}'
LANGUAGE c IMMUTABLE STRICT;
{{end}}`
)

var datumTypes = map[string]string{
	"error":       "text",
	"string":      "text",
	"[]byte":      "bytea",
	"int16":       "smallint",
	"uint16":      "smallint",
	"int32":       "integer",
	"uint32":      "integer",
	"int64":       "bigint",
	"int":         "bigint",
	"uint":        "bigint",
	"float32":     "real",
	"float64":     "double precision",
	"time.Time":   "timestamp with timezone",
	"bool":        "boolean",
	"[]string":    "text[]",
	"[]int16":     "smallint[]",
	"[]uint16":    "smallint[]",
	"[]int32":     "integer[]",
	"[]uint32":    "integer[]",
	"[]int64":     "bigint[]",
	"[]int":       "bigint[]",
	"[]uint":      "bigint[]",
	"[]float32":   "real[]",
	"[]float64":   "double precision[]",
	"[]bool":      "boolean[]",
	"[]time.Time": "timestamp with timezone[]",
	"TriggerRow":  "trigger",
}
