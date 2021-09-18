package version

import (
	"bytes"
	"fmt"
	"html/template"
	"runtime"
	"strings"
)

var (
	goVersion    = fmt.Sprintf("%s/%s %s", runtime.GOOS, runtime.GOARCH, runtime.Version())
	BuildVersion string
	BuildDate    string
	BuildUser    string
	Branch       string
	CommitID     string
	// 输出的版本信息模板
	tmpl = `
{{.program}}, version {{.version}} (branch: {{.branch}}, revision: {{.revision}})
    build user:       {{.user}}
    build date:       {{.buildDate}}
    go version:       {{.goVersion}}
	`
)

// 当使用 --version 时展示
func Version(program string) string {
	m := map[string]string{
		"program":   program,
		"version":   BuildVersion,
		"revision":  CommitID,
		"branch":    Branch,
		"user":      BuildUser,
		"buildDate": BuildDate,
		"goVersion": goVersion,
	}
	t := template.Must(template.New(program).Parse(tmpl))

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, program, m); err != nil {
		panic(err)
	}
	return strings.TrimSpace(buf.String())
}

// version接口响应
type Struct struct {
	Program   string `json:"program" example:"tailor"`                // 服务名称
	Version   string `json:"version" example:"v0.1.0"`                // 软件版本
	Branch    string `json:"branch" example:"master"`                 // 构建代码分支
	Revision  string `json:"revision"  example:"h36dj82j78"`          // 代码CommitID
	BuildUser string `json:"buildUser" example:"Jeyrce.Lu"`           // 构建用户
	BuildDate string `json:"buildDate" example:"2021-08-08 01:02:03"` // 构建时间
	GoVersion string `json:"goVersion" example:"linux/amd64 1.16.2"`  // 构建时go版本
}

func Context(program string) Struct {
	return Struct{
		Program:   program,
		Version:   BuildVersion,
		Branch:    Branch,
		Revision:  CommitID,
		BuildUser: BuildUser,
		BuildDate: BuildDate,
		GoVersion: goVersion,
	}
}
