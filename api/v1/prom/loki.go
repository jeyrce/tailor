package prom

import (
	"compress/gzip"
	"fmt"
	"io"
	"math"

	v1 "woqutech.com/tailor/api/v1"

	"github.com/cortexproject/cortex/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/grafana/loki/pkg/loghttp"
	"github.com/grafana/loki/pkg/logproto"
	loki_util "github.com/grafana/loki/pkg/util"
	"github.com/grafana/loki/pkg/util/unmarshal"
	unmarshal2 "github.com/grafana/loki/pkg/util/unmarshal/legacy"

	"woqutech.com/tailor/pkg/log"
	"woqutech.com/tailor/pkg/proxy"
)

// @Summary 转发到目标loki
// @Description 该接口供promtail的webhook使用
// @Produce application/x-protobuf
// @Param body body logproto.PushRequest true "日志流格式"
// @Success 200 {object} v1.Response{} 成功响应
// @Failure 400 {object} v1.Response{} "失败响应"
// @Failure 500 {object} v1.Response{} "失败响应"
// @Router /prom/push [POST]
// @Tags Prom
func handleLokiPush(c *gin.Context) {
	// 截获protobuf, 转换为json格式重组后发送给loki
	var (
		req             logproto.PushRequest
		contentType     = c.ContentType()
		body            io.Reader
		bodySize        = loki_util.NewSizeReader(c.Request.Body)
		contentEncoding = c.Request.Header.Get("Content-Encoding")
	)
	// bodySize should always reflect the compressed size of the request body
	switch contentEncoding {
	case "":
		body = bodySize
	case "snappy":
		body = bodySize
	case "gzip":
		gzipReader, err := gzip.NewReader(bodySize)
		if err != nil {
			log.Logger.Error(err)
			v1.ApiFailed(c, v1.DataMarshalError{err})
			return
		}
		defer gzipReader.Close()
		body = gzipReader
	default:
		err := fmt.Errorf("Content-Encoding %q not supported", contentEncoding)
		log.Logger.Error(err)
		v1.ApiFailed(c, v1.DataMarshalError{err})
		return
	}
	// Content-Type
	switch contentType {
	case "application/json":
		var err error
		// todo once https://github.com/weaveworks/common/commit/73225442af7da93ec8f6a6e2f7c8aafaee3f8840 is in Loki.
		// We can try to pass the body as bytes.buffer instead to avoid reading into another buffer.
		if loghttp.GetVersion(c.Request.URL.String()) == loghttp.VersionV1 {
			err = unmarshal.DecodePushRequest(body, &req)
		} else {
			err = unmarshal2.DecodePushRequest(body, &req)
		}
		if err != nil {
			log.Logger.Error(err)
			v1.ApiFailed(c, v1.DataMarshalError{err})
			return
		}
	default:
		// `application/x-protobuf`: expect snappy compression.
		if err := util.ParseProtoReader(
			c.Request.Context(),
			body,
			int(c.Request.ContentLength),
			math.MaxInt32,
			&req,
			util.RawSnappy,
		); err != nil {
			log.Logger.Error(err)
			v1.ApiFailed(c, v1.DataMarshalError{err})
			return
		}
	}
	proxy.GlobalRegistry.Dispatch(req)
	v1.ApiSucceed(c, nil)
}

// @Summary 注册监听对象
// @Description 接收来自管理节点的注册请求,将Loki实例注册到全局注册表
// @Produce application/json
// @Param body body proxy.AppCore true "body参数"
// @Success 200 {object} v1.Response{data=proxy.Application} 成功响应
// @Failure 400 {object} v1.Response{} "失败响应"
// @Failure 500 {object} v1.Response{} "失败响应"
// @Router /prom/app [POST]
// @Tags Prom
func handleAppRegister(c *gin.Context) {
	core := proxy.AppCore{}
	err := c.ShouldBind(&core)
	if err != nil {
		v1.ApiFailed(c, v1.InvalidParams{err})
	}
	err = proxy.GlobalRegistry.Register(core)
	if err != nil {
		v1.ApiFailed(c, err.(v1.ApiError))
	}
	v1.ApiSucceed(c, nil)
}

// @Summary 已注册服务下线
// @Description 通过传递服务名称,主动将已注册服务下线
// @Produce application/json
// @Param name body v1.AppIdentifier{} true "服务名称(服务推送url)"
// @Success 200 {object} v1.Response{} "成功响应"
// @Failure 400 {object} v1.Response{} "失败响应"
// @Failure 500 {object} v1.Response{} "失败响应"
// @Router /prom/app [DELETE]
// @Tags Prom
func handleAppOffline(c *gin.Context) {
	id := v1.AppIdentifier{}
	if err := c.ShouldBind(&id); err != nil {
		v1.ApiFailed(c, v1.InvalidParams{err})
	}
	proxy.GlobalRegistry.Cancel(id.Name)
	v1.ApiSucceed(c, nil)
}

// @Summary 查询所有已注册应用
// @Description 查询查询所有已注册应用, 因为注册行为是幂等的, 因此不提供接口查询自身是否已经注册
// @Produce application/json
// @Success 200 {object} v1.Response{data=[]proxy.Application} "成功响应"
// @Failure 400 {object} v1.Response{} "失败响应"
// @Failure 500 {object} v1.Response{} "失败响应"
// @Router /prom/app [GET]
// @Tags Prom
func handleListApp(c *gin.Context) {
	apps := proxy.GlobalRegistry.Fetch("*")
	v1.ApiSucceed(c, apps)
}
