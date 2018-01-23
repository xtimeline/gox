package httpx

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
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
	body          []byte
	query         url.Values
	header        http.Header
	cookies       []http.Cookie
	breaker       HttpBreaker
	ctx           context.Context
	operationName string
	customRequest *http.Request
}

func newRequestOptions() requestOptions {
	return requestOptions{
		header:        make(http.Header),
		query:         make(url.Values),
		cookies:       make([]http.Cookie, 0, 1),
		operationName: "http request",
	}
}

type RequestOption func(opts *requestOptions)

func OperationName(v string) RequestOption {
	return func(opts *requestOptions) {
		opts.operationName = v
	}
}

func Body(v []byte) RequestOption {
	return func(opts *requestOptions) {
		opts.body = v
	}
}

func QueryKV(key, val string) RequestOption {
	return func(opts *requestOptions) {
		opts.query.Add(key, val)
	}
}

func HeadKV(key, val string) RequestOption {
	return func(opts *requestOptions) {
		opts.header.Add(key, val)
	}
}

func CookieKV(key, val string) RequestOption {
	return func(opts *requestOptions) {
		opts.cookies = append(opts.cookies, http.Cookie{Name: key, Value: val})
	}
}

func Context(v context.Context) RequestOption {
	return func(opts *requestOptions) {
		opts.ctx = v
	}
}

func Breaker(v HttpBreaker) RequestOption {
	return func(opts *requestOptions) {
		opts.breaker = v
	}
}

func CustomRequest(v *http.Request) RequestOption {
	return func(opts *requestOptions) {
		opts.customRequest = v
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
	body, err := data.Marshal()
	if err != nil {
		return nil, err
	}
	opts = append(opts, HeadKV("Content-Type", "application/json"))
	return cli.Put(url, body, opts...)
}

func (cli *Client) PostJson(url string, data json.Map, opts ...RequestOption) (*HttpResponse, error) {
	body, err := data.Marshal()
	if err != nil {
		return nil, err
	}
	opts = append(opts, HeadKV("Content-Type", "application/json"))
	return cli.Post(url, body, opts...)
}

func (cli *Client) PostForm(url string, data url.Values, opts ...RequestOption) (*HttpResponse, error) {
	opts = append(opts, HeadKV("Content-Type", "application/json"))
	return cli.Post(url, []byte(data.Encode()), opts...)
}

func (cli *Client) doRequest(request *http.Request, ctx context.Context) (*HttpResponse, error) {
	type Pack struct {
		response *http.Response
		err      error
	}
	c := make(chan Pack, 1)
	go func() {
		response, err := cli.raw.Do(request)
		c <- Pack{response: response, err: err}
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

func (cli *Client) DoRequest(method, url string, opts ...RequestOption) (*HttpResponse, error) {
	reqOps := newRequestOptions()
	for _, opt := range opts {
		opt(&reqOps)
	}

	fn := func() (*HttpResponse, error) {
		var err error

		//
		// use custom request first
		//
		request := reqOps.customRequest

		if request == nil {
			//
			// new request
			//
			var bodyReader io.ReadCloser
			if reqOps.body != nil {
				bodyReader = ioutil.NopCloser(strings.NewReader(string(reqOps.body)))
			}

			request, err = http.NewRequest(method, url, bodyReader)
			if err != nil {
				return nil, err
			}

			//
			// config request
			//
			length := len(reqOps.body)
			if length != 0 {
				request.ContentLength = int64(length)
			}

			if len(reqOps.header) != 0 {
				request.Header = reqOps.header
			}

			if len(reqOps.query) != 0 {
				request.URL.RawQuery = reqOps.query.Encode()
			}

			for _, c := range reqOps.cookies {
				request.AddCookie(&c)
			}

			if reqOps.ctx != nil {
				request = request.WithContext(reqOps.ctx)
			}

			//
			// tracing
			//
			{
				tracer := opentracing.GlobalTracer()
				span, ctx := opentracing.StartSpanFromContext(request.Context(), reqOps.operationName)
				defer span.Finish()
				request = request.WithContext(ctx)

				host, port, err := net.SplitHostPort(request.URL.Host)
				if err == nil {
					ext.PeerHostname.Set(span, host)
					if v, err := strconv.Atoi(port); err != nil {
						ext.PeerPort.Set(span, uint16(v))
					}
				} else {
					ext.PeerHostname.Set(span, request.URL.Host)
				}
				tracer.Inject(
					span.Context(),
					opentracing.HTTPHeaders,
					opentracing.HTTPHeadersCarrier(request.Header),
				)
			}
		}

		return cli.doRequest(request, reqOps.ctx)
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
