package admin_handler

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/umfaka/tgfaka/internal/models"
	"github.com/umfaka/tgfaka/internal/services"
	"github.com/umfaka/tgfaka/internal/utils/functions"
	"github.com/umfaka/tgfaka/internal/utils/restful"
)

func CreateProductItems(c *gin.Context) {
	var requestData struct {
		ProductID uuid.UUID `json:"product_id"`
		Content   string    `json:"content"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		restful.ParamErr(c, "参数错误")
		return
	}

	var productItems []models.ProductItem
	lines := strings.Split(requestData.Content, "\n")
	for _, line := range lines {
		if !functions.IsWhitespace(line) {
			productItems = append(productItems, *models.NewProductItem(line, requestData.ProductID))
		}
	}

	err := services.CreateProductItems(productItems)
	if err != nil {
		restful.ParamErr(c, "创建失败")
		return
	}

	restful.Ok(c, "创建成功")
}

func DeleteProductItems(c *gin.Context) {
	var requestData struct {
		ProductID uuid.UUID `json:"product_id"`
		IDsString string    `json:"ids"`
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

	err = services.DeleteProductItems(ids)
	if err != nil {
		restful.ParamErr(c, "删除失败")
		return
	}

	// 更新库存
	services.UpdateProductInStockCount([]uuid.UUID{requestData.ProductID})

	restful.Ok(c, "删除成功")
}
