package utils

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// ResponseCode 统一 API 响应码。
// 0 表示成功；400-599 映射为对应 HTTP 状态码；1001 起为业务错误，HTTP 状态码仍为 200。
// 生产环境建议根据业务领域拆分更详细的错误码表。
type ResponseCode int

const (
	CodeSuccess           ResponseCode = 0
	CodeBadRequest        ResponseCode = 400
	CodeUnauthorized      ResponseCode = 401
	CodeForbidden         ResponseCode = 403
	CodeNotFound          ResponseCode = 404
	CodeTooManyRequests   ResponseCode = 429
	CodeInternalError     ResponseCode = 500
	CodeBusinessError     ResponseCode = 1001
)

// Response 统一 API 响应结构
type Response struct {
	Code    ResponseCode `json:"code"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
}

// Success 返回成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithStatus 返回指定 HTTP 状态码的成功响应
func SuccessWithStatus(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// Error 返回错误响应
func Error(c *gin.Context, code ResponseCode, message string) {
	// HTTP 状态码优先使用业务错误码的数值部分，但 1001 这类业务码映射为 200
	httpStatus := http.StatusOK
	if code >= 400 && code < 600 {
		httpStatus = int(code)
	}
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

// BadRequest 400 错误
func BadRequest(c *gin.Context, message string) {
	Error(c, CodeBadRequest, message)
}

// Unauthorized 401 错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, CodeUnauthorized, message)
}

// NotFound 404 错误
func NotFound(c *gin.Context, message string) {
	Error(c, CodeNotFound, message)
}

// InternalError 500 错误
func InternalError(c *gin.Context, message string) {
	Error(c, CodeInternalError, message)
}
