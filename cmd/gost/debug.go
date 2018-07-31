package main

import (
	"fmt"
	"os"
	"net/http"
	// "strconv"
	_ "net/http/pprof"

	"github.com/phuslu/glog"
	"runtime"
	// "github.com/jtolds/go-manhole"
)

// type MyType struct{ x int }

// func (m *MyType) Set(x int) { m.x = x }
// func (m *MyType) Get() int  { return m.x }

func init() {

	if port, ok := os.LookupEnv("GOST_PPROF"); ok {
		if _, ok = os.LookupEnv("GOST_MUTEX"); ok {
			glog.Infof("mutex pprof enabled")
			runtime.SetMutexProfileFraction(1)
		}
		pprofAddr := fmt.Sprintf("localhost:%s", port)
		glog.Infof("pprof listening on http://%v", pprofAddr)
		go http.ListenAndServe(pprofAddr, nil)
	}

	// if port, ok := os.LookupEnv("GOST_DEBUG"); ok {
	// 	manhole.RegisterType("MyType", MyType{})
	// 	manhole.RegisterType("MyType", MyType{})
	// 	if intPort, err := strconv.Atoi(port); err == nil {
	// 		glog.Infof("manhole listening on %v", port)
	// 		go manhole.ListenAndServe(intPort)
	// 	}
	// }
}
