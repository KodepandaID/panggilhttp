package test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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
