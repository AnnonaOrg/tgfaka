package admin_handler

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gopay/internal/exts/db"
	"gopay/internal/models"
	"gopay/internal/services"
	"gopay/internal/utils/functions"
	"gopay/internal/utils/restful"
)

type PaginationRequest struct {
	Page    int `json:"page"`
	PerPage int `json:"perPage"`
}

func (pr PaginationRequest) ToPagination() services.Pagination {
	return services.Pagination{
		Limit: pr.PerPage,
		Page:  pr.Page,
	}
}

func FetchList[T models.MyModel](c *gin.Context) {
	var paginationRequest PaginationRequest
	if err := c.ShouldBindBodyWith(&paginationRequest, binding.JSON); err != nil {
		restful.ParamErr(c)
		return
	}

	var filterParams map[string]interface{}
	if err := c.ShouldBindBodyWith(&filterParams, binding.JSON); err != nil {
		restful.ParamErr(c, "参数错误2")
		return
	}
	pagination := paginationRequest.ToPagination()

	query := db.DB
	query = models.ApplyFilters(query, filterParams)

	var entity T
	query = query.Order(entity.DefaultOrder())

	err := services.Paginate[T](&pagination, query)
	if err != nil {
		restful.ParamErr(c, "分页失败")
		return
	}

	if err != nil {
		restful.ParamErr(c, "获取失败")
		return
	}

	for i, entity := range pagination.Items {
		pagination.Items[i] = functions.StructToMap(entity, functions.StructToMapExcludeMode)
	}

	respData := map[string]interface{}{
		"items": pagination.Items,
		"total": pagination.Total,
	}
	restful.Ok(c, respData)
}

func DeleteEntities[T models.MyModel](c *gin.Context) {
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

	if err = services.DeleteEntities[T](ids); err != nil {
		restful.ParamErr(c, "删除失败")
		return
	}

	restful.Ok(c, "删除成功")
}
func DeleteAllEntities[T models.MyModel](c *gin.Context) {

	if err := services.DeleteAllEntities[T](); err != nil {
		restful.ParamErr(c, "删除失败")
		return
	}

	restful.Ok(c, "删除成功")
}
