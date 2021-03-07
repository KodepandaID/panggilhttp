package panggilhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"
)

// Get to set HTTP GET method.
// For the GET method, you can call this function more than 1, to merging the response body.
// You can use whitelist and blacklist args to filtering the response body.
// Whitelist to get field value from response body.
// Blacklist to ignore the field from response body.
func (c *Config) Get(url string, whitelist, blacklist []string) *Config {
	c.url = append(c.url, urlConfig{
		url:       url,
		method:    http.MethodGet,
		whitelist: whitelist,
		blacklist: blacklist,
	})

	return c
}

// Post to set HTTP POST method.
func (c *Config) Post(url string) *Config {
	c.url = append(c.url, urlConfig{
		url:    url,
		method: http.MethodPost,
	})

	return c
}

// Put to set HTTP PUT method.
func (c *Config) Put(url string) *Config {
	c.url = append(c.url, urlConfig{
		url:    url,
		method: http.MethodPut,
	})

	return c
}

// Patch to set HTTP PATCH method.
func (c *Config) Patch(url string) *Config {
	c.url = append(c.url, urlConfig{
		url:    url,
		method: http.MethodPatch,
	})

	return c
}

// Delete to set HTTP DELETE method.
func (c *Config) Delete(url string) *Config {
	c.url = append(c.url, urlConfig{
		url:    url,
		method: http.MethodDelete,
	})

	return c
}

// WithHeader to set HTTP headers.
func (c *Config) WithHeader(headers map[string]string) *Config {
	for key, val := range headers {
		c.req.Header.Set(key, val)
	}

	return c
}

// WithCookie to send HTTP cookies.
func (c *Config) WithCookie(cookies map[string]string) *Config {
	for key, val := range cookies {
		c.req.Header.SetCookie(key, val)
	}

	return c
}

// WithTimeout to set HTTP timeout in seconds.
// The default value is 1 seconds.
func (c *Config) WithTimeout(second time.Duration) *Config {
	if second < 1 {
		log.Fatal("Timeout value cannot be less than 1.")
	}

	c.timeout = time.Second * second

	return c
}

// WithFailRetry to retrying if HTTP call fails.
// Use 2 argument interval and attempt.
// interval args in miliseconds.
// attempt args is int how much to retry HTTP calls.
func (c *Config) WithFailRetry(interval time.Duration, attempt int) *Config {
	if interval < 1 || attempt < 1 {
		log.Fatal("Interval or Attempts value cannot be less than 1.")
	}

	c.retryInterval = time.Millisecond * interval
	c.retryAttempt = attempt

	return c
}

// SendJSON to send json data with POST, PUT or PATCH method.
func (c *Config) SendJSON(j map[string]interface{}) *Config {
	data, e := json.Marshal(j)
	if e != nil {
		log.Fatal("Failed to marshalling the JSON data")
	}

	c.req.Header.SetContentType("application/json")
	c.req.SetBody(data)

	return c
}

// SendFormData to send multipart/form-data with POST, PUT or PATCH method.
func (c *Config) SendFormData(fd map[string]string) *Config {
	if c.body.Len() == 0 {
		c.writer = multipart.NewWriter(&c.body)
	}

	for key, val := range fd {
		c.writer.WriteField(key, val)
	}

	return c
}

// SendFile to send file with POST, PUT or PATCH method.
func (c *Config) SendFile(key, filename string, file []byte) *Config {
	if c.body.Len() == 0 {
		c.writer = multipart.NewWriter(&c.body)
	}

	if file == nil {
		log.Fatal("File cannot be nil")
	}

	form, e := c.writer.CreateFormFile(key, filename)
	if e != nil {
		log.Fatalf("Create form error: %s", e)
	}

	r := bytes.NewReader(file)
	if _, e := io.Copy(form, r); e != nil {
		log.Fatalf("Write form file error: %s", e)
	}

	return c
}
