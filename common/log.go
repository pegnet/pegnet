package common

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/zpatrick/go-config"
	"os"
	"strings"
	"time"
)

// The point of this logging isn't really to solve the solution of logging, but to create something
// we can use to debug until someone comes in here and does something good.

//logging is enabled

var DoLogging bool = false
var start int64
var LogFile = os.Stdout

// Some parameters are really pricey to compute.  So use Do in those cases.
// common.Do(func() {
//    common.Logf("sortMessages", "%s", AReallyExpensiveThing())
// }
func Do(something func()) {
	if DoLogging {
		something()
	}
}

func Logf(Catagory string, format string, a ...interface{}) {
	if !DoLogging {
		return
	}
	elapse := time.Now().UnixNano() - start
	ts := humanize.Comma(elapse)
	data := fmt.Sprintf(format, a...)
	lines := strings.Split(data, "\n")
	str := fmt.Sprintf("%-20s %17s -- %s\n", Catagory, ts, lines[0])
	for _, v := range lines[1:] {
		str = str + fmt.Sprintf("%-20s %17s -- %s\n", Catagory, ts, v)
	}
	_, _ = fmt.Fprint(LogFile, str)
}

func InitLogs(Config *config.Config) {
	// If logging is set, we turn it on.  If there is any problem with the config file about logging, we just
	// don't log, so no need to check the errors.
	DoLogging, _ = Config.Bool("Debug.Logging")
	// Looking to see if a filename is specified for the output file for logging
	LogFile = os.Stdout // default to stdout
	filename, err := Config.String("Debug.LogFile")
	if DoLogging && err == nil && len(filename) > 0 {
		file, err := os.Create(filename)
		if err == nil && file != nil {
			LogFile = file
		}
	}
}
