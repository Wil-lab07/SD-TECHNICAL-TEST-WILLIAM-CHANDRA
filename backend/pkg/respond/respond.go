package respond

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BaseResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
	Meta    interface{}  `json:"meta,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PaginationMeta struct {
	CurrentPage  int `json:"current_page"`
	Limit        int `json:"limit"`
	TotalRecords int `json:"total_records"`
	TotalPages   int `json:"total_pages"`
}

func Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, BaseResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, BaseResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func PaginatedSuccess(c *gin.Context, message string, data interface{}, meta PaginationMeta) {
	c.JSON(http.StatusOK, BaseResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, BaseResponse{
		Success: false,
		Message: message,
		Error:   &ErrorDetail{Code: "BAD_REQUEST", Message: message},
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, BaseResponse{
		Success: false,
		Message: message,
		Error:   &ErrorDetail{Code: "UNAUTHORIZED", Message: message},
	})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, BaseResponse{
		Success: false,
		Message: message,
		Error:   &ErrorDetail{Code: "FORBIDDEN", Message: message},
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, BaseResponse{
		Success: false,
		Message: message,
		Error:   &ErrorDetail{Code: "NOT_FOUND", Message: message},
	})
}

func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, BaseResponse{
		Success: false,
		Message: message,
		Error:   &ErrorDetail{Code: "CONFLICT", Message: message},
	})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, BaseResponse{
		Success: false,
		Message: message,
		Error:   &ErrorDetail{Code: "INTERNAL_SERVER_ERROR", Message: message},
	})
}

func CustomError(c *gin.Context, statusCode int, code string, message string) {
	c.JSON(statusCode, BaseResponse{
		Success: false,
		Message: message,
		Error:   &ErrorDetail{Code: code, Message: message},
	})
}
