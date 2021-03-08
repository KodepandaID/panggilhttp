package panggilhttp

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/KodepandaID/panggilhttp/pkg/merging"
	"github.com/KodepandaID/panggilhttp/pkg/retry"
	"github.com/valyala/fasthttp"
)

var version = "panggilHTTP-v0.1.0"

// Config to set the configuration to calling an HTTP.
type Config struct {
	// HTTP configuration
	client  *fasthttp.Client
	req     *fasthttp.Request
	url     []urlConfig
	timeout time.Duration // in Seconds

	// HTTP request body
	body   bytes.Buffer
	writer *multipart.Writer

	// HTTP retry configuration
	retryInterval time.Duration // in Miliseconds
	retryAttempt  int           // How much retry to calling an HTTP
}

type urlConfig struct {
	url       string
	method    string
	whitelist []string // to get only whitelist field from the response body
	blacklist []string // to ignoring the blacklist field from the response body
}

// Response is a response structure
type Response struct {
	StatusCode int
	Headers    map[string]string
	Cookies    map[string]string
	Body       []byte
}

// New is an adapter to create new instance.
func New() *Config {
	req := fasthttp.AcquireRequest()

	return &Config{
		client: &fasthttp.Client{
			Name:                          version,
			NoDefaultUserAgentHeader:      true,
			ReadBufferSize:                4096,
			WriteBufferSize:               4096,
			DisableHeaderNamesNormalizing: true,
		},
		req: req,
	}
}

// Do to running an HTTP call.
// Response header and cookies can be returned if only use 1 call HTTP.
func (c *Config) Do() (Response, error) {
	var (
		headers    map[string]string
		cookies    map[string]string
		statusCode int
	)

	defer fasthttp.ReleaseRequest(c.req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	m := merging.New()

	for _, row := range c.url {
		c.req.SetRequestURI(row.url)
		c.req.Header.SetMethod(row.method)

		// If the request body is a multipart/form-data,
		// the writer will be closed.
		if c.body.Len() > 0 {
			c.writer.Close()

			c.req.Header.SetContentType(c.writer.FormDataContentType())
			c.req.SetBodyStream(&c.body, c.body.Len())
		}

		// If the HTTP request uses HTTP retry, if HTTP is failed will be retrying.
		finalResp, e := retry.New(&retry.Config{
			Timeouts: c.timeout,
			Attempts: c.retryAttempt,
			Interval: c.retryInterval,
		}).Do(c.req, resp, c.client)
		if e != nil && e.Error() != "Request Timeout" {
			return Response{}, e
		} else if e != nil && e.Error() == "Request Timeout" {
			return Response{
				StatusCode: http.StatusRequestTimeout,
			}, e
		}

		if len(c.url) == 1 {
			headers = convertHeader(&finalResp.Header)
			cookies = convertCookie(&finalResp.Header)
		}

		// If calling more than one URL, the response body will be merged.
		if len(row.whitelist) > 0 {
			m.MergeFromWhitelist(row.whitelist, finalResp.Body())
		} else if len(row.whitelist) == 0 {
			m.Merge(row.blacklist, finalResp.Body())
		}

		statusCode = finalResp.StatusCode()

		c.req.Reset()
		finalResp.Reset()
	}

	httpResponse := Response{
		StatusCode: statusCode,
		Headers:    headers,
		Cookies:    cookies,
		Body:       m.Get(),
	}

	return httpResponse, nil
}
