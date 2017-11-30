package http

import (
	"compress/gzip"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xtimeline/gox/json"
)

// https://github.com/sony/gobreaker
type HttpBreaker interface {
	Execute(func() (interface{}, error)) (interface{}, error)
}

type HttpClient struct {
	raw     *http.Client
	breaker HttpBreaker
}

type HttpValues struct {
	url.Values
}

func NewValues() HttpValues {
	v := HttpValues{
		make(url.Values),
	}
	return v
}

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
	if err := json.NewDecoder(bodyReader).Decode(out); err != nil {
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

func NewHttpClient(breaker HttpBreaker) *HttpClient {
	cfg := &tls.Config{
		InsecureSkipVerify: false,
	}
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		Dial:                dialer.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConnsPerHost: 150,
		TLSClientConfig:     cfg,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
	wrapper := &HttpClient{
		raw:     client,
		breaker: breaker,
	}
	return wrapper
}

func (c *HttpClient) NewRequest() *HttpRequest {
	httpRequest := &HttpRequest{
		httpClient: c,
		httpHeader: make(http.Header),
	}
	return httpRequest
}

type HttpRequest struct {
	httpClient *HttpClient
	httpHeader http.Header
}

func (r *HttpRequest) AddHeader(key, value string) *HttpRequest {
	r.httpHeader.Add(key, value)
	return r
}

func (r *HttpRequest) SetHeader(key, value string) *HttpRequest {
	r.httpHeader.Set(key, value)
	return r
}

func (r *HttpRequest) GetHeader(key string) string {
	return r.httpHeader.Get(key)
}

func (r *HttpRequest) DelHeader(key string) *HttpRequest {
	r.httpHeader.Del(key)
	return r
}

func (r *HttpRequest) newRawRequest(method string, url string, data []byte) (*http.Request, error) {
	var bodyReader io.ReadCloser
	if data != nil {
		bodyReader = ioutil.NopCloser(strings.NewReader(string(data)))
	}
	request, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	length := int64(len(data))
	if length != 0 {
		request.ContentLength = length
	}
	for key, value := range r.httpHeader {
		request.Header[key] = value
	}
	return request, nil
}

func (r *HttpRequest) doRawRequest(request *http.Request) (*HttpResponse, error) {
	response, err := r.httpClient.raw.Do(request)
	if err != nil {
		return nil, err
	}
	return &HttpResponse{response}, nil
}

func (r *HttpRequest) doRequest(method, url string, data []byte, query *HttpValues) (*HttpResponse, error) {
	fn := func() (interface{}, error) {
		request, err := r.newRawRequest(method, url, data)
		if err != nil {
			return nil, err
		}
		if query != nil {
			request.URL.RawQuery = query.Encode()
		}
		return r.doRawRequest(request)
	}

	if r.httpClient.breaker != nil {
		response, err := r.httpClient.breaker.Execute(fn)
		if err != nil {
			return nil, err
		}
		return response.(*HttpResponse), err
	}

	response, err := fn()
	if err != nil {
		return nil, err
	}
	return response.(*HttpResponse), err
}

func (r *HttpRequest) Post(url string, data []byte) (*HttpResponse, error) {
	return r.doRequest("POST", url, data, nil)
}

func (r *HttpRequest) Put(url string, data []byte) (*HttpResponse, error) {
	return r.doRequest("PUT", url, data, nil)
}

func (r *HttpRequest) Get(url string) (*HttpResponse, error) {
	return r.doRequest("GET", url, nil, nil)
}

func (r *HttpRequest) Query(url string, query HttpValues) (*HttpResponse, error) {
	return r.doRequest("GET", url, nil, &query)
}

func (r *HttpRequest) Delete(url string) (*HttpResponse, error) {
	return r.doRequest("DELETE", url, nil, nil)
}

func (r *HttpRequest) PutJson(url string, data json.Map) (*HttpResponse, error) {
	body, err := data.Marshal()
	if err != nil {
		return nil, err
	}
	return r.SetHeader("Content-Type", "application/json").Put(url, body)
}

func (r *HttpRequest) PostJson(url string, data json.Map) (*HttpResponse, error) {
	body, err := data.Marshal()
	if err != nil {
		return nil, err
	}
	return r.SetHeader("Content-Type", "application/json").Post(url, body)
}

func (r *HttpRequest) PostForm(url string, data HttpValues) (*HttpResponse, error) {
	return r.SetHeader("Content-Type", "application/x-www-form-urlencoded; param=value").Post(url, []byte(data.Encode()))
}
