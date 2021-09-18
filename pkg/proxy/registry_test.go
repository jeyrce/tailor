package proxy

import (
	"testing"
)

func TestName(t *testing.T) {
	r := NewRegistry()
	t.Logf("初始化之后: %d, %v", len(r.Apps), r.Apps)
	AppCore := AppCore{
		URL:      "http://127.0.0.1/loki/api/v1/push",
		CheckURL: "http://127.0.0.1/loki/api/v1/ready",
		Paths:    nil,
	}
	if err := r.Register(AppCore); err != nil {
		t.Fatal(err)
	}
	t.Logf("注册app: %d, %v", len(r.Apps), r.Apps)
	r.Cancel("*")
	t.Logf("重新置空: %d, %v", len(r.Apps), r.Apps)
}
