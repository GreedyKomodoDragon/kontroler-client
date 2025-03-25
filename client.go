package kontrolerclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

var contentTypeJSON = mime.TypeByExtension(".json")

type HTTPError struct {
	StatusCode int
	URL        string
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s: HTTP %d: %s", e.URL, e.StatusCode, e.Message)
}

type Client interface {
	CreateDag(ctx context.Context, dag Dag) error
	CreateDagRun(ctx context.Context, dagRun DagRunCreate) (*CreateDagRunResult, error)
	GetTaskDetails(ctx context.Context, runId, taskId int) (*TaskRunDetails, error)
	GetDagRun(ctx context.Context, runId int) (*DagRunAll, error)
	GetDagRunDetails(ctx context.Context, runId int) (*DagRun, error)
	StreamPodLogs(ctx context.Context, podUID string, logChan chan<- string, errChan chan<- error) error
}

type CreateDagRunResult struct {
	RunId int `json:"runId"`
}

type client struct {
	url            string
	httpClient     *http.Client
	authCookieName string
	username       string
	password       string
}

type ClientConfig struct {
	Url            string         `json:"Url"`
	Username       string         `json:"Username"`
	Password       string         `json:"Password"`
	AuthCookieName string         `json:"AuthCookieName"`
	Timeout        *time.Duration `json:"Timeout"`
}

func NewClient(config *ClientConfig) (Client, error) {
	httpClient := &http.Client{}
	if config.Timeout != nil {
		httpClient.Timeout = *config.Timeout
	}

	client := &client{
		url:            config.Url,
		httpClient:     httpClient,
		authCookieName: config.AuthCookieName,
		username:       config.Username,
		password:       config.Password,
	}

	if err := client.login(config.Username, config.Password); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *client) login(username, password string) error {
	jsonData, err := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(fmt.Sprintf("%s/api/v1/auth/login", c.url), contentTypeJSON, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var authCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == c.authCookieName {
			authCookie = cookie
			break
		}
	}

	if authCookie == nil {
		return fmt.Errorf("%s cookie not found", c.authCookieName)
	}

	c.httpClient.Transport = &cookieTransport{
		cookie:    authCookie,
		transport: http.DefaultTransport,
		client:    c,
	}

	return nil
}

func handleResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			URL:        resp.Request.URL.String(),
			Message:    fmt.Sprintf("unexpected status code"),
		}
	}

	if result != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("reading response body: %w", err)
		}

		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}

	return nil
}

func (c *client) CreateDag(ctx context.Context, dag Dag) error {
	if err := dag.Validate(); err != nil {
		return fmt.Errorf("validating dag: %w", err)
	}

	jsonData, err := json.Marshal(dag)
	if err != nil {
		return fmt.Errorf("marshaling dag: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/dag/create", c.url), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", contentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}

	return handleResponse(resp, nil)
}

func (c *client) CreateDagRun(ctx context.Context, dagRun DagRunCreate) (*CreateDagRunResult, error) {
	jsonData, err := json.Marshal(dagRun)
	if err != nil {
		return nil, fmt.Errorf("marshaling dag run: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/dag/run/create", c.url), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", contentTypeJSON)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	var result CreateDagRunResult
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) GetTaskDetails(ctx context.Context, runId, taskId int) (*TaskRunDetails, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/dag/run/task/%d/%d", c.url, runId, taskId), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	var result TaskRunDetails
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) GetDagRun(ctx context.Context, runId int) (*DagRunAll, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/dag/run/all/%d", c.url, runId), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	var result DagRunAll
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) GetDagRunDetails(ctx context.Context, runId int) (*DagRun, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/dag/run/%d", c.url, runId), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	var result DagRun
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *client) StreamPodLogs(ctx context.Context, podUID string, logChan chan<- string, errChan chan<- error) error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("parsing URL: %w", err)
	}

	// Convert scheme from http(s) to ws(s)
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}

	u.Path = "/ws/logs"
	u.RawQuery = url.Values{"pod": []string{podUID}}.Encode()

	// Get auth cookie from transport
	header := http.Header{}
	if transport, ok := c.httpClient.Transport.(*cookieTransport); ok && transport.cookie != nil {
		header.Add("Cookie", transport.cookie.String())
	} else {
		return fmt.Errorf("no authentication cookie found")
	}

	// Use custom dialer with headers
	dialer := websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, u.String(), header)
	if err != nil {
		return fmt.Errorf("connecting to websocket: %w", err)
	}

	// Start reading messages in a goroutine
	go func() {
		defer conn.Close()
		defer close(logChan)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, message, err := conn.ReadMessage()
				if err != nil {
					errChan <- fmt.Errorf("reading websocket message: %w", err)
					return
				}
				logChan <- string(message)
			}
		}
	}()

	return nil
}
