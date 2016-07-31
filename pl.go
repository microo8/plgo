package main

/*
#cgo CFLAGS: -Wall -Wpointer-arith -Wno-declaration-after-statement -Wendif-labels -Wmissing-format-attribute -Wformat-security -fno-strict-aliasing -fwrapv -fexcess-precision=standard -march=x86-64 -mtune=generic -O2 -pipe -fstack-protector-strong -fpic -I. -I./ -I/usr/include/postgresql/server -I/usr/include/postgresql/internal -D_FORTIFY_SOURCE=2 -D_GNU_SOURCE -I/usr/include/libxml2
#cgo LDFLAGS: -Wall -Wmissing-prototypes -Wpointer-arith -Wdeclaration-after-statement -Wendif-labels -Wmissing-format-attribute -Wformat-security -fno-strict-aliasing -fwrapv -fexcess-precision=standard -march=x86-64 -mtune=generic -O2 -pipe -fstack-protector-strong -fpic -L/usr/lib -Wl,-O1,--sort-common,--as-needed,-z,relro  -Wl,--as-needed -Wl,-rpath,'/usr/lib',--enable-new-dtags -shared

#include "plgo.h"

PG_FUNCTION_INFO_V1(plgo_example);
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

func main() {}

var SPI_conn bool

type Datum C.Datum
type FuncInfo C.FunctionCallInfoData
type Text *C.text
type Bytea *C.bytea
type SPIPlan C.SPIPlan

func Notice(text string) {
	C.notice(C.CString(text))
}

func (fcinfo *FuncInfo) Text(i uint) string {
	return TextToString(C.get_arg_text_p(fcinfo, C.uint(i)))
}

func (fcinfo *FuncInfo) Bytea(i uint) []byte {
	return ByteaToBytes(C.get_arg_bytea_p(fcinfo, C.uint(i)))
}

func (fcinfo *FuncInfo) Int16(i uint) int16 {
	return int16(C.get_arg_int16(fcinfo, C.uint(i)))
}

func (fcinfo *FuncInfo) Uint16(i uint) uint16 {
	return uint16(C.get_arg_uint16(fcinfo, C.uint(i)))
}

func (fcinfo *FuncInfo) Int32(i uint) int32 {
	return int32(C.get_arg_int32(fcinfo, C.uint(i)))
}

func (fcinfo *FuncInfo) Uint32(i uint) uint32 {
	return uint32(C.get_arg_uint32(fcinfo, C.uint(i)))
}

func (fcinfo *FuncInfo) Int64(i uint) int64 {
	return int64(C.get_arg_int64(fcinfo, C.uint(i)))
}

func (fcinfo *FuncInfo) Int(i uint) int {
	return int(C.get_arg_int64(fcinfo, C.uint(i)))
}

func (fcinfo *FuncInfo) Uint(i uint) uint {
	return uint(C.get_arg_uint32(fcinfo, C.uint(i)))
}

func TextToString(t Text) string {
	return C.GoString(C.text_to_cstring(t))
}

func ByteaToBytes(b Bytea) []byte {
	return C.GoBytes(b, C.varsize(b)-C.VARHDRSZ)
}

//PGVal returns the Postgresql C type from Golang type
func PGVal(val interface{}) Datum {
	switch v := val.(type) {
	case error:
		return (Datum)(C.cstring_to_datum(C.CString(v.Error())))
	case string:
		return (Datum)(C.cstring_to_datum(C.CString(v)))
	case int16:
		return (Datum)(C.int16_to_datum(C.int16(v)))
	case uint16:
		return (Datum)(C.uint16_to_datum(C.uint16(v)))
	case int32:
		return (Datum)(C.int32_to_datum(C.int32(v)))
	case uint32:
		return (Datum)(C.uint32_to_datum(C.uint32(v)))
	case int64:
		return (Datum)(C.int64_to_datum(C.int64(v)))
	case int:
		return (Datum)(C.int64_to_datum(C.int64(v)))
	case uint:
		return (Datum)(C.uint32_to_datum(C.uint32(v)))
	default:
		return (Datum)(C.void_datum())
	}
}

func (plan *SPIPlan) Close() error {
	if SPI_conn { //TODO this is not good
		if C.SPI_finish() != C.SPI_OK_FINISH {
			return errors.New("Error finish")
		}
	}
	return nil
}

func PLGoPrepare(query string, types []string) (*SPIPlan, error) {
	typeIds := make([]C.Oid, len(types))
	var typmod C.int32
	for i, t := range types {
		C.parseTypeString(C.CString(t), &typeIds[i], &typmod, C.false)
	}
	cquery := C.CString(query)
	if !SPI_conn { //TODO
		if C.SPI_connect() != C.SPI_OK_CONNECT {
			return nil, errors.New("can't connect")
		}
		SPI_conn = true
	}
	cplan := C.SPI_prepare(cquery, C.int(len(types)), &typeIds[0])
	plan := (*SPIPlan)(unsafe.Pointer(cplan))
	if plan != nil {
		return plan, nil
	} else {
		return nil, errors.New(fmt.Sprintf("SPI_prepare failed: %s", C.GoString(C.SPI_result_code_string(C.SPI_result))))
	}
}

func (plan *SPIPlan) Query(args ...interface{}) (*Rows, error) {
	values := make([]Datum, len(args))
	nulls := make([]C.char, len(args))
	for i, arg := range args {
		values[i] = PGVal(arg)
	}

	rv := C.SPI_execute_plan(plan, (*C.Datum)(unsafe.Pointer(&values[0])), &nulls[0], C.true, 0)
	if rv == C.SPI_OK_SELECT && C.SPI_processed > 0 {
		Notice(fmt.Sprintf("SPI_processed: %d", C.SPI_processed))
		return &Rows{
			tupleTable: C.SPI_tuptable,
			processed:  uint32(C.SPI_processed),
			current:    -1,
		}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("SPI_prepare failed: %s", C.GoString(C.SPI_result_code_string(C.SPI_result))))
	}
}

type Rows struct {
	tupleTable *C.SPITupleTable
	processed  uint32
	current    int
}

func (rows *Rows) Next() bool {
	rows.current++
	return rows.current < int(rows.processed)
}

func (rows *Rows) Scan(args ...interface{}) error {
	for i, arg := range args {
		Notice(fmt.Sprintf("current: %d", rows.current))
		val := C.get_col_as_datum(rows.tupleTable.vals, rows.tupleTable.tupdesc, C.uint32(rows.current), C.int(i))
		Notice("meh")
		switch targ := arg.(type) {
		case *string:
			*targ = C.GoString(C.datum_to_cstring(val))
		case *int16:
			*targ = int16(C.datum_to_int16(val))
		case *uint16:
			*targ = uint16(C.datum_to_uint16(val))
		case *int32:
			*targ = int32(C.datum_to_int32(val))
		case *uint32:
			*targ = uint32(C.datum_to_uint32(val))
		case *int64:
			*targ = int64(C.datum_to_int64(val))
		}
		convertAssign(arg, val)
	}
	return nil
}
