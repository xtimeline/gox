package json

import (
	"github.com/pkg/errors"
)

var ErrKeyMiss = errors.New("key is missing")
var ErrConvertFail = errors.New("convert fail")

func IsErrKeyMiss(err error) bool {
	return errors.Cause(err) == ErrKeyMiss
}

func IsErrConvertFail(err error) bool {
	return errors.Cause(err) == ErrConvertFail
}
