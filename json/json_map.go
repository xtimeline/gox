package json

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/xtimeline/gox/conv"
)

var (
	ErrKeyInvalid  = errors.New("key invalid")
	ErrConvertFail = errors.New("convert fail")
)

type Map map[string]interface{}

func (m Map) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

func (m Map) MustMarshal() []byte {
	data, _ := json.Marshal(m)
	return data
}

func (m Map) Unmarshal(data []byte) error {
	return json.Unmarshal(data, m)
}

func (m Map) Decode(reader io.Reader) error {
	decoder := json.NewDecoder(reader)
	decoder.UseNumber()
	return decoder.Decode(&m)
}

func (m Map) ItemCount(k string) (int, error) {
	if v, has := m[k]; has {
		if items, ok := v.([]interface{}); ok {
			return len(items), nil
		} else {
			return 0, ErrConvertFail
		}
	}
	return 0, ErrKeyInvalid
}

func (m Map) StringItem(k string, index int64) (string, error) {
	if v, has := m[k]; has {
		if items, ok := v.([]interface{}); ok {
			if s, ok := items[index].(string); ok {
				return s, nil
			}
		}
		return "", ErrConvertFail
	}
	return "", ErrKeyInvalid
}

func (m Map) Json(k string) (Map, error) {
	if v, has := m[k]; has {
		if m, ok := v.(map[string]interface{}); ok {
			return Map(m), nil
		} else {
			return nil, ErrConvertFail
		}
	}
	return nil, ErrKeyInvalid
}

func (m Map) MustJson(k string) Map {
	return Map(m[k].(map[string]interface{}))
}

func (m Map) Int64(k string) (int64, error) {
	if v, has := m[k]; has {
		if num, ok := v.(json.Number); ok {
			return num.Int64()
		} else {
			return 0, ErrConvertFail
		}
	}
	return 0, ErrKeyInvalid
}

func (m Map) MustInt64(k string) int64 {
	i, err := m[k].(json.Number).Int64()
	if err != nil {
		panic(err)
	}
	return i
}

func (m Map) MustParseString(k string) string {
	if v, has := m[k]; has {
		if s, ok := v.(string); ok {
			return s
		} else if n, ok := v.(json.Number); ok {
			return n.String()
		} else if i64, ok := v.(int64); ok {
			return conv.FormatInt64(i64)
		} else if i, ok := v.(int); ok {
			return conv.FormatInt(i)
		} else if i8, ok := v.(int8); ok {
			return conv.FormatInt8(i8)
		}
	}
	panic(-1)
}

func (m Map) String(k string) (string, error) {
	if v, has := m[k]; has {
		if s, ok := v.(string); ok {
			return s, nil
		} else {
			return "", ErrConvertFail
		}
	}
	return "", ErrKeyInvalid
}

func (m Map) MustString(k string) string {
	return m[k].(string)
}
