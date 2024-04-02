package admin_handler

import (
	"encoding/json"
	"regexp"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/umfaka/tgfaka/internal/exts/config"
	"github.com/umfaka/tgfaka/internal/utils/functions"
	"github.com/umfaka/tgfaka/internal/utils/restful"
)

func Setting(c *gin.Context) {
	items := config.ConfigToMap(config.GetSiteConfig())

	respData := map[string]interface{}{
		"items": items,
	}
	restful.Ok(c, respData)
}

func EditSetting(c *gin.Context) {
	var requestData struct {
		Key   string      `json:"key" binding:"required"`
		Value interface{} `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		restful.ParamErr(c, "参数错误")
		return
	}

	//单独校验 小数点钱包单位步长 和 小数点钱包最大增量
	if requestData.Key == "decimal_wallet_unit" {
		value, ok := requestData.Value.(string)
		if !ok {
			restful.ParamErr(c, "类型错误")
			return
		}
		re, _ := regexp.Compile(`^0\.0{0,5}1$`)
		if !re.MatchString(value) {
			restful.ParamErr(c, "格式错误")
			return
		}
	} else if requestData.Key == "decimal_wallet_max_increment" {
		value, ok := requestData.Value.(string)
		if !ok {
			restful.ParamErr(c, "类型错误")
			return
		}
		valueFloat, err := strconv.ParseFloat(value, 64)
		if err != nil {
			restful.ParamErr(c, "格式错误1")
			return
		}
		if valueFloat < 0.000000001 || valueFloat > 1 {
			restful.ParamErr(c, "格式错误2")
			return
		}
	} else if requestData.Key == "wallet_type" {
		valueFloat64, ok := requestData.Value.(float64)
		if !ok {
			restful.ParamErr(c, "类型错误")
			return
		}
		value := int(valueFloat64)
		if value != 1 && value != 2 {
			restful.ParamErr(c, "钱包类型错误")
			return
		}
	}

	if requestData.Key == "fixed_exchange_rate" {
		jsonData, err := json.Marshal(requestData.Value)
		if err != nil {
			restful.ParamErr(c, "json错误")
			return
		}
		requestData.Value = string(jsonData)
	}

	//DecimalWalletUnit         string         `validate:"numeric" json:"decimal_wallet_unit" desc:"小数点钱包单位步长,同时也是保留小数位数,如0.0001"`
	//DecimalWalletMaxIncrement string         `validate:"numeric" json:"decimal_wallet_max_increment" desc:"小数点钱包最大增量,如0.01"`

	siteConfig := config.GetSiteConfig()
	siteConfigMap := functions.StructToMap(siteConfig, functions.StructToMapExcludeMode)

	if _, exists := siteConfigMap[requestData.Key]; !exists {
		restful.ParamErr(c, "键不存在")
		return
	}

	siteConfigMap[requestData.Key] = requestData.Value
	var changedConfigData config.SiteConfigStruct
	err := functions.MapToStruct(siteConfigMap, &changedConfigData)
	if err != nil {
		restful.ParamErr(c, "设置失败:"+err.Error())
		return
	}

	err = config.SetSiteConfig(changedConfigData)
	if err != nil {
		restful.ParamErr(c, "保存失败")
		return
	}
	restful.Ok(c, "修改成功")
}
