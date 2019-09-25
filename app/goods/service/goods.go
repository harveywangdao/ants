package service

import (
	"context"
	"errors"
	"strings"

	"github.com/harveywangdao/ants/app/goods/model"
	"github.com/harveywangdao/ants/common"
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

func (s *Service) DeductStock(ctx context.Context, req *proto.DeductStockRequest) (*proto.DeductStockResponse, error) {
	if req.GoodsID == "" || req.OrderID == "" || req.PayID == "" {
		return nil, errors.New("param can not be null")
	}

	if req.Number == 0 {
		return nil, errors.New("deduct stock number is 0")
	}

	tx := s.db.Begin()
	if err := tx.Error; err != nil {
		return nil, err
	} else {
		// 查询库存
		var goods model.GoogsModel
		if err := tx.Where("goods_id = ?", req.GoodsID).First(&goods).Error; err != nil {
			logger.Error(err)
			tx.Rollback()
			return nil, err
		}

		if goods.Stock < int32(req.Number) {
			tx.Rollback()
			return &proto.DeductStockResponse{
				Code:    common.ErrStockIsNotEnough,
				CodeMsg: "stock is not enough",
			}, nil
		}

		if err := tx.Create(&model.PurchaseRecordModel{
			GoodsID: req.GoodsID,
			OrderID: req.OrderID,
			PayID:   req.PayID,
		}).Error; err != nil {
			logger.Error(err)
			tx.Rollback()

			if strings.Contains(err.Error(), "Duplicate entry") {
				return &proto.DeductStockResponse{
					Code:    common.ErrDeductStockRepeat,
					CodeMsg: "deduct stock repeat",
				}, nil
			}

			return nil, err
		}

		// 扣库存,能解决超卖问题,但是性能不高,适合并发少的情况
		result := tx.Exec("UPDATE goods_tb SET stock = stock - ? WHERE goods_id = ? AND stock >= ?", req.Number, req.GoodsID, req.Number)
		if err := result.Error; err != nil {
			logger.Error(err)
			tx.Rollback()
			return nil, err
		}

		logger.Debug("RowsAffected:", result.RowsAffected)

		if result.RowsAffected == 0 {
			tx.Rollback()
			return &proto.DeductStockResponse{
				Code:    common.ErrStockIsNotEnough,
				CodeMsg: "stock is not enough",
			}, nil
		}

		if err := tx.Commit().Error; err != nil {
			return nil, err
		}
	}

	return &proto.DeductStockResponse{
		CodeMsg: "deduct stock success",
	}, nil
}
