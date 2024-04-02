package models

import (
	"fmt"

	"github.com/umfaka/tgfaka/internal/utils"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// 使用指针可以方便的置空，使用原则：必须要判断是否为空的情况
type BalanceHistory struct {
	// ID        uint   `gorm:"primaryKey"`
	ID uuid.UUID `gorm:"type:uuid;primary_key;not null" json:"id"`
	// Status    int       `gorm:"default:1;not null" json:"status"` //1 未通知  2 已通知
	InfoHash string `gorm:"unique" json:"info_hash"`

	CreateTime int64 `gorm:"index;autoCreateTime;not null" json:"create_time"`

	Price     decimal.Decimal `gorm:"not null" json:"price"`
	PaidPrice decimal.Decimal `gorm:"default:0;not null" json:"paid_price"`

	BaseCurrency      string          `json:"base_currency"`
	BaseCurrencyPrice decimal.Decimal `json:"base_currency_price"`

	OrderID  *uuid.UUID `json:"order_id"` //gorm:"size:255;"
	Order    *Order     `gorm:"foreignKey:OrderID"`
	OrderObj *Order     `gorm:"-"`

	TGUsername string `gorm:"index;" json:"tg_username"`
	TGChatID   int64  `gorm:"index;not null" json:"tg_chat_id"`
	// TGMsgID    int64  `gorm:"default:0;not null" json:"tg_msg_id"` //推送消息id
	Note string `json:"note"`

	BalanceNum int64 `json:"balance_num"`
}

func (*BalanceHistory) TableName() string {
	return "balance_history"
}
func (*BalanceHistory) DefaultOrder() string {
	return "create_time DESC"
}
func (t *BalanceHistory) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	// create 后会把 foreignKey 对象丢掉
	// 赋予空对象，如果是指针则赋予nil，不然会插入数据库
	// 另一种做法是再建一个field用于储存，不赋予gorm标签
	//t.Order = nil
	t.Order = nil
	return
}
func NewBalanceHistory(price decimal.Decimal, baseCurrency string, baseCurrencyPrice decimal.Decimal, orderID *uuid.UUID, tgChatID int64, tgUsername string) *BalanceHistory {
	oid := orderID.String()
	if len(oid) == 0 {
		oid = uuid.New().String()
	}
	infoHash := utils.EncryptMd5(
		fmt.Sprintf("%d_%s", tgChatID, oid),
	)
	balanceHistory := &BalanceHistory{
		InfoHash:          infoHash,
		Price:             price,
		BaseCurrency:      baseCurrency,
		BaseCurrencyPrice: baseCurrencyPrice,
		OrderID:           orderID,
		TGChatID:          tgChatID,
		TGUsername:        tgUsername,
	}
	return balanceHistory
}
