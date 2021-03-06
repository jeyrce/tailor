# tailor

用于在一体机上 管理 `promtail` 及其配置文件的服务, 协助 promtail 收集日志传输给下游 Loki

- promtail 2.2.1

## (1)读写本地Promtail配置

`Promtail` 将自身收集到的日志信息push给下游 `Loki` 做进一步处理, 在此过程中Promtail通过自身的一份yml格式配置文件明确:

- 采集的日志文件目标
- 数据推送目标
- 采集的一些行为

而这些配置是需要根据业务端用户的操作进行变更的, `tailor` 核心功能就在于维护管理这份配置文件.

> 这一功能还可以使用在管理节点构建好配置后ssh传输到目标节点上供promtail使用,但是该方式不安全.

## (2)管理维护Loki对象

每个管理节点将至少运行一个 `Loki` 实例, 在实际生产中可能会运行多个Loki实例来组织业务, 此外会存在同一个物理节点需要被添加到多个管理节点中进行管控的需求, 因此对于主动 `push` 日志数据到Loki这样的工作模式来说,需要一个服务来维护管理推送对象.

- 当在管理节点开启一体机节点的日志采集任务,需要将对应Loki对象注册到推送列表
- 当管理节点无法连通或主动关闭采集,需要将对应Loki对象从列表剔除
- 分发: 不同管理节点推送不同采集任务的日志

Promtail 配置的官方解释如下:

```
# 警告：如果远程 Loki 服务器之一无法响应或响应
# 任何可重试的错误，这将影响发送日志到任何
# 其他配置的远程 Loki 服务器。 发送是在一个线程上完成的！
# 一般建议并行运行多个promtail客户端
# 如果你想发送到多个远程 Loki 实例。
```

> 对于我们来说开启多个promtail实例是不合适的, promtail 虽然支持推送给多个Loki实例,但是所有推送内容所有都是相同的,此项目可以接收promtail的推送, 根据不同管理节点的需要分别推送各自需要的日志

## (3)作为agent功能扩展

在后续迭代过程中, 该服务还可以充当类似于qdatamgr, 但是仅专注于监控采集的基础服务.

- 按需获取日志上下文
- 提供日志下载接口

