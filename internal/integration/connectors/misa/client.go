package misa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://testapi.meinvoice.vn/api/integration"
	defaultTimeout = 30 * time.Second
)

// Client talks to MISA meInvoice Open API. baseURL and httpClient are injectable so tests
// can point at httptest without hitting the real sandbox.
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

func (c *Client) GetToken(ctx context.Context, appID, taxCode, username, password string) (string, error) {
	body := tokenRequest{
		AppID:    appID,
		TaxCode:  taxCode,
		Username: username,
		Password: password,
	}
	var env apiEnvelope
	if err := c.postJSON(ctx, "/auth/token", "", taxCode, body, &env); err != nil {
		return "", err
	}
	if !env.Success {
		return "", authError(env.ErrorCode)
	}
	token, ok := env.Data.(string)
	if !ok || strings.TrimSpace(token) == "" {
		return "", ErrAuthFailed
	}
	return token, nil
}

// ListTemplates is a lightweight authenticated call used by Health.
func (c *Client) ListTemplates(ctx context.Context, token, taxCode string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/invoice/templates?invoiceWithCode=true&ticket=false", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("CompanyTaxCode", taxCode)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("misa: templates request: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode >= 400 {
		return providerError(fmt.Sprintf("http_%d", resp.StatusCode), "")
	}
	var env apiEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		// Some environments return a bare array; treat 2xx as healthy.
		return nil
	}
	if env.ErrorCode != "" && !env.Success {
		return providerError(env.ErrorCode, "")
	}
	return nil
}

func (c *Client) CreateInvoice(ctx context.Context, token, taxCode string, reqBody createInvoiceRequest) (*publishResponse, error) {
	var out publishResponse
	if err := c.postJSON(ctx, "/invoice", token, taxCode, reqBody, &out); err != nil {
		return nil, err
	}
	if out.ErrorCode != nil && *out.ErrorCode != "" {
		detail := ""
		if out.DescriptionErrorCode != nil {
			detail = *out.DescriptionErrorCode
		}
		return nil, providerError(*out.ErrorCode, detail)
	}
	for _, r := range out.PublishInvoiceResult {
		if r.ErrorCode != "" {
			return nil, providerError(r.ErrorCode, r.TransactionID)
		}
	}
	for _, r := range out.CreateInvoiceResult {
		if code, _ := r["ErrorCode"].(string); code != "" {
			return nil, providerError(code, "")
		}
	}
	if !out.Success && len(out.PublishInvoiceResult) == 0 && len(out.CreateInvoiceResult) == 0 {
		return nil, ErrProviderRejected
	}
	return &out, nil
}

func (c *Client) postJSON(ctx context.Context, path, bearerToken, taxCode string, body any, out any) error {
	raw, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("misa: marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(raw))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	if taxCode != "" {
		req.Header.Set("CompanyTaxCode", taxCode)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("misa: http: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("misa: read response: %w", err)
	}
	if resp.StatusCode >= 500 {
		return providerError(fmt.Sprintf("http_%d", resp.StatusCode), "")
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return ErrAuthFailed
	}
	if err := json.Unmarshal(respBody, out); err != nil {
		return fmt.Errorf("misa: decode response: %w", err)
	}
	return nil
}
