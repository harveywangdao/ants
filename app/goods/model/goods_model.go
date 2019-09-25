package model

import (
	"github.com/harveywangdao/ants/logger"
	"github.com/jinzhu/gorm"
	"time"
)

type GoogsModel struct {
	ID         int64     `gorm:"column:id"`
	GoodsID    string    `gorm:"column:goods_id"`
	SellerID   string    `gorm:"column:seller_id"`
	GoodsName  string    `gorm:"column:goods_name"`
	Price      float64   `gorm:"column:price"`
	Category   uint32    `gorm:"column:category"`
	Stock      int32     `gorm:"column:stock"`
	Brand      string    `gorm:"column:brand"`
	Remark     string    `gorm:"column:remark"`
	CreateTime time.Time `gorm:"column:create_time;-"`
	UpdateTime time.Time `gorm:"column:update_time;-"`
	IsDelete   uint8     `gorm:"column:is_delete"`
}

func (m GoogsModel) TableName() string {
	return "goods_tb"
}

func GetGoodsListByCategory(db *gorm.DB, category string) ([]*GoogsModel, error) {
	var goodsList []*GoogsModel
	err := db.Where("category = ?", category).Find(&goodsList).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return goodsList, nil
}

type PurchaseRecordModel struct {
	ID         int64     `gorm:"column:id"`
	GoodsID    string    `gorm:"column:goods_id"`
	OrderID    string    `gorm:"column:order_id"`
	PayID      string    `gorm:"column:pay_id"`
	Remark     string    `gorm:"column:remark"`
	CreateTime time.Time `gorm:"column:create_time;-"`
	UpdateTime time.Time `gorm:"column:update_time;-"`
	IsDelete   uint8     `gorm:"column:is_delete"`
}

func (m PurchaseRecordModel) TableName() string {
	return "purchase_record_tb"
}
