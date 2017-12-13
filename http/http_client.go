package http

import (
	"compress/gzip"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/xtimeline/gox/json"
)

var (
	ErrRequestTimeOut = errors.New("request time out")
)

// https://github.com/sony/gobreaker
type HttpBreaker interface {
	Execute(func() (interface{}, error)) (interface{}, error)
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

type HttpClientConfig struct {
	breaker        HttpBreaker
	disbaleTimeout bool
}

type HttpClient struct {
	raw       *http.Client
	transport *http.Transport
	cfg       HttpClientConfig
}

func NewHttpClient(cfg HttpClientConfig) *HttpClient {
	dialer := &net.Dialer{
		Timeout:   3 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		Dial:                dialer.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
		MaxIdleConnsPerHost: 150,
	}
	client := &http.Client{
		Transport: transport,
	}
	wrapper := &HttpClient{
		raw:       client,
		transport: transport,
		cfg:       cfg,
	}
	return wrapper
}

func NewDefaultHttpClient() *HttpClient {
	return NewHttpClient(HttpClientConfig{})
}

func (c *HttpClient) NewRequest() *HttpRequest {
	httpRequest := &HttpRequest{
		httpClient: c,
		httpHeader: make(http.Header),
		timeout:    2 * time.Second,
	}
	return httpRequest
}

func (c *HttpClient) NewRequestWithoutTimeout() *HttpRequest {
	httpRequest := &HttpRequest{
		httpClient: c,
		httpHeader: make(http.Header),
		timeout:    0,
	}
	return httpRequest
}

func (c *HttpClient) NewRequestWithTimeout(timeout time.Duration) *HttpRequest {
	httpRequest := &HttpRequest{
		httpClient: c,
		httpHeader: make(http.Header),
		timeout:    timeout,
	}
	return httpRequest
}

type HttpRequest struct {
	httpClient *HttpClient
	httpHeader http.Header
	timeout    time.Duration
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

func (r *HttpRequest) doRawRequestWithTimeout(request *http.Request, ctx context.Context) (*HttpResponse, error) {
	type Pack struct {
		response *http.Response
		err      error
	}
	c := make(chan Pack, 1)
	go func() {
		response, err := r.httpClient.raw.Do(request)
		c <- Pack{response: response, err: err}
	}()
	select {
	case <-ctx.Done():
		r.httpClient.transport.CancelRequest(request)
		<-c
		return nil, ErrRequestTimeOut
	case result := <-c:
		if result.err != nil {
			return nil, result.err
		}
		return &HttpResponse{result.response}, nil
	}
}

func (r *HttpRequest) doRawRequest(request *http.Request) (*HttpResponse, error) {
	response, err := r.httpClient.raw.Do(request)
	if err != nil {
		return nil, err
	}
	return &HttpResponse{response}, nil
}

func (r *HttpRequest) DoRequest(method, url string, data []byte, query *HttpValues) (*HttpResponse, error) {
	fn := func() (interface{}, error) {
		request, err := r.newRawRequest(method, url, data)
		if err != nil {
			return nil, err
		}
		if query != nil {
			request.URL.RawQuery = query.Encode()
		}
		if r.httpClient.cfg.disbaleTimeout == false && r.timeout > 0 {
			ctx, _ := context.WithTimeout(context.Background(), r.timeout)
			return r.doRawRequestWithTimeout(request, ctx)
		} else {
			return r.doRawRequest(request)
		}
	}

	if r.httpClient.cfg.breaker != nil {
		response, err := r.httpClient.cfg.breaker.Execute(fn)
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
	return r.DoRequest("POST", url, data, nil)
}

func (r *HttpRequest) Put(url string, data []byte) (*HttpResponse, error) {
	return r.DoRequest("PUT", url, data, nil)
}

func (r *HttpRequest) Patch(url string, data []byte) (*HttpResponse, error) {
	return r.DoRequest("PATCH", url, data, nil)
}

func (r *HttpRequest) Get(url string) (*HttpResponse, error) {
	return r.DoRequest("GET", url, nil, nil)
}

func (r *HttpRequest) Query(url string, query HttpValues) (*HttpResponse, error) {
	return r.DoRequest("GET", url, nil, &query)
}

func (r *HttpRequest) Delete(url string) (*HttpResponse, error) {
	return r.DoRequest("DELETE", url, nil, nil)
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
