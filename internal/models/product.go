package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;not null" json:"id"`
	Name        string    `gorm:"index;not null" json:"name"`
	Description string    `gorm:"not null" json:"description"`
	Status      int       `gorm:"default:0;not null" json:"status"`
	CreateTime  int64     `gorm:"index;autoCreateTime;not null" json:"create_time"`
	// 库存
	InStockCount int64 `gorm:"default:0;not null" json:"in_stock_count"`
	Priority     int64 `gorm:"default:0;not null" json:"priority"`

	Currency string          `gorm:"not null" json:"currency"`
	Price    decimal.Decimal `gorm:"not null" json:"price"`

	ProductItems []ProductItem `gorm:"constraint:OnDelete:CASCADE;"` // product_item有product_id外键联系，product被删除时会联级删除(仅限Delete函数)
	Orders       []Order       `gorm:"constraint:OnDelete:SET NULL;"`
}

func (*Product) TableName() string {
	return "product"
}
func (t *Product) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()

	return
}
func (*Product) DefaultOrder() string {
	return "priority DESC"
}
func NewProduct(name string, description string, currency string, price decimal.Decimal) *Product {
	product := &Product{
		Name:        name,
		Description: description,
		Currency:    currency,
		Price:       price,
	}
	return product
}
