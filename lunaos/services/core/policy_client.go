package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type UnixPolicyClient struct {
	socket string
	token  string
	client *http.Client
}

func NewUnixPolicyClient(socket, policyToken string) (*UnixPolicyClient, error) {
	if socket == "" || policyToken == "" {
		return nil, fmt.Errorf("policy socket and Core-to-Broker credential are required")
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 2 * time.Second}).DialContext(ctx, "unix", socket)
		},
		DisableKeepAlives: true,
	}
	return &UnixPolicyClient{
		socket: socket,
		token:  policyToken,
		client: &http.Client{Transport: transport, Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
	}, nil
}

func (p *UnixPolicyClient) Authorize(ctx context.Context, goalID string, req Request) (Decision, error) {
	return p.call(ctx, "/v1/grants", brokerRequest{GoalID: goalID, Action: req.Action, PrivacyClass: req.PrivacyClass})
}

func (p *UnixPolicyClient) Evaluate(ctx context.Context, goalID string, req Request, grant string) (Decision, error) {
	return p.call(ctx, "/v1/policy/evaluate", brokerRequest{GoalID: goalID, Action: req.Action, PrivacyClass: req.PrivacyClass, Grant: grant})
}

func (p *UnixPolicyClient) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://unix/v1/system/health", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	response, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("policy broker health returned HTTP %d", response.StatusCode)
	}
	decoder := json.NewDecoder(io.LimitReader(response.Body, 8193))
	decoder.DisallowUnknownFields()
	var result struct {
		Status string `json:"status"`
		Code   string `json:"code"`
	}
	if err := decoder.Decode(&result); err != nil {
		return fmt.Errorf("invalid policy broker health response: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("invalid trailing policy broker health data")
	}
	if result.Status != "ready" || result.Code != "BROKER_READY" {
		return fmt.Errorf("unexpected policy broker health code")
	}
	return nil
}

type brokerRequest struct {
	GoalID       string `json:"goal_id"`
	Action       string `json:"action"`
	PrivacyClass string `json:"privacy_class"`
	Grant        string `json:"grant,omitempty"`
}

func (p *UnixPolicyClient) call(ctx context.Context, path string, body brokerRequest) (Decision, error) {
	raw, err := json.Marshal(body)
	if err != nil {
		return Decision{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://unix"+path, bytes.NewReader(raw))
	if err != nil {
		return Decision{}, err
	}
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Content-Type", "application/json")
	response, err := p.client.Do(req)
	if err != nil {
		return Decision{}, err
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, 8193)
	decoder := json.NewDecoder(limited)
	decoder.DisallowUnknownFields()
	var decision Decision
	if err = decoder.Decode(&decision); err != nil {
		return Decision{}, fmt.Errorf("invalid policy response: %w", err)
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return Decision{}, fmt.Errorf("invalid trailing policy response data")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Decision{}, fmt.Errorf("policy returned HTTP %d", response.StatusCode)
	}
	if decision.Code == "" {
		return Decision{}, fmt.Errorf("policy response missing decision code")
	}
	return decision, nil
}
