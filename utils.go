package panggilhttp

import (
	"github.com/valyala/fasthttp"
)

func convertHeader(headers *fasthttp.ResponseHeader) map[string]string {
	m := make(map[string]string)
	headers.VisitAll(func(key, value []byte) {
		m[string(key)] = string(value)
	})

	return m
}

func convertCookie(cookies *fasthttp.ResponseHeader) map[string]string {
	m := make(map[string]string)
	cookies.VisitAllCookie(func(key, value []byte) {
		m[string(key)] = string(value)
	})

	return m
}
