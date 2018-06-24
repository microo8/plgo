package main

/*
#cgo CFLAGS: -I/usr/include/postgresql/server
#cgo LDFLAGS: -shared

#include "postgres.h"
#include "pgstat.h"
#include "postmaster/bgworker.h"
#include "storage/ipc.h"
#include "storage/latch.h"
#include "storage/proc.h"
#include "miscadmin.h"
#include "utils/elog.h"
#include "fmgr.h"

#define waitFlags WL_LATCH_SET | WL_TIMEOUT | WL_POSTMASTER_DEATH

PG_MODULE_MAGIC;

void _PG_init(void);
extern int backgroundWorkerMain();

void elog_log(char *string) {
	elog(LOG, string, "");
}

int wait_latch(long miliseconds) {
#if (PG_VERSION_NUM >= 100000)
		return WaitLatch(MyLatch, waitFlags, miliseconds, PG_WAIT_EXTENSION) & WL_POSTMASTER_DEATH;
#else
		return WaitLatch(&MyProc->procLatch, waitFlags, miliseconds) & WL_POSTMASTER_DEATH;
#endif
}

void reset_latch(void) {
#if (PG_VERSION_NUM >= 100000)
	ResetLatch(MyLatch);
#else
	ResetLatch(&MyProc->procLatch);
#endif
}

static volatile sig_atomic_t got_sigterm = false;
static volatile sig_atomic_t got_sighup = false;

int get_got_sigterm() {
	return (got_sigterm == true);
}

void background_main(Datum main_arg) pg_attribute_noreturn();

static void background_sigterm(SIGNAL_ARGS)
{
	int save_errno = errno;
	got_sigterm = true;
#if (PG_VERSION_NUM >= 100000)
	SetLatch(MyLatch);
#else
	if (MyProc != NULL)	SetLatch(&MyProc->procLatch);
#endif
	errno = save_errno;
}

static void background_sighup(SIGNAL_ARGS) {
	int	save_errno = errno;
	got_sighup = true;
#if PG_VERSION_NUM >= 100000
	SetLatch(MyLatch);
#else
	if (MyProc)
		SetLatch(&MyProc->procLatch);
#endif
	errno = save_errno;
}

void background_main(Datum main_arg) {
	pqsignal(SIGHUP, background_sighup);
	//pqsignal(SIGINT, SIG_IGN);
	pqsignal(SIGTERM, background_sigterm);
	BackgroundWorkerUnblockSignals();
	proc_exit(backgroundWorkerMain());
}


void _PG_init(void) {
	BackgroundWorker worker;
	MemSet(&worker, 0, sizeof(BackgroundWorker));
	worker.bgw_flags = BGWORKER_SHMEM_ACCESS;
	worker.bgw_start_time = BgWorkerStart_RecoveryFinished;
	snprintf(worker.bgw_name, BGW_MAXLEN, "GoBackgroundWorker");
#if PG_VERSION_NUM >= 100000
	sprintf(worker.bgw_library_name, "bgw");
	sprintf(worker.bgw_function_name, "backgroundWorkerMain");
#else
	worker.bgw_main = bg_mon_main;
#endif
	worker.bgw_restart_time = BGW_NEVER_RESTART;
	worker.bgw_main_arg = (Datum) 0;
#if PG_VERSION_NUM >= 90400
	worker.bgw_notify_pid = 0;
#endif
	RegisterBackgroundWorker(&worker);
}
*/
import "C"
import (
	"errors"
	"time"
)

func main() {}

//Log will log with the elog function
func Log(text string) {
	C.elog_log(C.CString(text))
}

//WaitLatch ...
func WaitLatch(d time.Duration) error {
	if rc := C.wait_latch(C.long(d.Nanoseconds() / 1000000)); rc != 0 {
		return errors.New("postmaster is dead")
	}
	C.reset_latch()
	return nil
}

//Tick ...
func Tick(d time.Duration) <-chan bool {
	ch := make(chan bool)
	Log("0")
	go func() {
		Log("1")
		for C.get_got_sigterm() == 0 {
			Log("2")
			if rc := C.wait_latch(C.long(d.Nanoseconds() / 1000000)); rc != 0 {
				Log("3")
				ch <- false
				return
			}
			Log("4")
			C.reset_latch()
			Log("5")
			ch <- true
			Log("6")
		}
		Log("7")
	}()
	Log("8")
	return ch
}
