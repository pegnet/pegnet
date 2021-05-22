package cmd

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	_ "net/http/pprof"

	log "github.com/sirupsen/logrus"
)

// StartProfiler runs the go pprof tool
// `go tool pprof http://localhost:6060/debug/pprof/profile`
// https://golang.org/pkg/net/http/pprof/
func StartProfiler(port int) {
	runtime.SetBlockProfileRate(int(time.Second.Nanoseconds()))
	url := fmt.Sprintf("localhost:%d", port)
	log.Infof("Profiling on %s", url)
	log.Println(http.ListenAndServe(url, nil))
}
