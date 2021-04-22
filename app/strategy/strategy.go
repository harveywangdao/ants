package main

import (
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

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Error(err)
			return err
		}
		logger.Error("ping fail, err:", string(body))
	}
	return nil
}

func do1() {
	sc, _ := NewStrategyClient(BaseEndpoint3)
	logger.Info(sc.Ping())
}

func main() {
	do1()
}
