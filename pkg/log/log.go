/**
全局日志配置,使用方式

(1) 在程序入口处初始化:
```
import "woqutech.com/tailor/pkg/log"

func init(){
	// 需要注意Parse需要发生在Init之前
	kingpin.Parse()
	log.Init()
}

func main() {
	log.Logger.Info("Jeyrce.Lu")
}
```

(2) 在其他文件使用时,无需再次初始化,再次初始化也可以因为设置了单例模式
```
package xx

import "woqutech.com/tailor/pkg/log"

func DoSomething() {
	log.Logger.Info("服务启动成功")
}
```
*/
package log

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"path"
	"strings"
	"sync"
)

const (
	jsonStyle    = "json"
	consoleStyle = "console"
	info         = "tailor.log"
	err          = "tailor.error.log"
)

var (
	logFormat = kingpin.Flag("log.format", "日志格式,支持: "+strings.Join([]string{jsonStyle, consoleStyle}, "|")).Short('l').Default(consoleStyle).Enum(jsonStyle, consoleStyle)
	logDir    = kingpin.Flag("log.dir", "日志存储路径").Default("/var/log/").String()
	// logger 单例模式
	once   sync.Once
	Logger *zap.SugaredLogger
	config = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
)

func Init() {
	once.Do(initLogger)
}

func initLogger() {
	var encoder zapcore.Encoder
	switch *logFormat {
	case jsonStyle:
		encoder = zapcore.NewJSONEncoder(config)
	case consoleStyle:
		encoder = zapcore.NewConsoleEncoder(config)
	default:
		encoder = zapcore.NewJSONEncoder(config)
	}
	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(newRotateWriter(info)), zap.DebugLevel),
		zapcore.NewCore(encoder, zapcore.AddSync(newRotateWriter(err)), zap.ErrorLevel),
	)
	logger := zap.New(core, zap.AddCaller())
	defer func() { _ = logger.Sync() }()
	Logger = logger.Sugar()
}

// 借助lumberjack进行日志轮转
func newRotateWriter(filename string) io.Writer {
	return &lumberjack.Logger{
		Filename:   path.Join(*logDir, filename), // 日志文件
		MaxSize:    20 * 1 << 20,                 // 保存的文件最大大小
		MaxAge:     7,                            // 保存天数
		MaxBackups: 20,                           // 保存日志数量
		LocalTime:  true,                         // 使用系统本地时间
		Compress:   false,                        // 开启gz压缩
	}
}
