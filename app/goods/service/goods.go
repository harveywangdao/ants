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
	"go.mongodb.org/mongo-driver/bson"
)

func (s *Service) AddGoodsIntoMgo(ctx context.Context, goods *model.GoogsModel) error {
	collection := s.Mongo.Database(s.Config.Mongo.DbName).Collection("product_col")

	// db
	dbs, err := s.Mongo.ListDatabases(context.Background(), bson.M{})
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("dbs:", dbs)

	// collection
	collections, err := s.Mongo.Database(s.Config.Mongo.DbName).ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("collections:", collections)

	// insert
	insertOneResult, err := collection.InsertOne(context.Background(), bson.M{
		"goods_id":   goods.GoodsID,
		"seller_id":  goods.SellerID,
		"goods_name": goods.GoodsName,
		"price":      goods.Price,
		"category":   goods.Category,
		"stock":      goods.Stock,
		"brand":      goods.Brand,
		"remark":     goods.Remark,
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("mgo insert result:", *insertOneResult)

	// query
	findOneResult := map[string]interface{}{}
	err = collection.FindOne(context.Background(), bson.M{
		"goods_id": goods.GoodsID,
	}).Decode(&findOneResult)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("mgo query result:", findOneResult)

	// update
	updateResult, err := collection.UpdateOne(context.Background(), bson.D{{"goods_id", goods.GoodsID}}, bson.D{
		{"$set", bson.D{{"goods_name", "青龙偃月刀"}, {"brand", "三国"}}},
	})
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("mgo update result:", *updateResult)

	findOneResult = map[string]interface{}{}
	err = collection.FindOne(context.Background(), bson.M{"goods_id": goods.GoodsID}).Decode(&findOneResult)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("mgo query after update result:", findOneResult)

	// delete
	deleteResult, err := collection.DeleteOne(context.Background(), bson.D{{"goods_id", goods.GoodsID}})
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("mgo delete result:", *deleteResult)

	findOneResult = map[string]interface{}{}
	err = collection.FindOne(context.Background(), bson.M{"goods_id": goods.GoodsID}).Decode(&findOneResult)
	if err != nil {
		logger.Error(err)
		return err
	}
	logger.Info("mgo query after delete result:", findOneResult)

	return nil
}

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

	// MongoDB
	s.AddGoodsIntoMgo(ctx, goods)

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
