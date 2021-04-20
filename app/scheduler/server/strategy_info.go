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
		Desc:     req.Desc,
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
	if req.StrategyId <= 0 {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}

	param := map[string]interface{}{
		"param": req.Param,
	}
	if err := s.db.Model(model.StrategyModel{}).Where("id = ?", req.StrategyId).Updates(param).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func (s *HttpService) DelStrategy(c *gin.Context) {
	strategyId, err := strconv.ParseInt(c.Query("strategyId"), 10, 64)
	if err != nil || strategyId <= 0 {
		AbortWithErrMsg(c, http.StatusBadRequest, "strategyId param error")
		return
	}

	if err := s.db.Where("id = ?", strategyId).Delete(model.StrategyModel{}).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func (s *HttpService) AddStrategyModel(c *gin.Context) {
	req := model.TemplateModel{}
	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.TemplateName == "" || req.StrategyId <= 0 {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}

	template := model.TemplateModel{
		StrategyId:   req.StrategyId,
		TemplateName: req.TemplateName,
		Param:        req.Param,
	}
	if err := s.db.Create(&template).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func (s *HttpService) QueryStrategyModels(c *gin.Context) {
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
	strategyId, err := strconv.ParseInt(c.Query("strategyId"), 10, 64)
	if err != nil || strategyId <= 0 {
		AbortWithErrMsg(c, http.StatusBadRequest, "strategyId param error")
		return
	}

	var list []*model.TemplateModel
	/*if err := s.db.Raw("SELECT * FROM strategy_template_tb WHERE strategy_id=? LIMIT ?,?", strategyId, offset, count).Scan(&list).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}*/

	if err := s.db.Where("strategy_id=?", strategyId).Limit(count).Offset(offset).Find(&list).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"templates": list,
	})
}

func (s *HttpService) UpdateStrategyModel(c *gin.Context) {
	req := model.TemplateModel{}
	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}
	if req.TemplateId <= 0 {
		AbortWithErrMsg(c, http.StatusBadRequest, "param can not be empty")
		return
	}

	param := map[string]interface{}{
		"param": req.Param,
	}
	if err := s.db.Model(model.TemplateModel{}).Where("id = ?", req.TemplateId).Updates(param).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}

func (s *HttpService) DelStrategyModel(c *gin.Context) {
	templateId, err := strconv.ParseInt(c.Query("templateId"), 10, 64)
	if err != nil || templateId <= 0 {
		AbortWithErrMsg(c, http.StatusBadRequest, "templateId param error")
		return
	}

	if err := s.db.Where("id = ?", templateId).Delete(model.TemplateModel{}).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}
