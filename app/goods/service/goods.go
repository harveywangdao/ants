package service

import (
	"context"
	"errors"

	"github.com/harveywangdao/ants/app/goods/model"
	"github.com/harveywangdao/ants/logger"
	proto "github.com/harveywangdao/ants/rpc/goods"
	"github.com/harveywangdao/ants/util"
)

func (s *Service) AddGoods(ctx context.Context, req *proto.AddGoodsRequest) (*proto.AddGoodsResponse, error) {
	if req.Name == "" {
		return nil, errors.New("goods name is null")
	}

	goodsID := util.GetUUID()

	goods := &model.GoogsModel{
		GoodsID:   goodsID,
		GoodsName: req.Name,
		Price:     req.Price,
		Category:  req.Category,
		Stock:     req.Stock,
		Brand:     req.Brand,
	}

	err := s.db.Create(goods).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &proto.AddGoodsResponse{
		GoodsID: goodsID,
	}, nil
}

func (s *Service) GetGoods(ctx context.Context, req *proto.GetGoodsRequest) (*proto.GetGoodsResponse, error) {
	if req.GoodsID == "" {
		return nil, errors.New("goodsID is null")
	}

	var goods model.GoogsModel
	err := s.db.Where("goods_id = ?", req.GoodsID).First(&goods).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Info(goods)

	resp := &proto.GetGoodsResponse{
		GoodsInfo: &proto.GoodsInfo{
			Name:     goods.GoodsName,
			Price:    goods.Price,
			Stock:    goods.Stock,
			Category: goods.Category,
			Brand:    goods.Brand,
		},
	}

	return resp, nil
}

func (s *Service) GetGoodsListByCategory(ctx context.Context, req *proto.GetGoodsListByCategoryRequest) (*proto.GetGoodsListByCategoryResponse, error) {
	var goodsList []*model.GoogsModel
	err := s.db.Select("goods_id").Where("category = ?", req.Category).Find(&goodsList).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	var goodsIDList []string
	for _, v := range goodsList {
		goodsIDList = append(goodsIDList, v.GoodsID)
	}

	return &proto.GetGoodsListByCategoryResponse{
		GoodsIDList: goodsIDList,
	}, nil
}

func (s *Service) ModifyGoodsInfo(ctx context.Context, req *proto.ModifyGoodsInfoRequest) (*proto.ModifyGoodsInfoResponse, error) {
	if req.GoodsID == "" {
		return nil, errors.New("goodsID is null")
	}

	param := map[string]interface{}{
		"goods_name": req.GoodsInfo.Name,
		"price":      req.GoodsInfo.Price,
		"category":   req.GoodsInfo.Category,
		"stock":      req.GoodsInfo.Stock,
		"brand":      req.GoodsInfo.Brand,
	}

	err := s.db.Model(model.GoogsModel{}).Where("goods_id = ?", req.GoodsID).Updates(param).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &proto.ModifyGoodsInfoResponse{
		CodeMsg: "modify success",
	}, nil
}

func (s *Service) DelGoods(ctx context.Context, req *proto.DelGoodsRequest) (*proto.DelGoodsResponse, error) {
	if req.GoodsID == "" {
		return nil, errors.New("goodsID is null")
	}

	err := s.db.Where("goods_id = ?", req.GoodsID).Delete(model.GoogsModel{}).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &proto.DelGoodsResponse{
		CodeMsg: "delete success",
	}, nil
}
