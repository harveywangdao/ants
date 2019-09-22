package service

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/harveywangdao/ants/logger"
	proto "github.com/harveywangdao/ants/rpc/goods"
)

type HandlerServiceFunc func([]byte) ([]byte, error)

type HttpServer struct {
	ServiceName string
	Port        string
}

func (h *HttpServer) StartHttpServer(httpService *HttpService) {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.POST("/ants/v1/"+h.ServiceName+"/:funcName", func(c *gin.Context) {
		funcName := c.Param("funcName")

		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			logger.Error(err)
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}
		logger.Debug("funcName:", funcName, "body:", string(body))

		//elem := reflect.ValueOf(&httpService).Elem()
		elem := reflect.ValueOf(httpService)

		myref := elem.Elem()
		typeOfType := myref.Type()
		for i := 0; i < myref.NumField(); i++ {
			field := myref.Field(i)
			logger.Debug(i, typeOfType.Field(i).Name, field.Type(), field.Interface())
		}

		methodExisted := false
		for i := 0; i < elem.NumMethod(); i++ {
			//logger.Info(elem.Method(i))
			//logger.Info(elem.Type().Method(i).Name)

			if funcName == elem.Type().Method(i).Name {
				methodExisted = true
				break
			}
		}

		if !methodExisted {
			c.JSON(http.StatusOK, gin.H{"error": "method not existed"})
			return
		}

		params := make([]reflect.Value, 1)
		params[0] = reflect.ValueOf(body)
		resp := elem.MethodByName(funcName).Call(params)

		if len(resp) != 2 {
			c.JSON(http.StatusOK, gin.H{"error": "method return param num error"})
			return
		}

		for i := 0; i < len(resp); i++ {
			logger.Info(resp[i].Interface())
		}

		errMsg, ok := resp[1].Interface().(error)
		if !ok {
			errMsg = errors.New("")
		}

		c.JSON(http.StatusOK, gin.H{
			"message": resp[0].Interface(),
			"error":   errMsg.Error(),
		})
	})

	if err := router.Run(":" + h.Port); err != nil {
		logger.Panicln(err)
	}
}

type HttpService struct {
	ServiceApp *Service
}

func (h *HttpService) AddGoods(reqData []byte) (interface{}, error) {
	req := &proto.AddGoodsRequest{}

	if err := json.Unmarshal(reqData, req); err != nil {
		logger.Error(err)
		return nil, err
	}

	return h.ServiceApp.AddGoods(context.Background(), req)
}

func (h *HttpService) GetGoods(reqData []byte) (interface{}, error) {
	req := &proto.GetGoodsRequest{}

	if err := json.Unmarshal(reqData, req); err != nil {
		logger.Error(err)
		return nil, err
	}

	return h.ServiceApp.GetGoods(context.Background(), req)
}

func (h *HttpService) GetGoodsListByCategory(reqData []byte) (interface{}, error) {
	req := &proto.GetGoodsListByCategoryRequest{}

	if err := json.Unmarshal(reqData, req); err != nil {
		logger.Error(err)
		return nil, err
	}

	return h.ServiceApp.GetGoodsListByCategory(context.Background(), req)
}

func (h *HttpService) ModifyGoodsInfo(reqData []byte) (interface{}, error) {
	req := &proto.ModifyGoodsInfoRequest{}

	if err := json.Unmarshal(reqData, req); err != nil {
		logger.Error(err)
		return nil, err
	}

	return h.ServiceApp.ModifyGoodsInfo(context.Background(), req)
}

func (h *HttpService) DelGoods(reqData []byte) (interface{}, error) {
	req := &proto.DelGoodsRequest{}

	if err := json.Unmarshal(reqData, req); err != nil {
		logger.Error(err)
		return nil, err
	}

	return h.ServiceApp.DelGoods(context.Background(), req)
}

func (h *HttpService) DeductStock(reqData []byte) (interface{}, error) {
	req := &proto.DeductStockRequest{}

	if err := json.Unmarshal(reqData, req); err != nil {
		logger.Error(err)
		return nil, err
	}

	return h.ServiceApp.DeductStock(context.Background(), req)
}
