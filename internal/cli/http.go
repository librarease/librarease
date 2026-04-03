package cli

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type APIClient struct {
	baseURL string
	cfg     *Config
	http    *http.Client
}

func NewAPIClient(cfg *Config) *APIClient {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: cfg.InsecureSkipVerify,
		},
	}
	return &APIClient{
		baseURL: cfg.BaseURL,
		cfg:     cfg,
		http: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: tr,
		},
	}
}

func (c *APIClient) Do(method, path string, query map[string][]string, body any) (*http.Response, []byte, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, nil, err
	}

	q := u.Query()
	for k, vv := range query {
		for _, v := range vv {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, u.String(), bodyReader)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.cfg.Token))
	}
	if c.cfg.ClientID != "" {
		req.Header.Set("X-Client-Id", c.cfg.ClientID)
	}
	if c.cfg.UID != "" {
		req.Header.Set("X-Uid", c.cfg.UID)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, err
	}
	return resp, data, nil
}

func decodeEnvelope(data []byte) (*EnvelopeResponse, error) {
	var env EnvelopeResponse
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, err
	}
	return &env, nil
}

func httpError(resp *http.Response, data []byte) error {
	msg := strings.TrimSpace(string(data))
	if msg == "" {
		msg = http.StatusText(resp.StatusCode)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err == nil {
		if em, ok := m["error"].(string); ok && em != "" {
			msg = em
		}
	}
	return fmt.Errorf("http %d: %s", resp.StatusCode, msg)
}

func parseRFC3339(v string) (time.Time, error) {
	return time.Parse(time.RFC3339, v)
}

