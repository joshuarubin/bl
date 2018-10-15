package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"golang.org/x/oauth2"
)

var token = oauth2.Token{
	AccessToken: os.Getenv("BITLY_ACCESS_TOKEN"),
	TokenType:   "Bearer",
}

func TestExtractToken(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	token.SetAuthHeader(req)

	if *extractToken(req) != token {
		t.Error("tokens don't match")
	}
}

func TestHandler(t *testing.T) {
	handler := Handler(2)

	for _, v := range []struct {
		url        string
		noAuth     bool
		statusCode int
		body       string
	}{{
		url:        "/",
		noAuth:     true,
		statusCode: http.StatusUnauthorized,
		body:       "Unauthorized\n",
	}, {
		url:        "/",
		statusCode: http.StatusOK,
	}, {
		url:        "/?unit=day&units=30&size=1&page=1",
		statusCode: http.StatusOK,
	}} {
		req, err := http.NewRequest("GET", v.url, nil)
		if err != nil {
			t.Fatal(err)
		}

		if !v.noAuth {
			token.SetAuthHeader(req)
		}

		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		resp := w.Result()

		if resp.StatusCode != v.statusCode {
			t.Errorf("unexpected status code %d != %d", resp.StatusCode, v.statusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		if v.body != "" && v.body != string(body) {
			t.Errorf("unexpected body %q != %q", v.body, string(body))
		}

		if v.statusCode != http.StatusOK {
			continue
		}

		if resp.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected content-type")
		}

		var results Results
		if err = json.Unmarshal(body, &results); err != nil {
			t.Fatal(err)
		}
	}
}
