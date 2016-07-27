/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

#include "postgres.h"
#include "fmgr.h"
#include "utils/builtins.h"

#ifdef PG_MODULE_MAGIC
PG_MODULE_MAGIC;
#endif

//the return value must be allocated trough palloc
void* ret(void *val, uint64 *size) {
    void *retDatum = palloc(*size);
    memcpy(retDatum, val, *size);
    return retDatum;
}

int varsize(void *var) {
    return VARSIZE(var);
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
