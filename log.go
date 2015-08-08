package extdirect
import (
	stdlog "log"
	"os"
)

type logger interface {
	Print(v ...interface{})
}

var log logger = stdlog.New(os.Stderr, "", stdlog.LstdFlags)

func SetLogger(l logger) {
	log = l
}