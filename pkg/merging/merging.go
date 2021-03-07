package merging

import (
	"encoding/json"
	"strconv"

	"github.com/valyala/fastjson"
)

// Config is an adapter to merging the response body.
type Config struct {
	ResponseBody []byte
	data         map[string]interface{}
}

// New to create a new instance.
func New() *Config {
	return &Config{
		ResponseBody: nil,
		data:         make(map[string]interface{}),
	}
}

// Merge to merging all the response body.
func (m *Config) Merge(blacklist []string, b []byte) {
	c := make(map[string]interface{})
	json.Unmarshal(b, &c)

	for field, row := range c {
		if find(blacklist, field) == false {
			m.data[field] = row
		}
	}
}

// MergeFromWhitelist to merge response body from whitelist field.
func (m *Config) MergeFromWhitelist(whitelist []string, b []byte) {
	v, _ := fastjson.ParseBytes(b)
	for _, field := range whitelist {
		fieldType := v.Get(field).Type()

		switch fieldType {
		case fastjson.TypeString:
			s, _ := strconv.Unquote(v.Get(field).String())
			m.data[field] = s
		case fastjson.TypeNumber:
			tmpFloat := v.GetFloat64(field)
			if tmpFloat == 0 {
				m.data[field] = v.GetInt64(field)
			} else {
				m.data[field] = tmpFloat
			}
		case fastjson.TypeTrue, fastjson.TypeFalse:
			m.data[field] = v.GetBool(field)
		case fastjson.TypeArray:
			arr := v.GetArray(field)
			if len(arr) == 0 {
				m.data[field] = []string{}
			} else {
				m.data[field] = v.GetArray(field)
			}
		case fastjson.TypeObject:
			m.data[field] = v.GetObject(field)
		case fastjson.TypeNull:
			m.data[field] = nil
		}
	}
}

// Get to get response body byte
func (m *Config) Get() []byte {
	body, _ := json.Marshal(m.data)

	return body
}

func find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}

	return false
}
