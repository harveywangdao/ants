package server

import (
	"net/http"
	"strconv"

	"github.com/harveywangdao/ants/app/scheduler/model"
	"github.com/harveywangdao/ants/app/scheduler/util/logger"

	"github.com/gin-gonic/gin"
)

func (s *HttpService) AddApikey(c *gin.Context) {
	req := model.ApiKeyModel{}
	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.UserID == "" || req.ApiKey == "" || req.SecretKey == "" || req.Exchange == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}

	apiKey := model.ApiKeyModel{
		UserID:     req.UserID,
		ApiKey:     req.ApiKey,
		SecretKey:  req.SecretKey,
		Passphrase: req.Passphrase,
		Exchange:   req.Exchange,
	}
	if err := s.db.Create(&apiKey).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func (s *HttpService) QueryUserApikeys(c *gin.Context) {
	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil || offset < 0 {
		AbortWithErrMsg(c, http.StatusBadRequest, "offset param error")
		return
	}

	count, err := strconv.Atoi(c.Query("count"))
	if err != nil || count <= 0 {
		AbortWithErrMsg(c, http.StatusBadRequest, "count param error")
		return
	}

	userId := c.Query("userId")
	if userId == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "userId can not be empty")
		return
	}

	var list []*model.ApiKeyModel
	if err := s.db.Raw("SELECT * FROM apikey_tb WHERE user_id=? LIMIT ?,?", userId, offset, count).Scan(&list).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"apiKeys": list,
	})
}
