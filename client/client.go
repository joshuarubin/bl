package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync/atomic"

	"golang.org/x/oauth2"
)

// Client for bitly api
type Client struct {
	client  *http.Client
	BaseURL string
}

// DefaultBaseURL is uses unless WithBaseURL is provided as an option to New
const DefaultBaseURL = "https://api-ssl.bitly.com/v4"

// Option valuess for New
type Option func(*Client)

// WithBaseURL overrides the default base url
func WithBaseURL(val string) Option {
	if val == "" {
		val = DefaultBaseURL
	}

	return func(c *Client) {
		c.BaseURL = val
	}
}

// New constructs a Client
func New(ctx context.Context, token *oauth2.Token, opts ...Option) Client {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	client.Transport = &transport{
		Transport: client.Transport,
	}

	ret := Client{
		client:  client,
		BaseURL: DefaultBaseURL,
	}

	for _, opt := range opts {
		opt(&ret)
	}

	return ret
}

// Calls returns the number of times this Client as made a call to the bitly api
func (c Client) Calls() int {
	return int(atomic.LoadInt32(&c.client.Transport.(*transport).Calls))
}

type clientError struct {
	Message string
	Code    int
}

func (e clientError) Error() string {
	return fmt.Sprintf("%s (%d)", e.Message, e.Code)
}

// Error writes the given error as the http response
func Error(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	if e, ok := err.(clientError); ok {
		http.Error(w, e.Message, e.Code)
		return
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (c Client) get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req = req.WithContext(ctx)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, clientError{
			Message: err.Error(),
			Code:    http.StatusBadGateway,
		}
	}

	if resp.StatusCode != http.StatusOK {
		var buf bytes.Buffer
		buf.ReadFrom(resp.Body) // #nosec
		resp.Body.Close()       // #nosec
		return nil, clientError{
			Message: buf.String(),
			Code:    resp.StatusCode,
		}
	}

	return resp, nil
}

// User retrives information for the current authenticated user
func (c Client) User(ctx context.Context) (*User, error) {
	resp, err := c.get(ctx, c.BaseURL+"/user")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var user User
	if err = json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// BitlinksByGroup retrieves a paginated collection of Bitlinks for a Group
func (c Client) BitlinksByGroup(ctx context.Context, groupGUID string, size, page int) (*Links, error) {
	url := fmt.Sprintf(
		"%s/groups/%s/bitlinks?page=%d&size=%d",
		c.BaseURL,
		url.PathEscape(groupGUID),
		page,
		size,
	)

	resp, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var links Links
	if err = json.NewDecoder(resp.Body).Decode(&links); err != nil {
		return nil, err
	}

	return &links, nil
}

// MetricsForBitlinkByCountry returns metrics about the countries referring
// click traffic to a single Bitlink
func (c Client) MetricsForBitlinkByCountry(ctx context.Context, bitlink string, unit string, units int) (*Metrics, error) {
	url := fmt.Sprintf(
		"%s/bitlinks/%s/countries?unit=%s&units=%d",
		c.BaseURL,
		url.PathEscape(bitlink),
		url.PathEscape(unit),
		units,
	)

	resp, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var metrics Metrics
	if err = json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, err
	}

	return &metrics, nil
}
