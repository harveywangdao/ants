package model

import (
	"github.com/harveywangdao/ants/logger"
	"github.com/jinzhu/gorm"
	"time"
)

type ArticleModel struct {
	ID         int64     `gorm:"column:id"`
	ArticleID  string    `gorm:"column:article_id"`
	UserID     string    `gorm:"column:user_id"`
	Title      string    `gorm:"column:title"`
	Content    string    `gorm:"column:content"`
	Tags       string    `gorm:"column:tags"`
	CreateTime time.Time `gorm:"column:create_time;-"`
	UpdateTime time.Time `gorm:"column:update_time;-"`
	IsDelete   uint8     `gorm:"column:is_delete"`
}

func (u ArticleModel) TableName() string {
	return "article_tb"
}

/*
page从0开始
*/
func GetArticlesByUserIDAndPage(db *gorm.DB, userID string, page, numPerPage int64) ([]*ArticleModel, error) {
	var articles []*ArticleModel
	if err := db.Where("user_id = ?", userID).Order("id desc").Limit(numPerPage).Offset(page * numPerPage).Find(&articles).Error; err != nil {
		logger.Error(err)
		return nil, err
	}

	return articles, nil
}

func GetArticlesByPage(db *gorm.DB, page, numPerPage int64) ([]*ArticleModel, error) {
	var articles []*ArticleModel
	if err := db.Where("is_delete = ?", 0).Order("id desc").Limit(numPerPage).Offset(page * numPerPage).Find(&articles).Error; err != nil {
		logger.Error(err)
		return nil, err
	}

	return articles, nil
}

func GetArticlesByPage2(db *gorm.DB, page, numPerPage int64) ([]*ArticleModel, error) {
	var article ArticleModel
	//`select * from article_tb where id<=(select id from article_tb where is_delete=0 oreder by id desc limit 1000000,1) and is_delete=0 oreder by id desc limit 10`
	if err := db.Select("id").Where("is_delete = ?", 0).Order("id desc").Offset(page * numPerPage).Take(&article).Error; err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Info("article.id:", article.ID)

	var articles []*ArticleModel
	if err := db.Where("id <= ? and is_delete = ?", article.ID, 0).Order("id desc").Limit(numPerPage).Find(&articles).Error; err != nil {
		logger.Error(err)
		return nil, err
	}

	return articles, nil
}

func GetArticlesByPage3(db *gorm.DB, lastId, numPerPage int64) ([]*ArticleModel, error) {
	logger.Info("lastId:", lastId)

	var articles []*ArticleModel
	if err := db.Where("id <= ? and is_delete = ?", lastId, 0).Order("id desc").Limit(numPerPage).Find(&articles).Error; err != nil {
		logger.Error(err)
		return nil, err
	}

	return articles, nil
}
