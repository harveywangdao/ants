package model

import (
	"time"

	"github.com/harveywangdao/ants/app/scheduler/util/logger"
	"github.com/jinzhu/gorm"
	//"github.com/shopspring/decimal"
)

type UserModel struct {
	ID          int64     `gorm:"column:id" json:"id"`
	UserID      string    `gorm:"column:user_id" json:"user_id"`
	OpenID      string    `gorm:"column:open_id" json:"open_id"`
	UnionID     string    `gorm:"column:union_id" json:"union_id"`
	SessionKey  string    `gorm:"column:session_key" json:"session_key"`
	Name        string    `gorm:"column:name" json:"name"`
	IdentityNo  string    `gorm:"column:identity_no" json:"identity_no"`
	Age         uint32    `gorm:"column:age" json:"age"`
	Gender      uint8     `gorm:"column:gender" json:"gender"`
	PhoneNumber string    `gorm:"column:phone_number" json:"phone_number"`
	Email       string    `gorm:"column:email" json:"email"`
	Remark      string    `gorm:"column:remark" json:"remark"`
	Password    string    `gorm:"column:password" json:"password"`
	CreateTime  time.Time `gorm:"column:create_time;-" json:"create_time"`
	UpdateTime  time.Time `gorm:"column:update_time;-" json:"update_time"`
	IsDelete    uint8     `gorm:"column:is_delete" json:"-"`
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

//Rate decimal.Decimal `gorm:"column:rate"`
type ApiKeyModel struct {
	ID         int64     `gorm:"column:id" json:"id"`
	UserID     string    `gorm:"column:user_id" json:"user_id"`
	ApiKey     string    `gorm:"column:api_key" json:"api_key"`
	SecretKey  string    `gorm:"column:secret_key" json:"secret_key"`
	Passphrase string    `gorm:"column:passphrase" json:"passphrase"`
	Exchange   string    `gorm:"column:exchange" json:"exchange"`
	CreateTime time.Time `gorm:"column:create_time;-" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;-" json:"update_time"`
	IsDelete   uint8     `gorm:"column:is_delete" json:"-"`
}

func (u ApiKeyModel) TableName() string {
	return "apikey_tb"
}

type StrategyModel struct {
	ID         int64     `gorm:"column:id" json:"id"`
	Strategy   string    `gorm:"column:strategy" json:"strategy"`
	Desc       string    `gorm:"column:desc" json:"desc"`
	Param      string    `gorm:"column:param" json:"param"`
	CreateTime time.Time `gorm:"column:create_time;-" json:"create_time"`
	UpdateTime time.Time `gorm:"column:update_time;-" json:"update_time"`
	IsDelete   uint8     `gorm:"column:is_delete" json:"-"`
}

func (u StrategyModel) TableName() string {
	return "strategy_tb"
}

type TemplateModel struct {
	ID           int64     `gorm:"column:id" json:"id"`
	StrategyId   int64     `gorm:"column:strategy_id" json:"strategy_id"`
	TemplateName string    `gorm:"column:template_name" json:"template_name"`
	Param        string    `gorm:"column:param" json:"param"`
	CreateTime   time.Time `gorm:"column:create_time;-" json:"create_time"`
	UpdateTime   time.Time `gorm:"column:update_time;-" json:"update_time"`
	IsDelete     uint8     `gorm:"column:is_delete" json:"-"`
}

func (u TemplateModel) TableName() string {
	return "strategy_template_tb"
}
