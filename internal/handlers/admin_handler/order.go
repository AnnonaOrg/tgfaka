package admin_handler

import (
	"github.com/gin-gonic/gin"
	"gopay/internal/services"
	"gopay/internal/utils/functions"
	"gopay/internal/utils/restful"
)

func ReleaseOrders(c *gin.Context) {
	var requestData struct {
		IDsString string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		restful.ParamErr(c, "参数错误")
		return
	}
	ids, err := functions.ParseIDsString(requestData.IDsString)
	if err != nil {
		restful.ParamErr(c, "id格式错误")
		return
	}

	err = services.ReleaseOrders(ids)
	if err != nil {
		restful.ParamErr(c, "释放失败")
		return
	}

	restful.Ok(c, "释放成功")
}
