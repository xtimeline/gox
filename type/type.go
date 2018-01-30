package t

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/xtimeline/gox/json"
)

func typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

func UInt64(v interface{}) (uint64, error) {
	if i, ok := v.(uint64); ok {
		return i, nil
	} else {
		return 0, errors.Wrap(ErrConvertFail, typeof(v))
	}
}

func Int64(v interface{}) (int64, error) {
	if i, ok := v.(int64); ok {
		return i, nil
	} else {
		return 0, errors.Wrap(ErrConvertFail, typeof(v))
	}
}

func Bool(v interface{}) (bool, error) {
	if i, ok := v.(bool); ok {
		return i, nil
	} else {
		return false, errors.Wrap(ErrConvertFail, typeof(v))
	}
}

func Float64(v interface{}) (float64, error) {
	if i, ok := v.(float64); ok {
		return i, nil
	} else {
		return 0, errors.Wrap(ErrConvertFail, typeof(v))
	}
}

func Int(v interface{}) (int, error) {
	if i, ok := v.(int); ok {
		return i, nil
	} else {
		return 0, errors.Wrap(ErrConvertFail, typeof(v))
	}
}

func Str(v interface{}) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	} else {
		return "", errors.Wrap(ErrConvertFail, typeof(v))
	}
}

func Array(v interface{}) ([]interface{}, error) {
	if items, ok := v.([]interface{}); ok {
		return items, nil
	}
	return nil, errors.Wrap(ErrConvertFail, typeof(v))
}

func Map(v interface{}) (json.Map, error) {
	if m, ok := v.(map[string]interface{}); ok {
		return m, nil
	} else if m, ok := v.(json.Map); ok {
		return m, nil
	} else if m, ok := v.(*map[string]interface{}); ok {
		return *m, nil
	} else if m, ok := v.(*json.Map); ok {
		return *m, nil
	} else {
		return nil, errors.Wrap(ErrConvertFail, typeof(v))
	}
}
