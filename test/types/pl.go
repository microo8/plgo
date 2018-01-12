package main

/*
#cgo CFLAGS: -I/usr/include/postgresql/server
#cgo LDFLAGS: -shared

#include "postgres.h"
#include "fmgr.h"
#include "pgtime.h"
#include "access/htup_details.h"
#include "catalog/pg_type.h"
#include "utils/builtins.h"
#include "utils/date.h"
#include "utils/timestamp.h"
#include "utils/array.h"
#include "utils/elog.h"
#include "executor/spi.h"
#include "parser/parse_type.h"
#include "commands/trigger.h"
#include "utils/rel.h"
#include "utils/lsyscache.h"

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

Datum get_arg(PG_FUNCTION_ARGS, uint i) {
	return PG_GETARG_DATUM(i);
}

HeapTuple get_heap_tuple(HeapTuple* ht, uint i) {
    return ht[i];
}

Datum get_col_as_datum(HeapTuple ht, TupleDesc td, int colnumber) {
    bool isNull;
    Datum ret = SPI_getbinval(ht, td, colnumber + 1, &isNull);
	if (isNull) PG_RETURN_VOID();
	return ret;
}

bool called_as_trigger(PG_FUNCTION_ARGS) {
	return CALLED_AS_TRIGGER(fcinfo);
}

Datum get_heap_getattr(HeapTuple ht, uint i, TupleDesc td) {
	bool isNull;
	Datum ret = heap_getattr(ht, i, td, &isNull);
	if (isNull) PG_RETURN_VOID();
	return ret;
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

Datum float4_to_datum(float val) {
	return Float4GetDatum(val);
}

Datum float8_to_datum(double val) {
	return Float8GetDatum(val);
}

Datum heap_tuple_to_datum(HeapTuple val) {
	return PointerGetDatum(val);
}

Datum pointer_get_datum(void* val) {
	return PointerGetDatum(val);
}

Datum array_to_datum(Oid element_type, Datum* vals, int size) {
	ArrayType *result;
	bool* isnull = (bool *)palloc0(sizeof(bool)*size);
    int dims[1];
    int lbs[1];
    int16 typlen;
    bool typbyval;
    char typalign;

	dims[0] = size;
	lbs[0] = 1;

    // get required info about the element type
    get_typlenbyvalalign(element_type, &typlen, &typbyval, &typalign);
	result = construct_md_array(vals, isnull, 1, dims, lbs,
                                element_type, typlen, typbyval, typalign);

    PG_RETURN_ARRAYTYPE_P(result);
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

float datum_to_float4(Datum val) {
	return DatumGetFloat4(val);
}

double datum_to_float8(Datum val) {
	return DatumGetFloat8(val);
}

HeapTuple datum_to_heap_tuple(Datum val) {
	return (HeapTuple) DatumGetPointer(val);
}

void* datum_to_pointer(Datum val) {
	return (void*) DatumGetPointer(val);
}


Datum* datum_to_array(Datum val, int* nelemsp) {
	ArrayType* array = DatumGetArrayTypeP(val);

    int16 typlen;
    bool typbyval;
    char typalign;
	Datum *result;
	bool *nullsp;

    get_typlenbyvalalign(ARR_ELEMTYPE(array), &typlen, &typbyval, &typalign);

	deconstruct_array(array, ARR_ELEMTYPE(array),
                      typlen, typbyval, typalign,
                      &result, &nullsp, nelemsp);
	return result;
}

char* unknown_to_char(Datum val) {
	return (char*)val;
}

//TriggerData functions/////////////////////////////////////////////
bool trigger_fired_before(TriggerEvent tg_event) {
	return TRIGGER_FIRED_BEFORE(tg_event);
}

bool trigger_fired_after(TriggerEvent tg_event) {
	return TRIGGER_FIRED_AFTER(tg_event);
}

bool trigger_fired_instead(TriggerEvent tg_event) {
	return TRIGGER_FIRED_INSTEAD(tg_event);
}

bool trigger_fired_for_row(TriggerEvent tg_event) {
	return TRIGGER_FIRED_FOR_ROW(tg_event);
}

bool trigger_fired_for_statement(TriggerEvent tg_event) {
	return TRIGGER_FIRED_FOR_STATEMENT(tg_event);
}

bool trigger_fired_by_insert(TriggerEvent tg_event) {
	return TRIGGER_FIRED_BY_INSERT(tg_event);
}

bool trigger_fired_by_update(TriggerEvent tg_event) {
	return TRIGGER_FIRED_BY_UPDATE(tg_event);
}

bool trigger_fired_by_delete(TriggerEvent tg_event) {
	return TRIGGER_FIRED_BY_DELETE(tg_event);
}

bool trigger_fired_by_truncate(TriggerEvent tg_event) {
	return TRIGGER_FIRED_BY_TRUNCATE(tg_event);
}

PG_FUNCTION_INFO_V1(ComplexIn);
PG_FUNCTION_INFO_V1(ComplexOut);
*/
import "C"
import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"
	"unsafe"
)

//TODO check all public things

//this has to be here
func main() {}

//Datum is the return type of postgresql
type Datum C.Datum

//DB represents the db connection, can be made only once
type DB struct{}

//Open returns DB connection and runs SPI_connect
func Open() (*DB, error) {
	if C.SPI_connect() != C.SPI_OK_CONNECT {
		return nil, errors.New("can't connect")
	}
	return new(DB), nil
}

//Close closes the DB connection
func (db *DB) Close() error {
	if C.SPI_finish() != C.SPI_OK_FINISH {
		return errors.New("Error closing DB")
	}
	return nil
}

//elogLevel Log level enum
type elogLevel int

//elogLevel constants
const (
	noticeLevel elogLevel = iota
	errorLevel
)

//elog represents the elog io.Writter to use with Logger
type elog struct {
	Level elogLevel
}

//Write is an notify implemented as io.Writter
func (e *elog) Write(p []byte) (n int, err error) {
	switch e.Level {
	case noticeLevel:
		C.elog_notice(C.CString(string(p)))
	case errorLevel:
		C.elog_error(C.CString(string(p)))
	}
	return len(p), nil
}

//NewNoticeLogger creates an logger that writes into NOTICE elog
func NewNoticeLogger(prefix string, flag int) *log.Logger {
	return log.New(&elog{Level: noticeLevel}, prefix, flag)
}

//NewErrorLogger creates an logger that writes into ERROR elog
func NewErrorLogger(prefix string, flag int) *log.Logger {
	return log.New(&elog{Level: errorLevel}, prefix, flag)
}

//funcInfo is the type of parameters that all functions get
type funcInfo C.FunctionCallInfoData

//CalledAsTrigger checks if the function is called as trigger
func (fcinfo *funcInfo) CalledAsTrigger() bool {
	return C.called_as_trigger((*C.struct_FunctionCallInfoData)(unsafe.Pointer(fcinfo))) == C.true
}

//TODO Scan must return argument also if the function is called as trigger

//Scan sets the args to the function parameter values (converted from PostgreSQL types to Go types)
func (fcinfo *funcInfo) Scan(args ...interface{}) error {
	for i, arg := range args {
		funcArg := C.get_arg((*C.struct_FunctionCallInfoData)(unsafe.Pointer(fcinfo)), C.uint(i))
		argOid := C.get_call_expr_argtype(fcinfo.flinfo.fn_expr, C.int(i))
		err := scanVal(argOid, "", funcArg, arg)
		if err != nil {
			return err
		}
	}
	return nil
}

//TriggerData returns Trigger data, if the function was called as trigger, else nil
func (fcinfo *funcInfo) TriggerData() *TriggerData {
	if !fcinfo.CalledAsTrigger() {
		return nil
	}
	trigdata := (*C.TriggerData)(unsafe.Pointer(fcinfo.context))
	return &TriggerData{
		tgEvent:    trigdata.tg_event,
		tgRelation: trigdata.tg_relation,
		tgTrigger:  trigdata.tg_trigger,
		OldRow:     newTriggerRow(trigdata.tg_relation.rd_att, trigdata.tg_trigtuple),
		NewRow:     newTriggerRow(trigdata.tg_relation.rd_att, trigdata.tg_newtuple),
	}
}

//TriggerData represents the data passed by the trigger manager
type TriggerData struct {
	tgEvent    C.TriggerEvent
	tgRelation C.Relation
	tgTrigger  *C.Trigger
	OldRow     *TriggerRow
	NewRow     *TriggerRow
}

//FiredBefore returns true if the trigger fired before the operation.
func (td *TriggerData) FiredBefore() bool {
	return C.trigger_fired_before(td.tgEvent) == C.true
}

//FiredAfter returns true if the trigger fired after the operation.
func (td *TriggerData) FiredAfter() bool {
	return C.trigger_fired_after(td.tgEvent) == C.true
}

//FiredInstead returns true if the trigger fired instead of the operation.
func (td *TriggerData) FiredInstead() bool {
	return C.trigger_fired_instead(td.tgEvent) == C.true
}

//FiredForRow returns true if the trigger fired for a row-level event.
func (td *TriggerData) FiredForRow() bool {
	return C.trigger_fired_for_row(td.tgEvent) == C.true
}

//FiredForStatement returns true if the trigger fired for a statement-level event.
func (td *TriggerData) FiredForStatement() bool {
	return C.trigger_fired_for_statement(td.tgEvent) == C.true
}

//FiredByInsert returns true if the trigger was fired by an INSERT command.
func (td *TriggerData) FiredByInsert() bool {
	return C.trigger_fired_by_insert(td.tgEvent) == C.true
}

//FiredByUpdate returns true if the trigger was fired by an UPDATE command.
func (td *TriggerData) FiredByUpdate() bool {
	return C.trigger_fired_by_update(td.tgEvent) == C.true
}

//FiredByDelete returns true if the trigger was fired by a DELETE command.
func (td *TriggerData) FiredByDelete() bool {
	return C.trigger_fired_by_delete(td.tgEvent) == C.true
}

//FiredByTruncate returns true if the trigger was fired by a TRUNCATE command.
func (td *TriggerData) FiredByTruncate() bool {
	return C.trigger_fired_by_truncate(td.tgEvent) == C.true
}

//TriggerRow is used in TriggerData as NewRow and OldRow
type TriggerRow struct {
	tupleDesc C.TupleDesc
	attrs     []C.Datum
}

func newTriggerRow(tupleDesc C.TupleDesc, heapTuple C.HeapTuple) *TriggerRow {
	row := &TriggerRow{tupleDesc, make([]C.Datum, int(tupleDesc.natts))}
	for i := 0; i < int(tupleDesc.natts); i++ {
		row.attrs[i] = C.get_heap_getattr(heapTuple, C.uint(i+1), tupleDesc)
	}
	return row
}

//Scan sets the args from the TriggerRow
func (row *TriggerRow) Scan(args ...interface{}) error {
	for i, arg := range args {
		oid := C.SPI_gettypeid(row.tupleDesc, C.int(i+1))
		typeName := C.SPI_gettype(row.tupleDesc, C.int(i+1))
		err := scanVal(oid, C.GoString(typeName), row.attrs[i], arg)
		if err != nil {
			return err
		}
	}
	return nil
}

//Set sets the i'th value in the row
func (row *TriggerRow) Set(i int, val interface{}) {
	row.attrs[i] = (C.Datum)(toDatum(val))
}

func makeArray(elemtype C.Oid, arg interface{}) Datum {
	s := reflect.ValueOf(arg)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	datums := make([]C.Datum, s.Len())
	for i := 0; i < s.Len(); i++ {
		datums[i] = (C.Datum)(toDatum(s.Index(i).Interface()))
	}
	return (Datum)(C.array_to_datum(elemtype, &datums[0], C.int(s.Len())))
}

func makeSlice(val C.Datum) []C.Datum {
	var clength C.int
	datumArray := C.datum_to_array(val, &clength)
	length := int(clength)
	slice := (*[1 << 30]C.Datum)(unsafe.Pointer(datumArray))[:length]
	return slice
}

//toDatum returns the Postgresql C type from Golang type
func toDatum(val interface{}) Datum {
	switch v := val.(type) {
	case unsafe.Pointer:
		return (Datum)(C.pointer_get_datum(v))
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
	case float32:
		return (Datum)(C.float4_to_datum(C.float(v)))
	case float64:
		return (Datum)(C.float8_to_datum(C.double(v)))
	case time.Time:
		return (Datum)(C.timetz_to_datum(C.TimestampTz((v.UTC().Unix() - 946684800) * int64(C.USECS_PER_SEC))))
	case bool:
		if v {
			return (Datum)(C.bool_to_datum(C.true))
		}
		return (Datum)(C.bool_to_datum(C.false))
	case []string:
		return makeArray(C.TEXTOID, v)
	case []int16:
		return makeArray(C.INT2OID, v)
	case []uint16:
		return makeArray(C.INT2OID, v)
	case []int32:
		return makeArray(C.INT4OID, v)
	case []uint32:
		return makeArray(C.INT4OID, v)
	case []int64:
		return makeArray(C.INT8OID, v)
	case []int:
		return makeArray(C.INT8OID, v)
	case []uint:
		return makeArray(C.INT8OID, v)
	case []float32:
		return makeArray(C.FLOAT4OID, v)
	case []float64:
		return makeArray(C.FLOAT8OID, v)
	case []bool:
		return makeArray(C.BOOLOID, v)
	case []time.Time:
		return makeArray(C.TIMESTAMPTZOID, v)
	case *TriggerRow:
		isNull := make([]C.bool, len(v.attrs))
		for i, attr := range v.attrs {
			if attr == (C.Datum)(toDatum(nil)) {
				isNull[i] = C.true
			} else {
				isNull[i] = C.false
			}
		}
		heapTuple := C.heap_form_tuple(v.tupleDesc, &v.attrs[0], &isNull[0])
		return (Datum)(C.heap_tuple_to_datum(heapTuple))
	default:
		return (Datum)(C.void_datum())
	}
}

//Stmt represents the prepared SQL statement
type Stmt struct {
	spiPlan C.SPIPlanPtr
	db      *DB
}

//Prepare prepares an SQL query and returns a Stmt that can be executed
//query - the SQL query
//types - an array of strings with type names from postgresql of the prepared query
func (db *DB) Prepare(query string, types []string) (*Stmt, error) {
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
		return &Stmt{spiPlan: cplan, db: db}, nil
	}
	return nil, fmt.Errorf("Prepare failed: %s", C.GoString(C.SPI_result_code_string(C.SPI_result)))
}

//Query executes the prepared Stmt with the provided args and returns
//multiple Rows result, that can be iterated
func (stmt *Stmt) Query(args ...interface{}) (*Rows, error) {
	valuesP, nullsP := spiArgs(args)
	rv := C.SPI_execute_plan(stmt.spiPlan, valuesP, nullsP, C.true, 0)
	if rv == C.SPI_OK_SELECT && C.SPI_processed > 0 {
		return newRows(C.SPI_tuptable.vals, C.SPI_tuptable.tupdesc, C.SPI_processed), nil
	}
	return nil, fmt.Errorf("Query failed: %s", C.GoString(C.SPI_result_code_string(C.SPI_result)))
}

//QueryRow executes the prepared Stmt with the provided args and returns one row result
func (stmt *Stmt) QueryRow(args ...interface{}) (*Row, error) {
	valuesP, nullsP := spiArgs(args)
	rv := C.SPI_execute_plan(stmt.spiPlan, valuesP, nullsP, C.false, 1)
	if rv >= C.int(0) && C.SPI_processed == 1 {
		return &Row{
			heapTuple: C.get_heap_tuple(C.SPI_tuptable.vals, C.uint(0)),
			tupleDesc: C.SPI_tuptable.tupdesc,
		}, nil
	}
	return nil, fmt.Errorf("QueryRow failed: %s", C.GoString(C.SPI_result_code_string(C.SPI_result)))
}

//Exec executes a prepared query Stmt with no result
func (stmt *Stmt) Exec(args ...interface{}) error {
	valuesP, nullsP := spiArgs(args)
	rv := C.SPI_execute_plan(stmt.spiPlan, valuesP, nullsP, C.false, 0)
	if rv >= C.int(0) && C.SPI_processed == 1 {
		return nil
	}
	return fmt.Errorf("Exec failed: %s", C.GoString(C.SPI_result_code_string(C.SPI_result)))
}

func spiArgs(args []interface{}) (valuesP *C.Datum, nullsP *C.char) {
	if len(args) > 0 {
		values := make([]Datum, len(args))
		nulls := make([]C.char, len(args))
		for i, arg := range args {
			values[i] = toDatum(arg)
			nulls[i] = C.char(' ')
		}
		valuesP = (*C.Datum)(unsafe.Pointer(&values[0]))
		nullsP = &nulls[0]
	}
	return valuesP, nullsP
}

//Rows represents the result of running a prepared Stmt with Query
type Rows struct {
	heapTuples []C.HeapTuple
	tupleDesc  C.TupleDesc
	processed  uint32
	current    C.HeapTuple
}

func newRows(heapTuples *C.HeapTuple, tupleDesc C.TupleDesc, processed C.uint64) *Rows {
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
		oid := C.SPI_gettypeid(rows.tupleDesc, C.int(i+1))
		typeName := C.SPI_gettype(rows.tupleDesc, C.int(i+1))
		err := scanVal(oid, C.GoString(typeName), val, arg)
		if err != nil {
			return err
		}
	}
	return nil
}

//Columns returns the names of columns
func (rows *Rows) Columns() ([]string, error) {
	var columns []string
	for i := 1; ; i++ {
		fname := C.SPI_fname(rows.tupleDesc, C.int(i))

		if fname == nil {
			break
		}
		if C.SPI_result == C.SPI_ERROR_NOATTRIBUTE {
			return nil, fmt.Errorf("Error getting column names")
		}
		columns = append(columns, C.GoString(fname))
	}
	return columns, nil
}

//Row represents a single row from running a query
type Row struct {
	tupleDesc C.TupleDesc
	heapTuple C.HeapTuple
}

//Scan scans the args from Row
func (row *Row) Scan(args ...interface{}) error {
	for i, arg := range args {
		val := C.get_col_as_datum(row.heapTuple, row.tupleDesc, C.int(i))
		oid := C.SPI_gettypeid(row.tupleDesc, C.int(i+1))
		typeName := C.SPI_gettype(row.tupleDesc, C.int(i+1))
		err := scanVal(oid, C.GoString(typeName), val, arg)
		if err != nil {
			return err
		}
	}
	return nil
}

func scanVal(oid C.Oid, typeName string, val C.Datum, arg interface{}) error {
	switch targ := arg.(type) {
	case unsafe.Pointer:
		targ = C.datum_to_pointer(val)
	case *string:
		switch oid {
		case C.TEXTOID:
			*targ = C.GoString(C.datum_to_cstring(val))
		case C.UNKNOWNOID:
			*targ = C.GoString(C.unknown_to_char(val))
		default:
			return fmt.Errorf("Column type is not text %s", typeName)
		}
	case *int16:
		switch oid {
		case C.INT2OID:
			*targ = int16(C.datum_to_int16(val))
		default:
			return fmt.Errorf("Column type is not int16 %s", typeName)
		}
	case *uint16:
		switch oid {
		case C.INT2OID:
			*targ = uint16(C.datum_to_uint16(val))
		default:
			return fmt.Errorf("Column type is not uint16 %s", typeName)
		}
	case *int32:
		switch oid {
		case C.INT4OID:
			*targ = int32(C.datum_to_int32(val))
		default:
			return fmt.Errorf("Column type is not int32 %s", typeName)
		}
	case *uint32:
		switch oid {
		case C.INT4OID:
			*targ = uint32(C.datum_to_uint32(val))
		default:
			return fmt.Errorf("Column type is not uint32 %s", typeName)
		}
	case *int64:
		switch oid {
		case C.INT8OID:
			*targ = int64(C.datum_to_int64(val))
		default:
			return fmt.Errorf("Column type is not int64 %s", typeName)
		}
	case *int:
		switch oid {
		case C.INT2OID:
			*targ = int(C.datum_to_int16(val))
		case C.INT4OID:
			*targ = int(C.datum_to_int32(val))
		case C.INT8OID:
			*targ = int(C.datum_to_int64(val))
		default:
			return fmt.Errorf("Column type is not int %s", typeName)
		}
	case *uint:
		switch oid {
		case C.INT2OID:
			*targ = uint(C.datum_to_int16(val))
		case C.INT4OID:
			*targ = uint(C.datum_to_int32(val))
		case C.INT8OID:
			*targ = uint(C.datum_to_int64(val))
		default:
			return fmt.Errorf("Column type is not uint %s", typeName)
		}
	case *bool:
		switch oid {
		case C.BOOLOID:
			*targ = C.datum_to_bool(val) == C.true
		default:
			return fmt.Errorf("Column type is not bool %s", typeName)
		}
	case *float32:
		switch oid {
		case C.FLOAT4OID:
			*targ = float32(C.datum_to_float4(val))
		default:
			return fmt.Errorf("Column type is not real %s", typeName)
		}
	case *float64:
		switch oid {
		case C.FLOAT8OID:
			*targ = float64(C.datum_to_float8(val))
		default:
			return fmt.Errorf("Column type is not double precision %s", typeName)
		}
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
			return fmt.Errorf("Unsupported time type %s", typeName)
		}
	case *[]string:
		slice := makeSlice(val)
		*targ = make([]string, len(slice))
		for i := range slice {
			err := scanVal(C.TEXTOID, "Text", slice[i], &((*targ)[i]))
			if err != nil {
				return err
			}
		}
	case *[]int16:
		slice := makeSlice(val)
		*targ = make([]int16, len(slice))
		for i := range slice {
			err := scanVal(C.INT2OID, "SMALLINT", slice[i], &((*targ)[i]))
			if err != nil {
				return err
			}
		}
	case *[]uint16:
		slice := makeSlice(val)
		*targ = make([]uint16, len(slice))
		for i := range slice {
			err := scanVal(C.INT2OID, "SMALLINT", slice[i], &((*targ)[i]))
			if err != nil {
				return err
			}
		}
	case *[]int32:
		slice := makeSlice(val)
		*targ = make([]int32, len(slice))
		for i := range slice {
			err := scanVal(C.INT4OID, "INT", slice[i], &((*targ)[i]))
			if err != nil {
				return err
			}
		}
	case *[]uint32:
		slice := makeSlice(val)
		*targ = make([]uint32, len(slice))
		for i := range slice {
			err := scanVal(C.INT4OID, "INT", slice[i], &((*targ)[i]))
			if err != nil {
				return err
			}
		}
	case *[]int64:
		slice := makeSlice(val)
		*targ = make([]int64, len(slice))
		for i := range slice {
			err := scanVal(C.INT8OID, "INT", slice[i], &((*targ)[i]))
			if err != nil {
				return err
			}
		}
	case *[]int:
		slice := makeSlice(val)
		*targ = make([]int, len(slice))
		for i := range slice {
			err := scanVal(C.INT8OID, "INT", slice[i], &((*targ)[i]))
			if err != nil {
				return err
			}
		}
	case *[]bool:
		slice := makeSlice(val)
		*targ = make([]bool, len(slice))
		for i := range slice {
			err := scanVal(C.BOOLOID, "BOOL", slice[i], &((*targ)[i]))
			if err != nil {
				return err
			}
		}
	case *[]float32:
		slice := makeSlice(val)
		*targ = make([]float32, len(slice))
		for i := range slice {
			err := scanVal(C.FLOAT4OID, "REAL", slice[i], &((*targ)[i]))
			if err != nil {
				return err
			}
		}
	case *[]float64:
		slice := makeSlice(val)
		*targ = make([]float64, len(slice))
		for i := range slice {
			err := scanVal(C.FLOAT8OID, "DOUBLE", slice[i], &((*targ)[i]))
			if err != nil {
				return err
			}
		}
	case *[]time.Time:
		slice := makeSlice(val)
		*targ = make([]time.Time, len(slice))
		for i := range slice {
			err := scanVal(C.TIMESTAMPTZOID, "TIMETZ", slice[i], &((*targ)[i]))
			if err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("Unsupported type in Scan (%s) %s", reflect.TypeOf(arg).String(), typeName)
	}
	return nil
}
