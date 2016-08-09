package main

/*
#cgo CFLAGS: -fPIC -I/usr/include/postgresql/server
#cgo LDFLAGS: -fPIC -shared

#include "postgres.h"
#include "fmgr.h"
#include "pgtime.h"
#include "catalog/pg_type.h"
#include "utils/builtins.h"
#include "utils/date.h"
#include "utils/timestamp.h"
#include "utils/elog.h"
#include "executor/spi.h"
#include "parser/parse_type.h"

#ifdef PG_MODULE_MAGIC
PG_MODULE_MAGIC;
#endif

int varsize(void *var) {
    return VARSIZE(var);
}

void elog_notice(char* string) {
    elog(NOTICE, string, "");
}

void elog_error(char* string) {
    elog(ERROR, string, "");
}

HeapTuple get_heap_tuple(HeapTuple* ht, uint i) {
    return ht[i];
}

Datum get_col_as_datum(HeapTuple ht, TupleDesc td, int colnumber) {
    bool isNull = true;
    return SPI_getbinval(ht, td, colnumber + 1, &isNull);
}


//Get value from function args/////////////////////////////////////////////
text* get_arg_text_p(PG_FUNCTION_ARGS, uint i) {
    return PG_GETARG_TEXT_P(i);
}

bytea* get_arg_bytea_p(PG_FUNCTION_ARGS, uint i) {
    return PG_GETARG_BYTEA_P(i);
}

int16 get_arg_int16(PG_FUNCTION_ARGS, uint i) {
    return PG_GETARG_INT16(i);
}

uint16 get_arg_uint16(PG_FUNCTION_ARGS, uint i) {
    return PG_GETARG_UINT32(i);
}

int32 get_arg_int32(PG_FUNCTION_ARGS, uint i) {
    return PG_GETARG_INT32(i);
}

uint32 get_arg_uint32(PG_FUNCTION_ARGS, uint i) {
    return PG_GETARG_UINT32(i);
}

int64 get_arg_int64(PG_FUNCTION_ARGS, uint i) {
    return PG_GETARG_INT64(i);
}

DateADT get_arg_date(PG_FUNCTION_ARGS, uint i) {
	return PG_GETARG_DATEADT(i);
}

Timestamp get_arg_time(PG_FUNCTION_ARGS, uint i) {
	return PG_GETARG_TIMESTAMP(i);
}

TimestampTz get_arg_timetz(PG_FUNCTION_ARGS, uint i) {
	return PG_GETARG_TIMESTAMPTZ(i);
}

bool get_arg_bool(PG_FUNCTION_ARGS, uint i) {
	return PG_GETARG_BOOL(i);
}

//val to datum//////////////////////////////////////////////////
Datum void_datum(){
    PG_RETURN_VOID();
}

Datum cstring_to_datum(char *val) {
    return CStringGetDatum(cstring_to_text(val));
}

Datum int16_to_datum(int16 val) {
    return Int16GetDatum(val);
}

Datum uint16_to_datum(uint16 val) {
    return UInt16GetDatum(val);
}

Datum int32_to_datum(int32 val) {
    return Int32GetDatum(val);
}

Datum uint32_to_datum(uint32 val) {
    return UInt32GetDatum(val);
}

Datum int64_to_datum(int64 val) {
    return Int64GetDatum(val);
}

Datum date_to_datum(DateADT val){
	return DateADTGetDatum(val);
}

Datum time_to_datum(TimeADT val){
	return TimestampGetDatum(val);
}

Datum timetz_to_datum(TimestampTz val) {
	return TimestampTzGetDatum(val);
}

Datum bool_to_datum(bool val) {
	return BoolGetDatum(val);
}

//Datum to val //////////////////////////////////////////////////////////
char* datum_to_cstring(Datum val) {
    return DatumGetCString(text_to_cstring((struct varlena *)val));
}

int16 datum_to_int16(Datum val) {
    return DatumGetInt16(val);
}

uint16 datum_to_uint16(Datum val) {
    return DatumGetUInt16(val);
}

int32 datum_to_int32(Datum val) {
    return DatumGetInt32(val);
}

uint32 datum_to_uint32(Datum val) {
    return DatumGetUInt32(val);
}

int64 datum_to_int64(Datum val) {
    return DatumGetInt64(val);
}

DateADT datum_to_date(Datum val) {
	return DatumGetDateADT(val);
}

Timestamp datum_to_time(Datum val) {
	return DatumGetTimestamp(val);
}

TimestampTz datum_to_timetz(Datum val) {
	return DatumGetTimestampTz(val);
}

bool datum_to_bool(Datum val) {
	return DatumGetBool(val);
}

//PG_FUNCTION declarations
#include "funcdec.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

//this hase to be here
func main() {}

//DB connection
type DB struct {
	lock sync.Mutex
}

//Returns DB connection and runs SPI_connect
func Open() (*DB, error) {
	if C.SPI_connect() != C.SPI_OK_CONNECT {
		return nil, errors.New("can't connect")
	}
	return new(DB), nil
}

//Closes the DB connection
func (db *DB) Close() error {
	db.lock.Lock()
	defer db.lock.Unlock()
	if C.SPI_finish() != C.SPI_OK_FINISH {
		return errors.New("Error closing DB")
	}
	return nil
}

//ELog represents the elog io.Writter to use with Logger
type ELogLevel int

const (
	NOTICE ELogLevel = iota
	ERROR
)

type ELog struct {
	lock  sync.Mutex
	level ELogLevel
}

//notify implemented as io.Writter
func (e *ELog) Write(p []byte) (n int, err error) {
	e.print(string(p))
	return len(p), nil
}

func (e *ELog) print(str string) {
	e.lock.Lock()
	defer e.lock.Unlock()
	switch e.level {
	case NOTICE:
		C.elog_notice(C.CString(str))
	case ERROR:
		C.elog_error(C.CString(str))
	}
}

func (e *ELog) Print(args ...interface{}) {
	e.print(fmt.Sprint(args...))
}

func (e *ELog) Printf(format string, args ...interface{}) {
	e.print(fmt.Sprintf(format, args...))
}

func (e *ELog) Println(args ...interface{}) {
	e.print(fmt.Sprintln(args...))
}

//FuncInfo is the type of parameters that all functions get
type FuncInfo C.FunctionCallInfoData

//Returns i'th parameter of the function and converts it from text to string TODO: error?
func (fcinfo *FuncInfo) Text(i uint) string {
	return C.GoString(C.text_to_cstring(C.get_arg_text_p(fcinfo, C.uint(i))))
}

//Returns i'th parameter of the function and converts it from bytea to []byte TODO: error?
func (fcinfo *FuncInfo) Bytea(i uint) []byte {
	b := C.get_arg_bytea_p(fcinfo, C.uint(i)) //TODO check this
	return C.GoBytes(b, C.varsize(b)-C.VARHDRSZ)
}

//Returns i'th parameter of the function and converts it to int16 TODO: error?
func (fcinfo *FuncInfo) Int16(i uint) int16 {
	return int16(C.get_arg_int16(fcinfo, C.uint(i)))
}

//Returns i'th parameter of the function and converts it to uint16 TODO: error?
func (fcinfo *FuncInfo) Uint16(i uint) uint16 {
	return uint16(C.get_arg_uint16(fcinfo, C.uint(i)))
}

//Returns i'th parameter of the function and converts it to int32 TODO: error?
func (fcinfo *FuncInfo) Int32(i uint) int32 {
	return int32(C.get_arg_int32(fcinfo, C.uint(i)))
}

//Returns i'th parameter of the function and converts it to uint32 TODO: error?
func (fcinfo *FuncInfo) Uint32(i uint) uint32 {
	return uint32(C.get_arg_uint32(fcinfo, C.uint(i)))
}

//Returns i'th parameter of the function and converts it to int64 TODO: error?
func (fcinfo *FuncInfo) Int64(i uint) int64 {
	return int64(C.get_arg_int64(fcinfo, C.uint(i)))
}

//Returns i'th parameter of the function and converts it to int TODO: error?
func (fcinfo *FuncInfo) Int(i uint) int {
	return int(C.get_arg_int64(fcinfo, C.uint(i)))
}

//Returns i'th parameter of the function and converts it to uint TODO: error?
func (fcinfo *FuncInfo) Uint(i uint) uint {
	return uint(C.get_arg_uint32(fcinfo, C.uint(i)))
}

func (fcinfo *FuncInfo) Date(i uint) time.Time {
	date := C.get_arg_date(fcinfo, C.uint(i))
	return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(date))
}

func (fcinfo *FuncInfo) Time(i uint) time.Time {
	t := C.get_arg_time(fcinfo, C.uint(i))
	return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Second * time.Duration(int64(t)/int64(C.USECS_PER_SEC)))
}

func (fcinfo *FuncInfo) TimeTz(i uint) time.Time {
	t := C.get_arg_timetz(fcinfo, C.uint(i))
	return time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Second * time.Duration(int64(t)/int64(C.USECS_PER_SEC))).Local()
}

func (fcinfo *FuncInfo) Bool(i uint) bool {
	return C.get_arg_bool(fcinfo, C.uint(i)) == C.true
}

//Datum is the return type of postgresql
type Datum C.Datum

//ToDatum returns the Postgresql C type from Golang type
func ToDatum(val interface{}) Datum {
	switch v := val.(type) {
	case error:
		return (Datum)(C.cstring_to_datum(C.CString(v.Error())))
	case string:
		return (Datum)(C.cstring_to_datum(C.CString(v)))
	case []byte:
		return *(*Datum)(unsafe.Pointer(&v[0]))
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
	case time.Time:
		return (Datum)(C.timetz_to_datum(C.TimestampTz(v.UTC().UnixNano())))
	case bool:
		if v {
			return (Datum)(C.bool_to_datum(C.true))
		} else {
			return (Datum)(C.bool_to_datum(C.false))
		}
	default:
		return (Datum)(C.void_datum())
	}
}

//Prepared SQL statement
type Plan struct {
	spi_plan C.SPIPlanPtr
	db       *DB
}

//Prepare prepares an SQL query and returns a Plan that can be executed
//query - the SQL query
//types - an array of strings with type names from postgresql of the prepared query
func (db *DB) Prepare(query string, types []string) (*Plan, error) {
	var typeIdsP *C.Oid
	if len(types) > 0 {
		typeIds := make([]C.Oid, len(types))
		var typmod C.int32
		for i, t := range types {
			C.parseTypeString(C.CString(t), &typeIds[i], &typmod, C.false)
		}
		typeIdsP = &typeIds[0]
	}
	cplan := C.SPI_prepare(C.CString(query), C.int(len(types)), typeIdsP)
	if cplan != nil {
		return &Plan{spi_plan: cplan, db: db}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("Prepare failed: %s", C.GoString(C.SPI_result_code_string(C.SPI_result))))
	}
}

//Query executes the prepared Plan with the provided args and returns
//multiple Rows result, that can be iterated
func (plan *Plan) Query(args ...interface{}) (*Rows, error) {
	valuesP, nullsP := spiArgs(args)
	plan.db.lock.Lock()
	defer plan.db.lock.Unlock()
	rv := C.SPI_execute_plan(plan.spi_plan, valuesP, nullsP, C.true, 0)
	if rv == C.SPI_OK_SELECT && C.SPI_processed > 0 {
		return newRows(C.SPI_tuptable.vals, C.SPI_tuptable.tupdesc, C.SPI_processed), nil
	} else {
		return nil, errors.New(fmt.Sprintf("Query failed: %s", C.GoString(C.SPI_result_code_string(C.SPI_result))))
	}
}

//Query executes the prepared Plan with the provided args and returns
//multiple Rows result, that can be iterated
func (plan *Plan) QueryRow(args ...interface{}) (*Row, error) {
	valuesP, nullsP := spiArgs(args)
	plan.db.lock.Lock()
	defer plan.db.lock.Unlock()
	rv := C.SPI_execute_plan(plan.spi_plan, valuesP, nullsP, C.false, 1)
	if rv >= C.int(0) && C.SPI_processed == 1 {
		return &Row{
			heapTuple: C.get_heap_tuple(C.SPI_tuptable.vals, C.uint(0)),
			tupleDesc: C.SPI_tuptable.tupdesc,
		}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("QueryRow failed: %s", C.GoString(C.SPI_result_code_string(C.SPI_result))))
	}
}

//Exec executes a prepared query Plan with no result
func (plan *Plan) Exec(args ...interface{}) error {
	valuesP, nullsP := spiArgs(args)
	plan.db.lock.Lock()
	defer plan.db.lock.Unlock()
	rv := C.SPI_execute_plan(plan.spi_plan, valuesP, nullsP, C.false, 0)
	if rv >= C.int(0) && C.SPI_processed == 1 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("Exec failed: %s", C.GoString(C.SPI_result_code_string(C.SPI_result))))
	}
}

func spiArgs(args ...interface{}) (valuesP *C.Datum, nullsP *C.char) {
	if len(args) > 0 {
		values := make([]Datum, len(args))
		nulls := make([]C.char, len(args))
		for i, arg := range args {
			values[i] = ToDatum(arg)
			nulls[i] = C.char(' ')
		}
		valuesP = (*C.Datum)(unsafe.Pointer(&values[0]))
		nullsP = &nulls[0]
	}
	return valuesP, nullsP
}

//Rows represents the result of running a prepared Plan with Query
type Rows struct {
	heapTuples []C.HeapTuple
	tupleDesc  C.TupleDesc
	processed  uint32
	current    C.HeapTuple
}

func newRows(heapTuples *C.HeapTuple, tupleDesc C.TupleDesc, processed C.uint32) *Rows {
	rows := &Rows{
		tupleDesc: tupleDesc,
		processed: uint32(processed),
	}
	rows.heapTuples = make([]C.HeapTuple, rows.processed)
	for i := 0; i < int(rows.processed); i++ {
		rows.heapTuples[i] = C.get_heap_tuple(heapTuples, C.uint(i))
	}
	return rows
}

//Next sets the Rows to another row, returs false if there isn't another
//must be first called to set the Rows to the first row
func (rows *Rows) Next() bool {
	if len(rows.heapTuples) == 0 {
		return false
	}
	rows.current = rows.heapTuples[0]
	rows.heapTuples = rows.heapTuples[1:]
	return true
}

//Scan takes pointers to variables that will be filled with the values of the current row
func (rows *Rows) Scan(args ...interface{}) error {
	for i, arg := range args {
		val := C.get_col_as_datum(rows.current, rows.tupleDesc, C.int(i))
		err := scanVal(C.SPI_gettypeid(rows.tupleDesc, C.int(i+1)), val, arg)
		if err != nil {
			return err
		}
	}
	return nil
}

type Row struct {
	tupleDesc C.TupleDesc
	heapTuple C.HeapTuple
}

func (row *Row) Scan(args ...interface{}) error {
	for i, arg := range args {
		val := C.get_col_as_datum(row.heapTuple, row.tupleDesc, C.int(i))
		err := scanVal(C.SPI_gettypeid(row.tupleDesc, C.int(i+1)), val, arg)
		if err != nil {
			return err
		}
	}
	return nil
}

func scanVal(oid C.Oid, val C.Datum, arg interface{}) error {
	switch targ := arg.(type) { //TODO error by converting
	case *string:
		if oid != C.TEXTOID {
			return errors.New(fmt.Sprintf("Column type is not text %s", oid))
		}
		*targ = C.GoString(C.datum_to_cstring(val))
	case *int16:
		if oid != C.INT2OID {
			return errors.New(fmt.Sprintf("Column type is not int16 %s", oid))
		}
		*targ = int16(C.datum_to_int16(val))
	case *uint16:
		if oid != C.INT2OID {
			return errors.New(fmt.Sprintf("Column type is not uint16 %s", oid))
		}
		*targ = uint16(C.datum_to_uint16(val))
	case *int32:
		if oid != C.INT4OID {
			return errors.New(fmt.Sprintf("Column type is not int32 %s", oid))
		}
		*targ = int32(C.datum_to_int32(val))
	case *uint32:
		if oid != C.INT4OID {
			return errors.New(fmt.Sprintf("Column type is not uint32 %s", oid))
		}
		*targ = uint32(C.datum_to_uint32(val))
	case *int64:
		if oid != C.INT8OID {
			return errors.New(fmt.Sprintf("Column type is not int64 %s", oid))
		}
		*targ = int64(C.datum_to_int64(val))
	case *int:
		if oid != C.INT2OID && oid != C.INT4OID && oid != C.INT8OID {
			return errors.New(fmt.Sprintf("Column type is not int %s", oid))
		}
		*targ = int(C.datum_to_int64(val))
	case *uint:
		if oid != C.INT2OID && oid != C.INT4OID && oid != C.INT8OID {
			return errors.New(fmt.Sprintf("Column type is not uint64 %s", oid))
		}
		*targ = uint(C.datum_to_uint32(val))
	case *bool:
		if oid != C.BOOLOID {
			return errors.New(fmt.Sprintf("Column type is not bool %s", oid))
		}
		*targ = C.datum_to_bool(val) == C.true
	case *time.Time:
		switch oid {
		case C.DATEOID:
			dateadt := C.datum_to_date(val)
			*targ = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(dateadt))
		case C.TIMESTAMPOID:
			t := C.datum_to_time(val)
			*targ = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Second * time.Duration(int64(t)/int64(C.USECS_PER_SEC)))
		case C.TIMESTAMPTZOID:
			t := C.datum_to_timetz(val)
			*targ = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Second * time.Duration(int64(t)/int64(C.USECS_PER_SEC))).Local()
		default:
			return errors.New(fmt.Sprintf("Unsupported time type %s", oid))
		}
	default:
		return errors.New(fmt.Sprintf("Unsupported type in Scan (%s) %s", reflect.TypeOf(arg).String(), oid))
	}
	return nil
}
