package service

import (
	"context"
	"errors"

	"github.com/harveywangdao/ants/app/user/model"
	"github.com/harveywangdao/ants/logger"
	userpb "github.com/harveywangdao/ants/rpc/user"
	"github.com/harveywangdao/ants/util"
	"github.com/jinzhu/gorm"
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

	logger.Info(user)

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

func (s *Service) GetUserIdByPhoneNumber(ctx context.Context, req *userpb.GetUserIdByPhoneNumberRequest) (*userpb.GetUserIdByPhoneNumberResponse, error) {
	if req.PhoneNumber == "" {
		return nil, errors.New("phone number is null")
	}

	var user model.UserModel
	/*err := s.db.Select("user_id, create_time, update_time, is_delete").Where("phone_number = ?", req.PhoneNumber).First(&user).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}*/

	err := s.db.Raw("SELECT user_id, create_time, update_time, is_delete FROM user_tb WHERE phone_number = ?", req.PhoneNumber).Scan(&user).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Info(user)

	return &userpb.GetUserIdByPhoneNumberResponse{
		UserID: user.UserID,
	}, nil
}

func (s *Service) GetUsersByName(ctx context.Context, req *userpb.GetUsersByNameRequest) (*userpb.GetUsersByNameResponse, error) {
	if req.Name == "" {
		return nil, errors.New("name is null")
	}

	var users []*model.UserModel
	/*err := s.db.Select("user_id, create_time, update_time, is_delete").Where("phone_number = ?", req.PhoneNumber).First(&user).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}*/

	err := s.db.Raw("SELECT user_id, create_time, update_time, is_delete FROM user_tb WHERE name = ?", req.Name).Scan(&users).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var userIDList []string
	for _, v := range users {
		logger.Info(*v)
		userIDList = append(userIDList, v.UserID)
	}

	return &userpb.GetUsersByNameResponse{
		UserIDList: userIDList,
	}, nil
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

func (s *Service) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	if req.WxUserInfo == nil {
		return nil, errors.New("WxUserInfo is null")
	}
	logger.Infof("WxUserInfo: %+v", req.WxUserInfo)

	if req.WxUserInfo.Code == "" {
		return nil, errors.New("code is null")
	}
	logger.Info("wx code:", req.WxUserInfo.Code)

	userInfo, err := code2Session(req.WxUserInfo.Code)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	if userInfo.Openid == "" {
		logger.Error("openid is null, wx code:", req.WxUserInfo.Code)
		return nil, errors.New("openid is null")
	}

	user := &model.UserModel{}
	if err := s.db.Where("open_id = ?", userInfo.Openid).First(user).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			logger.Error(err)
			return nil, err
		} else {
			user = &model.UserModel{
				UserID:     util.GetUUID(),
				OpenID:     userInfo.Openid,
				UnionID:    userInfo.Unionid,
				SessionKey: userInfo.SessionKey,
			}

			if err := s.db.Create(user).Error; err != nil {
				logger.Error(err)
				return nil, err
			}
		}
	} else {
		if user.SessionKey != userInfo.SessionKey {
			err = s.db.Model(model.UserModel{}).Where("open_id = ?", userInfo.Openid).Updates(map[string]interface{}{
				"session_key": userInfo.SessionKey,
			}).Error
			if err != nil {
				logger.Error(err)
				return nil, err
			}
			user.SessionKey = userInfo.SessionKey
		}
	}
	logger.Infof("%+v", user)

	return &userpb.LoginResponse{
		UserID:  user.UserID,
		CodeMsg: "login success",
	}, nil
}

func (s *Service) DecryptWxUserInfo(ctx context.Context, req *userpb.DecryptWxUserInfoRequest) (*userpb.DecryptWxUserInfoResponse, error) {
	if req.UserID == "" || req.WxUserInfo == nil {
		return nil, errors.New("params error")
	}
	logger.Infof("UserID: %s, WxUserInfo: %+v", req.UserID, req.WxUserInfo)

	user := &model.UserModel{}
	if err := s.db.Where("user_id = ?", req.UserID).First(user).Error; err != nil {
		logger.Error(err)
		return nil, err
	}

	if !checkWxUserInfoSign(req.WxUserInfo.RawData, user.SessionKey, req.WxUserInfo.Signature) {
		logger.Error("wx user info sign fail")
		return nil, errors.New("wx user info sign fail")
	}

	wxUserInfoData, err := decryptWxUserInfo(req.WxUserInfo.EncryptedData, req.WxUserInfo.Iv, user.SessionKey)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	logger.Info("wxUserInfoData:", string(wxUserInfoData))

	return &userpb.DecryptWxUserInfoResponse{
		WxUserInfoData: string(wxUserInfoData),
		CodeMsg:        "DecryptWxUserInfo success",
	}, nil
}

func (s *Service) AddUserInfo(ctx context.Context, req *userpb.AddUserInfoRequest) (*userpb.AddUserInfoResponse, error) {
	if req.UserID == "" || req.UserInfo == nil {
		return nil, errors.New("params error")
	}
	logger.Infof("UserID: %s, UserInfo: %+v", req.UserID, req.UserInfo)

	user := &model.UserModel{}
	if err := s.db.Where("user_id = ?", req.UserID).First(user).Error; err != nil {
		logger.Error(err)
		return nil, err
	}

	param := make(map[string]interface{})
	param["age"] = req.UserInfo.Age
	if req.UserInfo.Name != "" {
		param["name"] = req.UserInfo.Name
	}
	if req.UserInfo.IdentityNo != "" {
		param["identity_no"] = req.UserInfo.IdentityNo
	}
	if req.UserInfo.Gender == 0 || req.UserInfo.Gender == 1 {
		param["gender"] = req.UserInfo.Gender
	}
	if req.UserInfo.Email != "" {
		param["email"] = req.UserInfo.Email
	}
	if req.UserInfo.PhoneNumber != "" {
		param["phone_number"] = req.UserInfo.PhoneNumber
	}

	err := s.db.Model(model.UserModel{}).Where("user_id = ?", req.UserID).Updates(param).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &userpb.AddUserInfoResponse{
		CodeMsg: "AddUserInfo success",
	}, nil
}
