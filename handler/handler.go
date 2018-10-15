package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"jrubin.io/bl/client"
)

// Result represents the average number of clicks that all links received from a
// given country
type Result struct {
	Link                client.Link `json:"link"`
	Error               string      `json:"error,omitempty"`
	Unit                string      `json:"unit"`
	Units               int         `json:"units"`
	UnitReference       string      `json:"unit_reference"`
	AvgClicksPerCountry float32     `json:"avg_clicks_per_country"`
}

// Results contains the per-country results as well as the total number of links
// in the default group and the number of calls made to the bitly api
type Results struct {
	Pagination client.Pagination `json:"pagination"`
	APICalls   int               `json:"api_calls"`
	Results    []Result          `json:"results"`
}

func extractToken(r *http.Request) *oauth2.Token {
	var token oauth2.Token
	if auth := r.Header.Get("authorization"); auth != "" {
		if i := strings.Index(auth, " "); i != -1 && len(auth) > i+1 {
			token.TokenType = auth[:i]
			token.AccessToken = auth[i+1:]
		}
	}
	return &token
}

const timeout = 10 * time.Second

// Handler processes a request to provide the average user clicks per country
// over the last 30 days for a user's default group
func Handler(workers int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)

		if token.TokenType == "" || token.AccessToken == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// parse params

		unit := "day"
		if u := r.URL.Query().Get("unit"); u != "" {
			unit = u
		}

		units := 30
		if u, err := strconv.Atoi(r.URL.Query().Get("units")); err == nil {
			units = int(u)
		}

		size := 10
		if s, err := strconv.Atoi(r.URL.Query().Get("size")); err == nil {
			size = int(s)
		}

		page := 1
		if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
			page = int(p)
		}

		// set an upper limit on request duration
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		c := client.New(ctx, token)

		// fetch user info so we can get default group
		user, err := c.User(ctx)
		if err != nil {
			client.Error(w, err)
			return
		}

		// get the links in the default group
		var links *client.Links
		if links, err = c.BitlinksByGroup(ctx, user.DefaultGroupGUID, size, page); err != nil {
			client.Error(w, err)
			return
		}

		ret := Results{Pagination: links.Pagination}
		ret.Pagination.Prev = ""
		ret.Pagination.Next = ""

		// prep the worker pool
		results := make(chan Result)
		jobs := make(chan client.Link)
		for w := 0; w < workers; w++ {
			go worker(ctx, c, results, unit, units, jobs)
		}

		// queue all the jobs to fetch link info
		go func() {
			for _, link := range links.Links {
				jobs <- link
			}
		}()

		for i := 0; i < len(links.Links); i++ {
			ret.Results = append(ret.Results, <-results)
		}
		close(jobs) // ensure the workers return

		ret.APICalls = c.Calls()

		if err = json.NewEncoder(w).Encode(ret); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

func process(ctx context.Context, c client.Client, unit string, units int, link client.Link) Result {
	ret := Result{Link: link}

	// every job needs to be processed, if ctx is cancelled, the request
	// will return an error
	metrics, err := c.MetricsForBitlinkByCountry(ctx, link.ID, unit, units)
	if err != nil {
		ret.Error = err.Error()
		return ret
	}

	ret.Unit = metrics.Unit
	ret.Units = metrics.Units
	ret.UnitReference = metrics.UnitReference

	var clicks, countries int
	for _, m := range metrics.Metrics {
		clicks += m.Clicks
		countries++
	}

	if countries > 0 {
		ret.AvgClicksPerCountry = float32(clicks) / float32(countries)
	}

	return ret
}

func worker(ctx context.Context, c client.Client, results chan<- Result, unit string, units int, jobs <-chan client.Link) {
	for link := range jobs {
		results <- process(ctx, c, unit, units, link)
	}
}
