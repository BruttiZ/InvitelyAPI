package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type SupabaseClient struct {
	URL        string
	AnonKey    string
	ServiceKey string
	httpClient *http.Client
}

type SupabaseAuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int          `json:"expires_in"`
	User         SupabaseUser `json:"user"`
}

type SupabaseUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func NewSupabaseClient(url, anonKey string, serviceKey ...string) *SupabaseClient {
	key := ""
	if len(serviceKey) > 0 {
		key = serviceKey[0]
	}

	return &SupabaseClient{
		URL:        strings.TrimRight(url, "/"),
		AnonKey:    anonKey,
		ServiceKey: key,
		httpClient: &http.Client{Timeout: 12 * time.Second},
	}
}

func (c *SupabaseClient) Login(ctx context.Context, email, password string) (SupabaseAuthResponse, error) {
	var out SupabaseAuthResponse
	err := c.post(ctx, "/auth/v1/token?grant_type=password", c.AnonKey, map[string]string{
		"email":    email,
		"password": password,
	}, &out)

	return out, err
}

func (c *SupabaseClient) Register(ctx context.Context, email, password string) (SupabaseAuthResponse, error) {
	if c.ServiceKey != "" {
		var adminUser SupabaseUser
		if err := c.post(ctx, "/auth/v1/admin/users", c.ServiceKey, map[string]any{
			"email":         email,
			"password":      password,
			"email_confirm": true,
		}, &adminUser); err != nil {
			return SupabaseAuthResponse{}, err
		}

		return c.Login(ctx, email, password)
	}

	var out SupabaseAuthResponse
	err := c.post(ctx, "/auth/v1/signup", c.AnonKey, map[string]string{
		"email":    email,
		"password": password,
	}, &out)

	return out, err
}

func (c *SupabaseClient) User(ctx context.Context, token string) (SupabaseUser, error) {
	var out SupabaseUser
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.URL+"/auth/v1/user", nil)
	if err != nil {
		return out, err
	}
	req.Header.Set("apikey", c.AnonKey)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return out, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return out, fmt.Errorf("supabase auth returned status %d", res.StatusCode)
	}

	return out, json.NewDecoder(res.Body).Decode(&out)
}

func (c *SupabaseClient) post(ctx context.Context, path, key string, payload any, out any) error {
	if c.URL == "" || key == "" {
		return errors.New("supabase is not configured")
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", key)
	req.Header.Set("Authorization", "Bearer "+key)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		var errPayload struct {
			Message string `json:"msg"`
			Error   string `json:"error_description"`
		}
		_ = json.NewDecoder(res.Body).Decode(&errPayload)
		if errPayload.Message != "" {
			return errors.New(errPayload.Message)
		}
		if errPayload.Error != "" {
			return errors.New(errPayload.Error)
		}
		return fmt.Errorf("supabase returned status %d", res.StatusCode)
	}

	return json.NewDecoder(res.Body).Decode(out)
}
