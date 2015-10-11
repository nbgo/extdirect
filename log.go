package extdirect
import (
	stdlog "log"
	"os"
	"github.com/Sirupsen/logrus"
	"strings"
	"runtime/debug"
)

type logger interface {
	Print(v ...interface{})
}

const (
	logLevelInfo = "info: "
	logLevelWarn = "warn: "
)

var log logger = stdlog.New(os.Stderr, "", stdlog.LstdFlags)

func SetLogger(l logger) {
	log = l
}

type LogrusLogger struct {
	L *logrus.Entry
}
func (this *LogrusLogger) Print(v ...interface{}) {
	if len(v) == 0 {
		return
	}
	l := this.L
	if len(v) == 1 {
		if err, errOk := v[0].(error); errOk {
			if err2, err2Ok := err.(*ErrDirectActionMethod); err2Ok {
				l = l.WithFields(logrus.Fields{
					"action": err2.Action,
					"method": err2.Method,
					"stack": string(debug.Stack()),
				})
			}
			l.Error(err)
		} else {
			l.Debug(v[0])
		}
	} else {
		if len(v) > 2 {
			l = l.WithFields(v[2].(map[string]interface{}))
		}

		m := strings.TrimSpace(v[1].(string))
		logLevel := v[0].(string)
		switch logLevel {
		case logLevelInfo: l.Info(m)
		case logLevelWarn: l.Warn(m)
		default: stdlog.Panicf("Unknow log message type: '%v'", logLevel)
		}
	}
}