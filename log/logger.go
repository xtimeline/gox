package l

import (
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"

	logrus "github.com/Sirupsen/logrus"
	e "github.com/xtimeline/gox/errors"
)

type Logger struct {
	rawLogger *logrus.Logger
}

func NewLogger() *Logger {
	rawLogger := logrus.New()
	rawLogger.Out = os.Stdout
	rawLogger.Formatter = &logrus.JSONFormatter{}
	rawLogger.Level = logrus.DebugLevel
	return &Logger{
		rawLogger: rawLogger,
	}
}

func (l *Logger) LogX(depth int, err error) error {
	if _, ok := err.(e.E); !ok {
		err = e.New(e.CodeUnknownError, err.Error(), nil)
	}
	ex, _ := err.(e.E)
	id := ex.Id()
	code := ex.Code()
	logged := ex.IsLogged()
	var detail map[string]interface{}
	if !logged {
		detail = ex.Detail()
	}
	entry := l.withPos(depth+1, detail).WithField("code", code).WithField("eid", id)
	if !logged {
		entry = entry.WithField("cap", 1)
	}
	if ex.IsError() {
		entry.Error(ex.Message())
	} else {
		entry.Warn(ex.Message())
	}
	ex.MarkLogged()
	return err
}

func (l *Logger) Log(err error) error {
	return l.LogX(1, err)
}

func (l *Logger) New(code int, message string, detail map[string]interface{}) error {
	return l.LogX(1, e.New(code, message, detail))
}

func (l *Logger) NewUnk(message string) error {
	return l.LogX(1, e.New(e.CodeUnknownError, message, nil))
}

func (l *Logger) Debug(format string, o map[string]interface{}, v ...interface{}) {
	l.withPos(1, o).Debugf(format, v...)
}

func (l *Logger) Info(format string, o map[string]interface{}, v ...interface{}) {
	l.withPos(1, o).Infof(format, v...)
}

func (l *Logger) Warn(format string, o map[string]interface{}, v ...interface{}) {
	l.withPos(1, o).Warnf(format, v...)
}

func (l *Logger) Error(format string, o map[string]interface{}, v ...interface{}) {
	l.withPos(1, o).Errorf(format, v...)
}

func (l *Logger) PithyDebug(format string, o map[string]interface{}, v ...interface{}) {
	l.withFields(o).Debugf(format, v...)
}

func (l *Logger) PithyInfo(format string, o map[string]interface{}, v ...interface{}) {
	l.withFields(o).Infof(format, v...)
}

func (l *Logger) PithyWarn(format string, o map[string]interface{}, v ...interface{}) {
	l.withFields(o).Warnf(format, v...)
}

func (l *Logger) PithyError(format string, o map[string]interface{}, v ...interface{}) {
	l.withFields(o).Errorf(format, v...)
}

func (l *Logger) SetLevel(level string) {
	if lvl, ok := levelMap[strings.ToLower(level)]; ok {
		l.rawLogger.Level = lvl
	}
}

func (l *Logger) SetOutput(out io.Writer) {
	l.rawLogger.Out = out
}

func (l *Logger) withPos(depth int, o map[string]interface{}) *logrus.Entry {
	return l.withFields(o).WithField("pos", filePos(depth+1))
}

func (l *Logger) withFields(o map[string]interface{}) *logrus.Entry {
	return l.rawLogger.WithFields(o)
}

var levelMap = map[string]logrus.Level{
	"trace": logrus.DebugLevel,
	"debug": logrus.DebugLevel,
	"info":  logrus.InfoLevel,
	"warn":  logrus.WarnLevel,
	"error": logrus.ErrorLevel,
	"fatal": logrus.FatalLevel,
	"panic": logrus.PanicLevel,
}

func filePos(depth int) string {
	_, file, line, ok := runtime.Caller(depth + 1)
	if !ok {
		file = "???"
		line = 0
	} else {
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				file = file[i+1:]
				break
			}
		}
	}
	return file + ":" + strconv.FormatInt(int64(line), 10)
}
