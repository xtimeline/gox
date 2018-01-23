package httpx

import (
	"compress/gzip"
	stdjson "encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/xtimeline/gox/json"
)

type HttpResponse struct {
	*http.Response
}

func (r *HttpResponse) readJson(out interface{}) error {
	defer r.Body.Close()
	var err error
	var bodyReader io.Reader
	if r.Header.Get("Content-Encoding") == "gzip" {
		bodyReader, err = gzip.NewReader(r.Body)
		if err != nil {
			return err
		}
	} else {
		bodyReader = r.Body
	}
	decoder := stdjson.NewDecoder(bodyReader)
	decoder.UseNumber()
	if err := decoder.Decode(out); err != nil {
		return err
	}
	return nil
}

func (r *HttpResponse) ReadBytes() ([]byte, error) {
	defer r.Body.Close()
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return data, err
}

func (r *HttpResponse) ReadJsons() ([]json.Map, error) {
	items := []json.Map{}
	err := r.readJson(&items)
	return items, err
}

func (r *HttpResponse) ReadJson() (json.Map, error) {
	jsonMap := json.Map{}
	err := r.readJson(&jsonMap)
	return jsonMap, err
}

func (r *HttpResponse) ReadObject(o interface{}) error {
	return r.readJson(o)
}
