// Package neshan isolates the external v5 reverse-geocoding contract.
package neshan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var ErrUnavailable = errors.New("map_provider_unavailable")

type Result struct {
	Address  string `json:"formatted_address"`
	Province string `json:"state"`
	City     string `json:"city"`
	Status   string `json:"status"`
}
type Reverse struct {
	key, endpoint string
	client        *http.Client
}

func New(key string) *Reverse { return NewWithEndpoint(key, "https://api.neshan.org/v5/reverse", nil) }

// Endpoint override is dependency injection for isolated tests; never a request parameter.
func NewWithEndpoint(key, endpoint string, client *http.Client) *Reverse {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	}
	return &Reverse{key: key, endpoint: endpoint, client: client}
}
func (r *Reverse) Lookup(ctx context.Context, lat, lng float64) (Result, error) {
	u, e := url.Parse(r.endpoint)
	if e != nil {
		return Result{}, ErrUnavailable
	}
	q := u.Query()
	q.Set("lat", strconv.FormatFloat(lat, 'f', 6, 64))
	q.Set("lng", strconv.FormatFloat(lng, 'f', 6, 64))
	u.RawQuery = q.Encode()
	request, e := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if e != nil {
		return Result{}, ErrUnavailable
	}
	request.Header.Set("Api-Key", r.key)
	request.Header.Set("Accept", "application/json")
	response, e := r.client.Do(request)
	if e != nil {
		return Result{}, ErrUnavailable
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return Result{}, ErrUnavailable
	}
	body, e := io.ReadAll(io.LimitReader(response.Body, 32769))
	if e != nil || len(body) > 32768 {
		return Result{}, ErrUnavailable
	}
	var result Result
	if json.Unmarshal(body, &result) != nil || result.Status != "OK" || strings.TrimSpace(result.Address) == "" || len(result.Address) > 6000 || len(result.Province) > 400 || len(result.City) > 400 {
		return Result{}, ErrUnavailable
	}
	return result, nil
}
func (r *Reverse) String() string { return fmt.Sprint("Neshan reverse adapter (credentials private)") }
