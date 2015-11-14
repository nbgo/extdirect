package extdirect
import (
	stdlog "log"
	"os"
	"github.com/Sirupsen/logrus"
	"strings"
	"github.com/nbgo/fail"
	"fmt"
)

type logger interface {
	Print(v ...interface{})
}

const (
	logLevelInfo = "info: "
	logLevelWarn = "warn: "
)

var log logger = stdlog.New(os.Stderr, "", stdlog.LstdFlags)

// SetLogger sets custom logger for package.
func SetLogger(l logger) {
	log = l
}

// LogrusLogger is a logrus implementation of internal logger.
type LogrusLogger struct {
	L *logrus.Entry
}

// Print implements internal logger Print method for logrus.
func (logrusWrapper *LogrusLogger) Print(v ...interface{}) {
	defer func() {
		if err := recover(); err != nil {
			stdlog.Panicf("Logging error: %v", err)
		}
	}()
	if len(v) == 0 {
		return
	}
	l := logrusWrapper.L
	if len(v) == 1 {
		if err, errOk := v[0].(error); errOk {
			if err2, err2Ok := fail.GetOriginalError(err).(*ErrDirectActionMethod); err2Ok {
				var stackTrace string
				if err3, err3Ok := err2.Err.(error); err3Ok {
					stackTrace = fail.GetStackTrace(err3)
				}

				if stackTrace == "" {
					stackTraceSkip := 1
					if err2.isPanic {
						stackTraceSkip = 4
					}
					stackTrace = fail.StackTrace(stackTraceSkip)
				}

				l = l.WithFields(logrus.Fields{
					"action": err2.Action,
					"method": err2.Method,
					"stack": stackTrace,
				})
			} else {
				stackTrace := fail.GetStackTrace(err)
				if stackTrace == "" {
					stackTrace = fail.StackTrace(1)
				}

				l = l.WithFields(logrus.Fields{
					"stack": stackTrace,
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
		default: panic(fmt.Errorf("Unknown log message type: '%v.", logLevel))
		}
	}
}