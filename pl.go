package main

/*
#cgo CFLAGS: -Wall -Wpointer-arith -Wno-declaration-after-statement -Wendif-labels -Wmissing-format-attribute -Wformat-security -fno-strict-aliasing -fwrapv -fexcess-precision=standard -march=x86-64 -mtune=generic -O2 -pipe -fstack-protector-strong -fpic -I. -I./ -I/usr/include/postgresql/server -I/usr/include/postgresql/internal -D_FORTIFY_SOURCE=2 -D_GNU_SOURCE -I/usr/include/libxml2
#cgo LDFLAGS: -Wall -Wmissing-prototypes -Wpointer-arith -Wdeclaration-after-statement -Wendif-labels -Wmissing-format-attribute -Wformat-security -fno-strict-aliasing -fwrapv -fexcess-precision=standard -march=x86-64 -mtune=generic -O2 -pipe -fstack-protector-strong -fpic -L/usr/lib -Wl,-O1,--sort-common,--as-needed,-z,relro  -Wl,--as-needed -Wl,-rpath,'/usr/lib',--enable-new-dtags -shared

#include "plgo.h"

PG_FUNCTION_INFO_V1(plgo_example);
*/
import "C"
import "unsafe"

func main() {}

type Datum *C.Datum
type FuncInfo C.FunctionCallInfoData
type Text *C.text
type Bytea *C.bytea

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

//PGVal returns the Postgresql C type from Golang type (currently implements just stringtotext)
func PGVal(val interface{}) Datum {
	var size uintptr
	switch v := val.(type) {
	case string:
		return (Datum)(unsafe.Pointer(C.cstring_to_text(C.CString(v))))
	default:
		size = unsafe.Sizeof(val)
		psize := (*C.uint64)(unsafe.Pointer(size))
		pval := val.(*C.void)
		return (Datum)(C.ret(pval, psize))
	}
}
