package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

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

type GridStrategy struct {
	client       *futures.Client
	Symbol       string
	PositionSide string
	LongTrades   sync.Map
	ShortTrades  sync.Map

	HedgeLong  float64
	HedgeShort float64
}

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

	// {Asset:"USDT", InitialMargin:"2.73550490", MaintMargin:"0.27355049", MarginBalance:"6.26307065", MaxWithdrawAmount:"3.52756575", OpenOrderInitialMargin:"0.00000000", PositionInitialMargin:"2.73550490", UnrealizedProfit:"0.02344900", WalletBalance:"6.23962165"}
	for i := 0; i < len(account.Assets); i++ {
		if account.Assets[i].Asset == "USDT" {
			//logger.Infof("%#v", account.Assets[i])
		}
	}

	// {Isolated:false, Leverage:"10", InitialMargin:"2.73550490", MaintMargin:"0.27355049", OpenOrderInitialMargin:"0", PositionInitialMargin:"2.73550490", Symbol:"DOGEUSDT", UnrealizedProfit:"0.02344900", EntryPrice:"0.273316", MaxNotional:"100000", PositionSide:"LONG", PositionAmt:"100"}
	for i := 0; i < len(account.Positions); i++ {
		if account.Positions[i].Symbol == g.Symbol && account.Positions[i].PositionSide == positionSide {
			//logger.Infof("%s: %#v", g.Symbol, account.Positions[i])
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
		return SELL
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
		return BUY
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
			if highIndex-lowIndex <= 2 {
				logger.Info("当前在震荡")
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

func (g *GridStrategy) DoLong() error {
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

	interval := "1m"
	limit := 30
	rates, err := g.GetKlines(g.Symbol, interval, limit)
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
		chunk := 30.0
		order, err := g.Trade(futures.SideTypeBuy, 0, chunk, futures.PositionSideTypeLong)
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

func (g *GridStrategy) DoShort() error {
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

	interval := "1m"
	limit := 30
	rates, err := g.GetKlines(g.Symbol, interval, limit)
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
		chunk := 30.0
		order, err := g.Trade(futures.SideTypeSell, 0, chunk, futures.PositionSideTypeShort)
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
	interval := "1m"
	limit := 30
	rates, err := g.GetKlines(g.Symbol, interval, limit)
	if err != nil {
		logger.Error(err)
		return err
	}
	op := g.bottomtop(rates)
	logger.Info("op:", op)

	chunk := 30.0
	if op == SELL {
		if g.HedgeLong >= chunk && g.HedgeLong >= g.HedgeShort {
			_, err := g.Trade(futures.SideTypeSell, 0, chunk, futures.PositionSideTypeLong)
			if err != nil {
				logger.Error(err)
				return err
			}
			g.HedgeLong -= chunk
		}

		if g.HedgeShort <= g.HedgeLong && g.HedgeShort < 5*chunk {
			_, err = g.Trade(futures.SideTypeSell, 0, chunk, futures.PositionSideTypeShort)
			if err != nil {
				logger.Error(err)
				return err
			}
			g.HedgeShort += chunk
		}
	} else if op == BUY {
		if g.HedgeShort >= g.HedgeLong && g.HedgeLong < 5*chunk {
			_, err := g.Trade(futures.SideTypeBuy, 0, chunk, futures.PositionSideTypeLong)
			if err != nil {
				logger.Error(err)
				return err
			}
			g.HedgeLong += chunk
		}

		if g.HedgeShort >= chunk && g.HedgeLong <= g.HedgeShort {
			_, err = g.Trade(futures.SideTypeBuy, 0, chunk, futures.PositionSideTypeShort)
			if err != nil {
				logger.Error(err)
				return err
			}
			g.HedgeShort -= chunk
		}
	}

	fmt.Println("做多数量:", g.HedgeLong)
	fmt.Println("做空数量:", g.HedgeShort)
	return nil
}

func (g *GridStrategy) RunTask() error {
	for {
		g.GetTrades()
		g.GetAccountInfo()
		if g.PositionSide == "LONG" {
			g.DoLong()
		} else if g.PositionSide == "SHORT" {
			g.DoShort()
		} else if g.PositionSide == "BOTH" {
			g.DoLong()
			g.DoShort()
		} else if g.PositionSide == "HEDGE" {
			g.Hedge()
		}
		time.Sleep(time.Second * 60)
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

	flag.Parse()

	grid := &GridStrategy{
		client:       futures.NewClient(ApiKey, SecretKey),
		Symbol:       *symbol,
		PositionSide: *positionSide,
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
