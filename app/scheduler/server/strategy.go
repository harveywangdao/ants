package server

import (
	"net/http"
	"strconv"

	"github.com/harveywangdao/ants/app/scheduler/model"
	"github.com/harveywangdao/ants/app/scheduler/util/logger"

	"github.com/gin-gonic/gin"
)

func (s *HttpService) AddStrategy(c *gin.Context) {
	req := model.StrategyModel{}
	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.Strategy == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}

	strategy := model.StrategyModel{
		Strategy: req.Strategy,
		Param:    req.Param,
	}
	if err := s.db.Create(&strategy).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func (s *HttpService) QueryStrategies(c *gin.Context) {
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

	var list []*model.StrategyModel
	if err := s.db.Raw("SELECT * FROM strategy_tb LIMIT ?,?", offset, count).Scan(&list).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"strategies": list,
	})
}

func (s *HttpService) UpdateStrategy(c *gin.Context) {
	req := model.StrategyModel{}
	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.Strategy == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}

	param := map[string]interface{}{
		"param": req.Param,
	}
	if err := s.db.Model(model.StrategyModel{}).Where("strategy = ?", req.Strategy).Updates(param).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func (s *HttpService) DelStrategy(c *gin.Context) {
	req := model.StrategyModel{}
	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.Strategy == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}

	if err := s.db.Where("strategy = ?", req.Strategy).Delete(model.StrategyModel{}).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}
