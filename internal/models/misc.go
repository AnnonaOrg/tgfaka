package models

import (
	"database/sql/driver"
	"errors"
)

import (
	"encoding/json"
)

// Value 和 Scan方法用来处理非标准sql类型
// Value()定义了我们的值如何储存到数据库中
// Scan()定义数据怎么从数据库中取出转换为我们的自定义类型
// 当您使用指针接收器定义 Value 方法时，仅当 JSONField 是指针或结构中的字段是指向 JSONField 的指针时才会调用该方法。

type JSONField map[string]interface{} // 不能直接使用，需要解指针(*w.BalanceData)["field"]，并且使用前要判断是否为nil

func (j *JSONField) Scan(value interface{}) error {
	if value == nil {
		*j = make(map[string]interface{})
		return nil
	}

	var byteValue []byte

	// 区分db中的数据是string还是bytes
	switch v := value.(type) {
	case []byte:
		byteValue = v
	case string:
		byteValue = []byte(v)
	default:
		return errors.New("unsupported type for JSONMap")
	}

	return json.Unmarshal(byteValue, j)
}
func (j JSONField) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

type MyModel interface {
	DefaultOrder() string
	TableName() string
}
