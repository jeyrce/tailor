package proxy

import (
	"sync"
	"time"

	"github.com/grafana/loki/pkg/logproto"
	"github.com/grafana/loki/pkg/logql"
	"gopkg.in/alecthomas/kingpin.v2"

	v1 "woqutech.com/tailor/api/v1"
	"woqutech.com/tailor/pkg/log"
)

var (
	once           sync.Once
	GlobalRegistry *Registry

	promMaxLokiNum = kingpin.Flag("prom.max-loki-num", "允许推送的Loki最大数量, 0表示不限制").Default("3").Int()
)

const (
	DefaultTTL    = time.Second * 60 // 默认单次续约的有效时间
	FilenameLabel = "filename"       // 日志流文件名标签
)

// 注册表
type Registry struct {
	Apps map[string]*Application // 应用实例, key作为实例身份标识
	lock sync.RWMutex            // 读写锁, 防止并发问题
}

// 单例模式维护一个全局注册表
func NewRegistry() *Registry {
	once.Do(func() {
		registry := Registry{
			Apps: make(map[string]*Application),
			lock: sync.RWMutex{},
		}
		GlobalRegistry = &registry
	})
	return GlobalRegistry
}

// 接收注册请求维护到注册表(覆盖)
func (r *Registry) Register(a AppCore) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	var app Application
	if *promMaxLokiNum > 0 {
		// 添加后数量会超限
		if _, in := r.Apps[a.URL]; !in && len(r.Apps) >= *promMaxLokiNum {
			return new(v1.AppNumLimited)
		}
	}
	app = Application{
		AppCore: a,
		AppMeta: AppMeta{RegAt: time.Now().Unix(), Failed: 0},
		lock:    sync.RWMutex{},
	}
	r.Apps[app.Name()] = &app
	return nil
}

// 按名称查询应用, 其中 * 查询所有
func (r *Registry) Fetch(name string) []*Application {
	r.lock.Lock()
	defer r.lock.Unlock()
	var apps = make([]*Application, 0)
	for _, app := range r.Apps {
		if name == app.Name() {
			apps = append(apps, app)
			break
		}
		if name == "*" {
			apps = append(apps, app)
		}
	}
	return apps
}

// 将服务从注册表移除,以免在整体上拖累整个转发进度
func (r *Registry) Cancel(name string) {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.Apps, name)
}

// 数据分发给各个loki对象
func (r *Registry) Dispatch(streams logproto.PushRequest) {
	r.lock.Lock()
	defer r.lock.Unlock()
	// 如果app状态是ok并且当前日志路径在app的监控对象中, 则推送对应的流数据
	wg := sync.WaitGroup{}
	var count = 0
	for _, app := range r.Apps {
		appS := make([]logproto.Stream, 0)
		for _, stream := range streams.Streams {
			labels, err := logql.ParseLabels(stream.Labels)
			if err != nil {
				log.Logger.Warnf("[跳过]解析label失败: %s", err)
				continue
			}
			if app.Match(labels.Get(FilenameLabel)) {
				appS = append(appS, stream)
			}
		}
		if len(appS) > 0 {
			wg.Add(1)
			// 并发推送这些日志
			go func(app *Application, ss []logproto.Stream) {
				err := app.Push(ss)
				if err != nil {
					log.Logger.Errorf("[%s]推送失败: %s", app.URL, err)
				}
				wg.Done()
			}(app, appS)
			count++
		}
	}
	wg.Wait()
	log.Logger.Infof("共计[%d]条记录转发", count)
}
