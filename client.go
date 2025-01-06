package kontrolerclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"time"
)

var contentTypeJSON = mime.TypeByExtension(".json")

type Client interface {
	CreateDag(Dag) error
	CreateDagRun(dagRun DagRun) (*CreateDagRunResult, error)
}

type CreateDagRunResult struct {
	RunId int `json:"runId"`
}

type client struct {
	url            string
	httpClient     *http.Client
	authCookieName string
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
	}

	return nil
}

func (c *client) CreateDag(dag Dag) error {
	if err := dag.Validate(); err != nil {
		return err
	}

	jsonData, err := json.Marshal(dag)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Post(fmt.Sprintf("%s/api/v1/dag/create", c.url), contentTypeJSON, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create DAG, status code: %d", resp.StatusCode)
	}

	return nil
}

func (c *client) CreateDagRun(dagRun DagRun) (*CreateDagRunResult, error) {
	jsonData, err := json.Marshal(dagRun)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(fmt.Sprintf("%s/api/v1/dag/run/create", c.url), contentTypeJSON, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create DAG RUN, status code: %d", resp.StatusCode)
	}

	var result CreateDagRunResult
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
