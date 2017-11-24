package errors

import (
	"fmt"
	"time"
)

const (
	CodeError        = 50000000
	CodeUnknownError = 99999999
)

func New(code int, message string, detail map[string]interface{}) error {
	return E{
		id:      time.Now().UnixNano(), // TODO: use snowflake?
		code:    code,
		message: message,
		detail:  detail,
		logged:  new(bool),
	}
}

func GetCode(err error) int {
	if x, ok := err.(E); ok {
		return x.Code()
	}
	return CodeUnknownError
}


type E struct {
	id      int64
	code    int
	message string
	detail  map[string]interface{}
	logged  *bool
}

func (err E) Code() int {
	return err.code
}

func (err E) Error() string {
	if err.detail == nil {
		return err.message
	} else {
		return fmt.Sprintf("%s: %v", err.message, err.detail)
	}
}

func (err E) Message() string {
	return err.message
}

func (err E) Detail() map[string]interface{} {
	return err.detail
}

func (err E) Id() int64 {
	return err.id
}

func (err E) IsLogged() bool {
	return *err.logged
}

func (err E) MarkLogged() {
	*err.logged = true
}

func (err E) IsError() bool {
	return err.code >= CodeError
}
