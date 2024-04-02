package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/umfaka/tgfaka/internal/exts/config"
	"gorm.io/gorm"
)

// 属性为指针的时候，不会赋予默认值，json报marshal会跳过nil
type Transfer struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;not null" json:"id"`
	Status int       `gorm:"default:1;not null" json:"status"`
	Cate   uint      `gorm:"default:0;not null" json:"cate"` // 0.未分类 1.小数点钱包交易 2.非小数点钱包交易 3.用户的固定钱包交易

	Currency      string          `gorm:"not null" json:"currency"`
	Network       string          `gorm:"not null" json:"network"`
	TransactionID string          `gorm:"index;not null" json:"transaction_id"` // 不使用unique了,支出和收入重复了
	Price         decimal.Decimal `gorm:"not null" json:"price"`

	CreateTime int64 `gorm:"index;not null" json:"create_time"`

	FromAddress string `gorm:"index;not null" json:"from_address"`
	ToAddress   string `gorm:"index;not null" json:"to_address"`

	OrderID  *uuid.UUID `json:"order_id"`
	Order    *Order     `gorm:"foreignKey:OrderID"`
	OrderObj *Order     `gorm:"-"`

	WalletID  uuid.UUID `gorm:"" json:"wallet_id"`
	Wallet    Wallet    `gorm:"foreignKey:WalletID"`
	WalletObj Wallet    `gorm:"-"`
}

func (*Transfer) TableName() string {
	return "transfer"
}

func (t *Transfer) BeforeCreate(tx *gorm.DB) (err error) {
	// create 后会把foreignKey对象丢掉
	// 赋予空对象，如果是指针则赋予nil，不然会插入数据库
	// 另一种做法是再建一个field用于储存，不赋予gorm标签
	//t.Wallet = Wallet{}
	//t.Order = nil
	t.ID = uuid.New()
	return
}
func (*Transfer) DefaultOrder() string {
	return "create_time DESC"
}
func NewTransfer(transactionID string, currency config.Currency, network config.Network, fromAddress string, toAddress string, price decimal.Decimal, createTime int64) *Transfer {
	transfer := &Transfer{
		TransactionID: transactionID,
		Currency:      string(currency),
		Network:       string(network),
		FromAddress:   fromAddress,
		ToAddress:     toAddress,
		Price:         price,
		CreateTime:    createTime,
	}
	return transfer
}
