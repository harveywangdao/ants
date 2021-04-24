package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/harveywangdao/ants/logger"
)

func init() {
	logger.SetHandlers(logger.Console)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	logger.SetLevel(logger.INFO)
}

/*
https://vapi.binance.com 欧式期权 /vapi/v1/ping
https://dapi.binance.com 币本位合约 /dapi/v1/ping
https://fapi.binance.com U本位合约 /fapi/v1/ping
https://api.binance.com 现货/杠杆/币安宝/矿池 /api/v3/ping
*/

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
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	sc := &StrategyClient{
		endpoint: endpoint,
		client: &http.Client{
			Transport: transport,
		},

		apikey:    "xx",
		secretkey: "xx",
	}
	return sc, nil
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

func (s *StrategyClient) Ping() error {
	body, err := s.GET("/api/v1/ping")
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info(string(body))
	return nil
}

func (s *StrategyClient) Time() error {
	body, err := s.GET("/api/v1/time")
	if err != nil {
		logger.Error(err)
		return err
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
	body, err := s.GET("/api/v1/exchangeInfo")
	if err != nil {
		logger.Error(err)
		return err
	}

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
		if strings.Contains(ei.Symbols[i].Symbol, "BTC") {
			logger.Info(ei.Symbols[i])
		}
	}

	return nil
}

func (s *StrategyClient) Depth() error {
	body, err := s.GET("/api/v1/depth?symbol=ETHBTC&limit=10")
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info(string(body))
	return nil
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
	sc, err := NewStrategyClient(BaseEndpoint)
	if err != nil {
		logger.Fatal(err)
		return
	}

	sc.Ping()
	sc.Time()
	sc.GetExchangeInfo()
	sc.Depth()
	sc.QueryTrades()
	sc.Klines()
	sc.AvgPrice()
}

func main() {
	do1()
}
