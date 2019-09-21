package model

import (
	"github.com/harveywangdao/ants/logger"
	"github.com/jinzhu/gorm"
	"time"
)

type GoogsModel struct {
	ID          int64     `gorm:"column:id"`
	UserID      string    `gorm:"column:user_id"`
	Name        string    `gorm:"column:name"`
	IdentityNo  string    `gorm:"column:identity_no"`
	Age         uint32    `gorm:"column:age"`
	Gender      uint8     `gorm:"column:gender"`
	PhoneNumber string    `gorm:"column:phone_number"`
	Email       string    `gorm:"column:email"`
	Remark      string    `gorm:"column:remark"`
	CreateTime  time.Time `gorm:"column:create_time;-"`
	UpdateTime  time.Time `gorm:"column:update_time;-"`
	IsDelete    uint8     `gorm:"column:is_delete"`
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
