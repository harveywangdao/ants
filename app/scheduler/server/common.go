package server

import (
	"github.com/gin-gonic/gin"
)

type ErrResult struct {
	Msg     string `json:"msg"`
	ErrCode string `json:"errCode"`
}

func AbortWithErrMsg(c *gin.Context, code int, msg string) {
	c.JSON(code, &ErrResult{
		Msg: msg,
	})
	c.Abort()
}
