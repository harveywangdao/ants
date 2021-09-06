package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adshao/go-binance/v2/common"
	"github.com/adshao/go-binance/v2/futures"
	"github.com/harveywangdao/ants/logger"
)

func init() {
	logger.SetHandlers(logger.Console)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)
	logger.SetLevel(logger.INFO)
}

const (
	ApiKey    = "KtnHyOpZzRgetPtTlxkdkCck1DlVumUvBUCEtSmAzxItVdXKigsugS11rteCRYLh"
	SecretKey = "TLo3hxjoGCVHBCyq6GUOA7QyoWCrV35VUOZaybVlVPfsHpJy6T2AkjlnH8mdFlJr"

	_interval = "1m"
	_limit    = 30
)

type GridStrategy struct {
	client       *futures.Client
	Symbol       string
	PositionSide string
	LongTrades   sync.Map
	ShortTrades  sync.Map

	HedgeLong  float64
	HedgeShort float64

	tradeInterval int64 //s

	maxDistance   int
	maxChunks     int
	profit        float64
	intervalPrice float64

	chunk             float64
	maxAmount         float64
	stopWin           float64
	stopLoss          float64
	longCloseOrderId  int64
	shortCloseOrderId int64
}

type TradeInfo struct {
	Type string

	OpenPrice   float64
	OpenAmount  float64
	OpenTime    time.Time
	OpenOrderId int64

	ClosePrice   float64
	CloseAmount  float64
	CloseTime    time.Time
	CloseOrderId int64
}

var (
	Fibonacci = []float64{1.0, 1.0, 2.0, 3.0, 5.0, 8.0, 13.0, 21.0, 34.0}
)

func (g *GridStrategy) DoLong() error {
	entryPrice, positionAmt, err := g.getEntryPriceAndAmt("LONG")
	if err != nil {
		return err
	}

	nowPrice, err := g.GetNewestPrice()
	if err != nil {
		logger.Error(err)
		return err
	}

	fmt.Printf("做多: entryPrice: %v, positionAmt: %v\n", entryPrice, positionAmt)
	if positionAmt > 0.0 {
		if nowPrice > entryPrice {
			profitRate := 100.0 * (nowPrice - entryPrice) / entryPrice
			fmt.Printf("做多: 盈利 %f USDT, 幅度:%f%%\n", (nowPrice-entryPrice)*positionAmt, profitRate)
		} else {
			lossRate := 100.0 * (nowPrice - entryPrice) / entryPrice
			fmt.Printf("做多: 亏损 %f USDT, 幅度:%f%%\n", (nowPrice-entryPrice)*positionAmt, lossRate)

			if positionAmt > 0.0 && -lossRate >= g.stopLoss {
				_, err := g.Trade(futures.SideTypeSell, 0, positionAmt, futures.PositionSideTypeLong)
				if err != nil {
					logger.Error(err)
					return err
				}
			}
		}
	}

	rates, err := g.GetKlines(g.Symbol, _interval, _limit+1)
	if err != nil {
		logger.Error(err)
		return err
	}
	op := g.bottomtop(rates[:len(rates)-1])
	fmt.Println("op:", op)

	nowPrice, err = g.GetNewestPrice()
	if err != nil {
		logger.Error(err)
		return err
	}

	if positionAmt > 0.0 {
		n := int(positionAmt / g.chunk)
		if n > len(Fibonacci)-1 {
			n = len(Fibonacci) - 1
		}

		nextPrice := entryPrice * (1 - (g.intervalPrice*Fibonacci[n])/100.0)
		fmt.Println("next price:", nextPrice)
		if positionAmt >= g.maxAmount || nowPrice > nextPrice {
			op = WAIT
		}
	}
	if op == BUY {
		_, err := g.Trade(futures.SideTypeBuy, 0, g.chunk, futures.PositionSideTypeLong)
		if err != nil {
			logger.Error(err)
			return err
		}
		g.longCloseOrderId = 0
	}

	if g.longCloseOrderId == 0 {
		entryPrice, positionAmt, err := g.getEntryPriceAndAmt("LONG")
		if err != nil {
			return err
		}
		if positionAmt > 0.0 {
			price := truncFloat(entryPrice*(1.0+g.stopWin/100.0), 5)
			order, err := g.TradeLimit(futures.SideTypeSell, price, positionAmt, futures.PositionSideTypeLong)
			if err != nil {
				apiError, ok := err.(*common.APIError)
				if ok && apiError.Code == 2022 {
					g.longCloseOrderId = 1
				}
				logger.Error(err)
				return err
			}
			g.longCloseOrderId = order.OrderID
		}
	}

	return nil
}

func (g *GridStrategy) DoShort() error {
	entryPrice, positionAmt, err := g.getEntryPriceAndAmt("SHORT")
	if err != nil {
		return err
	}

	nowPrice, err := g.GetNewestPrice()
	if err != nil {
		logger.Error(err)
		return err
	}

	positionAmt = -positionAmt
	fmt.Printf("做空: entryPrice: %v, positionAmt: %v\n", entryPrice, positionAmt)

	if positionAmt > 0.0 {
		if nowPrice > entryPrice {
			lossRate := 100.0 * (entryPrice - nowPrice) / entryPrice
			fmt.Printf("做空: 亏损 %f USDT, 幅度:%f%%\n", (entryPrice-nowPrice)*positionAmt, lossRate)

			if positionAmt > 0.0 && -lossRate >= g.stopLoss {
				_, err := g.Trade(futures.SideTypeBuy, 0, positionAmt, futures.PositionSideTypeShort)
				if err != nil {
					logger.Error(err)
					return err
				}
			}
		} else {
			profitRate := 100.0 * (entryPrice - nowPrice) / entryPrice
			fmt.Printf("做空: 盈利 %f USDT, 幅度:%f%%\n", (entryPrice-nowPrice)*positionAmt, profitRate)
		}
	}

	rates, err := g.GetKlines(g.Symbol, _interval, _limit+1)
	if err != nil {
		logger.Error(err)
		return err
	}
	op := g.bottomtop(rates[:len(rates)-1])
	fmt.Println("op:", op)

	nowPrice, err = g.GetNewestPrice()
	if err != nil {
		logger.Error(err)
		return err
	}

	if positionAmt > 0.0 {
		n := int(positionAmt / g.chunk)
		if n > len(Fibonacci)-1 {
			n = len(Fibonacci) - 1
		}

		nextPrice := entryPrice * (1 + (g.intervalPrice*Fibonacci[n])/100.0)
		fmt.Println("next price:", nextPrice)
		if positionAmt >= g.maxAmount || nowPrice < nextPrice {
			op = WAIT
		}
	}

	if op == SELL {
		_, err := g.Trade(futures.SideTypeSell, 0, g.chunk, futures.PositionSideTypeShort)
		if err != nil {
			logger.Error(err)
			return err
		}
		g.shortCloseOrderId = 0
	}

	if g.shortCloseOrderId == 0 {
		entryPrice, positionAmt, err := g.getEntryPriceAndAmt("SHORT")
		if err != nil {
			return err
		}
		positionAmt = -positionAmt
		if positionAmt > 0.0 {
			price := truncFloat(entryPrice*(1.0-g.stopWin/100.0), 5)
			order, err := g.TradeLimit(futures.SideTypeBuy, price, positionAmt, futures.PositionSideTypeShort)
			if err != nil {
				apiError, ok := err.(*common.APIError)
				if ok && apiError.Code == 2022 {
					g.shortCloseOrderId = 1
				}
				logger.Error(err)
				return err
			}
			g.shortCloseOrderId = order.OrderID
		}
	}

	return nil
}

func (g *GridStrategy) DoLong3() error {
	long, err := g.Position(futures.PositionSideType("LONG"))
	if err != nil {
		logger.Error(err)
		return err
	}
	entryPrice, err := strconv.ParseFloat(long.EntryPrice, 64)
	if err != nil {
		logger.Error(err)
		return err
	}
	positionAmt, err := strconv.ParseFloat(long.PositionAmt, 64)
	if err != nil {
		logger.Error(err)
		return err
	}

	nowPrice, err := g.GetNewestPrice()
	if err != nil {
		logger.Error(err)
		return err
	}

	fmt.Printf("做多: entryPrice: %v, positionAmt: %v\n", entryPrice, positionAmt)
	if nowPrice > entryPrice {
		profitRate := 100.0 * (nowPrice - entryPrice) / entryPrice
		fmt.Printf("做多: 盈利 %f USDT, 幅度:%f%%\n", (nowPrice-entryPrice)*positionAmt, profitRate)

		if positionAmt > 0.0 && profitRate >= g.stopWin {
			_, err := g.Trade(futures.SideTypeSell, 0, positionAmt, futures.PositionSideTypeLong)
			if err != nil {
				logger.Error(err)
				return err
			}
		}
	} else {
		lossRate := 100.0 * (nowPrice - entryPrice) / entryPrice
		fmt.Printf("做多: 亏损 %f USDT, 幅度:%f%%\n", (nowPrice-entryPrice)*positionAmt, lossRate)

		if positionAmt > 0.0 && -lossRate >= g.stopLoss {
			_, err := g.Trade(futures.SideTypeSell, 0, positionAmt, futures.PositionSideTypeLong)
			if err != nil {
				logger.Error(err)
				return err
			}
		}
	}

	rates, err := g.GetKlines(g.Symbol, _interval, _limit)
	if err != nil {
		logger.Error(err)
		return err
	}
	op := g.bottomtop(rates)
	fmt.Println("op:", op)

	nowPrice, err = g.GetNewestPrice()
	if err != nil {
		logger.Error(err)
		return err
	}

	if op == BUY {
		if positionAmt > 0.0 {
			n := int(positionAmt/g.chunk) / 2
			if n > len(Fibonacci)-1 {
				n = len(Fibonacci) - 1
			}
			if positionAmt >= g.maxAmount || nowPrice > entryPrice-g.intervalPrice*Fibonacci[n] {
				return nil
			}
		}
		_, err := g.Trade(futures.SideTypeBuy, 0, g.chunk, futures.PositionSideTypeLong)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	return nil
}

func (g *GridStrategy) DoShort3() error {
	short, err := g.Position(futures.PositionSideType("SHORT"))
	if err != nil {
		logger.Error(err)
		return err
	}
	entryPrice, err := strconv.ParseFloat(short.EntryPrice, 64)
	if err != nil {
		logger.Error(err)
		return err
	}
	positionAmt, err := strconv.ParseFloat(short.PositionAmt, 64)
	if err != nil {
		logger.Error(err)
		return err
	}

	nowPrice, err := g.GetNewestPrice()
	if err != nil {
		logger.Error(err)
		return err
	}

	positionAmt = -positionAmt
	fmt.Printf("做空: entryPrice: %v, positionAmt: %v\n", entryPrice, positionAmt)
	if nowPrice > entryPrice {
		lossRate := 100.0 * (entryPrice - nowPrice) / entryPrice
		fmt.Printf("做空: 亏损 %f USDT, 幅度:%f%%\n", (entryPrice-nowPrice)*positionAmt, lossRate)

		if positionAmt > 0.0 && -lossRate >= g.stopLoss {
			_, err := g.Trade(futures.SideTypeBuy, 0, positionAmt, futures.PositionSideTypeShort)
			if err != nil {
				logger.Error(err)
				return err
			}
		}
	} else {
		profitRate := 100.0 * (entryPrice - nowPrice) / entryPrice
		fmt.Printf("做空: 盈利 %f USDT, 幅度:%f%%\n", (entryPrice-nowPrice)*positionAmt, profitRate)

		if positionAmt > 0.0 && profitRate >= g.stopWin {
			_, err := g.Trade(futures.SideTypeBuy, 0, positionAmt, futures.PositionSideTypeShort)
			if err != nil {
				logger.Error(err)
				return err
			}
		}
	}

	rates, err := g.GetKlines(g.Symbol, _interval, _limit)
	if err != nil {
		logger.Error(err)
		return err
	}
	op := g.bottomtop(rates)
	fmt.Println("op:", op)

	nowPrice, err = g.GetNewestPrice()
	if err != nil {
		logger.Error(err)
		return err
	}

	if op == SELL {
		if positionAmt > 0.0 {
			n := int(positionAmt/g.chunk) / 2
			if n > len(Fibonacci)-1 {
				n = len(Fibonacci) - 1
			}
			if positionAmt >= g.maxAmount || nowPrice < entryPrice+g.intervalPrice*Fibonacci[n] {
				return nil
			}
		}
		_, err := g.Trade(futures.SideTypeSell, 0, g.chunk, futures.PositionSideTypeShort)
		if err != nil {
			logger.Error(err)
			return err
		}
	}

	return nil
}

func (g *GridStrategy) DoLong2() error {
	openOrders, err := g.GetOpenOrders()
	if err != nil {
		logger.Error(err)
		return err
	}

	minOpenPrice := 0.0
	remainCount := 0
	g.LongTrades.Range(func(key, value interface{}) bool {
		orderId := key.(int64)
		_, ok := openOrders[orderId]
		if !ok {
			g.LongTrades.Delete(orderId)
		} else {
			remainCount++

			in := value.(*TradeInfo)
			if in.OpenPrice < minOpenPrice {
				minOpenPrice = in.OpenPrice
			}
		}
		return true
	})

	fmt.Println("剩余未完成交易:", remainCount)
	g.LongTrades.Range(func(key, value interface{}) bool {
		fmt.Printf("%+v\n\n", value)
		return true
	})

	rates, err := g.GetKlines(g.Symbol, _interval, _limit)
	if err != nil {
		logger.Error(err)
		return err
	}
	op := g.bottomtop(rates)
	logger.Info("op:", op)

	nowPrice, err := g.GetNewestPrice()
	if err != nil {
		logger.Error(err)
		return err
	}

	if op == BUY && remainCount < 10 && minOpenPrice-0.0005 >= nowPrice {
		order, err := g.Trade(futures.SideTypeBuy, 0, g.chunk, futures.PositionSideTypeLong)
		if err != nil {
			logger.Error(err)
			return err
		}
		price, err := strconv.ParseFloat(order.AvgPrice, 64)
		if err != nil {
			logger.Error(err)
			return err
		}
		amount, err := strconv.ParseFloat(order.ExecutedQuantity, 64)
		if err != nil {
			logger.Error(err)
			return err
		}

		ti := &TradeInfo{
			Type:        string(order.PositionSide),
			OpenPrice:   price,
			OpenAmount:  amount,
			OpenTime:    time.Unix(order.Time/1000, 0),
			OpenOrderId: order.OrderID,

			ClosePrice:  price + 0.002,
			CloseAmount: amount,
		}

		limitOrder, err := g.TradeLimit(futures.SideTypeSell, ti.ClosePrice, ti.CloseAmount, futures.PositionSideTypeLong)
		if err != nil {
			logger.Error(err)
			return err
		}
		ti.CloseOrderId = limitOrder.OrderID

		g.LongTrades.Store(ti.CloseOrderId, ti)
	}

	return nil
}

func (g *GridStrategy) DoShort2() error {
	openOrders, err := g.GetOpenOrders()
	if err != nil {
		logger.Error(err)
		return err
	}

	maxOpenPrice := 0.0
	remainCount := 0
	g.ShortTrades.Range(func(key, value interface{}) bool {
		orderId := key.(int64)
		_, ok := openOrders[orderId]
		if !ok {
			g.ShortTrades.Delete(orderId)
		} else {
			remainCount++

			in := value.(*TradeInfo)
			if in.OpenPrice > maxOpenPrice {
				maxOpenPrice = in.OpenPrice
			}
		}
		return true
	})

	fmt.Println("剩余未完成交易:", remainCount)
	g.ShortTrades.Range(func(key, value interface{}) bool {
		fmt.Printf("%+v\n\n", value)
		return true
	})

	rates, err := g.GetKlines(g.Symbol, _interval, _limit)
	if err != nil {
		logger.Error(err)
		return err
	}
	op := g.bottomtop(rates)
	logger.Info("op:", op)

	nowPrice, err := g.GetNewestPrice()
	if err != nil {
		logger.Error(err)
		return err
	}

	if op == SELL && remainCount < 10 && maxOpenPrice+0.0005 <= nowPrice {
		order, err := g.Trade(futures.SideTypeSell, 0, g.chunk, futures.PositionSideTypeShort)
		if err != nil {
			logger.Error(err)
			return err
		}
		price, err := strconv.ParseFloat(order.AvgPrice, 64)
		if err != nil {
			logger.Error(err)
			return err
		}
		amount, err := strconv.ParseFloat(order.ExecutedQuantity, 64)
		if err != nil {
			logger.Error(err)
			return err
		}

		ti := &TradeInfo{
			Type:        string(order.PositionSide),
			OpenPrice:   price,
			OpenAmount:  amount,
			OpenTime:    time.Unix(order.Time/1000, 0),
			OpenOrderId: order.OrderID,

			ClosePrice:  price - 0.002,
			CloseAmount: amount,
		}

		limitOrder, err := g.TradeLimit(futures.SideTypeBuy, ti.ClosePrice, ti.CloseAmount, futures.PositionSideTypeShort)
		if err != nil {
			logger.Error(err)
			return err
		}
		ti.CloseOrderId = limitOrder.OrderID

		g.ShortTrades.Store(ti.CloseOrderId, ti)
	}

	return nil
}

func (g *GridStrategy) Hedge() error {
	openOrders, err := g.GetOpenOrders()
	if err != nil {
		logger.Error(err)
		return err
	}

	longMinOpenPrice := math.MaxFloat64
	longRemainCount := 0
	g.LongTrades.Range(func(key, value interface{}) bool {
		orderId := key.(int64)
		_, ok := openOrders[orderId]
		if !ok {
			g.LongTrades.Delete(orderId)
		} else {
			longRemainCount++

			in := value.(*TradeInfo)
			if in.OpenPrice < longMinOpenPrice {
				longMinOpenPrice = in.OpenPrice
			}
		}
		return true
	})

	fmt.Println("做多剩余未完成交易:", longRemainCount)
	g.LongTrades.Range(func(key, value interface{}) bool {
		fmt.Printf("%+v\n\n", value)
		return true
	})

	shortMaxOpenPrice := 0.0
	shortRemainCount := 0
	g.ShortTrades.Range(func(key, value interface{}) bool {
		orderId := key.(int64)
		_, ok := openOrders[orderId]
		if !ok {
			g.ShortTrades.Delete(orderId)
		} else {
			shortRemainCount++

			in := value.(*TradeInfo)
			if in.OpenPrice > shortMaxOpenPrice {
				shortMaxOpenPrice = in.OpenPrice
			}
		}
		return true
	})

	fmt.Println("做空剩余未完成交易:", shortRemainCount)
	g.ShortTrades.Range(func(key, value interface{}) bool {
		fmt.Printf("%+v\n\n", value)
		return true
	})

	rates, err := g.GetKlines(g.Symbol, _interval, _limit+1)
	if err != nil {
		logger.Error(err)
		return err
	}
	op := g.bottomtop(rates[:len(rates)-1])
	logger.Info("op:", op)

	nowPrice, err := g.GetNewestPrice()
	if err != nil {
		logger.Error(err)
		return err
	}

	if op == WAIT {
		if longRemainCount-shortRemainCount > g.maxDistance {
			op = SELL
			logger.Info("change op:", op)
		} else if shortRemainCount-longRemainCount > g.maxDistance {
			op = BUY
			logger.Info("change op:", op)
		}
	}

	if op == SELL {
		if shortRemainCount > 0 {
			if shortRemainCount-longRemainCount >= g.maxDistance {
				return nil
			}
			if shortRemainCount >= g.maxChunks {
				return nil
			}
			if shortMaxOpenPrice+g.intervalPrice*Fibonacci[shortRemainCount] >= nowPrice {
				return nil
			}
		}

		order, err := g.Trade(futures.SideTypeSell, 0, g.chunk, futures.PositionSideTypeShort)
		if err != nil {
			logger.Error(err)
			return err
		}
		price, err := strconv.ParseFloat(order.AvgPrice, 64)
		if err != nil {
			logger.Error(err)
			return err
		}
		amount, err := strconv.ParseFloat(order.ExecutedQuantity, 64)
		if err != nil {
			logger.Error(err)
			return err
		}

		profit := g.profit
		if longRemainCount > shortRemainCount {
			profit = float64(longRemainCount-shortRemainCount) * g.profit
		}

		price = truncFloat(price, 5)
		ti := &TradeInfo{
			Type:        string(order.PositionSide),
			OpenPrice:   price,
			OpenAmount:  amount,
			OpenTime:    time.Unix(order.Time/1000, 0),
			OpenOrderId: order.OrderID,

			ClosePrice:  price - profit,
			CloseAmount: amount,
		}

		limitOrder, err := g.TradeLimit(futures.SideTypeBuy, ti.ClosePrice, ti.CloseAmount, futures.PositionSideTypeShort)
		if err != nil {
			logger.Error(err)
			return err
		}
		ti.CloseOrderId = limitOrder.OrderID

		g.ShortTrades.Store(ti.CloseOrderId, ti)
	} else if op == BUY {
		if longRemainCount > 0 {
			if longRemainCount-shortRemainCount >= g.maxDistance {
				return nil
			}
			if longRemainCount >= g.maxChunks {
				return nil
			}
			if longMinOpenPrice-g.intervalPrice*Fibonacci[longRemainCount] <= nowPrice {
				return nil
			}
		}

		order, err := g.Trade(futures.SideTypeBuy, 0, g.chunk, futures.PositionSideTypeLong)
		if err != nil {
			logger.Error(err)
			return err
		}
		price, err := strconv.ParseFloat(order.AvgPrice, 64)
		if err != nil {
			logger.Error(err)
			return err
		}
		amount, err := strconv.ParseFloat(order.ExecutedQuantity, 64)
		if err != nil {
			logger.Error(err)
			return err
		}

		profit := g.profit
		if shortRemainCount > longRemainCount {
			profit = float64(shortRemainCount-longRemainCount) * g.profit
		}

		price = truncFloat(price, 5)
		ti := &TradeInfo{
			Type:        string(order.PositionSide),
			OpenPrice:   price,
			OpenAmount:  amount,
			OpenTime:    time.Unix(order.Time/1000, 0),
			OpenOrderId: order.OrderID,

			ClosePrice:  price + profit,
			CloseAmount: amount,
		}

		limitOrder, err := g.TradeLimit(futures.SideTypeSell, ti.ClosePrice, ti.CloseAmount, futures.PositionSideTypeLong)
		if err != nil {
			logger.Error(err)
			return err
		}
		ti.CloseOrderId = limitOrder.OrderID

		g.LongTrades.Store(ti.CloseOrderId, ti)
	}

	return nil
}

func (g *GridStrategy) Hedge2() error {
	rates, err := g.GetKlines(g.Symbol, _interval, _limit)
	if err != nil {
		logger.Error(err)
		return err
	}
	op := g.bottomtop(rates)
	logger.Info("op:", op)

	if op == SELL {
		if g.HedgeLong >= g.chunk && g.HedgeLong >= g.HedgeShort {
			_, err := g.Trade(futures.SideTypeSell, 0, g.chunk, futures.PositionSideTypeLong)
			if err != nil {
				logger.Error(err)
				return err
			}
			g.HedgeLong -= g.chunk
		}

		if g.HedgeShort <= g.HedgeLong && g.HedgeShort < float64(g.maxDistance)*g.chunk {
			_, err = g.Trade(futures.SideTypeSell, 0, g.chunk, futures.PositionSideTypeShort)
			if err != nil {
				logger.Error(err)
				return err
			}
			g.HedgeShort += g.chunk
		}
	} else if op == BUY {
		if g.HedgeShort >= g.HedgeLong && g.HedgeLong < float64(g.maxDistance)*g.chunk {
			_, err := g.Trade(futures.SideTypeBuy, 0, g.chunk, futures.PositionSideTypeLong)
			if err != nil {
				logger.Error(err)
				return err
			}
			g.HedgeLong += g.chunk
		}

		if g.HedgeShort >= g.chunk && g.HedgeLong <= g.HedgeShort {
			_, err = g.Trade(futures.SideTypeBuy, 0, g.chunk, futures.PositionSideTypeShort)
			if err != nil {
				logger.Error(err)
				return err
			}
			g.HedgeShort -= g.chunk
		}
	}

	fmt.Println("做多数量:", g.HedgeLong)
	fmt.Println("做空数量:", g.HedgeShort)
	return nil
}

func (g *GridStrategy) RunTask() error {
	for {
		g.GetTrades()
		//g.GetAccountInfo()
		if g.PositionSide == "LONG" {
			g.DoLong()
		} else if g.PositionSide == "SHORT" {
			g.DoShort()
		} else if g.PositionSide == "BOTH" {
			nowPrice, err := g.GetNewestPrice()
			if err != nil {
				logger.Error(err)
				return err
			}
			fmt.Printf("%s now price: %v\n", g.Symbol, nowPrice)

			availableBalance, err := g.GetAvailableBalance("USDT")
			if err != nil {
				logger.Error(err)
				return err
			}
			fmt.Printf("available balance: %v\n\n", availableBalance)

			g.DoLong()
			fmt.Println()
			g.DoShort()
		} else if g.PositionSide == "HEDGE" {
			g.Hedge()
		} else {
			rates, err := g.GetKlines(g.Symbol, _interval, _limit)
			if err != nil {
				logger.Error(err)
				return err
			}
			op := g.bottomtop(rates)
			logger.Info("op:", op)
			fmt.Println()
			g.GetOpenOrders()
		}
		time.Sleep(time.Second * time.Duration(g.tradeInterval))
	}
	return nil
}

func main() {
	symbol := flag.String("symbol", "DOGEUSDT", "symbol")
	orderId := flag.Int64("orderId", 0, "orderId")
	action := flag.String("action", "", "action")

	positionSide := flag.String("positionSide", "", "positionSide")
	sideType := flag.String("sideType", "BUY", "sideType")
	price := flag.Float64("price", 0.0, "price")
	amount := flag.Float64("amount", 0.0, "amount")
	tradeInterval := flag.Int64("tradeInterval", 20, "tradeInterval")
	maxDistance := flag.Int("maxDistance", 2, "maxDistance")
	maxChunks := flag.Int("maxChunks", 1, "maxChunks")
	profit := flag.Float64("profit", 0.002, "profit")

	chunk := flag.Float64("chunk", 30.0, "chunk")
	maxAmount := flag.Float64("maxAmount", 200.0, "maxAmount")
	intervalPrice := flag.Float64("intervalPrice", 0.5, "intervalPrice %")
	stopWin := flag.Float64("stopWin", 1.0, "stopWin %")
	stopLoss := flag.Float64("stopLoss", 20.0, "stopLoss %")

	flag.Parse()

	grid := &GridStrategy{
		client:        futures.NewClient(ApiKey, SecretKey),
		Symbol:        *symbol,
		PositionSide:  *positionSide,
		tradeInterval: *tradeInterval, //间隔
		chunk:         *chunk,         //单元数量
		maxDistance:   *maxDistance,   //做多做空最大单元距离
		maxChunks:     *maxChunks,     //最大单元
		profit:        *profit,
		intervalPrice: *intervalPrice,
		maxAmount:     *maxAmount,
		stopWin:       *stopWin,
		stopLoss:      *stopLoss,
	}

	if *action == "cancel" {
		grid.CancelOrder(*symbol, *orderId)
		return
	} else if *action == "order" {
		grid.TradeLimit(futures.SideType(strings.ToUpper(*sideType)), *price, *amount, futures.PositionSideType(strings.ToUpper(*positionSide)))
		return
	}

	grid.RunTask()
}
