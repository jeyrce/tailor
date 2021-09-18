package v1

// 服务下线app身份参数
type AppIdentifier struct {
	Name string `json:"name" example:"http://127.0.0.1:20015/loki/loki/api/v1/push"`
}
