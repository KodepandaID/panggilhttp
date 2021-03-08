package test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/KodepandaID/panggilhttp"
	"github.com/stretchr/testify/assert"
)

func TestMethodGET(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{"message":"ping"}`))
	}))
	defer ts.Close()

	client := panggilhttp.New()

	_, e := client.
		Get(ts.URL, nil, nil).
		Do()
	if e != nil {
		t.Fatal(e)
	}
}

func TestMethodGETWithMerge(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/hotel-destination" {
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte(`{"id_hotel": 25,"name": "Hotel California","destination_id": 123}`))
		} else if r.URL.Path == "/destinations" {
			w.WriteHeader(http.StatusOK)
			w.Header().Add("Content-Type", "application/json")
			w.Write([]byte(`{"destination_id": 123,"destinations": ["LAX", "SFO", "OAK"], "flights": [{"plane": "ABC", "departured": "09:00"}, {"plane": "DEF", "departured": "07:00"}], "informations": {"total_population": 11000, "total_land_area": 120000, "average_temperatures": {"morning": "20c", "night": "13c"}}}`))
		}
	}))
	defer ts.Close()

	client := panggilhttp.New()

	resp, e := client.
		Get(ts.URL+"/hotel-destination", nil, nil).
		Get(ts.URL+"/destinations", []string{"flights", "informations"}, nil).
		Do()
	if e != nil {
		t.Fatal(e)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("HTTP request failed")
	}

	type Flights struct {
		Plane      string `json:"plane"`
		Departured string `json:"departured"`
	}

	type Informations struct {
		TotalPopulation     int `json:"total_population"`
		AverageTemperatures struct {
			Morning string `json:"morning"`
			Night   string `json:"night"`
		} `json:"average_temperatures"`
	}

	type Hotels struct {
		IDHotel       int          `json:"id_hotel"`
		DestinationID int          `json:"destination_id"`
		Name          string       `json:"name"`
		Destinations  []string     `json:"destinations"`
		Flights       []Flights    `json:"flights"`
		Informations  Informations `json:"informations"`
	}

	var h Hotels
	json.Unmarshal(resp.Body, &h)

	assert.Equal(t, 25, h.IDHotel)
	assert.Equal(t, 123, h.DestinationID)
	assert.Equal(t, 2, len(h.Flights))
	assert.Equal(t, 11000, h.Informations.TotalPopulation)
	assert.Equal(t, "20c", h.Informations.AverageTemperatures.Morning)
}

func TestMethodPOST(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusServiceUnavailable)
			t.Errorf("Expected ‘POST’ request, got ‘%s’", r.Method)
		}

		body, e := ioutil.ReadAll(r.Body)
		if e != nil {
			t.Fatal(e)
		}

		c := make(map[string]interface{})
		json.Unmarshal(body, &c)

		if c["username"] != "administrator" || c["password"] != "password" {
			w.WriteHeader(http.StatusBadRequest)
			t.Error("Request body not match")
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
		t.Fatal(e)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("HTTP request failed")
	}
}

func TestMethodPOSTWithFormData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusServiceUnavailable)
			t.Errorf("Expected ‘POST’ request, got ‘%s’", r.Method)
		}

		if e := r.ParseMultipartForm(r.ContentLength); e != nil {
			t.Error(e)
		}

		if r.FormValue("username") != "administrator" || r.FormValue("password") != "password" {
			w.WriteHeader(http.StatusBadRequest)
			t.Error("Request body not match")
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
		t.Fatal(e)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("HTTP request failed")
	}
}

func TestMethodPOSTWithSendFile(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusServiceUnavailable)
			t.Errorf("Expected ‘POST’ request, got ‘%s’", r.Method)
		}

		if e := r.ParseForm(); e != nil {
			t.Error(e)
		}

		f, h, e := r.FormFile("image")
		if e != nil {
			t.Error(e)
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
			t.Error("File not match")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wd, _ := os.Getwd()
	f, e := ioutil.ReadFile(wd + "/assets/person.jpg")
	if e != nil {
		t.Fatal(e)
	}

	client := panggilhttp.New()

	resp, e := client.
		Post(ts.URL).
		SendFile("image", "person.jpg", f).
		Do()
	if e != nil {
		t.Fatal(e)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("HTTP request failed")
	}
}

func TestMethodPUT(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			w.WriteHeader(http.StatusServiceUnavailable)
			t.Errorf("Expected ‘PUT’ request, got ‘%s’", r.Method)
		}

		body, e := ioutil.ReadAll(r.Body)
		if e != nil {
			t.Fatal(e)
		}

		c := make(map[string]interface{})
		json.Unmarshal(body, &c)

		if c["username"] != "administrator" || c["password"] != "password" {
			w.WriteHeader(http.StatusBadRequest)
			t.Error("Request body not match")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := panggilhttp.New()

	resp, e := client.
		Put(ts.URL).
		SendJSON(map[string]interface{}{
			"username": "administrator",
			"password": "password",
		}).
		Do()
	if e != nil {
		t.Fatal(e)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("HTTP request failed")
	}
}

func TestMethodPATCH(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusServiceUnavailable)
			t.Errorf("Expected ‘PATCH’ request, got ‘%s’", r.Method)
		}

		body, e := ioutil.ReadAll(r.Body)
		if e != nil {
			t.Fatal(e)
		}

		c := make(map[string]interface{})
		json.Unmarshal(body, &c)

		if c["username"] != "administrator" || c["password"] != "password" {
			w.WriteHeader(http.StatusBadRequest)
			t.Error("Request body not match")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := panggilhttp.New()

	resp, e := client.
		Patch(ts.URL).
		SendJSON(map[string]interface{}{
			"username": "administrator",
			"password": "password",
		}).
		Do()
	if e != nil {
		t.Fatal(e)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("HTTP request failed")
	}
}

func TestMethodDELETE(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusServiceUnavailable)
			t.Errorf("Expected ‘DELETE’ request, got ‘%s’", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := panggilhttp.New()

	resp, e := client.
		Delete(ts.URL).
		Do()
	if e != nil {
		t.Fatal(e)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal("HTTP request failed")
	}
}

func TestRequestTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{"message":"ping"}`))
	}))
	defer ts.Close()

	client := panggilhttp.New()

	_, e := client.
		Get(ts.URL, nil, nil).
		Do()

	assert.Equal(t, e.Error(), "Request Timeout")
}

func TestWithHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "123456" {
			w.WriteHeader(http.StatusUnauthorized)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{"message":"ping"}`))
	}))
	defer ts.Close()

	client := panggilhttp.New()

	resp, e := client.
		WithHeader(map[string]string{
			"Authorization": "123456",
		}).
		Get(ts.URL, nil, nil).
		Do()
	if e != nil {
		t.Fatal(e)
	}

	assert.Equal(t, 200, resp.StatusCode)
}

func TestWithCookies(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, e := r.Cookie("Authorization")
		if e != nil {
			t.Fatal(e)
		}
		if cookie.Name != "Authorization" {
			w.WriteHeader(http.StatusUnauthorized)
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write([]byte(`{"message":"ping"}`))
	}))
	defer ts.Close()

	client := panggilhttp.New()

	resp, e := client.
		WithCookie(map[string]string{
			"Authorization": "123456",
		}).
		Get(ts.URL, nil, nil).
		Do()
	if e != nil {
		t.Fatal(e)
	}

	assert.Equal(t, 200, resp.StatusCode)
}

func TestWithTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
	}))
	defer ts.Close()

	client := panggilhttp.New()

	_, e := client.
		WithTimeout(2).
		Get(ts.URL, nil, nil).
		Do()

	assert.Equal(t, e.Error(), "Request Timeout")
}

func TestRetryFail(t *testing.T) {
	attempts := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		attempts++
	}))
	defer ts.Close()

	client := panggilhttp.New()

	client.
		Get(ts.URL, nil, nil).
		WithFailRetry(500, 3).
		Do()

	assert.Equal(t, 3, attempts)
}
