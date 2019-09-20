package service

import (
	"context"
	"errors"

	"github.com/harveywangdao/ants/app/user/model"
	"github.com/harveywangdao/ants/logger"
	userpb "github.com/harveywangdao/ants/rpc/user"
	"github.com/harveywangdao/ants/util"
)

func (s *Service) AddUser(ctx context.Context, req *userpb.AddUserRequest) (*userpb.AddUserResponse, error) {
	if req.PhoneNumber == "" {
		return nil, errors.New("phone number is null")
	}

	userID := util.GetUUID()

	user := &model.UserModel{
		UserID:      userID,
		Name:        req.Name,
		IdentityNo:  req.IdentityNo,
		Age:         req.Age,
		Gender:      uint8(req.Gender),
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
	}

	err := s.db.Create(user).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &userpb.AddUserResponse{
		UserID: userID,
	}, nil
}

func (s *Service) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	if req.UserID == "" {
		return nil, errors.New("userID is null")
	}

	var user model.UserModel
	err := s.db.Where("user_id = ?", req.UserID).First(&user).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	resp := &userpb.GetUserResponse{
		UserInfo: &userpb.UserInfo{
			Name:        user.Name,
			IdentityNo:  user.IdentityNo,
			Age:         user.Age,
			Gender:      uint32(user.Gender),
			PhoneNumber: user.PhoneNumber,
			Email:       user.Email,
		},
	}

	return resp, nil
}

func (s *Service) ModifyUserInfo(ctx context.Context, req *userpb.ModifyUserInfoRequest) (*userpb.ModifyUserInfoResponse, error) {
	if req.UserID == "" {
		return nil, errors.New("userID is null")
	}

	param := map[string]interface{}{
		"name":        req.UserInfo.Name,
		"identity_no": req.UserInfo.IdentityNo,
		"age":         req.UserInfo.Age,
		"gender":      req.UserInfo.Gender,
		"email":       req.UserInfo.Email,
	}

	err := s.db.Model(model.UserModel{}).Where("user_id = ?", req.UserID).Updates(param).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	/*err := s.db.Table("user_tb").Where("user_id = ?", req.UserID).Updates(param).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}*/

	return &userpb.ModifyUserInfoResponse{
		CodeMsg: "modify success",
	}, nil
}

func (s *Service) DelUser(ctx context.Context, req *userpb.DelUserRequest) (*userpb.DelUserResponse, error) {
	if req.UserID == "" {
		return nil, errors.New("userID is null")
	}

	err := s.db.Where("user_id = ?", req.UserID).Delete(model.UserModel{}).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &userpb.DelUserResponse{
		CodeMsg: "delete success",
	}, nil
}
