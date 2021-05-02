package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/harveywangdao/ants/logger"
)

const (
	KlineStateUp = iota
	KlineStateDown
	KlineStateFlat
)

func (g *GridStrategy) KlineState() error {
	// 1m 3m 5m 15m 30m 1h 2h 4h 6h 8h 12h 1d 3d 1w 1M
	klines, err := g.client.NewKlinesService().Symbol(g.Symbol).Interval("1m").Limit(6).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}
	n := len(klines)
	if n != 6 {
		return fmt.Errorf("klines count error")
	}
	// 最后一个k线不准确需要去掉
	n--
	klines = klines[:n]

	rates := make([]KlineData, n)
	down := 0
	for i := 0; i < n; i++ {
		openp, err := strconv.ParseFloat(klines[i].Open, 64)
		if err != nil {
			logger.Error(err)
			return err
		}
		closep, err := strconv.ParseFloat(klines[i].Close, 64)
		if err != nil {
			logger.Error(err)
			return err
		}
		rates[i].Rate = (closep - openp) / openp
		rates[i].Open = openp
		rates[i].Close = closep
		if rates[i].Rate < 0 {
			down++
		}
	}
	logger.Info("rates:", rates)
	logger.Info("down:", down)

	return nil
}
