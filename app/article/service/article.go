package service

import (
	"context"
	"errors"
	"strings"

	"github.com/harveywangdao/ants/app/article/model"
	"github.com/harveywangdao/ants/logger"
	articlepb "github.com/harveywangdao/ants/rpc/article"
	"github.com/harveywangdao/ants/util"
)

func (s *Service) AddArticle(ctx context.Context, req *articlepb.AddArticleRequest) (*articlepb.AddArticleResponse, error) {
	if req.UserID == "" || req.ArticleInfo == nil || req.ArticleInfo.Content == "" {
		return nil, errors.New("param lost")
	}

	articleID := util.GetUUID()
	article := &model.ArticleModel{
		ArticleID: articleID,
		UserID:    req.UserID,
		Title:     req.ArticleInfo.Title,
		Content:   req.ArticleInfo.Content,
		Tags:      strings.Join(req.ArticleInfo.Tags, "&&"),
	}

	err := s.db.Create(article).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &articlepb.AddArticleResponse{
		ArticleID: articleID,
	}, nil
}

func (s *Service) GetArticle(ctx context.Context, req *articlepb.GetArticleRequest) (*articlepb.GetArticleResponse, error) {
	if req.ArticleID == "" {
		return nil, errors.New("lost articleID")
	}

	var article model.ArticleModel
	err := s.db.Where("article_id = ?", req.ArticleID).First(&article).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	logger.Info(article)

	resp := &articlepb.GetArticleResponse{
		ArticleInfo: &articlepb.ArticleInfo{
			ArticleID: article.ArticleID,
			Title:     article.Title,
			Content:   article.Content,
			Tags:      strings.Split(article.Tags, "&&"),
		},
	}

	return resp, nil
}

func (s *Service) GetArticleList(ctx context.Context, req *articlepb.GetArticleListRequest) (*articlepb.GetArticleListResponse, error) {
	if req.NumPerPage <= 0 {
		return nil, errors.New("param error")
	}

	var articles []*model.ArticleModel
	var err error
	if req.LastArticleID == 0 {
		if req.Page < 0 {
			return nil, errors.New("param error")
		}

		articles, err = model.GetArticlesByPage(s.db, req.Page, req.NumPerPage)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
	} else if req.LastArticleID > 0 {
		articles, err = model.GetArticlesByPage3(s.db, req.LastArticleID, req.NumPerPage)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
	}

	resp := &articlepb.GetArticleListResponse{}
	for _, article := range articles {
		articleInfo := &articlepb.ArticleInfo{
			ArticleID: article.ArticleID,
			Title:     article.Title,
			Content:   article.Content,
			Tags:      strings.Split(article.Tags, "&&"),
		}
		resp.ArticleInfos = append(resp.ArticleInfos, articleInfo)
	}

	return resp, nil
}

func (s *Service) ModifyArticleInfo(ctx context.Context, req *articlepb.ModifyArticleInfoRequest) (*articlepb.ModifyArticleInfoResponse, error) {
	if req.ArticleInfo == nil || req.ArticleInfo.ArticleID == "" {
		return nil, errors.New("lost articleID")
	}

	param := map[string]interface{}{
		"title":   req.ArticleInfo.Title,
		"content": req.ArticleInfo.Content,
		"tags":    strings.Join(req.ArticleInfo.Tags, "&&"),
	}

	err := s.db.Model(model.ArticleModel{}).Where("article_id = ?", req.ArticleInfo.ArticleID).Updates(param).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &articlepb.ModifyArticleInfoResponse{
		CodeMsg: "modify success",
	}, nil
}

func (s *Service) DelArticle(ctx context.Context, req *articlepb.DelArticleRequest) (*articlepb.DelArticleResponse, error) {
	if req.ArticleID == "" {
		return nil, errors.New("lost articleID")
	}

	err := s.db.Where("article_id = ?", req.ArticleID).Delete(model.ArticleModel{}).Error
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return &articlepb.DelArticleResponse{
		CodeMsg: "delete success",
	}, nil
}
