package v1

// 不会重复的错误码

const (
	// 少数几个特殊的错误码
	ok = iota
	systemError
	invalidParams
)

const (
	// 依次递增的错误码
	dataMarshalError = 10000 + iota
	appNumLimited
)
