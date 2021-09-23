package client

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	UserAgent   = "tailor"
	JSONHeader  = "application/json"
	ProtoHeader = "application/x-protobuf"
)

var (
	promPushTimeout = kingpin.Flag("prom.push-timeout", "推送Loki的超时设置,单位:s").Default("3").Int()
)

// 将日志流推送给loki
func SendLog(url string, buf []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*promPushTimeout))
	defer cancel()
	req, err := http.NewRequest("POST", url, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", ProtoHeader)
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("X-Scope-OrgID", "")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	all, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode/100 != 2 {
		return errors.New(string(all))
	}
	return nil
}
