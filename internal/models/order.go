package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// 使用指针可以方便的置空，使用原则：必须要判断是否为空的情况
type Order struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;not null" json:"id"`
	Status     int       `gorm:"default:0;not null" json:"status"` // 0，待支付 1,已支付 -1,超时 -2.强行关闭
	CreateTime int64     `gorm:"index;autoCreateTime;not null" json:"create_time"`
	EndTime    int64     `gorm:"index" json:"end_time"` // 结束时间,订单完成,则标记为支付时间
	//MerchantOrderID string    `gorm:"unique;index" json:"merchant_order_id"` // 用来防止用户重复创建订单，系统订单时为空

	Currency string `gorm:"not null" json:"currency"`
	Network  string `gorm:"not null" json:"network"`

	Price     decimal.Decimal `gorm:"not null" json:"price"`
	PaidPrice decimal.Decimal `gorm:"default:0;not null" json:"paid_price"`
	//PriceID        *decimal.Decimal `json:"price_id"`
	PriceIDForLock *string `gorm:"unique" json:"price_id_for_lock"` // 字符串，钱包-网络-货币-价格，从数据库层级防止重复价格

	BaseCurrency      string          `json:"base_currency"`
	BaseCurrencyPrice decimal.Decimal `json:"base_currency_price"`

	WalletID      uuid.UUID `gorm:"" json:"wallet_id"`
	Wallet        Wallet    `gorm:"foreignKey:WalletID"`
	WalletAddress string    `gorm:"" json:"wallet_address"`
	WalletType    int       `gorm:"" json:"wallet_type"`

	ProductID uuid.UUID `gorm:"" json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductID"`

	TGUsername string `gorm:"index;" json:"tg_username"`
	TGChatID   int64  `gorm:"index;not null" json:"tg_chat_id"`
	TGMsgID    int64  `gorm:"index;not null" json:"tg_msg_id"`

	//NotifyType string `json:"notify_type"`
	NotifyStatus uint `gorm:"default:0" json:"notify_status"`
	//notify_info = db.Column(JSONB, default={})

	Note string `gorm:"" json:"note"` //备注

	//UserID uuid.UUID `json:"user_id"`
	//User   User      `gorm:"foreignKey:UserID"`
	BuyNum int64 `gorm:"" json:"buy_num"`

	ProductItem []ProductItem `gorm:"constraint:OnDelete:SET NULL;"`
	Transfers   []Transfer    `gorm:"constraint:OnDelete:SET NULL;"`
}

func (*Order) TableName() string {
	return "order"
}
func (*Order) DefaultOrder() string {
	return "create_time DESC"
}
func (t *Order) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	//t.EndTime = time.Now().Unix() + int64(config.SiteConfig.OrderExpireDuration.Seconds())
	return
}
func NewOrder(endTime int64, currency string, network string, price decimal.Decimal, priceIDForLock *string, baseCurrency string, baseCurrencyPrice decimal.Decimal, walletID uuid.UUID, walletAddress string, walletType int, productID uuid.UUID, tgChatID int64, tgUsername string) *Order {
	order := &Order{
		EndTime:           endTime,
		Currency:          currency,
		Network:           network,
		Price:             price,
		PriceIDForLock:    priceIDForLock,
		BaseCurrency:      baseCurrency,
		BaseCurrencyPrice: baseCurrencyPrice,
		WalletID:          walletID,
		WalletAddress:     walletAddress,
		WalletType:        walletType,
		ProductID:         productID,
		TGChatID:          tgChatID,
		TGUsername:        tgUsername,
	}
	return order
}
