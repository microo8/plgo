package main

const (
	plgo        = "plgo"
	triggerData = "TriggerData"
	triggerRow  = "TriggerRow"

	//TODO triggers (will have TriggerData arg, so must be extracted from fcinfo)

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
