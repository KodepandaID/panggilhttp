package test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KodepandaID/panggilhttp"
)

func BenchmarkMethodGET(*testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{"message":"ping"}`))
	}))
	defer ts.Close()

	client := panggilhttp.New()
	client.
		Get(ts.URL+"/ping", nil, nil).
		Do()
}

func BenchmarkMethodPOST(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusServiceUnavailable)
			b.Errorf("Expected ‘POST’ request, got ‘%s’", r.Method)
		}

		body, e := ioutil.ReadAll(r.Body)
		if e != nil {
			b.Fatal(e)
		}

		c := make(map[string]interface{})
		json.Unmarshal(body, &c)

		if c["username"] != "administrator" || c["password"] != "password" {
			w.WriteHeader(http.StatusBadRequest)
			b.Error("Request body not match")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := panggilhttp.New()

	resp, e := client.
		Post(ts.URL).
		SendJSON(map[string]interface{}{
			"username": "administrator",
			"password": "password",
		}).
		Do()
	if e != nil {
		b.Fatal(e)
	}

	if resp.StatusCode != http.StatusOK {
		b.Fatal("HTTP request failed")
	}
}

func BenchmarkMethodPOSTWithFormData(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusServiceUnavailable)
			b.Errorf("Expected ‘POST’ request, got ‘%s’", r.Method)
		}

		if e := r.ParseMultipartForm(r.ContentLength); e != nil {
			b.Error(e)
		}

		if r.FormValue("username") != "administrator" || r.FormValue("password") != "password" {
			w.WriteHeader(http.StatusBadRequest)
			b.Error("Request body not match")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := panggilhttp.New()

	resp, e := client.
		Post(ts.URL).
		SendFormData(map[string]string{
			"username": "administrator",
			"password": "password",
		}).
		Do()
	if e != nil {
		b.Fatal(e)
	}

	if resp.StatusCode != http.StatusOK {
		b.Fatal("HTTP request failed")
	}
}
