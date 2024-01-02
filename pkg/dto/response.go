package dto

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func ResponseBadRequest(ctx *gin.Context, err error) {
	rsp := &Response{
		Code:    http.StatusBadRequest,
		Message: err.Error(),
	}
	ctx.JSON(http.StatusBadRequest, rsp)
}

func ResponseInternalServerError(ctx *gin.Context, err error) {
	rsp := &Response{
		Code:    http.StatusInternalServerError,
		Message: err.Error(),
	}
	ctx.JSON(http.StatusInternalServerError, rsp)
}

func ResponseSuccess(ctx *gin.Context, data interface{}) {
	if data == nil {
		ctx.Status(http.StatusOK)
	} else {
		ctx.JSON(http.StatusOK, data)
	}
}
