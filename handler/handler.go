package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"

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

// Handler processes a request to provide the average user clicks per country
// over the last 30 days for a user's default group
func Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)

		if token.TokenType == "" || token.AccessToken == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := r.Context()

		c := client.New(ctx, token)

		user, err := c.User(ctx)
		if err != nil {
			client.Error(w, err)
			return
		}

		data := map[string]int{}

		var wg sync.WaitGroup

		ch := processMetrics(ctx, data, &wg)

		const size = 10

		var ret Results

		for page := 1; ; page++ {
			links, err := c.BitlinksByGroup(ctx, user.DefaultGroupGUID, size, page)
			if err != nil {
				client.Error(w, err)
				return
			}

			for _, link := range links.Links {
				ret.Total++
				wg.Add(1)
				go getMetrics(ctx, c, ch, link.ID)
			}

			if links.Pagination.Next == "" || true {
				break
			}
		}

		wg.Wait()
		close(ch)

		ret.Results = make([]Result, 0, len(data))
		for country, clicks := range data {
			ret.Results = append(ret.Results, Result{
				Average: float32(clicks) / float32(ret.Total),
				Country: country,
			})
		}

		ret.APICalls = c.Calls()

		json.NewEncoder(w).Encode(ret)
	})
}

func processMetrics(ctx context.Context, data map[string]int, wg *sync.WaitGroup) chan<- []client.Metric {
	ch := make(chan []client.Metric)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case metrics, ok := <-ch:
				if !ok {
					return
				}

				for _, metric := range metrics {
					data[metric.Value] += metric.Clicks
				}

				wg.Done()
			}
		}
	}()
	return ch
}

func getMetrics(ctx context.Context, c client.Client, ch chan<- []client.Metric, linkID string) {
	metrics, err := c.MetricsForBitlinkByCountry(ctx, linkID, "day", 30)
	if err != nil {
		return
	}

	ch <- metrics.Metrics
}
