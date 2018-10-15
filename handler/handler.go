package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2"
	"jrubin.io/bl/client"
)

// Result represents the average number of clicks that all links received from a
// given country
type Result struct {
	Average float32 `json:"average"`
	Country string  `json:"country"`
}

// Results contains the per-country results as well as the total number of links
// in the default group and the number of calls made to the bitly api
type Results struct {
	Total    int      `json:"total"`
	APICalls int      `json:"api_calls"`
	Results  []Result `json:"results"`
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

const timeout = 1 * time.Minute

// Handler processes a request to provide the average user clicks per country
// over the last 30 days for a user's default group
func Handler(workers int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)

		if token.TokenType == "" || token.AccessToken == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		unit := "day"
		if u := r.URL.Query().Get("unit"); u != "" {
			unit = u
		}

		units := 30
		if u, err := strconv.Atoi(r.URL.Query().Get("units")); err == nil {
			units = int(u)
		}

		// set an upper limit on request duration
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		c := client.New(ctx, token)

		user, err := c.User(ctx)
		if err != nil {
			client.Error(w, err)
			return
		}

		var wg sync.WaitGroup
		var processErr error
		data := map[string]int{}
		results := processMetrics(data, &processErr, &wg)

		const size = 10

		// prep the worker pool
		jobs := make(chan string)
		for w := 0; w < workers; w++ {
			go worker(ctx, c, results, unit, units, jobs)
		}

		var ret Results
		for page := 1; ; page++ {
			var links *client.Links
			if links, err = c.BitlinksByGroup(ctx, user.DefaultGroupGUID, size, page); err != nil {
				client.Error(w, err)
				return
			}

			for _, link := range links.Links {
				ret.Total++
				wg.Add(1)
				jobs <- link.ID
			}

			if links.Pagination.Next == "" || true {
				break
			}
		}

		close(jobs) // ensure the workers return

		// set up an event that can be monitored for processing completion
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(results) // ensure the processor func returns
			close(done)
		}()

		// wait for either a timeout or processing completion
		select {
		case <-ctx.Done():
			http.Error(w, ctx.Err().Error(), http.StatusGatewayTimeout)
			return
		case <-done:
		}

		if processErr != nil {
			client.Error(w, processErr)
			return
		}

		// build a sorted list of countries
		countries := make([]string, 0, len(data))
		for country := range data {
			countries = append(countries, country)
		}
		sort.Strings(countries)

		// calculate the average clicks
		ret.Results = make([]Result, 0, len(data))
		for _, country := range countries {
			clicks := data[country]
			ret.Results = append(ret.Results, Result{
				Average: float32(clicks) / float32(ret.Total),
				Country: country,
			})
		}

		ret.APICalls = c.Calls()

		if err = json.NewEncoder(w).Encode(ret); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

type metricsOrError struct {
	Metrics []client.Metric
	Error   error
}

func processMetrics(data map[string]int, err *error, wg *sync.WaitGroup) chan<- metricsOrError {
	results := make(chan metricsOrError)
	go func() {
		// don't check for context cancellation as workers already do that
		// all results need to be processed to ensue no leaked worker goroutines
		for metrics := range results {
			if metrics.Error != nil {
				*err = metrics.Error
				wg.Done()
				continue
			}

			for _, metric := range metrics.Metrics {
				data[metric.Value] += metric.Clicks
			}

			wg.Done()
		}
	}()
	return results
}

func worker(ctx context.Context, c client.Client, results chan<- metricsOrError, unit string, units int, jobs <-chan string) {
	for linkID := range jobs {
		// every job needs to be processed, if ctx is cancelled, the request
		// will return an error
		metrics, err := c.MetricsForBitlinkByCountry(ctx, linkID, unit, units)
		if err != nil {
			results <- metricsOrError{Error: err}
			continue
		}

		results <- metricsOrError{Metrics: metrics.Metrics}
	}
}
