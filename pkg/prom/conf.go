package prom

import (
	"net/http"
	"path"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	promConfigPath = kingpin.Flag("prom.config-path", "Promtail配置文件路径").Default("/etc/promtail/config.yml").String()
	promTargetDir  = kingpin.Flag("prom.target-dir", "目标日志文件file_sd_config对应目录").Default("/etc/promtail/target/").String()
	promBaseUrl    = kingpin.Flag("prom.base-url", "promtail的api基础地址").Default("http://127.0.0.1:15004/promtail").String()
)

// promtail 管理器
type Promtail struct {
	PathSet []string `json:"pathSet"` // 需要监听的日志路径
}

// 定时构建当前日志监听配置(覆盖原本)
/**
如何计算配置文件是否需要重新覆写?
(1) 将所有路径配置起来
*/
func (p *Promtail) BuildTarget() error {
	// 计算当前需要监听的文件
	// 写入配置
	return nil
}

// 检查是否就绪
func (p *Promtail) Ready() bool {
	api := path.Join(*promBaseUrl, "/ready")
	get, err := http.Get(api)
	if err != nil {
		return false
	}
	defer func() { _ = get.Body.Close() }()
	return get.StatusCode/100 == 2
}
