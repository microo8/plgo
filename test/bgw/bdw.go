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
extern int BackgroundWorkerMain();

void elog_log(char *string) {
	elog(LOG, string, "");
}

int wait_latch(long miliseconds) {
	char text[1000];
	sprintf(text, "start waiting for max %ld miliseconds", miliseconds);
	elog_log(text);
#if (PG_VERSION_NUM >= 100000)
		return WaitLatch(MyLatch, waitFlags, miliseconds, PG_WAIT_EXTENSION) & WL_POSTMASTER_DEATH;
#else
		return WaitLatch(MyLatch, waitFlags, miliseconds) & WL_POSTMASTER_DEATH;
#endif
}

void reset_latch(void) {
	ResetLatch(MyLatch);
}

static volatile sig_atomic_t got_sigterm = false;

int get_got_sigterm() {
	return (got_sigterm == true);
}

void background_main(Datum main_arg) pg_attribute_noreturn();

static void background_sigterm(SIGNAL_ARGS)
{
	int save_errno = errno;
	got_sigterm = true;
	if (MyProc != NULL) {
		SetLatch(&MyProc->procLatch);
	}
	errno = save_errno;
}

static void background_sighup(SIGNAL_ARGS) {
	//CronJobCacheValid = false;

	if (MyProc != NULL)
	{
		SetLatch(&MyProc->procLatch);
	}
}

void background_main(Datum main_arg) {
	pqsignal(SIGHUP, background_sighup);
	pqsignal(SIGINT, SIG_IGN);
	pqsignal(SIGTERM, background_sigterm);
	BackgroundWorkerUnblockSignals();

	int rc;

	//if (BackgroundWorkerMain()) {
	while (!got_sigterm) {
		rc = WaitLatch(MyLatch, waitFlags, 1000L);
		ResetLatch(MyLatch);

		if (rc & WL_POSTMASTER_DEATH) {
			elog_log("bgw exited, postmaster is dead");
			proc_exit(1);
		}
		reset_latch();
		elog_log("Hello from C!!!!!!!!!!!!!");
	}

	elog_log("quit BackgroundWorkerMain");
	proc_exit(0);
}


void _PG_init(void) {
	BackgroundWorker worker;

	MemSet(&worker, 0, sizeof(BackgroundWorker));
	worker.bgw_flags = BGWORKER_SHMEM_ACCESS;
	worker.bgw_start_time = BgWorkerStart_RecoveryFinished;
	snprintf(worker.bgw_library_name, BGW_MAXLEN, "bgw");
	snprintf(worker.bgw_function_name, BGW_MAXLEN, "background_main");
	snprintf(worker.bgw_name, BGW_MAXLEN, "GoBackgroundWorker");
	worker.bgw_restart_time = BGW_NEVER_RESTART;
	worker.bgw_main_arg = (Datum) 0;
	worker.bgw_notify_pid = 0;
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

//WaitLatch will wait the duration, but can be interrupted, so use this and not time.Sleep
func WaitLatch(d time.Duration) error {
	if rc := C.wait_latch(C.long(d.Nanoseconds() / 1000000)); rc != 0 {
		return errors.New("postmaster is dead")
	}
	C.reset_latch()
	return nil
}
