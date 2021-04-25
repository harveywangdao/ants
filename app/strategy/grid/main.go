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
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
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

	PositionSide string
	Symbol       string  `opt:"symbol,BTUSDT"`
	Direction    float64 `opt:"direction,1"` // 网格方向 up 1, down -1

	GridNum         int     `opt:"grid_num,10"`         // 网格节点数量 10
	GridPointAmount float64 `opt:"grid_point_amount,1"` // 网格节点下单量 1
	GridPointDis    float64 `opt:"grid_point_dis,20"`   // 网格节点间距 20
	GridCovDis      float64 `opt:"grid_cov_dis,50"`     // 网格节点平仓价差 50
}

func (g *GridStrategy) OnInit() error {
	logger.Infof("PositionSide: %v", g.PositionSide)
	logger.Infof("Symbol: %v", g.Symbol)
	logger.Infof("Direction: %v", g.Direction)
	logger.Infof("GridNum: %v", g.GridNum)
	logger.Infof("GridPointAmount: %v", g.GridPointAmount)
	logger.Infof("GridPointDis: %v", g.GridPointDis)
	logger.Infof("GridCovDis: %v", g.GridCovDis)
	return nil
}

func (g *GridStrategy) OnTick() error {
	res, err := g.client.NewDepthService().Symbol(g.Symbol).Limit(5).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("depth asks:", res.Asks)
	logger.Info("depth bids:", res.Bids)

	nowAskPrice, nowBidPrice := res.Asks[0], res.Bids[0]
	logger.Infof("nowAskPrice=%v, nowBidPrice=%v", nowAskPrice, nowBidPrice)

	// 1m 3m 5m 15m 30m 1h 2h 4h 6h 8h 12h 1d 3d 1w 1M
	klines, err := g.client.NewKlinesService().Symbol(g.Symbol).Interval("1m").Limit(50).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}
	for i := 0; i < len(klines); i++ {
		logger.Info(*klines[i])
	}

	g.Trade(futures.SideTypeSell, 0, g.GridPointAmount)
	time.Sleep(time.Second * 5)
	g.Trade(futures.SideTypeBuy, 0, g.GridPointAmount)

	return nil
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

func (g *GridStrategy) GetBalance(asset string) float64 {
	balances, err := g.client.NewGetBalanceService().Do(context.Background())
	if err != nil {
		logger.Error(err)
		return 0.0
	}
	for i := 0; i < len(balances); i++ {
		if balances[i].Asset == asset {
			logger.Infof("%#v", balances[i])
			availableBalance, err := strconv.ParseFloat(balances[i].AvailableBalance, 64)
			if err != nil {
				logger.Error("availableBalance parse float64 fail, err", err)
				return 0.0
			}
			logger.Info("availableBalance:", availableBalance)
			return availableBalance
		}
	}
	return 0.0
}

func (g *GridStrategy) Account() error {
	account, err := g.client.NewGetAccountService().Do(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}
	for i := 0; i < len(account.Assets); i++ {
		if account.Assets[i].Asset == "USDT" {
			logger.Infof("%#v", account.Assets[i])
		}
	}
	for i := 0; i < len(account.Positions); i++ {
		if account.Positions[i].Symbol == g.Symbol && account.Positions[i].PositionSide == futures.PositionSideType(g.PositionSide) {
			logger.Infof("%s: %#v", g.Symbol, account.Positions[i])
		}
	}
	return nil
}

func (g *GridStrategy) Run() error {
	for {
		g.OnTick()
		time.Sleep(2 * time.Second)
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
	}
	grid.OnInit()
	grid.Run()
}
