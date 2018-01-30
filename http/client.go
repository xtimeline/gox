package httpx

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	"github.com/xtimeline/gox/json"
)

var (
	ErrRequestTimeOut = errors.New("request time out")
)

type Client struct {
	raw       *http.Client
	transport *http.Transport
}

type clientOptions struct {
}

type ClientOption func(opts *clientOptions)

func NewClient(opts ...ClientOption) *Client {
	cliOps := clientOptions{}
	for _, opt := range opts {
		opt(&cliOps)
	}
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
	wrapper := &Client{
		raw:       client,
		transport: transport,
	}
	return wrapper
}

type requestOptions struct {
	query       url.Values
	breaker     HttpBreaker
	request     *http.Request
	testHandler http.Handler
}

func newRequestOptions(request *http.Request) requestOptions {
	return requestOptions{
		query:   make(url.Values),
		request: request,
	}
}

type RequestOption func(opts *requestOptions) error

func Body(v []byte) RequestOption {
	return func(opts *requestOptions) error {
		if v != nil {
			opts.request.Body = ioutil.NopCloser(bytes.NewBuffer(v))
			opts.request.ContentLength = int64(len(v))
		}
		return nil
	}
}

func QueryKV(key, val string) RequestOption {
	return func(opts *requestOptions) error {
		opts.query.Add(key, val)
		return nil
	}
}

func HeadKV(key, val string) RequestOption {
	return func(opts *requestOptions) error {
		opts.request.Header.Add(key, val)
		return nil
	}
}

func CookieKV(key, val string) RequestOption {
	return func(opts *requestOptions) error {
		opts.request.AddCookie(&http.Cookie{Name: key, Value: val})
		return nil
	}
}

func Context(v context.Context) RequestOption {
	return func(opts *requestOptions) error {
		if v != nil {
			opts.request = opts.request.WithContext(v)
		}
		return nil
	}
}

func Breaker(v HttpBreaker) RequestOption {
	return func(opts *requestOptions) error {
		opts.breaker = v
		return nil
	}
}

func TestHandler(v http.Handler) RequestOption {
	return func(opts *requestOptions) error {
		opts.testHandler = v
		return nil
	}
}

func (cli *Client) Post(url string, body []byte, opts ...RequestOption) (*HttpResponse, error) {
	opts = append(opts, Body(body))
	return cli.DoRequest("POST", url, opts...)
}

func (cli *Client) Put(url string, body []byte, opts ...RequestOption) (*HttpResponse, error) {
	opts = append(opts, Body(body))
	return cli.DoRequest("PUT", url, opts...)
}

func (cli *Client) Patch(url string, body []byte, opts ...RequestOption) (*HttpResponse, error) {
	opts = append(opts, Body(body))
	return cli.DoRequest("PATCH", url, opts...)
}

func (cli *Client) Get(url string, opts ...RequestOption) (*HttpResponse, error) {
	return cli.DoRequest("GET", url, opts...)
}

func (cli *Client) Delete(url string, opts ...RequestOption) (*HttpResponse, error) {
	return cli.DoRequest("DELETE", url, opts...)
}

func (cli *Client) PutJson(url string, data json.Map, opts ...RequestOption) (*HttpResponse, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	opts = append(opts, HeadKV("Content-Type", "application/json"))
	return cli.Put(url, body, opts...)
}

func (cli *Client) PostJson(url string, data json.Map, opts ...RequestOption) (*HttpResponse, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	opts = append(opts, HeadKV("Content-Type", "application/json"))
	return cli.Post(url, body, opts...)
}

func (cli *Client) PatchJson(url string, data json.Map, opts ...RequestOption) (*HttpResponse, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	opts = append(opts, HeadKV("Content-Type", "application/json"))
	return cli.Patch(url, body, opts...)
}

func (cli *Client) PutForm(url string, data url.Values, opts ...RequestOption) (*HttpResponse, error) {
	opts = append(opts, HeadKV("Content-Type", "application/x-www-form-urlencoded; param=value"))
	return cli.Put(url, []byte(data.Encode()), opts...)
}

func (cli *Client) PostForm(url string, data url.Values, opts ...RequestOption) (*HttpResponse, error) {
	opts = append(opts, HeadKV("Content-Type", "application/x-www-form-urlencoded; param=value"))
	return cli.Post(url, []byte(data.Encode()), opts...)
}

func (cli *Client) PatchForm(url string, data url.Values, opts ...RequestOption) (*HttpResponse, error) {
	opts = append(opts, HeadKV("Content-Type", "application/x-www-form-urlencoded; param=value"))
	return cli.Patch(url, []byte(data.Encode()), opts...)
}

func (cli *Client) sendRequest(request *http.Request, ctx context.Context, testHandler http.Handler) (*HttpResponse, error) {
	type Pack struct {
		response *http.Response
		err      error
	}
	c := make(chan Pack, 1)
	go func() {
		if testHandler == nil {
			response, err := cli.raw.Do(request)
			c <- Pack{response: response, err: err}
		} else {
			recorder := httptest.NewRecorder()
			testHandler.ServeHTTP(recorder, request)
			c <- Pack{response: recorder.Result(), err: nil}
		}
	}()
	select {
	case <-ctx.Done():
		cli.transport.CancelRequest(request)
		<-c
		return nil, ErrRequestTimeOut
	case result := <-c:
		if result.err != nil {
			return nil, result.err
		}
		return &HttpResponse{result.response}, nil
	}
}

func (cli *Client) Do(request *http.Request, opts ...RequestOption) (*HttpResponse, error) {
	//
	// config request
	//
	reqOps := newRequestOptions(request)
	for _, opt := range opts {
		err := opt(&reqOps)
		if err != nil {
			return nil, err
		}
	}

	fn := func() (*HttpResponse, error) {
		//
		// encodes query params
		//
		if len(reqOps.query) != 0 {
			request.URL.RawQuery = reqOps.query.Encode()
		}

		//
		// request config done and sent it
		//
		return cli.sendRequest(request, request.Context(), reqOps.testHandler)
	}

	if reqOps.breaker != nil {
		done, err := reqOps.breaker.Allow()
		if err != nil {
			return nil, err
		}
		response, err := fn()
		done(err == nil && response.StatusCode < http.StatusInternalServerError)
		return response, err
	}

	response, err := fn()
	if err != nil {
		return nil, err
	}
	return response, err
}

func (cli *Client) DoRequest(method, url string, opts ...RequestOption) (*HttpResponse, error) {
	//
	// new request
	//
	request, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	return cli.Do(request, opts...)
}
