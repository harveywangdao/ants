package model

import (
	//"github.com/harveywangdao/ants/logger"
	//"github.com/jinzhu/gorm"
	"time"
)

type OrderModel struct {
	ID         int64     `gorm:"column:id"`
	OrderID    string    `gorm:"column:order_id"`
	SellerID   string    `gorm:"column:seller_id"`
	BuyerID    string    `gorm:"column:buyer_id"`
	GoodsID    string    `gorm:"column:goods_id"`
	GoodsName  string    `gorm:"column:goods_name"`
	Count      uint32    `gorm:"column:count"`
	Price      float64   `gorm:"column:price"`
	Pay        float64   `gorm:"column:pay"`
	Status     uint8     `gorm:"column:status"`
	Remark     string    `gorm:"column:remark"`
	CreateTime time.Time `gorm:"column:create_time;-"`
	UpdateTime time.Time `gorm:"column:update_time;-"`
	IsDelete   uint8     `gorm:"column:is_delete"`
}

func (m OrderModel) TableName() string {
	return "order_tb"
}
