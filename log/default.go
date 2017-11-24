package log

import (
	"io"

	e "github.com/xtimeline/gox/errors"
)

var (
	def = NewLogger()
)

func LogX(depth int, err error) error {
	return def.LogX(depth+1, err)
}

func Log(err error) error {
	return def.LogX(1, err)
}

func New(code int, message string, detail map[string]interface{}) error {
	return def.LogX(1, e.New(code, message, detail))
}

func NewUnk(message string) error {
	return def.LogX(1, e.New(e.CodeUnknownError, message, nil))
}

func Debug(format string, o map[string]interface{}, v ...interface{}) {
	def.withPos(1, o).Debugf(format, v...)
}

func Info(format string, o map[string]interface{}, v ...interface{}) {
	def.withPos(1, o).Infof(format, v...)
}

func Warn(format string, o map[string]interface{}, v ...interface{}) {
	def.withPos(1, o).Warnf(format, v...)
}

func Error(format string, o map[string]interface{}, v ...interface{}) {
	def.withPos(1, o).Errorf(format, v...)
}

func PithyDebug(format string, o map[string]interface{}, v ...interface{}) {
	def.withFields(o).Debugf(format, v...)
}

func PithyInfo(format string, o map[string]interface{}, v ...interface{}) {
	def.withFields(o).Infof(format, v...)
}

func PithyWarn(format string, o map[string]interface{}, v ...interface{}) {
	def.withFields(o).Warnf(format, v...)
}

func PithyError(format string, o map[string]interface{}, v ...interface{}) {
	def.withFields(o).Errorf(format, v...)
}

func SetLevel(level string) {
	def.SetLevel(level)
}

func SetOutput(out io.Writer) {
	def.SetOutput(out)
}
