package main

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/adshao/go-binance/v2/futures"
	"github.com/harveywangdao/ants/logger"
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

func (g *GridStrategy) GetDepth() error {
	res, err := g.client.NewDepthService().Symbol(g.Symbol).Limit(5).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("depth asks:", res.Asks)
	logger.Info("depth bids:", res.Bids)

	nowAskPrice, err := strconv.ParseFloat(res.Asks[0].Price, 64)
	if err != nil {
		logger.Error("nowAskPrice parse float64 fail, err", err)
		return err
	}
	nowBidPrice, err := strconv.ParseFloat(res.Bids[0].Price, 64)
	if err != nil {
		logger.Error("nowBidPrice parse float64 fail, err", err)
		return err
	}
	logger.Infof("nowAskPrice=%v, nowBidPrice=%v", nowAskPrice, nowBidPrice)
	return nil
}

func (g *GridStrategy) GetNewestPrice() (float64, error) {
	symbolPrices, err := g.client.NewListPricesService().Symbol(g.Symbol).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return 0.0, err
	}
	if len(symbolPrices) == 0 {
		return 0.0, fmt.Errorf("can not get newest price")
	}
	price, err := strconv.ParseFloat(symbolPrices[0].Price, 64)
	if err != nil {
		logger.Error(err)
		return 0.0, err
	}
	return price, nil
}

/*
1.双向持仓模式下 positionSide LONG/SHORT
2.双向持仓模式下 reduceOnly 不接受此参数
*/
func (g *GridStrategy) Trade(sideType futures.SideType, price, amount float64, positionSide futures.PositionSideType) (*futures.Order, error) {
	service := g.client.NewCreateOrderService().
		Symbol(g.Symbol).Quantity(fmt.Sprint(amount)).Side(sideType).Type(futures.OrderTypeMarket)
	if price > 0 {
		service = service.Price(fmt.Sprint(price))
	}

	// SHORT BOTH
	service = service.PositionSide(positionSide)
	resp, err := service.Do(context.Background())
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	orderInfo, err := g.client.NewGetOrderService().
		Symbol(g.Symbol).OrderID(resp.OrderID).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Infof("order=%#v", orderInfo)

	return orderInfo, nil
}

func (g *GridStrategy) TradeLimit(sideType futures.SideType, price, amount float64, positionSide futures.PositionSideType) (*futures.Order, error) {
	service := g.client.NewCreateOrderService().
		Symbol(g.Symbol).Quantity(fmt.Sprint(amount)).Side(sideType).Type(futures.OrderTypeLimit)
	if price > 0 {
		service = service.Price(fmt.Sprint(price))
	}

	// SHORT BOTH
	service = service.PositionSide(positionSide).TimeInForce(futures.TimeInForceTypeGTC)
	resp, err := service.Do(context.Background())
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	orderInfo, err := g.client.NewGetOrderService().
		Symbol(g.Symbol).OrderID(resp.OrderID).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Infof("order=%#v", orderInfo)

	return orderInfo, nil
}

func (g *GridStrategy) GetAvailableBalance(asset string) (float64, error) {
	balances, err := g.client.NewGetBalanceService().Do(context.Background())
	if err != nil {
		logger.Error(err)
		return 0.0, err
	}
	for i := 0; i < len(balances); i++ {
		if balances[i].Asset == asset {
			//logger.Infof("%#v", balances[i])
			availableBalance, err := strconv.ParseFloat(balances[i].AvailableBalance, 64)
			if err != nil {
				logger.Error(err)
				return 0.0, err
			}
			return availableBalance, nil
		}
	}
	return 0.0, fmt.Errorf("can not find %s balance", asset)
}

func (g *GridStrategy) Position(positionSide futures.PositionSideType) (*futures.AccountPosition, error) {
	account, err := g.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	for i := 0; i < len(account.Positions); i++ {
		if account.Positions[i].Symbol == g.Symbol && account.Positions[i].PositionSide == positionSide {
			return account.Positions[i], nil
		}
	}
	return nil, fmt.Errorf("not existed")
}

func (g *GridStrategy) Position2() (*futures.AccountPosition, *futures.AccountPosition, error) {
	account, err := g.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		logger.Error(err)
		return nil, nil, err
	}

	var long, short *futures.AccountPosition
	for i := 0; i < len(account.Positions); i++ {
		if account.Positions[i].Symbol == g.Symbol {
			if account.Positions[i].PositionSide == futures.PositionSideType("LONG") {
				long = account.Positions[i]
			} else if account.Positions[i].PositionSide == futures.PositionSideType("SHORT") {
				short = account.Positions[i]
			}
			if long != nil && short != nil {
				break
			}
		}
	}
	return long, short, nil
}

func (g *GridStrategy) GetAccountInfo() error {
	nowPrice, err := g.GetNewestPrice()
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Infof("%s now price: %v", g.Symbol, nowPrice)

	availableBalance, err := g.GetAvailableBalance("USDT")
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Infof("available balance: %v", availableBalance)

	long, short, err := g.Position2()
	if err != nil {
		logger.Error(err)
		return err
	}

	if long != nil {
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

		logger.Infof("做多: entryPrice: %v, positionAmt: %v", entryPrice, positionAmt)
		if nowPrice > entryPrice {
			logger.Infof("做多: 盈利 %f USDT, 幅度:%f%%", (nowPrice-entryPrice)*positionAmt, 100.0*(nowPrice-entryPrice)/entryPrice)
		} else {
			logger.Infof("做多: 亏损 %f USDT, 幅度:%f%%", (nowPrice-entryPrice)*positionAmt, 100.0*(nowPrice-entryPrice)/entryPrice)
		}
	}

	if short != nil {
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

		logger.Infof("做空: entryPrice: %v, positionAmt: %v", entryPrice, positionAmt)
		if nowPrice > entryPrice {
			logger.Infof("做空: 亏损 %f USDT, 幅度:%f%%", (nowPrice-entryPrice)*positionAmt, 100.0*(entryPrice-nowPrice)/entryPrice)
		} else {
			logger.Infof("做空: 盈利 %f USDT, 幅度:%f%%", (nowPrice-entryPrice)*positionAmt, 100.0*(entryPrice-nowPrice)/entryPrice)
		}
	}
	fmt.Println()
	return nil
}

func (g *GridStrategy) CancelOrder(symbol string, orderId int64) error {
	ret, err := g.client.NewCancelOrderService().Symbol(symbol).OrderID(orderId).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}
	fmt.Printf("%+v\n", *ret)
	return nil
}

func (g *GridStrategy) GetTrades() error {
	trades, err := g.client.NewListAccountTradeService().Symbol(g.Symbol).Limit(10).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}
	fmt.Println("历史成交:", len(trades))
	for i := 0; i < len(trades); i++ {
		fmt.Println("OrderID:", trades[i].OrderID)
		fmt.Println("Symbol:", trades[i].Symbol)
		fmt.Println("PositionSide:", trades[i].PositionSide)
		fmt.Println("Side:", trades[i].Side)
		fmt.Println("Price:", trades[i].Price)
		fmt.Println("Quantity:", trades[i].Quantity)
		fmt.Println("Commission:", trades[i].Commission)
		fmt.Println("Time:", time.Unix(trades[i].Time/1000, 0))

		fmt.Println("\n")
	}
	return nil
}

func (g *GridStrategy) GetOpenOrders() (map[int64]*futures.Order, error) {
	trades, err := g.client.NewListOpenOrdersService().Symbol(g.Symbol).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	m := make(map[int64]*futures.Order)
	fmt.Println("当前挂单:", len(trades))
	for i := 0; i < len(trades); i++ {
		m[trades[i].OrderID] = trades[i]

		fmt.Println("OrderID:", trades[i].OrderID)
		fmt.Println("Symbol:", trades[i].Symbol)
		fmt.Println("PositionSide:", trades[i].PositionSide)
		fmt.Println("Side:", trades[i].Side)
		fmt.Println("Price:", trades[i].Price)
		fmt.Println("OrigQuantity:", trades[i].OrigQuantity)
		fmt.Println("Status:", trades[i].Status)
		fmt.Println("Type:", trades[i].Type)
		fmt.Println("WorkingType:", trades[i].WorkingType)
		fmt.Println("Time:", time.Unix(trades[i].Time/1000, 0))
		fmt.Println("\n")
	}
	return m, nil
}

func (g *GridStrategy) GetKlines(symbol, interval string, limit int) ([]KlineData, error) {
	// 1m 3m 5m 15m 30m 1h 2h 4h 6h 8h 12h 1d 3d 1w 1M
	klines, err := g.client.NewKlinesService().Symbol(symbol).Interval(interval).Limit(limit).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	n := len(klines)
	if n != limit {
		return nil, fmt.Errorf("klines count error")
	}

	rates := make([]KlineData, n)
	for i := 0; i < n; i++ {
		openp, err := strconv.ParseFloat(klines[i].Open, 64)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		closep, err := strconv.ParseFloat(klines[i].Close, 64)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		highp, err := strconv.ParseFloat(klines[i].High, 64)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		lowp, err := strconv.ParseFloat(klines[i].Low, 64)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
		rates[i].Rate = (closep - openp) / openp
		rates[i].High = highp
		rates[i].Low = lowp
		rates[i].Open = openp
		rates[i].Close = closep
		rates[i].OpenTime = time.Unix(klines[i].OpenTime/1000, 0)   //秒
		rates[i].CloseTime = time.Unix(klines[i].CloseTime/1000, 0) //秒
	}

	return rates, nil
}

func (g *GridStrategy) bottomtop(klines []KlineData) Operate {
	n := len(klines)
	if n == 0 {
		return WAIT
	}

	var hpoints []int
	var lpoints []int
	for start := 0; start < n-1; start++ {
		highest, lowest := 0.0, math.MaxFloat64
		var hpos, lpos int

		for i := start; i < n; i++ {
			if klines[i].High > highest {
				highest = klines[i].High
				hpos = i + 1
			}
			if klines[i].Low < lowest {
				lowest = klines[i].Low
				lpos = i + 1
			}
		}

		hpoints = append(hpoints, hpos)
		lpoints = append(lpoints, lpos)
	}

	if hpoints[0] == n {
		logger.Info("突破新高")
		return WAIT
	}
	if hpoints[0] == n-1 {
		logger.Info("突破新高,下降初期1")
		return SELL
	}
	if hpoints[0] == n-2 {
		logger.Info("突破新高,下降初期2")
	}
	if hpoints[0] == n-3 {
		logger.Info("突破新高,下降初期3")
	}

	if lpoints[0] == n {
		logger.Info("突破新低")
		return WAIT
	}
	if lpoints[0] == n-1 {
		logger.Info("突破新低,上升初期1")
		return BUY
	}
	if lpoints[0] == n-2 {
		logger.Info("突破新低,上升初期2")
	}
	if lpoints[0] == n-3 {
		logger.Info("突破新低,上升初期3")
	}

	var highpoints []int
	point := -1
	repeat := 0
	for i := 0; i < len(hpoints); i++ {
		if point != hpoints[i] {
			if repeat >= 5 {
				highpoints = append(highpoints, point)
			}
			repeat = 0
		}
		point = hpoints[i]
		repeat++
	}
	if repeat >= 5 {
		highpoints = append(highpoints, point)
	}

	var lowpoints []int
	point = -1
	repeat = 0
	for i := 0; i < len(lpoints); i++ {
		if point != lpoints[i] {
			if repeat >= 5 {
				lowpoints = append(lowpoints, point)
			}
			repeat = 0
		}
		point = lpoints[i]
		repeat++
	}
	if repeat >= 5 {
		lowpoints = append(lowpoints, point)
	}

	highIndex := -1
	for i := 0; i < len(highpoints); i++ {
		logger.Infof("波峰,price: %f, 倒数第%d个", klines[highpoints[i]-1].High, n-highpoints[i]+1)
		if highpoints[i] > highIndex {
			highIndex = highpoints[i]
		}
	}
	if highIndex != -1 {
		logger.Infof("最近的波峰,price: %f, 倒数第%d个", klines[highIndex-1].High, n-highIndex+1)
	} else {
		logger.Info("无波峰,当前可能在波动")
	}

	lowIndex := -1
	for i := 0; i < len(lowpoints); i++ {
		logger.Infof("波谷,price: %f, 倒数第%d个", klines[lowpoints[i]-1].Low, n-lowpoints[i]+1)
		if lowpoints[i] > lowIndex {
			lowIndex = lowpoints[i]
		}
	}
	if lowIndex != -1 {
		logger.Infof("最近的波谷,price: %f, 倒数第%d个", klines[lowIndex-1].Low, n-lowIndex+1)
	} else {
		logger.Info("无波谷,当前可能在波动")
	}

	if highIndex != -1 && lowIndex != -1 {
		if highIndex >= lowIndex {
			if highIndex-lowIndex <= 5 {
				logger.Info("当前在震荡")
				return WAIT
			}

			if highIndex == n {
				logger.Info("当前是波峰")
			} else if highIndex == n-1 {
				logger.Info("当前是波峰,刚开始下降")
				return SELL
			} else {
				logger.Info("当前是下降阶段")
			}
		} else {
			if lowIndex-highIndex <= 2 {
				logger.Info("当前在震荡")
				return WAIT
			}

			if lowIndex == n {
				logger.Info("当前是波谷")
			} else if lowIndex == n-1 {
				logger.Info("当前是波谷,刚开始上升")
				return BUY
			} else {
				logger.Info("当前是上升阶段")
			}
		}
	}
	return WAIT
}

func truncFloat(f float64, num int) float64 {
	unit := math.Pow10(num)
	a := f * unit
	b := math.Trunc(a)
	//fmt.Println(unit, a, b)
	return b / unit
}

func (g *GridStrategy) getEntryPriceAndAmt(side string) (float64, float64, error) {
	position, err := g.Position(futures.PositionSideType(side))
	if err != nil {
		logger.Error(err)
		return 0.0, 0.0, err
	}
	entryPrice, err := strconv.ParseFloat(position.EntryPrice, 64)
	if err != nil {
		logger.Error(err)
		return 0.0, 0.0, err
	}
	positionAmt, err := strconv.ParseFloat(position.PositionAmt, 64)
	if err != nil {
		logger.Error(err)
		return 0.0, 0.0, err
	}
	return entryPrice, positionAmt, nil
}
