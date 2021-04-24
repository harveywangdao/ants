package dataloader

import (
	. "github.com/harveywangdao/ants/app/strategy/grid/crex"
	"time"
)

type DataLoader interface {
	Setup(start time.Time, end time.Time) error
	ReadOrderBooks() []*OrderBook
	ReadRecords(limit int) []*Record
	HasMoreData() bool
}
