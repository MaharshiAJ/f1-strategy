package openf1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const defaultBaseURL = "https://api.openf1.org/v1"

type Session struct {
	SessionKey  int       `json:"session_key"`
	StartDate   time.Time `json:"date_start"`
	CountryName string    `json:"country_name"`
	Location    string    `json:"location"`
	SessionName string    `json:"session_name"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type SessionsFilter struct {
	Year int
	Type string
}

func performRequest[T any](ctx context.Context, c *Client, endpoint string, query url.Values) ([]T, error) {
	u, err := url.Parse(c.baseURL)

	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}

	u = u.JoinPath(endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)

	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	req.URL.RawQuery = query.Encode()

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		reader := io.LimitReader(resp.Body, 200)
		body, _ := io.ReadAll(reader)
		return nil, fmt.Errorf("bad status %d: %s", resp.StatusCode, body)
	}

	var output []T

	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return nil, fmt.Errorf("decode output: %w", err)
	}

	return output, nil
}

func (c *Client) Sessions(ctx context.Context, f SessionsFilter) ([]Session, error) {
	q := url.Values{}
	q.Set("year", strconv.Itoa(f.Year))

	if f.Type != "" {
		q.Set("session_name", f.Type)
	}

	sessions, err := performRequest[Session](ctx, c, "sessions", q)

	if err != nil {
		return nil, fmt.Errorf("sessions: %w", err)
	}

	return sessions, nil
}
