package index

import (
	"github.com/gin-gonic/gin"

	v1 "woqutech.com/tailor/api/v1"
	"woqutech.com/tailor/pkg/version"
)

// @Summary 查询tailor应用是否正常
// @Description 响应2xx则代表当前应用正常
// @Produce application/json
// @Success 200 {object} v1.Response{} "成功响应"
// @Failure 400 {object} v1.Response{} "失败响应"
// @Failure 500 {object} v1.Response{} "失败响应"
// @Router /index/ready [GET]
// @Tags Index
func handleServerReady(c *gin.Context) {
	v1.ApiSucceed(c, nil)
}

// @Summary 查询该软件版本信息
// @Description 等同于 --version 方式, 但是通过api返回
// @Produce application/json
// @Success 200 {object} v1.Response{data=version.Struct{}} "成功响应"
// @Failure 400 {object} v1.Response{} "失败响应"
// @Failure 500 {object} v1.Response{} "失败响应"
// @Router /index/version [GET]
// @Tags Index
func handleVersion(c *gin.Context) {
	v1.ApiSucceed(c, version.Context("tailor"))
}
