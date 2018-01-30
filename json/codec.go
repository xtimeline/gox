package json

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/ugorji/go/codec"
)

type unixTime struct{}

func (x unixTime) ConvertExt(v interface{}) interface{} {
	switch v2 := v.(type) {
	case time.Time:
		return v2.UTC().Unix()
	case *time.Time:
		return v2.UTC().Unix()
	default:
		panic(fmt.Sprintf("unsupported format for time conversion: expecting time.Time; got %T", v))
	}
}

func (x unixTime) UpdateExt(dest interface{}, v interface{}) {
	tt := dest.(*time.Time)
	switch v2 := v.(type) {
	case int64:
		*tt = time.Unix(0, v2).UTC()
	case uint64:
		*tt = time.Unix(0, int64(v2)).UTC()
	default:
		panic(fmt.Sprintf("unsupported format for time conversion: expecting int64/uint64; got %T", v))
	}
}

var jsonHandler *codec.JsonHandle

func init() {
	jsonHandler = new(codec.JsonHandle)
	timeType := reflect.TypeOf(time.Time{})
	jsonHandler.SetInterfaceExt(timeType, 1, unixTime{})
}

func NewEncoder(w io.Writer) *codec.Encoder {
	return codec.NewEncoder(w, jsonHandler)
}

func NewDecoder(r io.Reader) *codec.Decoder {
	return codec.NewDecoder(r, jsonHandler)
}

func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
