package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductItem struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;not null" json:"id"`
	Status     int       `gorm:"default:1;not null" json:"status"` //1未出售，-1已出售，0待支付                                                //0未出售，1已出售，-1待支付
	CreateTime int64     `gorm:"index;autoCreateTime;not null" json:"create_time"`
	Content    string    `gorm:"not null" json:"content"`

	EndLockTime *uint `gorm:"" json:"end_lock_time"`

	ProductID uuid.UUID `gorm:"" json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductID"`

	OrderID *uuid.UUID `json:"order_id"` // 只有在锁定或已完成交易才存在
	Order   *Order     `gorm:"foreignKey:OrderID"`
}

func (*ProductItem) TableName() string {
	return "product_item"
}

func (t *ProductItem) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	return
}
func (*ProductItem) DefaultOrder() string {
	return "create_time DESC"
}
func NewProductItem(content string, productID uuid.UUID) *ProductItem {
	product := &ProductItem{
		Content:   content,
		ProductID: productID,
	}
	return product
}
