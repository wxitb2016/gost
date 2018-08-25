package main

import (
	"fmt"
	"os"
	"flag"
	"net/http"
	"strings"
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
		var addr string
		if !strings.Contains(port, ":") {
			addr = fmt.Sprintf("localhost:%s", port)
		} else {
			addr = port
		}
		pprofAddr := addr
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

	glogFlagShim(map[string]string{
		"v":                "3",
		"logtostderr":      "true",
	})
}

// Copied from https://github.com/urfave/cli/issues/269#issuecomment-255516642
func glogFlagShim(fakeVals map[string]string) {
	flag.VisitAll(func(fl *flag.Flag) {
		if val, ok := fakeVals[fl.Name]; ok {
			fl.Value.Set(val)
		}
	})
}
