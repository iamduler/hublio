package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"hublio/internal/identity/application"
	"hublio/internal/identity/domain"
	"hublio/internal/platform/env"
)

// OAuthExchanger exchanges authorization codes with Google / Microsoft / GitHub.
type OAuthExchanger struct {
	httpClient *http.Client
	google     oauthClientConfig
	microsoft  oauthClientConfig
	github     oauthClientConfig
}

type oauthClientConfig struct {
	ClientID     string
	ClientSecret string
}

func NewOAuthExchanger() *OAuthExchanger {
	return &OAuthExchanger{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		google: oauthClientConfig{
			ClientID:     env.GetEnv("GOOGLE_OAUTH_CLIENT_ID", ""),
			ClientSecret: env.GetEnv("GOOGLE_OAUTH_CLIENT_SECRET", ""),
		},
		microsoft: oauthClientConfig{
			ClientID:     env.GetEnv("MICROSOFT_OAUTH_CLIENT_ID", ""),
			ClientSecret: env.GetEnv("MICROSOFT_OAUTH_CLIENT_SECRET", ""),
		},
		github: oauthClientConfig{
			ClientID:     env.GetEnv("GITHUB_OAUTH_CLIENT_ID", ""),
			ClientSecret: env.GetEnv("GITHUB_OAUTH_CLIENT_SECRET", ""),
		},
	}
}

func (e *OAuthExchanger) Configured(provider domain.OAuthProvider) bool {
	cfg := e.config(provider)
	return cfg.ClientID != "" && cfg.ClientSecret != ""
}

func (e *OAuthExchanger) config(provider domain.OAuthProvider) oauthClientConfig {
	switch provider {
	case domain.OAuthProviderGoogle:
		return e.google
	case domain.OAuthProviderMicrosoft:
		return e.microsoft
	case domain.OAuthProviderGitHub:
		return e.github
	default:
		return oauthClientConfig{}
	}
}

func (e *OAuthExchanger) Exchange(
	ctx context.Context,
	provider domain.OAuthProvider,
	code, codeVerifier, redirectURI string,
) (*application.OAuthProfile, error) {
	if !e.Configured(provider) {
		return nil, fmt.Errorf("oauth provider %s is not configured", provider)
	}
	switch provider {
	case domain.OAuthProviderGoogle:
		return e.exchangeGoogle(ctx, code, codeVerifier, redirectURI)
	case domain.OAuthProviderMicrosoft:
		return e.exchangeMicrosoft(ctx, code, codeVerifier, redirectURI)
	case domain.OAuthProviderGitHub:
		return e.exchangeGitHub(ctx, code, codeVerifier, redirectURI)
	default:
		return nil, domain.ErrInvalidOAuthProvider
	}
}

func (e *OAuthExchanger) exchangeGoogle(ctx context.Context, code, codeVerifier, redirectURI string) (*application.OAuthProfile, error) {
	token, err := e.tokenRequest(ctx, "https://oauth2.googleapis.com/token", url.Values{
		"client_id":     {e.google.ClientID},
		"client_secret": {e.google.ClientSecret},
		"code":          {code},
		"code_verifier": {codeVerifier},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	})
	if err != nil {
		return nil, err
	}
	var userinfo struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Name          string `json:"name"`
	}
	if err := e.getJSON(ctx, "https://openidconnect.googleapis.com/v1/userinfo", token.AccessToken, &userinfo); err != nil {
		return nil, err
	}
	return &application.OAuthProfile{
		Provider: domain.OAuthProviderGoogle,
		Subject:  userinfo.Sub,
		Email:    userinfo.Email,
		FullName: userinfo.Name,
		Verified: userinfo.EmailVerified,
	}, nil
}

func (e *OAuthExchanger) exchangeMicrosoft(ctx context.Context, code, codeVerifier, redirectURI string) (*application.OAuthProfile, error) {
	token, err := e.tokenRequest(ctx, "https://login.microsoftonline.com/common/oauth2/v2.0/token", url.Values{
		"client_id":     {e.microsoft.ClientID},
		"client_secret": {e.microsoft.ClientSecret},
		"code":          {code},
		"code_verifier": {codeVerifier},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {redirectURI},
	})
	if err != nil {
		return nil, err
	}
	var userinfo struct {
		Sub               string `json:"sub"`
		Email             string `json:"email"`
		PreferredUsername string `json:"preferred_username"`
		Name              string `json:"name"`
	}
	if err := e.getJSON(ctx, "https://graph.microsoft.com/oidc/userinfo", token.AccessToken, &userinfo); err != nil {
		return nil, err
	}
	email := userinfo.Email
	if email == "" {
		email = userinfo.PreferredUsername
	}
	return &application.OAuthProfile{
		Provider: domain.OAuthProviderMicrosoft,
		Subject:  userinfo.Sub,
		Email:    email,
		FullName: userinfo.Name,
		Verified: email != "",
	}, nil
}

func (e *OAuthExchanger) exchangeGitHub(ctx context.Context, code, codeVerifier, redirectURI string) (*application.OAuthProfile, error) {
	token, err := e.tokenRequest(ctx, "https://github.com/login/oauth/access_token", url.Values{
		"client_id":     {e.github.ClientID},
		"client_secret": {e.github.ClientSecret},
		"code":          {code},
		"code_verifier": {codeVerifier},
		"redirect_uri":  {redirectURI},
	})
	if err != nil {
		return nil, err
	}
	var user struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := e.getJSON(ctx, "https://api.github.com/user", token.AccessToken, &user); err != nil {
		return nil, err
	}
	email := user.Email
	verified := false
	if email == "" {
		var emails []struct {
			Email    string `json:"email"`
			Primary  bool   `json:"primary"`
			Verified bool   `json:"verified"`
		}
		if err := e.getJSON(ctx, "https://api.github.com/user/emails", token.AccessToken, &emails); err != nil {
			return nil, err
		}
		for _, item := range emails {
			if item.Primary && item.Verified {
				email = item.Email
				verified = true
				break
			}
		}
		if email == "" {
			for _, item := range emails {
				if item.Verified {
					email = item.Email
					verified = true
					break
				}
			}
		}
	} else {
		verified = true
	}
	fullName := user.Name
	if fullName == "" {
		fullName = user.Login
	}
	return &application.OAuthProfile{
		Provider: domain.OAuthProviderGitHub,
		Subject:  fmt.Sprintf("%d", user.ID),
		Email:    email,
		FullName: fullName,
		Verified: verified && email != "",
	}, nil
}

type oauthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error"`
	ErrorDesc   string `json:"error_description"`
}

func (e *OAuthExchanger) tokenRequest(ctx context.Context, endpoint string, form url.Values) (*oauthTokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	res, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	var token oauthTokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("invalid token response: %w", err)
	}
	if token.AccessToken == "" {
		msg := token.ErrorDesc
		if msg == "" {
			msg = token.Error
		}
		if msg == "" {
			msg = string(body)
		}
		return nil, fmt.Errorf("token exchange failed: %s", msg)
	}
	return &token, nil
}

func (e *OAuthExchanger) getJSON(ctx context.Context, endpoint, accessToken string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "hublio-oauth")
	res, err := e.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return err
	}
	if res.StatusCode >= 300 {
		return fmt.Errorf("provider userinfo failed: status %d", res.StatusCode)
	}
	return json.Unmarshal(body, dest)
}
