package extdirect
import (
	stdlog "log"
	"os"
)

var log *stdlog.Logger = stdlog.New(os.Stderr, "extdirect ", stdlog.LstdFlags)

func SetLogger(logger *stdlog.Logger) {
	log = logger
}