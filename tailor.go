package main

import (
	"path"

	"github.com/gin-gonic/gin"
	"gopkg.in/alecthomas/kingpin.v2"

	"woqutech.com/tailor/pkg/proxy"
	"woqutech.com/tailor/task"

	registry "woqutech.com/tailor/api"
	"woqutech.com/tailor/docs"
	"woqutech.com/tailor/middle"
	log "woqutech.com/tailor/pkg/log"
	"woqutech.com/tailor/pkg/version"
)

const (
	program = "tailor"
)

var (
	listenAddr  = kingpin.Flag("web.listen-address", "服务监听配置").Default(":15100").String()
	routePrefix = kingpin.Flag("web.route-prefix", "统一路由前缀").Default(program).String()
)

func init() {
	// note: 许多模块都需要使用flag, 此处应当率先初始化
	kingpin.HelpFlag.Short('h')
	kingpin.Version(version.Version(program))
	kingpin.Parse()
	// 初始化全局logger
	log.Init()
	// 初始化全局注册表
	proxy.NewRegistry()
	// 配置gin的一些属性
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	// swagger 的一些配置
	docs.SwaggerInfo.Title = program
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/tailor/api/v1"
	docs.SwaggerInfo.Description = "读写一体机上promtail配置"
	docs.SwaggerInfo.Host = "10.10.168.77:15100"
	docs.SwaggerInfo.Schemes = []string{"http"}
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Logger.Errorf("服务出现重启: %v", err)
		}
	}()
	task.StartAll()
	defer task.StopAll()
	router := middle.Registry(gin.New())
	registry.V1(router.Group(path.Join([]string{*routePrefix, "/api/v1"}...)))
	registry.V2(router.Group(path.Join([]string{*routePrefix, "/api/v2"}...)))
	log.Logger.Infof("启动服务: %s", *listenAddr)
	if err := router.Run(*listenAddr); err != nil {
		log.Logger.Panicf("服务启动出现错误: %v", err)
	}
}
