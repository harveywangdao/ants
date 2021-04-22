package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/harveywangdao/ants/logger"
)

const (
	BaseEndpoint  = "https://api.binance.com"
	BaseEndpoint1 = "https://api1.binance.com"
	BaseEndpoint2 = "https://api2.binance.com"
	BaseEndpoint3 = "https://api3.binance.com"
)

type ErrResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type StrategyClient struct {
	endpoint string
	client   *http.Client

	apikey    string
	secretkey string
}

func NewStrategyClient(endpoint string) (*StrategyClient, error) {
	sc := &StrategyClient{
		endpoint: endpoint,
		client:   &http.Client{},

		apikey:    "xx",
		secretkey: "xx",
	}
	return sc, nil
}

func (s *StrategyClient) Ping() error {
	req, err := http.NewRequest(http.MethodGet, s.endpoint+"/api/v1/ping", nil)
	if err != nil {
		logger.Error(err)
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("ping fail, err:", string(body))
	}

	logger.Info(string(body))

	return nil
}

func (s *StrategyClient) Time() error {
	req, err := http.NewRequest(http.MethodGet, s.endpoint+"/api/v1/time", nil)
	if err != nil {
		logger.Error(err)
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("err:", string(body))
	}

	logger.Info(string(body))

	return nil
}

type RateLimit struct {
	RateLimitType string `json:"rateLimitType"`
	Interval      string `json:"interval"`
	IntervalNum   int    `json:"intervalNum"`
	Limit         int    `json:"limit"`
}

type FilterInfo struct {
	FilterType string `json:"filterType"`
	MinPrice   string `json:"minPrice"`
	MaxPrice   string `json:"maxPrice"`
	TickSize   string `json:"tickSize"`
}

type SymbolInfo struct {
	Symbol              string       `json:"symbol"`
	Status              string       `json:"status"`
	BaseAsset           string       `json:"baseAsset"`
	BaseAssetPrecision  int          `json:"baseAssetPrecision"`
	QuoteAsset          string       `json:"quoteAsset"`
	QuotePrecision      int          `json:"quotePrecision"`
	QuoteAssetPrecision int          `json:"quoteAssetPrecision"`
	Filters             []FilterInfo `json:"filters"`
}

type ExchangeInfo struct {
	RateLimits []RateLimit  `json:"rateLimits"`
	Symbols    []SymbolInfo `json:"symbols"`
}

func (s *StrategyClient) GetExchangeInfo() error {
	req, err := http.NewRequest(http.MethodGet, s.endpoint+"/api/v1/exchangeInfo", nil)
	if err != nil {
		logger.Error(err)
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("err:", string(body))
	}

	//logger.Info(string(body))

	ei := ExchangeInfo{}
	if err := json.Unmarshal(body, &ei); err != nil {
		logger.Error(err)
		return err
	}

	logger.Info(ei.RateLimits)
	logger.Info(len(ei.Symbols))

	for i := 0; i < 20; i++ {
		logger.Info(ei.Symbols[i])
	}

	return nil
}

func (s *StrategyClient) Depth() error {
	req, err := http.NewRequest(http.MethodGet, s.endpoint+"/api/v1/depth?symbol=ETHBTC&limit=100", nil)
	if err != nil {
		logger.Error(err)
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("err:", string(body))
	}

	logger.Info(string(body))

	return nil
}

func do1() {
	sc, _ := NewStrategyClient(BaseEndpoint)
	sc.Ping()
	sc.Time()
	sc.GetExchangeInfo()
	sc.Depth()
}

func main() {
	do1()
}
