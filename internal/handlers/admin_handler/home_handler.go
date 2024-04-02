package admin_handler

import (
	"html/template"

	"github.com/gin-gonic/gin"
)

// 后台主页
func HomeHandle(c *gin.Context) {
	title := "后台"
	homeTemplate = homeTemplate.Funcs(template.FuncMap{"unescaped": unescaped})
	homeTemplate.Execute(c.Writer, map[string]interface{}{
		"title": title,
		// "icp":   "",
		// "CoreAPIRouter": template.HTML(osenv.GetCoreApiUrl()),
	})
}
