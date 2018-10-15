package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"golang.org/x/oauth2"
)

var token = oauth2.Token{
	AccessToken: os.Getenv("BITLY_ACCESS_TOKEN"),
}

func TestError(t *testing.T) {
	for _, v := range []struct {
		error      error
		message    string
		statusCode int
	}{{
		error:      nil,
		message:    "",
		statusCode: 200,
	}, {
		error:      fmt.Errorf("another error"),
		message:    "another error\n",
		statusCode: http.StatusInternalServerError,
	}, {
		error: clientError{
			Message: "client error",
			Code:    http.StatusBadGateway,
		},
		message:    "client error\n",
		statusCode: http.StatusBadGateway,
	}} {
		w := httptest.NewRecorder()
		if v.error != nil && v.error.Error() == "" {
			t.Fatal("empty error")
		}
		Error(w, v.error)
		resp := w.Result()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if v.message != string(body) {
			t.Errorf("unexpected body %q != %q", v.message, string(body))
		}
		if resp.StatusCode != v.statusCode {
			t.Errorf("unexpected status code %d != %d", resp.StatusCode, v.statusCode)
		}
	}
}

func TestClient(t *testing.T) {
	ctx := context.Background()
	client := New(ctx, &token, WithBaseURL(""))
	user, err := client.User(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if user.Name == "" {
		t.Error("did not get user name")
	}
	if client.Calls() != 1 {
		t.Errorf("unexpected number of api calls %d != 1", client.Calls())
	}

	links, err := client.BitlinksByGroup(ctx, user.DefaultGroupGUID, 10, 1)
	if err != nil {
		t.Fatal(err)
	}

	if client.Calls() != 2 {
		t.Errorf("unexpected number of api calls %d != 1", client.Calls())
	}

	if len(links.Links) == 0 {
		t.Fatal("didn't get any links")
	}

	link := links.Links[0]

	metrics, err := client.MetricsForBitlinkByCountry(ctx, link.ID, "day", 30)
	if err != nil {
		t.Fatal(err)
	}

	if metrics.Facet != "countries" {
		t.Error("did not get metrics")
	}
}
