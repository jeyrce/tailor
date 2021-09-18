package v1

import (
	"fmt"
)

// 系统错误
type SystemError E

func (s SystemError) RawError() E {
	return E(s)
}

func (s SystemError) Error() string {
	return "系统出现错误"
}

func (s SystemError) Code() int32 {
	return systemError
}

// 参数不合法
type InvalidParams E

func (i InvalidParams) Error() string {
	return fmt.Sprintf("请求参数不合法: %s", i.RawError())
}

func (i InvalidParams) Code() int32 {
	return invalidParams
}

func (i InvalidParams) RawError() E {
	return E(i)
}

// 数据转换错误
type DataMarshalError E

func (d DataMarshalError) RawError() E {
	return E(d)
}

func (d DataMarshalError) Error() string {
	return "日志数据转换失败"
}

func (d DataMarshalError) Code() int32 {
	return dataMarshalError
}

// loki数量超过限制
type AppNumLimited E

func (a AppNumLimited) Error() string {
	return "应用注册数量超过限制"
}

func (a AppNumLimited) Code() int32 {
	return appNumLimited
}

func (a AppNumLimited) RawError() E {
	return E(a)
}
