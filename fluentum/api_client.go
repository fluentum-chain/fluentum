package fluentum

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// NodeStats represents the structure of the node stats response.
type NodeStats struct {
	BlockHeight        int     `json:"block_height"`
	Transactions24h    float64 `json:"transactions_24h"`
	ActiveValidators   int     `json:"active_validators"`
	NetworkUtilization float64 `json:"network_utilization"`
	AverageBlockTime   float64 `json:"average_block_time"`
	NetworkSecurity    float64 `json:"network_security"`
}

// FetchNodeStats fetches stats for node 1 from the Fluentum API using the provided API key.
func FetchNodeStats(apiKey string) (*NodeStats, error) {
	url := "https://www.fluentum.biz.id/api/stats/node/1"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var stats NodeStats
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}
