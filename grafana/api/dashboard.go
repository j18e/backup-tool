package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

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
