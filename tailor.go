package main

import (
	"path"

	"github.com/gin-gonic/gin"
	"gopkg.in/alecthomas/kingpin.v2"
	registry "woqutech.com/tailor/api"
	"woqutech.com/tailor/middle"
	log "woqutech.com/tailor/pkg/log"
	"woqutech.com/tailor/pkg/version"
)

const (
	program = "tailor"
)

var (
	listenAddr     = kingpin.Flag("web.listen-address", "服务监听配置").Default(":15100").String()
	routePrefix    = kingpin.Flag("web.route-prefix", "统一路由前缀").Default(program).String()
	promConfigPath = kingpin.Flag("prom.config-path", "Promtail配置文件路径").Default("/etc/promtail/config.yml").String()
	promMaxLokiNum = kingpin.Flag("prom.max-loki-num", "允许推送的Loki最大数量, 0表示不限制").Default("3").Int()
	// promtail 设置的10s超时,因此此处超时 = 重试次数 x 每次超时, 这个总时间不宜超过10s
	promMaxRetries  = kingpin.Flag("prom.max-retries", "推送Loki的最大尝试次数").Default("3").Int()
	promPushTimeout = kingpin.Flag("prom.push-timeout", "推送Loki的超时设置,单位:s").Default("3").Int()
	promTargetDir   = kingpin.Flag("prom.target-dir", "目标日志文件file_sd_config对应目录").Default("/etc/promtail/target/").String()
	promMaxTargets  = kingpin.Flag("prom.max-targets", "允许建立的最大任务数").Default("30").Int()
)

func init() {
	kingpin.HelpFlag.Short('h')
	kingpin.Version(version.Version(program))
	kingpin.Parse()
	log.Init()
}

func main() {
	router := middle.Registry(gin.New())
	gin.DisableConsoleColor()
	registry.V1(router.Group(path.Join([]string{*routePrefix, "/api/v1"}...)))
	registry.V2(router.Group(path.Join([]string{*routePrefix, "/api/v2"}...)))
	log.Logger.Infof("启动服务: %s", *listenAddr)
	if err := router.Run(*listenAddr); err != nil {
		log.Logger.Panicf("服务启动出现错误: %s", err)
	}
}
