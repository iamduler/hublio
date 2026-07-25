package nhanh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://pos.open.nhanh.vn/v3.0"
	defaultTimeout = 30 * time.Second
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}
	return &Client{baseURL: baseURL, httpClient: httpClient}
}

type auth struct {
	AppID       string
	BusinessID  string
	AccessToken string
}

func (c *Client) Ping(ctx context.Context, a auth) error {
	// product/list with empty filters is a cheap authenticated call.
	var out apiResponse
	if err := c.post(ctx, "/product/list", a, map[string]any{
		"filters":   map[string]any{},
		"paginator": map[string]any{"size": 1},
	}, &out); err != nil {
		return err
	}
	return checkAPI(out)
}

func (c *Client) GetRetailBill(ctx context.Context, a auth, billID int64) (*retailBill, error) {
	var out apiResponse
	if err := c.post(ctx, "/bill/retail", a, map[string]any{
		"filters":   map[string]any{"id": billID},
		"paginator": map[string]any{"size": 1},
	}, &out); err != nil {
		return nil, err
	}
	if err := checkAPI(out); err != nil {
		return nil, err
	}
	bills, err := decodeBills(out.Data)
	if err != nil {
		return nil, err
	}
	if len(bills) == 0 {
		return nil, ErrNotFound
	}
	return &bills[0], nil
}

type listBillsFilter struct {
	FromDate string
	PageSize int
}

func (c *Client) ListRetailBills(ctx context.Context, a auth, f listBillsFilter) ([]retailBill, error) {
	filters := map[string]any{}
	if strings.TrimSpace(f.FromDate) != "" {
		// Overlap-friendly lower bound; provider date formats vary — pass through as-is.
		filters["fromDate"] = f.FromDate
	}
	size := f.PageSize
	if size <= 0 {
		size = 50
	}
	var out apiResponse
	if err := c.post(ctx, "/bill/retail", a, map[string]any{
		"filters":   filters,
		"paginator": map[string]any{"size": size},
	}, &out); err != nil {
		return nil, err
	}
	if err := checkAPI(out); err != nil {
		return nil, err
	}
	return decodeBills(out.Data)
}

func (c *Client) UpdateOrderStatus(ctx context.Context, a auth, orderID int64, status int) error {
	var out apiResponse
	if err := c.post(ctx, "/order/edit", a, map[string]any{
		"orders": []map[string]any{{
			"id":     orderID,
			"status": status,
		}},
	}, &out); err != nil {
		return err
	}
	return checkAPI(out)
}

func (c *Client) post(ctx context.Context, path string, a auth, body any, out *apiResponse) error {
	q := url.Values{}
	q.Set("appId", a.AppID)
	q.Set("businessId", a.BusinessID)
	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("nhanh: marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path+"?"+q.Encode(), bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", a.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("nhanh: http: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("nhanh: read: %w", err)
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ErrAuthFailed
	}
	if resp.StatusCode >= 500 {
		return providerError(fmt.Sprintf("http_%d", resp.StatusCode))
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("nhanh: decode: %w", err)
	}
	return nil
}

func checkAPI(out apiResponse) error {
	if out.Code == 1 {
		return nil
	}
	switch out.ErrorCode {
	case "ERR_INVALID_ACCESS_TOKEN", "ERR_INVALID_APP_ID", "ERR_INVALID_BUSINESS_ID":
		return fmt.Errorf("%w: %s", ErrAuthFailed, out.ErrorCode)
	case "":
		return ErrProviderRejected
	default:
		return providerError(out.ErrorCode)
	}
}

func decodeBills(data any) ([]retailBill, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	var list []retailBill
	if err := json.Unmarshal(raw, &list); err == nil {
		return list, nil
	}
	// Some responses wrap list under a key.
	var wrap map[string]any
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return nil, fmt.Errorf("nhanh: unexpected bill data shape")
	}
	return nil, ErrNotFound
}

func parseAuth(config, secret map[string]any) (auth, string, error) {
	a := auth{
		AppID:       anyToString(first(config, "app_id", "appId")),
		BusinessID:  anyToString(first(config, "business_id", "businessId")),
		AccessToken: stringField(secret, "access_token", "accessToken", "token"),
	}
	baseURL := stringField(config, "base_url", "baseUrl")
	if a.AppID == "" || a.BusinessID == "" || a.AccessToken == "" {
		return auth{}, "", ErrMissingCredentials
	}
	return a, baseURL, nil
}

func first(m map[string]any, keys ...string) any {
	if m == nil {
		return nil
	}
	for _, k := range keys {
		if v, ok := m[k]; ok {
			return v
		}
	}
	return nil
}

func stringField(m map[string]any, keys ...string) string {
	return anyToString(first(m, keys...))
}

func anyToString(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case json.Number:
		return t.String()
	default:
		return strings.TrimSpace(fmt.Sprint(t))
	}
}

func int64Field(m map[string]any, keys ...string) int64 {
	v := first(m, keys...)
	switch t := v.(type) {
	case float64:
		return int64(t)
	case int64:
		return t
	case int:
		return int64(t)
	case string:
		n, _ := strconv.ParseInt(strings.TrimSpace(t), 10, 64)
		return n
	case json.Number:
		n, _ := t.Int64()
		return n
	default:
		return 0
	}
}
