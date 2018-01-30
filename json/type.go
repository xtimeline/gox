package json

import (
	"fmt"

	"github.com/pkg/errors"
)

func typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

func UInt64(m Map, key string) (uint64, error) {
	if v, has := m[key]; has {
		if i, ok := v.(uint64); ok {
			return i, nil
		} else {
			return 0, errors.Wrap(ErrConvertFail, typeof(v))
		}
	}
	return 0, errors.Wrap(ErrKeyMiss, key)
}

func Int64(m Map, key string) (int64, error) {
	if v, has := m[key]; has {
		if i, ok := v.(int64); ok {
			return i, nil
		} else {
			return 0, errors.Wrap(ErrConvertFail, typeof(v))
		}
	}
	return 0, errors.Wrap(ErrKeyMiss, key)
}

func Bool(m Map, key string) (bool, error) {
	if v, has := m[key]; has {
		if i, ok := v.(bool); ok {
			return i, nil
		} else {
			return false, errors.Wrap(ErrConvertFail, typeof(v))
		}
	}
	return false, errors.Wrap(ErrKeyMiss, key)
}

func Float64(m Map, key string) (float64, error) {
	if v, has := m[key]; has {
		if i, ok := v.(float64); ok {
			return i, nil
		} else {
			return 0, errors.Wrap(ErrConvertFail, typeof(v))
		}
	}
	return 0, errors.Wrap(ErrKeyMiss, key)
}

func Int(m Map, key string) (int, error) {
	if v, has := m[key]; has {
		if i, ok := v.(int); ok {
			return i, nil
		} else {
			return 0, errors.Wrap(ErrConvertFail, typeof(v))
		}
	}
	return 0, errors.Wrap(ErrKeyMiss, key)
}

func Str(m Map, key string) (string, error) {
	if v, has := m[key]; has {
		if s, ok := v.(string); ok {
			return s, nil
		} else {
			return "", errors.Wrap(ErrConvertFail, typeof(v))
		}
	}
	return "", errors.Wrap(ErrKeyMiss, key)
}

func Json(m Map, key string) (Map, error) {
	if v, has := m[key]; has {
		if m, ok := v.(map[string]interface{}); ok {
			return m, nil
		} else if m, ok := v.(Map); ok {
			return m, nil
		} else if m, ok := v.(*map[string]interface{}); ok {
			return *m, nil
		} else if m, ok := v.(*Map); ok {
			return *m, nil
		} else {
			return nil, errors.Wrap(ErrConvertFail, typeof(v))
		}
	}
	return nil, errors.Wrap(ErrKeyMiss, key)
}

func Count(m Map, key string) (int, error) {
	if v, has := m[key]; has {
		if items, ok := v.([]interface{}); ok {
			return len(items), nil
		} else {
			return 0, errors.Wrap(ErrConvertFail, typeof(v))
		}
	}
	return 0, errors.Wrap(ErrKeyMiss, key)
}

func Array(m Map, key string) ([]interface{}, error) {
	if v, has := m[key]; has {
		if items, ok := v.([]interface{}); ok {
			return items, nil
		}
		return nil, errors.Wrap(ErrConvertFail, typeof(v))
	}
	return nil, errors.Wrap(ErrKeyMiss, key)
}
