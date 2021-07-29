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
{{.program}}, version {{.version}} (branch: {{.branch}}, revision: {{.commitID}})
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
		"buildUser": BuildUser,
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
