#include "postgres.h"
#include "fmgr.h"
#include "utils/builtins.h"
#include "utils/elog.h"
#include "executor/spi.h"
#include "parser/parse_type.h"

#ifdef PG_MODULE_MAGIC
PG_MODULE_MAGIC;
#endif

typedef struct _SPI_plan SPIPlan;

HeapTuple get_heap_tuble(HeapTuple *ht, uint i) {
        return ht[i];
}

int varsize(void *var) {
        return VARSIZE(var);
}

void notice(char* string) {
        elog(NOTICE, string, "");
}

Datum get_col_as_datum(HeapTuple* ht, TupleDesc td, uint32 rownumber, int colnumber) {
        bool isNull = true;
        return SPI_getbinval(ht[rownumber], td, colnumber + 1, &isNull);
}

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
