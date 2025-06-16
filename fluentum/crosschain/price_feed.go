package crosschain

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"net/http"
	"time"
)

// PriceFeed interface for getting gas prices and network conditions
type PriceFeed interface {
	GetGasPrice(ctx context.Context) (*big.Int, error)
	GetNetworkCongestion(ctx context.Context) (float64, error)
}

// EtherscanPriceFeed implements PriceFeed for Ethereum mainnet
type EtherscanPriceFeed struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewEtherscanPriceFeed creates a new Etherscan price feed
func NewEtherscanPriceFeed(apiKey string) *EtherscanPriceFeed {
	return &EtherscanPriceFeed{
		apiKey: apiKey,
		baseURL: "https://api.etherscan.io/api",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetGasPrice gets the current gas price from Etherscan
func (pf *EtherscanPriceFeed) GetGasPrice(ctx context.Context) (*big.Int, error) {
	url := pf.baseURL + "?module=gastracker&action=gasoracle&apikey=" + pf.apiKey

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pf.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  struct {
			SafeLow   string `json:"SafeGasPrice"`
			Standard  string `json:"ProposeGasPrice"`
			Fast      string `json:"FastGasPrice"`
		} `json:"result"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if result.Status != "1" {
		return nil, errors.New("etherscan API error: " + result.Message)
	}

	// Use standard gas price
	gasPrice := new(big.Int)
	gasPrice.SetString(result.Result.Standard, 10)
	return gasPrice, nil
}

// GetNetworkCongestion gets the current network congestion level
func (pf *EtherscanPriceFeed) GetNetworkCongestion(ctx context.Context) (float64, error) {
	url := pf.baseURL + "?module=gastracker&action=gasestimate&gasprice=20000000000&apikey=" + pf.apiKey

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := pf.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var result struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Result  string `json:"result"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return 0, err
	}

	if result.Status != "1" {
		return 0, errors.New("etherscan API error: " + result.Message)
	}

	// Calculate congestion based on gas estimate
	estimate := new(big.Int)
	estimate.SetString(result.Result, 10)

	// Normalize congestion to 0-1 range
	// Higher estimate means higher congestion
	maxEstimate := big.NewInt(1000000) // 1M gas
	congestion := new(big.Float).Quo(
		new(big.Float).SetInt(estimate),
		new(big.Float).SetInt(maxEstimate),
	)

	congestionFloat, _ := congestion.Float64()
	if congestionFloat > 1.0 {
		congestionFloat = 1.0
	}

	return congestionFloat, nil
} 