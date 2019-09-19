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

	return &userpb.GetUserResponse{}, nil
}

func (s *Service) DelUser(ctx context.Context, req *userpb.DelUserRequest) (*userpb.DelUserResponse, error) {

	return &userpb.DelUserResponse{}, nil
}

func (s *Service) ModifyUserInfo(ctx context.Context, req *userpb.ModifyUserInfoRequest) (*userpb.ModifyUserInfoResponse, error) {

	return &userpb.ModifyUserInfoResponse{}, nil
}
