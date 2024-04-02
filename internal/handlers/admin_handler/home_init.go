package admin_handler

import (
	"html/template"

	adminTPL "github.com/umfaka/tgfaka/internal/admin_templates"
)

var (
	homeTemplate     *template.Template
	notfoundTemplate *template.Template
)

func init() {
	var err error
	// 管理首页模版
	homeTemplate, err = template.ParseFS(adminTPL.TPL, "index.html")
	if err != nil {
		panic(err)
	}

	// 404模版
	notfoundTemplate, err = template.ParseFS(adminTPL.TPL, "404.html")
	if err != nil {
		panic(err)
	}
}

func unescaped(x string) interface{} { return template.HTML(x) }
