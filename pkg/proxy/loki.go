package proxy

import (
	"net/http"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/grafana/loki/pkg/logproto"
	"gopkg.in/alecthomas/kingpin.v2"

	"woqutech.com/tailor/pkg/client"
	"woqutech.com/tailor/pkg/log"
)

var (
	// promtail 设置的10s超时,因此此处超时 = 重试次数 x 每次超时, 这个总时间不宜超过10s
	promMaxRetries = kingpin.Flag("prom.max-retries", "推送Loki的最大尝试次数").Default("3").Int()
)

// 需要传递的核心参数
type AppCore struct {
	URL      string   `json:"url"`      // 目标loki的推送地址
	CheckURL string   `json:"checkUrl"` // 主动探活的api
	Paths    []string `json:"paths"`    // 该loki对象需要监听的日志文件
}

// 应用的元信息
type AppMeta struct {
	RegAt  int64 `json:"lastTime"` // 最后注册时间
	Failed int64 `json:"failed"`   // 连接失败次数
}

type Application struct {
	AppCore
	AppMeta
	lock sync.RWMutex `swaggerignore:"true"`
}

// app的名称
func (app *Application) Name() string {
	return app.URL
}

// 将数据推送给目标实例
func (app *Application) Push(ss []logproto.Stream) error {
	// 此处仅实现push客户端方法, 至于是否需要push由上层决定
	app.lock.Lock()
	defer app.lock.Unlock()
	times := *promMaxRetries
	for times > 0 {
		req := logproto.PushRequest{Streams: ss}
		buf, err := proto.Marshal(&req)
		if err != nil {
			return err
		}
		err = client.SendLog(app.URL, snappy.Encode(nil, buf))
		if err != nil {
			log.Logger.Errorf("重试推送: %v", err)
			times--
			app.Failed++
			time.Sleep(time.Millisecond * 100 * time.Duration(app.Failed))
			continue
		}
		app.Failed = 0
		break
	}
	return nil
}

// 主动检查loki的存活状态
// GET app.CheckURL 得到2xx状态码则认为目标服务正常
func (app *Application) Ready() bool {
	app.lock.Lock()
	defer app.lock.Unlock()
	resp, err := http.Get(app.CheckURL)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode/100 == 2
}

// 检查日志是否已经被监听
func (app *Application) hasTarget(path string) bool {
	app.lock.Lock()
	defer app.lock.Unlock()
	for _, old := range app.Paths {
		match, err := doublestar.PathMatch(old, path)
		if err != nil {
			continue
		}
		if match {
			return true
		}
	}
	return false
}

// 添加日志监听
func (app *Application) AddTarget(path string) error {
	app.lock.Lock()
	defer app.lock.Unlock()
	// 如果原本的应用已经涵盖了需要添加的路径, 则跳过
	// 例如: 原本存在 /tmp/*.log, 现在尝试添加 /tmp/test.log 则跳过
	if !app.hasTarget(path) {
		// 如果将要新增的路径包含一些原本的路径, 则将原本路径移除, 添加新路径
		// 例如: 原本存在 /tmp/1.log, /tmp/2.log, /var/log/message, 现在尝试添加 /tmp/*.log
		//       则应当将 /tmp/1.log和/tmp/2.log移除,写入/tmp/*.log
		app.Paths = append([]string{path}, app.notMatchedPaths(path)...)
	}
	return nil
}

// 查找目标路径不包含的路径
func (app *Application) notMatchedPaths(path string) []string {
	var newPaths = make([]string, 0)
	for _, target := range app.Paths {
		match, err := doublestar.PathMatch(path, target)
		if err != nil {
			newPaths = append(newPaths, target)
			continue
		}
		if !match {
			newPaths = append(newPaths, target)
		}
	}
	return newPaths
}

// 移除日志监听
func (app *Application) RemoveTarget(path string) error {
	app.lock.Lock()
	defer app.lock.Unlock()
	app.Paths = app.notMatchedPaths(path)
	return nil
}

// 清空日志监听
func (app *Application) ClearTarget() error {
	app.lock.Lock()
	defer app.lock.Unlock()
	app.Paths = make([]string, 0)
	return nil
}

// 判断日志文件是否match当前应用
// 例如: 告警配置了 /tmp/*.log, 当前日志流数据的 filename=/tmp/test.log 则匹配
func (app *Application) Match(filename string) bool {
	app.lock.Lock()
	defer app.lock.Unlock()
	for _, pattern := range app.Paths {
		match, err := doublestar.PathMatch(pattern, filename)
		if err != nil {
			continue
		}
		if match {
			return match
		}
	}
	return false
}
