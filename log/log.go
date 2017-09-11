package log

import (
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"

	logrus "github.com/Sirupsen/logrus"
	e "github.com/xtimeline/gox/errors"
)

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

func LogX(depth int, err error) error {
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
	entry := withPos(depth+1, detail).WithField("code", code).WithField("eid", id)
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

func Log(err error) error {
	return LogX(1, err)
}

func New(code int, message string, detail map[string]interface{}) error {
	return LogX(1, e.New(code, message, detail))
}

func NewUnk(message string) error {
	return LogX(1, e.New(e.CodeUnknownError, message, nil))
}

func Debug(format string, o map[string]interface{}, v ...interface{}) {
	withPos(1, o).Debugf(format, v...)
}

func Info(format string, o map[string]interface{}, v ...interface{}) {
	withPos(1, o).Infof(format, v...)
}

func Warn(format string, o map[string]interface{}, v ...interface{}) {
	withPos(1, o).Warnf(format, v...)
}

func Error(format string, o map[string]interface{}, v ...interface{}) {
	withPos(1, o).Errorf(format, v...)
}

func PithyDebug(format string, o map[string]interface{}, v ...interface{}) {
	withFields(o).Debugf(format, v...)
}

func PithyInfo(format string, o map[string]interface{}, v ...interface{}) {
	withFields(o).Infof(format, v...)
}

func PithyWarn(format string, o map[string]interface{}, v ...interface{}) {
	withFields(o).Warnf(format, v...)
}

func PithyError(format string, o map[string]interface{}, v ...interface{}) {
	withFields(o).Errorf(format, v...)
}

func SetLevel(level string) {
	logrus.SetLevel(levelMap[strings.ToLower(level)])
}

func SetOutput(out io.Writer) {
	logrus.SetOutput(out)
}

func withPos(depth int, o map[string]interface{}) *logrus.Entry {
	return withFields(o).WithField("pos", filePos(depth+1))
}

func withFields(o map[string]interface{}) *logrus.Entry {
	return logrus.WithFields(o)
}

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.DebugLevel)
}
