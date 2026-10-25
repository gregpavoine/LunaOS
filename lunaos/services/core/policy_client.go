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

func NewUnixPolicyClient(socket, ownerToken string) (*UnixPolicyClient, error) {
	if socket == "" || ownerToken == "" {
		return nil, fmt.Errorf("policy socket and owner token are required")
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 2 * time.Second}).DialContext(ctx, "unix", socket)
		},
		DisableKeepAlives: true,
	}
	return &UnixPolicyClient{
		socket: socket,
		token:  ownerToken,
		client: &http.Client{Transport: transport, Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
	}, nil
}

func (p *UnixPolicyClient) Authorize(ctx context.Context, goalID string, req Request) (Decision, error) {
	return p.call(ctx, "/v1/grants", brokerRequest{GoalID: goalID, Action: req.Action, PrivacyClass: req.PrivacyClass})
}

func (p *UnixPolicyClient) Evaluate(ctx context.Context, goalID string, req Request, grant string) (Decision, error) {
	return p.call(ctx, "/v1/policy/evaluate", brokerRequest{GoalID: goalID, Action: req.Action, PrivacyClass: req.PrivacyClass, Grant: grant})
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
