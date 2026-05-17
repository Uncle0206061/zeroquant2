// Package response 提供统一的 API 响应格式
// 响应格式：{ "code": 0, "message": "success", "data": {...} }
package response

import (
	"github.com/gin-gonic/gin"
)

// Code 响应码定义
const (
	CodeSuccess       = 0    // 成功
	CodeInvalidParam = 40001 // 参数错误
	CodeUnauthorized = 40101 // 未认证
	CodeForbidden   = 40102 // 无权限
	CodeNotFound   = 40401 // 资源不存在
	CodeServerErr  = 50001 // 服务器内部错误
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 响应码：0=成功，非0=失败
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessMsg 带消息的成功响应（data 为空）
func SuccessMsg(c *gin.Context, message string) {
	c.JSON(200, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    nil,
	})
}

// Fail 失败响应
func Fail(c *gin.Context, code int, message string) {
	c.JSON(200, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// FailWithData 带数据的失败响应
func FailWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(200, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// InvalidParam 参数错误
func InvalidParam(c *gin.Context, message string) {
	Fail(c, CodeInvalidParam, message)
}

// Unauthorized 未认证
func Unauthorized(c *gin.Context, message string) {
	Fail(c, CodeUnauthorized, message)
}

// Forbidden 无权限
func Forbidden(c *gin.Context, message string) {
	Fail(c, CodeForbidden, message)
}

// NotFound 资源不存在
func NotFound(c *gin.Context, message string) {
	Fail(c, CodeNotFound, message)
}

// ServerError 服务器内部错误
func ServerError(c *gin.Context, message string) {
	Fail(c, CodeServerErr, message)
}