package task

import (
	"time"

	"woqutech.com/tailor/pkg/log"
	"woqutech.com/tailor/pkg/proxy"
)

const (
	threshold = 10 // 允许连接失败的次数, 超过将剔除对应的注册信息
)

// 定时维护注册表
type RegistryLoop struct {
	registry       *proxy.Registry
	healthInterval time.Duration // 探活周期
	buildInternal  time.Duration // 配置周期
	stop           chan struct{} // 停止信号
}

// 停止周期任务
func (r *RegistryLoop) Stop() {
	r.stop <- struct{}{}
}

// - 定时清理无效loki实例
// - 定时重写promtail的配置
func (r *RegistryLoop) Run() {
	health := time.NewTicker(r.healthInterval)
	defer health.Stop()
	build := time.NewTicker(r.buildInternal)
	defer build.Stop()
	for {
		select {
		case <-r.stop:
			return
		case point := <-health.C:
			apps := proxy.GlobalRegistry.Fetch("*")
			pending := make([]string, 0)
			for _, app := range apps {
				if !app.Ready() {
					app.Failed++
					if app.Failed > threshold {
						log.Logger.Infof("[%s]连接失败,即将删除", app.Name())
						pending = append(pending, app.Name())
					}
				}
			}
			for _, name := range pending {
				proxy.GlobalRegistry.Cancel(name)
			}
			log.Logger.Infow("一轮探活结束",
				"total", len(apps),
				"deleted", len(pending),
				"duration", time.Since(point).Milliseconds(),
			)
		case <-build.C:
			log.Logger.Infow("重构监听配置")
			if err := proxy.GlobalRegistry.Build(); err != nil {
				log.Logger.Errorf("构建失败: %v", err)
			}
		}
	}
}
