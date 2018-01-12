package main

/*
#cgo CFLAGS: -I/usr/include/postgresql/server
#cgo LDFLAGS: -shared

#include "postgres.h"
#include "fmgr.h"
#include "storage/latch.h"
#include "storage/proc.h"
#include "postmaster/bgworker.h"

int get_got_sigterm();
*/
import "C"
import (
	"time"
)

//export BackgroundWorkerMain
func BackgroundWorkerMain() C.int {
	Log("Starting GoBackgroundWorker")

	for C.get_got_sigterm() == 0 {

		Log("sleep")
		if err := WaitLatch(time.Second); err != nil {
			Log(err.Error())
			return 1
		}
		Log("end")

		Log("Hello World from GoLang!!! Yeahhhhh!!!!" + time.Now().Format(time.UnixDate))
	}

	Log("Finishing GoBackgroundWorker")
	return 0
}
