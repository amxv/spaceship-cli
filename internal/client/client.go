package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://spaceship.dev/api"

type Client struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

type Problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

type APIError struct {
	Status      int
	Code        string
	OperationID string
	Problem     Problem
	Body        string
}

func (e *APIError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "spaceship api error (status %d)", e.Status)
	if e.Code != "" {
		fmt.Fprintf(&b, ", code %s", e.Code)
	}
	if e.Problem.Title != "" {
		fmt.Fprintf(&b, ": %s", e.Problem.Title)
	}
	if e.Problem.Detail != "" {
		fmt.Fprintf(&b, " - %s", e.Problem.Detail)
	}
	if e.Problem.Title == "" && e.Problem.Detail == "" && e.Body != "" {
		fmt.Fprintf(&b, ": %s", e.Body)
	}
	return b.String()
}

func New(apiKey, apiSecret string) *Client {
	return &Client{
		baseURL:   defaultBaseURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetDomainList(ctx context.Context, take, skip int, orderBy string) (map[string]any, error) {
	query := map[string]string{
		"take": fmt.Sprintf("%d", take),
		"skip": fmt.Sprintf("%d", skip),
	}
	if orderBy != "" {
		query["orderBy"] = orderBy
	}

	var resp map[string]any
	if err := c.doJSON(ctx, http.MethodGet, "/v1/domains", query, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetDomainInfo(ctx context.Context, domain string) (map[string]any, error) {
	var resp map[string]any
	if err := c.doJSON(ctx, http.MethodGet, "/v1/domains/"+domain, nil, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetResourceRecordsList(ctx context.Context, domain string, take, skip int, orderBy string) (map[string]any, error) {
	query := map[string]string{
		"take": fmt.Sprintf("%d", take),
		"skip": fmt.Sprintf("%d", skip),
	}
	if orderBy != "" {
		query["orderBy"] = orderBy
	}

	var resp map[string]any
	if err := c.doJSON(ctx, http.MethodGet, "/v1/dns/records/"+domain, query, nil, &resp); err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) SaveResourceRecords(ctx context.Context, domain string, payload map[string]any) error {
	return c.doJSON(ctx, http.MethodPut, "/v1/dns/records/"+domain, nil, payload, nil)
}

func (c *Client) DeleteResourceRecords(ctx context.Context, domain string, payload []map[string]any) error {
	return c.doJSON(ctx, http.MethodDelete, "/v1/dns/records/"+domain, nil, payload, nil)
}

func (c *Client) doJSON(ctx context.Context, method, apiPath string, query map[string]string, reqBody any, out any) error {
	base, err := url.Parse(c.baseURL)
	if err != nil {
		return fmt.Errorf("invalid base url: %w", err)
	}

	base.Path = strings.TrimRight(base.Path, "/") + "/" + strings.TrimLeft(apiPath, "/")
	q := base.Query()
	for key, value := range query {
		q.Set(key, value)
	}
	base.RawQuery = q.Encode()

	var body io.Reader
	if reqBody != nil {
		buf := bytes.NewBuffer(nil)
		if err := json.NewEncoder(buf).Encode(reqBody); err != nil {
			return fmt.Errorf("failed to encode request: %w", err)
		}
		body = buf
	}

	req, err := http.NewRequestWithContext(ctx, method, base.String(), body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("X-API-Secret", c.apiSecret)
	if reqBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 300 {
		return decodeAPIError(resp)
	}

	if out == nil || resp.StatusCode == http.StatusNoContent {
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

func decodeAPIError(resp *http.Response) error {
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	apiErr := &APIError{
		Status:      resp.StatusCode,
		Code:        resp.Header.Get("spaceship-error-code"),
		OperationID: resp.Header.Get("spaceship-operation-id"),
		Body:        strings.TrimSpace(string(bodyBytes)),
	}

	var problem Problem
	if len(bodyBytes) > 0 {
		_ = json.Unmarshal(bodyBytes, &problem)
		apiErr.Problem = problem
	}

	return apiErr
}
