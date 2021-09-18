package task

import (
	"time"

	"gopkg.in/alecthomas/kingpin.v2"

	"woqutech.com/tailor/pkg/proxy"
)

type Task interface {
	Run()
	Stop()
}

var (
	promRegisterTTL   = kingpin.Flag("prom.register-ttl", "自动检测周期(单位s)").Default("60").Int()
	promBuildInterval = kingpin.Flag("prom.build-interval", "配置构建周期(单位s)").Default("60").Int()
)

// 全局需要开启的任务
var tasks []Task

// 开启所有定时任务
func StartAll() {
	tasks = append(tasks,
		&RegistryLoop{
			registry:       proxy.GlobalRegistry,
			healthInterval: time.Duration(*promRegisterTTL) * time.Second,
			buildInternal:  time.Duration(*promBuildInterval) * time.Second,
			stop:           make(chan struct{}, 1),
		},
	)
	for _, t := range tasks {
		go func(task Task) { task.Run() }(t)
	}
}

// 关闭所有定时任务
func StopAll() {
	for _, t := range tasks {
		t.Stop()
	}
}
