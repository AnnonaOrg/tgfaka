package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/umfaka/tgfaka/internal/exts/config"
	"github.com/umfaka/tgfaka/internal/services"
	"github.com/umfaka/tgfaka/internal/utils/functions"
	"github.com/umfaka/tgfaka/internal/utils/restful"
)

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		//校验cookie是否存在
		cookie, err := c.Request.Cookie("admin_token")
		if err != nil {
			restful.UnLoginErr(c)
			c.Abort()
			return
		}
		token := cookie.Value

		//校验token是否成功解码
		claims, err := functions.DecodeToken(token, config.SiteSecret)
		if err != nil {
			restful.UnLoginErr(c)
			c.Abort()
			return
		}

		//校验一致性
		if token != services.GetAdminLoginTokenSession() {
			restful.UnLoginErr(c)
			c.Abort()
			return
		}

		role, _ := claims["role"].(string)
		if role != "admin" {
			restful.UnLoginErr(c)
			c.Abort()
			return
		}

		c.Next()
	}
}
