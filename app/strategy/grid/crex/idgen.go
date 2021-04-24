package crex

import (
	"github.com/harveywangdao/ants/app/strategy/grid/crex/utils"
	"strconv"
)

var idGen *utils.IdGenerate

func SetIdGenerate(g *utils.IdGenerate) {
	idGen = g
}

func GenOrderId() string {
	id := idGen.Next()
	return strconv.Itoa(int(id))
}
