package exchanges

import (
	"fmt"
	. "github.com/harveywangdao/ants/app/strategy/grid/crex"
	"github.com/harveywangdao/ants/app/strategy/grid/crex/exchanges/binancefutures"
	"github.com/harveywangdao/ants/app/strategy/grid/crex/exchanges/bitmex"
	"github.com/harveywangdao/ants/app/strategy/grid/crex/exchanges/bybit"
	"github.com/harveywangdao/ants/app/strategy/grid/crex/exchanges/deribit"
	"github.com/harveywangdao/ants/app/strategy/grid/crex/exchanges/hbdm"
	"github.com/harveywangdao/ants/app/strategy/grid/crex/exchanges/hbdmswap"
	"github.com/harveywangdao/ants/app/strategy/grid/crex/exchanges/okexfutures"
	"github.com/harveywangdao/ants/app/strategy/grid/crex/exchanges/okexswap"
)

func NewExchange(name string, opts ...ApiOption) Exchange {
	params := &Parameters{}

	for _, opt := range opts {
		opt(params)
	}

	return NewExchangeFromParameters(name, params)
}

func NewExchangeFromParameters(name string, params *Parameters) Exchange {
	switch name {
	case BinanceFutures:
		return binancefutures.NewBinanceFutures(params)
	case BitMEX:
		return bitmex.NewBitMEX(params)
	case Deribit:
		return deribit.NewDeribit(params)
	case Bybit:
		return bybit.NewBybit(params)
	case Hbdm:
		return hbdm.NewHbdm(params)
	case HbdmSwap:
		return hbdmswap.NewHbdmSwap(params)
	case OkexFutures:
		return okexfutures.NewOkexFutures(params)
	case OkexSwap:
		return okexswap.NewOkexSwap(params)
	default:
		panic(fmt.Sprintf("new exchange error [%v]", name))
	}
}
