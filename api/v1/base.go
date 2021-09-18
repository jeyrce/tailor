package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 响应基本格式
type Response struct {
	Succeed bool        `json:"succeed" example:"true"` // 响应是否成功
	Code    int32       `json:"code" example:"0"`       // 响应业务码, 正常时为0
	Message string      `json:"message" example:"-"`    // 当出现问题时给出错误提示
	Data    interface{} `json:"data"`                   // 当正常响应时返回对应数据
}

// 成功的响应
func ApiSucceed(c *gin.Context, data interface{}) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, Response{
		Code:    ok,
		Succeed: true,
		Message: "",
		Data:    data,
	})
}

// 失败的响应
func ApiFailed(c *gin.Context, e ApiError) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.JSON(http.StatusOK, Response{
		Code:    e.Code(),
		Succeed: false,
		Message: e.Error(),
		Data:    nil,
	})
}

// 项目中自定义的api错误
type ApiError interface {
	Error() string // 自定义错误的message
	Code() int32   // 自定义错误码
	RawError() E   // 包含的原始错误
}

// 错误列表
type E []error
