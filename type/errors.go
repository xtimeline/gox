package t

import (
	"github.com/pkg/errors"
)

var ErrConvertFail = errors.New("convert fail")

func IsErrConvertFail(err error) bool {
	return errors.Cause(err) == ErrConvertFail
}
