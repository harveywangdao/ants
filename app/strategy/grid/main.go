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

func init() {
	logger.SetHandlers(logger.Console)
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)
	logger.SetLevel(logger.INFO)
}

const (
	ApiKey    = "KtnHyOpZzRgetPtTlxkdkCck1DlVumUvBUCEtSmAzxItVdXKigsugS11rteCRYLh"
	SecretKey = "TLo3hxjoGCVHBCyq6GUOA7QyoWCrV35VUOZaybVlVPfsHpJy6T2AkjlnH8mdFlJr"

	PositionSide    = "LONG"
	Symbol          = "DOGEUSDT"
	Direction       = 1      // 方向: 1(UP) -1(Down)
	GridNum         = 10     // 网格节点数量 10
	GridPointAmount = 20.0   // 网格节点下单量 1
	GridPointDis    = 0.0002 // 网格节点间距 20
	GridCovDis      = 0.0005 // 网格节点平仓价差 50
)

type GridStrategy struct {
	client *futures.Client

	PositionSide    string
	Symbol          string  `opt:"symbol,BTUSDT"`
	Direction       float64 `opt:"direction,1"`         // 网格方向 up 1, down -1
	GridNum         int     `opt:"grid_num,10"`         // 网格节点数量 10
	GridPointAmount float64 `opt:"grid_point_amount,1"` // 网格节点下单量 1
	GridPointDis    float64 `opt:"grid_point_dis,20"`   // 网格节点间距 20
	GridCovDis      float64 `opt:"grid_cov_dis,50"`     // 网格节点平仓价差 50

	state          string
	WinRate        float64
	StopRate       float64
	NormalWaveRate float64
}

func (g *GridStrategy) OnTick() error {
	availableBalance, err := g.getAvailableBalance("USDT")
	if err != nil {
		logger.Error(err)
		return err
	}
	position, err := g.Position()
	if err != nil {
		logger.Error(err)
		return err
	}
	entryPrice, err := strconv.ParseFloat(position.EntryPrice, 64)
	if err != nil {
		logger.Error(err)
		return err
	}
	positionAmt, err := strconv.ParseFloat(position.PositionAmt, 64)
	if err != nil {
		logger.Error(err)
		return err
	}
	nowPrice, err := g.getNewestPrice()
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Infof("availableBalance: %v, nowPrice: %v, entryPrice: %v, positionAmt: %v", availableBalance, nowPrice, entryPrice, positionAmt)
	// 止损
	logger.Infof("止损 (entryPrice-nowPrice)/entryPrice=%f, g.StopRate=%f", (entryPrice-nowPrice)/entryPrice, g.StopRate)
	if entryPrice > nowPrice && (entryPrice-nowPrice)/entryPrice > g.StopRate {
		g.Trade(futures.SideTypeSell, 0, 10000*g.GridPointAmount)
		return nil
	}

	// 止盈
	logger.Infof("止盈 (nowPrice-entryPrice)/nowPrice=%f, g.WinRate=%f", (nowPrice-entryPrice)/nowPrice, g.WinRate)
	if positionAmt > 0 && nowPrice > entryPrice && (nowPrice-entryPrice)/nowPrice > g.WinRate {
		g.Trade(futures.SideTypeSell, 0, positionAmt)
		return nil
	}

	// 10倍杠杆
	logger.Info("购买金额 (positionAmt+g.GridPointAmount)*nowPrice=", (positionAmt+g.GridPointAmount)*nowPrice)
	logger.Info("剩余金额 availableBalance*9=", availableBalance*9)
	if g.Minute1Buy() == nil && (positionAmt+g.GridPointAmount)*nowPrice < availableBalance*9 {
		g.Trade(futures.SideTypeBuy, 0, positionAmt+g.GridPointAmount)
	}

	fmt.Println()

	return nil
}

type KlineData struct {
	Open  float64
	Close float64
	Rate  float64
}

func (g *GridStrategy) Minute1Buy() error {
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

	rates := make([]KlineData, n)
	for i := 0; i < n; i++ {
		logger.Infof("%#v", klines[i])
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
	}
	logger.Info(rates)

	if (rates[3].Close-rates[0].Open)/rates[0].Open < -0.01 && rates[4].Rate > -g.NormalWaveRate {
		return nil
	}

	down := 0
	for i := 0; i < n-2; i++ {
		if rates[i].Rate < 0 {
			down++
			continue
		}
		if rates[i].Rate > g.NormalWaveRate {
			logger.Errorf("rate > NormalWaveRate, rate: %f, NormalWaveRate: %f", rates[i].Rate, g.NormalWaveRate)
			return fmt.Errorf("not buy point")
		}
	}
	if down <= 2 {
		logger.Error("down count <= 2, down:", down)
		return fmt.Errorf("not buy point")
	}

	if rates[n-2].Rate < -g.NormalWaveRate {
		logger.Errorf("last two rate < -NormalWaveRate, rate: %f, NormalWaveRate: %f", rates[n-2].Rate, -g.NormalWaveRate)
		return fmt.Errorf("not buy point")
	}

	if rates[n-1].Rate < -g.NormalWaveRate*5 {
		logger.Errorf("last one rate < -NormalWaveRate*5, rate: %f, NormalWaveRate: %f", rates[n-1].Rate, -g.NormalWaveRate*5)
		return fmt.Errorf("not buy point")
	}

	return nil
}

func (g *GridStrategy) Minute5() {

}

func (g *GridStrategy) Minute15() {

}

func (g *GridStrategy) Hour() {

}

func (g *GridStrategy) Day() {

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

func (g *GridStrategy) getNewestPrice() (float64, error) {
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
func (g *GridStrategy) Trade(sideType futures.SideType, price, amount float64) (*futures.Order, error) {
	service := g.client.NewCreateOrderService().
		Symbol(g.Symbol).Quantity(fmt.Sprint(amount)).Side(sideType).Type(futures.OrderTypeMarket)
	if price > 0 {
		service = service.Price(fmt.Sprint(price))
	}

	// SHORT BOTH
	service = service.PositionSide(futures.PositionSideType(g.PositionSide))
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

func (g *GridStrategy) getAvailableBalance(asset string) (float64, error) {
	balances, err := g.client.NewGetBalanceService().Do(context.Background())
	if err != nil {
		logger.Error(err)
		return 0.0, err
	}
	for i := 0; i < len(balances); i++ {
		if balances[i].Asset == asset {
			logger.Infof("%#v", balances[i])
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

func (g *GridStrategy) Position() (*futures.AccountPosition, error) {
	account, err := g.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	// {Asset:"USDT", InitialMargin:"2.73550490", MaintMargin:"0.27355049", MarginBalance:"6.26307065", MaxWithdrawAmount:"3.52756575", OpenOrderInitialMargin:"0.00000000", PositionInitialMargin:"2.73550490", UnrealizedProfit:"0.02344900", WalletBalance:"6.23962165"}
	for i := 0; i < len(account.Assets); i++ {
		if account.Assets[i].Asset == "USDT" {
			logger.Infof("%#v", account.Assets[i])
		}
	}

	// {Isolated:false, Leverage:"10", InitialMargin:"2.73550490", MaintMargin:"0.27355049", OpenOrderInitialMargin:"0", PositionInitialMargin:"2.73550490", Symbol:"DOGEUSDT", UnrealizedProfit:"0.02344900", EntryPrice:"0.273316", MaxNotional:"100000", PositionSide:"LONG", PositionAmt:"100"}
	for i := 0; i < len(account.Positions); i++ {
		if account.Positions[i].Symbol == g.Symbol && account.Positions[i].PositionSide == futures.PositionSideType(g.PositionSide) {
			logger.Infof("%s: %#v", g.Symbol, account.Positions[i])
			return account.Positions[i], nil
		}
	}
	return nil, fmt.Errorf("not existed")
}

func (g *GridStrategy) Run() error {
	g.OnTick()
	tk := time.NewTicker(1 * time.Minute)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			g.OnTick()
		}
	}
	return nil
}

func main() {
	logger.Info("grid strategy start")

	grid := &GridStrategy{
		client: futures.NewClient(ApiKey, SecretKey),

		PositionSide:    PositionSide,
		Symbol:          Symbol,
		Direction:       Direction,
		GridNum:         GridNum,
		GridPointAmount: GridPointAmount,
		GridPointDis:    GridPointDis,
		GridCovDis:      GridCovDis,

		WinRate:        0.05,
		StopRate:       0.1,
		NormalWaveRate: 0.003,
	}
	grid.Run()
}
