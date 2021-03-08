package merging

import (
	"encoding/json"
	"regexp"
	"strconv"

	"github.com/valyala/fastjson"
)

var floatRegex = regexp.MustCompile("^[+-]?([0-9]+[.][0-9]+)$")
var intRegex = regexp.MustCompile("^[+-]?([0-9]+)$")

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
			if floatRegex.MatchString(v.Get(field).String()) {
				m.data[field] = v.GetInt64(field)
			}

			if intRegex.MatchString(v.Get(field).String()) {
				m.data[field] = v.GetFloat64(field)
			}
		case fastjson.TypeTrue, fastjson.TypeFalse:
			m.data[field] = v.GetBool(field)
		case fastjson.TypeArray:
			slice := v.GetArray(field)
			if len(slice) == 0 {
				m.data[field] = []string{}
			} else {
				// This function is used to solve problems in FastJSON GetArray,
				// values cannot be unmarshaled when using a FastJSON GetArray.
				switch sliceCheckType(slice[0]) {
				case "string":
					m.data[field] = sliceString(slice)
				case "integer":
					m.data[field] = sliceInteger(slice)
				case "float":
					m.data[field] = sliceFloat(slice)
				case "boolean":
					m.data[field] = sliceBoolean(slice)
				case "object":
					m.data[field] = sliceObject(slice)
				}
			}
		case fastjson.TypeObject:
			// Same with fastjson.GetArray,
			// this function has the same problem after unmarshalled.
			m.data[field] = objects(v.GetObject(field))
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

func sliceCheckType(value *fastjson.Value) string {
	switch value.Type() {
	case fastjson.TypeString:
		return "string"
	case fastjson.TypeNumber:
		if floatRegex.MatchString(value.String()) {
			return "float"
		} else if intRegex.MatchString(value.String()) {
			return "integer"
		}
	case fastjson.TypeTrue, fastjson.TypeFalse:
		return "boolean"
	case fastjson.TypeArray:
		return "array"
	case fastjson.TypeObject:
		return "object"
	}

	return ""
}

func sliceString(values []*fastjson.Value) []string {
	slices := make([]string, 0)
	for _, row := range values {
		s, _ := strconv.Unquote(row.String())
		slices = append(slices, s)
	}

	return slices
}

func sliceInteger(values []*fastjson.Value) []int {
	slices := make([]int, 0)
	for _, row := range values {
		num, _ := row.Int()
		slices = append(slices, num)
	}

	return slices
}

func sliceFloat(values []*fastjson.Value) []float64 {
	slices := make([]float64, 0)
	for _, row := range values {
		num, _ := row.Float64()
		slices = append(slices, num)
	}

	return slices
}

func sliceBoolean(values []*fastjson.Value) []bool {
	slices := make([]bool, 0)
	for _, row := range values {
		b, _ := row.Bool()
		slices = append(slices, b)
	}

	return slices
}

func sliceObject(values []*fastjson.Value) []interface{} {
	slices := make([]interface{}, 0)

	for _, row := range values {
		c := make(map[string]interface{}, 0)
		json.Unmarshal([]byte(row.String()), &c)

		slices = append(slices, c)
	}

	return slices
}

func objects(values *fastjson.Object) map[string]interface{} {
	slices := make(map[string]interface{}, 0)
	values.Visit(func(key []byte, v *fastjson.Value) {
		switch v.Type() {
		case fastjson.TypeString:
			s, _ := strconv.Unquote(v.String())
			slices[string(key)] = s
		case fastjson.TypeNumber:
			if floatRegex.MatchString(v.String()) {
				slices[string(key)] = v.GetFloat64()
			} else if intRegex.MatchString(v.String()) {
				slices[string(key)] = v.GetInt64()
			}
		case fastjson.TypeTrue, fastjson.TypeFalse:
			slices[string(key)] = v.GetBool()
		case fastjson.TypeArray:
			slices[string(key)] = v.GetArray()
		case fastjson.TypeObject:
			slices[string(key)] = objects(v.GetObject())
		}
	})

	return slices
}
