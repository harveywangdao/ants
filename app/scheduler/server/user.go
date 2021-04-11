package server

import (
	"net/http"
	"strconv"

	"github.com/harveywangdao/ants/app/scheduler/model"
	"github.com/harveywangdao/ants/app/scheduler/util"
	"github.com/harveywangdao/ants/app/scheduler/util/logger"

	"github.com/gin-gonic/gin"
)

type AddUserReq struct {
	Name        string `json:"name"`
	IdentityNo  string `json:"identityNo"`
	Age         uint32 `json:"age"`
	Gender      uint8  `json:"gender"`
	PhoneNumber string `json:"phoneNumber"`
	Email       string `json:"email"`
	Remark      string `json:"remark"`
}

func (s *HttpService) AddUser(c *gin.Context) {
	req := AddUserReq{}
	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	if req.Email == "" {
		AbortWithErrMsg(c, http.StatusBadRequest, "email can not be empty")
		return
	}

	user := &model.UserModel{
		UserID:      util.GetUUID(),
		Name:        req.Name,
		IdentityNo:  req.IdentityNo,
		Age:         req.Age,
		Gender:      uint8(req.Gender),
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		Remark:      req.Remark,
		Password:    util.GetUUID(),
	}
	if err := s.db.Create(user).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userId":   user.UserID,
		"password": user.Password,
	})
}

func (s *HttpService) QueryUserList(c *gin.Context) {
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

	var users []*model.UserModel
	if err := s.db.Raw("SELECT * FROM user_tb LIMIT ?,?", offset, count).Scan(&users).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, "query user list fail: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}

func (s *HttpService) UserLogin(c *gin.Context) {

}

type ChangePasswordReq struct {
	UserID      string `json:"userId"`
	PhoneNumber string `json:"phonNumber"`
	Email       string `json:"email"`
	Password    string `gorm:"password"`
}

func (s *HttpService) ChangePassword(c *gin.Context) {
	req := ChangePasswordReq{}
	if err := c.BindJSON(&req); err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusBadRequest, err.Error())
		return
	}

	param := map[string]interface{}{
		"password": req.Password,
	}
	if err := s.db.Model(model.UserModel{}).Where("user_id = ?", req.UserID).Updates(param).Error; err != nil {
		logger.Error(err)
		AbortWithErrMsg(c, http.StatusInternalServerError, err.Error())
		return
	}
}
