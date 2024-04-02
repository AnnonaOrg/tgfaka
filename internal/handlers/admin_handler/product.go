package admin_handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gopay/internal/models"
	"gopay/internal/services"
	"gopay/internal/utils/functions"
	"gopay/internal/utils/restful"
)

func CreateProduct(c *gin.Context) {
	var requestData struct {
		Name        string          `json:"name" binding:"required"`
		Description string          `json:"description"`
		Currency    string          `json:"currency" binding:"required"`
		Price       decimal.Decimal `json:"price" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		restful.ParamErr(c, "参数错误")
		return
	}
	product := models.NewProduct(requestData.Name, requestData.Description, requestData.Currency, requestData.Price)

	err := services.CreateProduct(product)
	if err != nil {
		restful.ParamErr(c, "创建失败")
		return
	}

	restful.Ok(c, "创建成功")
}

func EditProduct(c *gin.Context) {
	var requestData struct {
		ID          *uuid.UUID       `json:"id" binding:"required"`
		Status      *uint            `json:"status" `
		Priority    *int64           `json:"priority" `
		Name        *string          `json:"name"`
		Description *string          `json:"description"`
		Currency    *string          `json:"currency"`
		Price       *decimal.Decimal `json:"price"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		restful.ParamErr(c, "参数错误")
		return
	}

	updateMap := functions.StructToMap(requestData, functions.StructToMapExcludeMode, "id")
	err := services.UpdateProduct(*requestData.ID, updateMap)
	if err != nil {
		restful.ParamErr(c, "编辑失败")
		return
	}

	restful.Ok(c, "编辑成功")
}

//func DeleteProducts(c *gin.Context) {
//	var requestData struct {
//		IDsString string `json:"ids"`
//	}
//	if err := c.ShouldBindJSON(&requestData); err != nil {
//		restful.ParamErr(c, "参数错误")
//		return
//	}
//
//	ids, err := functions.ParseIDsString(requestData.IDsString)
//	if err != nil {
//		restful.ParamErr(c, "id格式错误")
//		return
//	}
//
//	err = services.DeleteProducts(ids)
//	if err != nil {
//		restful.ParamErr(c, "删除失败")
//		return
//	}
//	restful.Ok(c, "删除成功")
//}
