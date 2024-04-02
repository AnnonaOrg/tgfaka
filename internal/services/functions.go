package services

import (
	"errors"

	"github.com/google/uuid"
	"github.com/umfaka/tgfaka/internal/exts/db"
	"github.com/umfaka/tgfaka/internal/models"
	"gorm.io/gorm"
)

type Pagination struct {
	Limit     int           `json:"limit"`
	Page      int           `json:"page"`
	Total     int64         `json:"total"`
	TotalPage int64         `json:"total_page"`
	Items     []interface{} `json:"items"`
}

// !!!struct的函数不能用泛型
func Paginate[T any](pagination *Pagination, query *gorm.DB) error {
	// 不能放在offset的后面
	var total int64
	if err := query.Model(new(T)).Count(&total).Error; err != nil {
		return err
	}

	query = query.Scopes(func(db *gorm.DB) *gorm.DB {
		if pagination.Page < 1 {
			pagination.Page = 1
		}
		if pagination.Limit == 0 {
			pagination.Limit = 10
		}
		offset := (pagination.Page - 1) * pagination.Limit
		return db.Offset(offset).Limit(pagination.Limit)
	})

	var modelItems []T
	result := query.Find(&modelItems)
	if result.Error != nil {
		return result.Error
	}

	for _, modelItem := range modelItems {
		pagination.Items = append(pagination.Items, modelItem)
	}

	pagination.Total = total
	pagination.TotalPage = (pagination.Total + int64(pagination.Limit) - 1) / int64(pagination.Limit)

	return nil
}

func CreateEntity[T models.MyModel](entity *T) error {
	result := db.DB.Create(&entity)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
func CreateEntities[T models.MyModel](entity []T) error {
	result := db.DB.Create(&entity)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
func UpdateEntity[T models.MyModel](entityID uuid.UUID, updateMap map[string]interface{}) error {
	var entity T
	result := db.DB.Model(entity).Where("id=?", entityID).Updates(updateMap)
	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return errors.New("影响0行")
	}
	return nil
}
func DeleteEntities[T models.MyModel](ids []uuid.UUID) error {
	var entity T
	result := db.DB.Model(entity).Delete("id in ?", ids)
	if result.RowsAffected == 0 {
		return errors.New("not_found")
	}

	return nil
}
func DeleteAllEntities[T models.MyModel]() error {
	var entity T
	result := db.DB.Model(entity).Where("1=1").Delete("1=1")
	if result.RowsAffected == 0 {
		return errors.New("not_found")
	}

	return nil
}
