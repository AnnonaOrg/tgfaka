package models

import (
	"fmt"
	"github.com/google/uuid"
	"gopay/internal/exts/db"
	"gopay/internal/utils/functions"
	"gorm.io/gorm"
	"reflect"
	"slices"
	"strings"
)

type Operation func(*gorm.DB, string, interface{}) *gorm.DB

var allowOrderFields = []string{
	"create_time",
	"end_time",
	"price",
	"paid_price",
	"priority",
}

var fieldOperations = map[string]Operation{
	"status":          applyAndEquals,
	"network":         applyAndEquals,
	"currency":        applyAndEquals,
	"cate":            applyAndEquals,
	"timestamp_range": applyBetween,

	"id":             applyOrEquals,
	"address":        applyOrEquals,
	"from_address":   applyOrEquals,
	"to_address":     applyOrEquals,
	"order_id":       applyOrEquals,
	"transaction_id": applyOrEquals,
	"wallet_id":      applyOrEquals,
	"product_id":     applyOrEquals,
	"wallet_address": applyOrEquals,
	"tg_username":    applyOrEquals,
	//"keyword":         ApplyKeywordSearch,

}

//"order_id", "product_id", "wallet_id", "transaction_id", "wallet_address", "address"

func applyGreaterThan(query *gorm.DB, field string, value interface{}) *gorm.DB {
	return query.Where(field+" > ?", value)
}

func applyAndEquals(query *gorm.DB, field string, value interface{}) *gorm.DB {
	return applyEqualsWithType(query, field, value, true)
}
func applyOrEquals(query *gorm.DB, field string, value interface{}) *gorm.DB {
	return applyEqualsWithType(query, field, value, false)
}
func applyEqualsWithType(query *gorm.DB, field string, value interface{}, isAnd bool) *gorm.DB {
	valueString, ok := value.(string)
	var fieldResult = field
	var valueResult = value

	if ok {
		if (field == "id" || strings.Contains(field, "_id")) && field != "transaction_id" {
			result, err := uuid.Parse(valueString)
			if err != nil {
				// 如果含有id且值不为uuid
				fieldResult = "1"
				valueResult = "0"

			} else {
				// 如果含有id且值为uuid
				fieldResult = field
				valueResult = result.String()
			}
		}
	}

	if isAnd {
		return query.Where(fieldResult+" = ?", valueResult)
	} else {
		return query.Or(fieldResult+" = ?", valueResult)
	}

}

func applyBetween(query *gorm.DB, field string, value interface{}) *gorm.DB {
	field = "create_time"
	strValue, ok := value.(string)
	if !ok {
		return query
	}
	rangeParts := strings.Split(strValue, ",")
	if len(rangeParts) != 2 {
		return query
	}
	if rangeParts[0] == "" {
		return query.Where(field+" < ?", rangeParts[1])
	} else if rangeParts[1] == "" {
		return query.Where(field+" > ?", rangeParts[0])
	} else {
		return query.Where(field+" BETWEEN ? AND ?", rangeParts[0], rangeParts[1])
	}
}

//	func ApplyKeywordSearch(query *gorm.DB, _ string, keyword interface{}) *gorm.DB {
//		searchFields := []string{"id"}
//		keywordString, ok := keyword.(string)
//		if !ok {
//			return query
//		}
//		for _, searchField := range searchFields {
//			query = query.Or(fmt.Sprintf("%s = ?", searchField), keywordString)
//		}
//		return query
//	}

func applyOrder(query *gorm.DB, orderField string, orderDir string) *gorm.DB {
	return query.Order(fmt.Sprintf("%s %s", orderField, orderDir))
}

func ApplyFilters(query *gorm.DB, filterParams map[string]interface{}) *gorm.DB {
	var orQuery *gorm.DB

	for field, value := range filterParams {
		// nil跳过
		if value == nil {
			continue
		}
		// 空字符串跳过
		if str, ok := value.(string); ok && str == "" {
			continue
		}

		var operation Operation
		operation, ok := fieldOperations[field]
		if !ok {
			continue
		}

		if reflect.ValueOf(operation).Pointer() == reflect.ValueOf(applyOrEquals).Pointer() {
			if orQuery == nil {
				orQuery = operation(db.DB, field, value)
			} else {
				orQuery = operation(orQuery, field, value)
			}
		} else {
			query = operation(query, field, value)
		}

	}
	if orQuery != nil {
		query = query.Where(orQuery)
	}

	// 应用排序
	var orderData struct {
		OrderBy  string `json:"orderBy"`
		OrderDir string `json:"orderDir"`
	}
	if err := functions.MapToStruct(filterParams, &orderData); err == nil {
		if slices.Contains(allowOrderFields, orderData.OrderBy) && (orderData.OrderDir == "desc" || orderData.OrderDir == "asc") {
			query = query.Order(fmt.Sprintf("%s %s", orderData.OrderBy, orderData.OrderDir))
		}
	}

	return query
}
