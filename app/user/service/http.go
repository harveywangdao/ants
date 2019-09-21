package service

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/harveywangdao/ants/logger"
	userpb "github.com/harveywangdao/ants/rpc/user"
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
			c.JSON(http.StatusOK, gin.H{"message": err.Error()})
			return
		}
		logger.Debug("funcName:", funcName, "body:", string(body))

		elem := reflect.ValueOf(&httpService).Elem()
		params := make([]reflect.Value, 1)
		params[0] = reflect.ValueOf(body)
		resp := elem.MethodByName(funcName).Call(params)

		c.JSON(http.StatusOK, gin.H{
			"message": resp[0].Interface(),
			"error":   resp[1].Interface(),
		})
	})

	if err := router.Run(":" + h.Port); err != nil {
		logger.Panicln(err)
	}
}

type HttpService struct {
	ServiceApp *Service
}

func (h *HttpService) AddUser(reqData []byte) (interface{}, error) {
	req := &userpb.AddUserRequest{}

	if err := json.Unmarshal(reqData, req); err != nil {
		logger.Error(err)
		return nil, err
	}

	return h.ServiceApp.AddUser(context.Background(), req)
}

func (h *HttpService) GetUser(reqData []byte) (interface{}, error) {
	req := &userpb.GetUserRequest{}

	if err := json.Unmarshal(reqData, req); err != nil {
		logger.Error(err)
		return nil, err
	}

	return h.ServiceApp.GetUser(context.Background(), req)
}

func (h *HttpService) GetUserIdByPhoneNumber(reqData []byte) (interface{}, error) {
	req := &userpb.GetUserIdByPhoneNumberRequest{}

	if err := json.Unmarshal(reqData, req); err != nil {
		logger.Error(err)
		return nil, err
	}

	return h.ServiceApp.GetUserIdByPhoneNumber(context.Background(), req)
}

func (h *HttpService) GetUsersByName(reqData []byte) (interface{}, error) {
	req := &userpb.GetUsersByNameRequest{}

	if err := json.Unmarshal(reqData, req); err != nil {
		logger.Error(err)
		return nil, err
	}

	return h.ServiceApp.GetUsersByName(context.Background(), req)
}

func (h *HttpService) ModifyUserInfo(reqData []byte) (interface{}, error) {
	req := &userpb.ModifyUserInfoRequest{}

	if err := json.Unmarshal(reqData, req); err != nil {
		logger.Error(err)
		return nil, err
	}

	return h.ServiceApp.ModifyUserInfo(context.Background(), req)
}

func (h *HttpService) DelUser(reqData []byte) (interface{}, error) {
	req := &userpb.DelUserRequest{}

	if err := json.Unmarshal(reqData, req); err != nil {
		logger.Error(err)
		return nil, err
	}

	return h.ServiceApp.DelUser(context.Background(), req)
}
