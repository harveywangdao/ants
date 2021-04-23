package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	//"strings"

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

	/*
	  BTCUSDT
	  ETHUSDT
	  BNBUSDT
	  LTCUSDT
	  EOSUSDT
	  ETCUSDT
	*/
	for i := 0; i < len(ei.Symbols); i++ {
		if ei.Symbols[i].Symbol == "BTCUSDT" {
			logger.Info(ei.Symbols[i])
		}
	}

	return nil
}

func (s *StrategyClient) Depth() error {
	req, err := http.NewRequest(http.MethodGet, s.endpoint+"/api/v1/depth?symbol=ETHBTC&limit=10", nil)
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

func (s *StrategyClient) GET(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, s.endpoint+url, nil)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http get fail, code: %d, err: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (s *StrategyClient) QueryTrades() error {
	data, err := s.GET("/api/v1/trades?symbol=BTCUSDT&limit=20")
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info(string(data))

	return nil
}

func (s *StrategyClient) Klines() error {
	now := time.Now().Unix()
	startTime := (now - 60*100) * 1000
	endTime := now * 1000

	/*
	 * 1m
	 * 3m
	 * 5m
	 * 15m
	 * 30m
	 * 1h
	 * 2h
	 * 4h
	 * 6h
	 * 8h
	 * 12h
	 * 1d
	 * 3d
	 * 1w
	 * 1M
	 */
	data, err := s.GET(fmt.Sprintf("/api/v1/klines?symbol=%s&interval=%s&startTime=%d&endTime=%d&limit=%d", "BTCUSDT", "1m", startTime, endTime, 20))
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info(string(data))

	return nil
}

func (s *StrategyClient) AvgPrice() error {
	data, err := s.GET(fmt.Sprintf("/api/v3/avgPrice?symbol=%s", "BTCUSDT"))
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info(string(data))

	return nil
}

func do1() {
	sc, _ := NewStrategyClient(BaseEndpoint)
	//sc.Ping()
	sc.Time()
	//sc.GetExchangeInfo()
	//sc.Depth()
	//sc.QueryTrades()
	sc.Klines()
	sc.AvgPrice()
}

func main() {
	do1()
}
