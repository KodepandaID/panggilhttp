package test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
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

func BenchmarkMethodPOSTWithSendFile(b *testing.B) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusServiceUnavailable)
			b.Errorf("Expected ‘POST’ request, got ‘%s’", r.Method)
		}

		if e := r.ParseForm(); e != nil {
			b.Error(e)
		}

		f, h, e := r.FormFile("image")
		if e != nil {
			b.Error(e)
		}
		defer f.Close()

		// I want to check if the file size same as my local file in /assets/person.jpg.
		// Local filesize is 39KB
		kbSize := h.Size / 1000

		// To check mime-type, the mime-type of the local file is image/jpeg
		fileHeader := make([]byte, h.Size)
		f.Read(fileHeader)
		mime := http.DetectContentType(fileHeader)

		if h.Filename != "person.jpg" || kbSize != 39 || mime != "image/jpeg" {
			b.Error("File not match")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wd, _ := os.Getwd()
	f, e := ioutil.ReadFile(wd + "/assets/person.jpg")
	if e != nil {
		b.Fatal(e)
	}

	client := panggilhttp.New()

	resp, e := client.
		Post(ts.URL).
		SendFile("image", "person.jpg", f).
		Do()
	if e != nil {
		b.Fatal(e)
	}

	if resp.StatusCode != http.StatusOK {
		b.Fatal("HTTP request failed")
	}
}
