package proxy

import (
	"crypto/md5"
	"encoding/hex"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/bmatcuk/doublestar"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/grafana/loki/pkg/logproto"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"

	"woqutech.com/tailor/pkg/client"
	"woqutech.com/tailor/pkg/log"
)

var (
	// promtail 设置的10s超时,因此此处超时 = 重试次数 x 每次超时, 这个总时间不宜超过10s
	promMaxRetries = kingpin.Flag("prom.max-retries", "推送Loki的最大尝试次数").Default("3").Int()
	promConfigPath = kingpin.Flag("prom.config-path", "Promtail配置文件路径").Default("/etc/promtail/config.yml").String()
	promTargetDir  = kingpin.Flag("prom.target-dir", "目标日志文件file_sd_config对应目录").Default("/etc/promtail/target/").String()
	promBaseUrl    = kingpin.Flag("prom.base-url", "promtail的api基础地址").Default("http://127.0.0.1:15004/promtail").String()
)

// 需要传递的核心参数
type AppCore struct {
	URL      string            `json:"url"`      // 目标loki的推送地址
	CheckURL string            `json:"checkUrl"` // 主动探活的api
	Paths    []string          `json:"paths"`    // 该loki对象需要监听的日志文件
	Labels   map[string]string `json:"labels"`   // 该目标希望携带的标签对
}

// 应用的元信息
type AppMeta struct {
	Hash   string `json:"-" yaml:"-"` // 上次构建完成后文件hash
	RegAt  int64  `json:"lastTime"`   // 最后注册时间
	Failed int64  `json:"failed"`     // 连接失败次数
}

type Application struct {
	AppCore
	AppMeta
	lock sync.RWMutex `swaggerignore:"true"`
}

func NewAPP(a AppCore) *Application {
	return &Application{
		AppCore: a,
		AppMeta: AppMeta{RegAt: time.Now().Unix(), Failed: 0},
		lock:    sync.RWMutex{},
	}
}

// app的名称: 0.0.0.0:8080 或者 [::1]:443 格式
func (app *Application) Name() string {
	uri, err := url.Parse(app.URL)
	if err != nil {
		return "0.0.0.0"
	}
	return uri.Host
}

// 将数据推送给目标实例
func (app *Application) Push(ss []logproto.Stream) error {
	// 此处仅实现push客户端方法, 至于是否需要push由上层决定
	app.lock.Lock()
	defer app.lock.Unlock()
	// 为该目标loki添加他的标签对

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

// 构建自身服务发现文件
func (app *Application) Build() error {
	// fixme: 当前传过来的label除了targetIP其他都不起作用
	app.lock.Lock()
	defer app.lock.Unlock()
	var targets = make([]Target, 0, len(app.Paths))
	for _, p := range app.Paths {
		targets = append(targets, Target{
			Targets: []string{"localhost"},
			Labels: Labels{
				Path:     p,
				TargetIP: app.Labels[IPLabel],
			},
		})
	}
	sort.SliceStable(targets, func(i, j int) bool { return targets[i].Labels.Path < targets[j].Labels.Path })
	marshal, err := yaml.Marshal(targets)
	if err != nil {
		return err
	}
	m := md5.New()
	m.Write(marshal)
	fingerprint := hex.EncodeToString(m.Sum(nil))
	if app.Hash != fingerprint {
		if err := os.WriteFile(path.Join(*promTargetDir, app.Name()+".yml"), marshal, 0666); err != nil {
			return err
		}
		app.Hash = fingerprint
	}
	return nil
}

// 固定日志监听文件的格式
type Labels struct {
	Path     string `json:"__path__" yaml:"__path__"`
	TargetIP string `json:"targetIP" yaml:"targetIP"`
}

// 创建每个日志监听文件格式
type Target struct {
	Targets []string `json:"targets" yaml:"targets"`
	Labels  Labels   `json:"labels" yaml:"labels"`
}
