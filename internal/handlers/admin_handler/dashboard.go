package admin_handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/shopspring/decimal"
	"github.com/umfaka/tgfaka/internal/exts/config"
	"github.com/umfaka/tgfaka/internal/exts/db"
	"github.com/umfaka/tgfaka/internal/models"
	"github.com/umfaka/tgfaka/internal/services"
	"github.com/umfaka/tgfaka/internal/utils/functions"
	"github.com/umfaka/tgfaka/internal/utils/restful"
)

func TokenLogin(c *gin.Context) {
	token := c.Query("token")

	//检查是否一致
	if token != services.GetAdminLoginUrlSession() {
		restful.ParamErr(c, "invalid_token0")
		return
	}

	claims, err := functions.DecodeToken(token, config.SiteSecret)
	if err != nil {
		restful.ParamErr(c, "invalid_token1")
		return
	}

	loginUsername, ok := claims["loginUsername"].(string)
	if !ok {
		restful.ParamErr(c, "invalid_token2")
		return
	}
	if loginUsername != fmt.Sprintf("%d", config.GetSiteConfig().AdminTGID) {
		restful.ParamErr(c, "invalid_token3")
		return
	}

	//更新登录session
	adminToken := services.SetAdminLoginTokenSession(time.Hour * 12)

	//删除url session
	services.ClearAdminLoginUrlSession()

	//更新cookie
	c.SetCookie("admin_token", adminToken, 0, "/", "", true, true)

	c.Redirect(http.StatusFound, "/page/admin/index.html")
}

func Logout(c *gin.Context) {
	services.ClearAdminLoginTokenSession()
	restful.Ok(c, "退出成功")
}

func Info(c *gin.Context) {
	var respData = map[string]interface{}{
		"currencies":        config.GetCurrencies(),
		"crypto_currencies": config.GetCryptoCurrencies(),
		"networks":          config.GetAllNetworks(),
	}

	restful.Ok(c, respData)
}

func Dashboard(c *gin.Context) {
	todayStartTimestamp := (functions.TruncateToStartOfDay(time.Now())).Unix()
	todayEndTimestamp := (functions.TruncateToEndOfDay(time.Now())).Unix()
	yesterdayStartTimestamp := (functions.TruncateToStartOfDay(time.Now().AddDate(0, 0, -1))).Unix()
	yesterdayEndTimestamp := (functions.TruncateToEndOfDay(time.Now().AddDate(0, 0, -1))).Unix()

	todayIncome, err := services.GetOrderIncomeByTimestampRange(todayStartTimestamp, todayEndTimestamp)
	if err != nil {
		restful.ParamErr(c, err)
		return
	}
	yesterdayIncome, err := services.GetOrderIncomeByTimestampRange(yesterdayStartTimestamp, yesterdayEndTimestamp)
	if err != nil {
		restful.ParamErr(c, err)
		return
	}

	realTimeExchangeData := functions.StructToMap(config.ExchangeRateData, functions.StructToMapExcludeMode)
	var fixExchangeData map[string]interface{}
	err = json.Unmarshal([]byte(config.SiteConfig.FixedExchangeRate), &fixExchangeData)
	if err != nil {
		fixExchangeData = map[string]interface{}{"err": "err"}
	}

	var respData = map[string]interface{}{
		"today_income":            todayIncome,
		"yesterday_income":        yesterdayIncome,
		"real_time_exchange_data": realTimeExchangeData,
		"fix_exchange_data":       fixExchangeData,
		"enable_fix_exchange":     config.SiteConfig.EnableFixExchangeRate,
	}
	restful.Ok(c, respData)
}
func DashboardChart(c *gin.Context) {
	var filterParams map[string]interface{}
	if err := c.ShouldBindBodyWith(&filterParams, binding.JSON); err != nil {
		restful.ParamErr(c, "参数错误")
		return
	}

	var startTimestamp, endTimestamp int64
	if _, ok := filterParams["timestamp_range"]; !ok {

		startOfToday := functions.TruncateToStartOfDay(time.Now())

		startOfPeriod := startOfToday.AddDate(0, 0, -6)
		startTimestamp = startOfPeriod.Unix()

		endOfPeriod := startOfToday.AddDate(0, 0, 1).Add(-time.Second)
		endTimestamp = endOfPeriod.Unix()
	} else {
		_, err := fmt.Sscanf(filterParams["timestamp_range"].(string), "%d,%d", &startTimestamp, &endTimestamp)
		if err != nil {
			restful.ParamErr(c, "时间格式错误")
			return
		}
		startTimestamp = functions.TruncateToStartOfDay(time.Unix(startTimestamp, 0)).Unix()
		endTimestamp = functions.TruncateToEndOfDay(time.Unix(endTimestamp, 0)).Unix()
	}
	filterParams["timestamp_range"] = fmt.Sprintf("%d,%d", startTimestamp, endTimestamp)

	var orders []models.Order
	query := models.ApplyFilters(db.DB, filterParams).Where("status=1")
	if result := query.Find(&orders); result.Error != nil {
		restful.ParamErr(c, "获取订单错误")
		return
	}

	startTime := time.Unix(startTimestamp, 0)
	endTime := time.Unix(endTimestamp, 0)

	var data EChartDataStruct
	data.TextStyle.FontSize = 15
	data.YAxis = []YAxisStruct{{Min: "dataMin", Max: "dataMax", Type: "value"}, {Min: "dataMin", Max: "dataMax", Type: "value"}}
	data.Tooltip.AxisPointer.Type = "cross"
	data.Tooltip.Trigger = "axis"

	orderPriceSumSeries := Series{Data: []interface{}{}, Name: "订单收入", Type: "line", YAxisIndex: 1}
	var totalOrderPriceSum decimal.Decimal
	orderCountSeries := Series{Data: []interface{}{}, Name: "完成订单", Type: "line", YAxisIndex: 0}
	var totalOrderCount int
	for t := startTime; t.Before(endTime); t = t.AddDate(0, 0, 1) {
		dayStart := t
		dayEnd := t.AddDate(0, 0, 1).Add(-time.Second)
		startTimestampTemp := dayStart.Unix()
		endTimestampTemp := dayEnd.Unix()

		// 订单收入
		orderPriceSumTemp := decimal.Zero
		for _, order := range orders {
			if order.CreateTime > startTimestampTemp && order.CreateTime < endTimestampTemp {
				convertedPrice, err := config.ConvertCurrencyPrice(order.Price, config.Currency(order.Currency), config.CNY)
				if err != nil {
					restful.ParamErr(c, "获取汇率失败")
					return
				}
				orderPriceSumTemp = orderPriceSumTemp.Add(convertedPrice)
			}
		}
		orderPriceSumSeries.Data = append(orderPriceSumSeries.Data, orderPriceSumTemp)
		totalOrderPriceSum = totalOrderPriceSum.Add(orderPriceSumTemp)

		// 完成订单
		orderCountTemp := 0
		for _, order := range orders {
			if order.CreateTime > startTimestampTemp && order.CreateTime < endTimestampTemp {
				orderCountTemp = orderCountTemp + 1
			}
		}
		orderCountSeries.Data = append(orderCountSeries.Data, orderCountTemp)
		totalOrderCount = totalOrderCount + orderCountTemp

		// x坐标轴标签
		data.XAxis.Data = append(data.XAxis.Data, dayStart.In(config.Loc).Format("01-02"))
	}
	data.Series = append(data.Series, orderPriceSumSeries, orderCountSeries)

	data.Legend = Legend{
		Data:      []string{"订单收入", "完成订单"},
		Formatter: fmt.Sprintf("function (name) { const data_filter={'订单收入': %s, '完成订单':%d };return name + '\\n' + data_filter[name];}", totalOrderPriceSum, totalOrderCount),
		Padding:   []int{10, 20, 10, 20},
	}

	var respData = functions.StructToMap(data, functions.StructToMapExcludeMode)
	restful.Ok(c, respData)
}
