package model

import (
	"time"

	"github.com/harveywangdao/ants/app/scheduler/util/logger"

	"github.com/jinzhu/gorm"
	//"github.com/shopspring/decimal"
)

type UserModel struct {
	ID          int64     `gorm:"column:id"`
	UserID      string    `gorm:"column:user_id"`
	OpenID      string    `gorm:"column:open_id"`
	UnionID     string    `gorm:"column:union_id"`
	SessionKey  string    `gorm:"column:session_key"`
	Name        string    `gorm:"column:name"`
	IdentityNo  string    `gorm:"column:identity_no"`
	Age         uint32    `gorm:"column:age"`
	Gender      uint8     `gorm:"column:gender"`
	PhoneNumber string    `gorm:"column:phone_number"`
	Email       string    `gorm:"column:email"`
	Remark      string    `gorm:"column:remark"`
	Password    string    `gorm:"column:password"`
	CreateTime  time.Time `gorm:"column:create_time;-"`
	UpdateTime  time.Time `gorm:"column:update_time;-"`
	IsDelete    uint8     `gorm:"column:is_delete"`
}

func (u UserModel) TableName() string {
	return "user_tb"
}

func GetUserByName(db *gorm.DB, name string) ([]*UserModel, error) {
	var users []*UserModel
	err := db.Where("name = ?", name).Find(&users).Error
	if err != nil {
		logger.Println(err)
		return nil, err
	}
	return users, nil
}
