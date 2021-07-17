package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
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

type GridStrategy struct {
	client *futures.Client

	Symbol string
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
	return nil
}

func (g *GridStrategy) GetTrades() error {
	trades, err := g.client.NewListAccountTradeService().Symbol(g.Symbol).Limit(20).Do(context.Background())
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

func (g *GridStrategy) GetOpenOrders() error {
	trades, err := g.client.NewListOpenOrdersService().Symbol(g.Symbol).Do(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}

	fmt.Println("当前挂单:", len(trades))
	for i := 0; i < len(trades); i++ {
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
	return nil
}

func (g *GridStrategy) RunTask() error {
	for {
		g.GetTrades()
		g.GetOpenOrders()
		g.GetAccountInfo()
		time.Sleep(time.Second * 10)
	}
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

func main() {
	symbol := flag.String("symbol", "DOGEUSDT", "symbol")
	orderId := flag.Int64("orderId", 0, "orderId")
	action := flag.String("action", "", "action")

	positionSide := flag.String("positionSide", "LONG", "positionSide")
	sideType := flag.String("sideType", "BUY", "sideType")
	price := flag.Float64("price", 0.0, "price")
	amount := flag.Float64("amount", 0.0, "amount")

	flag.Parse()

	grid := &GridStrategy{
		client: futures.NewClient(ApiKey, SecretKey),
		Symbol: *symbol,
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
