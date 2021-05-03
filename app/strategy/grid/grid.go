package main

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/harveywangdao/ants/logger"
)

const (
	KlineStateUp = iota
	KlineStateDown
	KlineStateFlat
)

var (
	Tolerate = 0.1 // <0.5
)

type Operate int

func (o Operate) String() string {
	if o == WAIT {
		return "WAIT"
	} else if o == BUY {
		return "BUY"
	} else if o == SELL {
		return "SELL"
	} else {
		return "WAIT"
	}
}

const (
	WAIT Operate = iota
	BUY
	SELL
)

type KlineData struct {
	Open  float64
	Close float64
	High  float64
	Low   float64
	Rate  float64

	OpenTime  time.Time
	CloseTime time.Time
}

func (g *GridStrategy) KlineState(interval string, limit int) (Operate, error) {
	// 1m 3m 5m 15m 30m 1h 2h 4h 6h 8h 12h 1d 3d 1w 1M
	klines, err := g.client.NewKlinesService().Symbol(g.Symbol).Interval(interval).Limit(limit).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return WAIT, err
	}
	n := len(klines)
	if n != limit {
		return WAIT, fmt.Errorf("klines count error")
	}
	// 最后一个k线不准确需要去掉
	n--
	klines = klines[:n]

	rates := make([]KlineData, n)
	highest, lowest := 0.0, math.MaxFloat64
	highestIndex, lowestIndex := 0, 0
	down := 0
	for i := 0; i < n; i++ {
		openp, err := strconv.ParseFloat(klines[i].Open, 64)
		if err != nil {
			logger.Error(err)
			return WAIT, err
		}
		closep, err := strconv.ParseFloat(klines[i].Close, 64)
		if err != nil {
			logger.Error(err)
			return WAIT, err
		}
		highp, err := strconv.ParseFloat(klines[i].High, 64)
		if err != nil {
			logger.Error(err)
			return WAIT, err
		}
		lowp, err := strconv.ParseFloat(klines[i].Low, 64)
		if err != nil {
			logger.Error(err)
			return WAIT, err
		}
		rates[i].Rate = (closep - openp) / openp
		rates[i].High = highp
		rates[i].Low = lowp
		rates[i].Open = openp
		rates[i].Close = closep
		rates[i].OpenTime = time.Unix(klines[i].OpenTime/1000, 0)   //秒
		rates[i].CloseTime = time.Unix(klines[i].CloseTime/1000, 0) //秒
		if rates[i].Rate < 0 {
			down++
		}

		if rates[i].High > highest {
			highest = rates[i].High
			highestIndex = i + 1
		}
		if rates[i].Low < lowest {
			lowest = rates[i].Low
			lowestIndex = i + 1
		}
	}
	//totalRate := (rates[n-1].Close - rates[0].Open) / rates[0].Open
	//logger.Info("rates:", rates)
	//logger.Info("down:", down)
	//logger.Infof("totalRate: %f%%", totalRate*100)
	logger.Infof("highest: %f, highestIndex: %d", highest, highestIndex)
	logger.Infof("lowest: %f, lowestIndex: %d", lowest, lowestIndex)

	nowPrice, err := g.getNewestPrice()
	if err != nil {
		logger.Error(err)
		return WAIT, err
	}

	// 突破高
	if nowPrice > highest {
		logger.Infof("突破高 nowPrice:%f > highest:%f", nowPrice, highest)
		return BUY, nil
	}

	// 突破低
	if nowPrice < lowest {
		logger.Infof("突破低 nowPrice:%f < lowest:%f", nowPrice, lowest)
		return SELL, nil
	}

	if lowestIndex < highestIndex {
		// 1.最大值刚好是最新价格  -->  无
		if highestIndex == n {
			logger.Infof("highestIndex == n")
			return WAIT, nil
		}

		// 2.当前价格离最高价近    -->  卖
		//if (highest-nowPrice)/highest < Tolerate {
		if float64(n-highestIndex)/float64(n) < Tolerate {
			logger.Infof("下降初期 (n-highestIndex)/n:%f < Tolerate:%f", float64(n-highestIndex)/float64(n), Tolerate)
			return SELL, nil
		}

		// 3.当前价格离最低价近    -->  买
		//if (nowPrice-lowest)/(highest-lowest) <Tolerate
		/*if float64(n-highestIndex)/float64(n) > (0.5 - Tolerate) {
			return BUY, nil
		}*/

	} else if lowestIndex > highestIndex {
		// 1.最小值刚好是最新价格  -->  无
		if lowestIndex == n {
			logger.Infof("lowestIndex == n")
			return WAIT, nil
		}

		// 2.当前价格离最高价近    -->  卖
		/*if float64(n-lowestIndex)/float64(n) > (0.5 - Tolerate) {
			return SELL, nil
		}*/

		// 3.当前价格离最低价近    -->  买
		if float64(n-lowestIndex)/float64(n) < Tolerate {
			logger.Infof("上升初期 (n-lowestIndex)/n:%f < Tolerate:%f", float64(n-lowestIndex)/float64(n), Tolerate)
			return BUY, nil
		}

	}
	return WAIT, nil
}

func (g *GridStrategy) KlineState2(interval string, limit int) (Operate, error) {
	// 1m 3m 5m 15m 30m 1h 2h 4h 6h 8h 12h 1d 3d 1w 1M
	klines, err := g.client.NewKlinesService().Symbol(g.Symbol).Interval(interval).Limit(limit).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return WAIT, err
	}
	n := len(klines)
	if n != limit {
		return WAIT, fmt.Errorf("klines count error")
	}
	// 最后一个k线不准确需要去掉
	n--
	klines = klines[:n]

	rates := make([]KlineData, n)
	highest, lowest := 0.0, math.MaxFloat64
	highestIndex, lowestIndex := 0, 0
	down := 0
	for i := 0; i < n; i++ {
		openp, err := strconv.ParseFloat(klines[i].Open, 64)
		if err != nil {
			logger.Error(err)
			return WAIT, err
		}
		closep, err := strconv.ParseFloat(klines[i].Close, 64)
		if err != nil {
			logger.Error(err)
			return WAIT, err
		}
		highp, err := strconv.ParseFloat(klines[i].High, 64)
		if err != nil {
			logger.Error(err)
			return WAIT, err
		}
		lowp, err := strconv.ParseFloat(klines[i].Low, 64)
		if err != nil {
			logger.Error(err)
			return WAIT, err
		}
		rates[i].Rate = (closep - openp) / openp
		rates[i].High = highp
		rates[i].Low = lowp
		rates[i].Open = openp
		rates[i].Close = closep
		rates[i].OpenTime = time.Unix(klines[i].OpenTime/1000, 0)   //秒
		rates[i].CloseTime = time.Unix(klines[i].CloseTime/1000, 0) //秒
		if rates[i].Rate < 0 {
			down++
		}

		if rates[i].High > highest {
			highest = rates[i].High
			highestIndex = i + 1
		}
		if rates[i].Low < lowest {
			lowest = rates[i].Low
			lowestIndex = i + 1
		}
	}
	totalRate := (rates[n-1].Close - rates[0].Open) / rates[0].Open
	//logger.Info("rates:", rates)
	logger.Info("down:", down)
	logger.Infof("totalRate: %f%%", totalRate*100)
	logger.Infof("highest: %f, highestIndex: %d", highest, highestIndex)
	logger.Infof("lowest: %f, lowestIndex: %d", lowest, lowestIndex)

	return WAIT, nil
}
