package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("component", "grafana-api")

// SearchResult is the response from a dashboard query in Grafana. Example JSON
// respresentation: [ { "id": 8, "uid": "sd", "title": "Some dashboard", "uri":
// "db/some-dashboard", "url": "/d/sd/some-dashboard", "slug": "", "type":
// "dash-db", "tags": [], "isStarred": false ]
type SearchResult struct {
	ID    int    `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
	URI   string `json:"uri"`
	URL   string `json:"url"`
}

// Client provides functionality for interacting with a Grafana server's API.
type Client interface {
	Ping() error
	SearchDashboards() ([]*SearchResult, error)
	GetDashboard(string) ([]byte, error)
}

type client struct {
	cli     http.Client
	token   string // Grafana API key with read permissions
	address string
}

// NewClient creates, tests and returns a Client.
func NewClient(addr, token string) (Client, error) {
	logger.Debug("creating new client")
	cli := &client{
		cli:     http.Client{Timeout: time.Second * 5},
		token:   token,
		address: addr,
	}
	if err := cli.Ping(); err != nil {
		return cli, fmt.Errorf("pinging %s: %w", cli.address, err)
	}
	return cli, nil
}

// Ping tests the connection to Grafana, verifying that the server is
// available. It does not verify that the client's token is functional.
func (c *client) Ping() error {
	logger.Debug("testing connection to Grafana")
	uri := c.address + "/api/health"
	res, err := c.cli.Get(uri)
	if err != nil {
		return fmt.Errorf("calling %s: %w", uri, err)
	}
	defer res.Body.Close()
	if res.StatusCode > 399 {
		return fmt.Errorf("calling %s got status: %d", uri, res.StatusCode)
	}
	return nil
}

func (c *client) doRequestWithAuth(path string) (*http.Response, error) {
	// create the request, add auth
	uri := c.address + path
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return &http.Response{}, fmt.Errorf("creating request to %s: %w", uri, err)
	}
	req.Header.Add("Authorization", "Bearer "+c.token)

	// do the request, check for errors
	res, err := c.cli.Do(req)
	if err != nil {
		return res, fmt.Errorf("requesting %s: %w", uri, err)
	} else if res.StatusCode > 399 {
		res.Body.Close()
		return res, fmt.Errorf("requesting %s: got status code %d", uri, res.StatusCode)
	}
	return res, nil
}
