package grafana

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("component", "grafana")

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

// SearchDashborads performs a search for all dashboards stored in Grafana's
// database.
func (c *client) SearchDashboards() ([]*SearchResult, error) {
	logger.Debug("searching dashboards")
	var srx []*SearchResult
	res, err := c.doRequestWithAuth("/api/search?type=dash-db")
	if err != nil {
		return srx, err
	}
	defer res.Body.Close()

	if err := json.NewDecoder(res.Body).Decode(&srx); err != nil {
		return srx, fmt.Errorf("decoding response: %w", err)
	}

	return srx, nil
}

// GetDashboard retreives the byte string of the JSOn respresentation of a
// given dashboard (identified by UID).
func (c *client) GetDashboard(uid string) ([]byte, error) {
	logger.Debug("getting dashboard with uid " + uid)
	res, err := c.doRequestWithAuth("/api/dashboards/uid/" + uid)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	// read and return byte response
	bs, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return bs, fmt.Errorf("reading response: %w", err)
	}
	return bs, nil
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
