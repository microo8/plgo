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
int wait_latch(long);
void reset_latch(void);
*/
import "C"
import (
	"errors"
	"net/http"
	"os"
	"sync"
	"time"
)

//export backgroundWorkerMain
func backgroundWorkerMain() C.int {
	if err := BGWTest(); err != nil {
		Log("BGWTest error: " + err.Error())
		return 1
	}
	return 0
}

//BGWTest ...
func BGWTest() error {
	Log("Starting GoBackgroundWorker")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})
	Log(http.ListenAndServe(":8080", nil).Error())
	for C.get_got_sigterm() == 0 {
		if rc := C.wait_latch(C.long(time.Second.Nanoseconds() / 1000000)); rc != 0 {
			return errors.New("postmaster is dead")
		}
		C.reset_latch()
		Log("Hello World from GoLang!!! Yeahhhhh!!!!" + time.Now().Format(time.UnixDate))
		var wg sync.WaitGroup
		wg.Add(1)
		touch(&wg, "/tmp/mehmehmeh")
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go touch(&wg, "/tmp/mehmeh")
		}
		wg.Wait()
	}
	Log("Finishing GoBackgroundWorker")
	return nil
}

func touch(wg *sync.WaitGroup, path string) {
	defer wg.Done()
	f, err := os.Create(path)
	if err != nil {
		Log("touch: " + err.Error())
		return
	}
	defer f.Close()
	f.Write([]byte("123"))
	Log("touch")
}
