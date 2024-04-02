package admin_handler

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"gopay/internal/exts/config"
	"gopay/internal/utils/functions"
	"reflect"
)

func init() {
	// 注册binding validator给json bind使用
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		panic("获取validate错误")
	}
	err := v.RegisterValidation("network", networkValidationFunc)
	if err != nil {
		panic(err)
	}

}

func networkValidationFunc(fl validator.FieldLevel) bool {
	if fl.Field().Kind() == reflect.String {
		value := fl.Field().String()
		networks := config.GetAllNetworks()
		return functions.SliceContainString(networks, value)
	}
	return false
}
