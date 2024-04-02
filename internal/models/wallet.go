package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/umfaka/tgfaka/internal/exts/config"
	"gorm.io/gorm"
)

// 钱包删除，order的wallet_id置空，这样schedule就查不到这个订单，晾在那里等待过期就行，同时不能联级删除订单，因为释放订单连着product_item释放
type Wallet struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Status     int       `gorm:"default:0;not null" json:"status"` // 0,占用 1,固定金额钱包,空闲 2.小数点尾数钱包 3.固定钱包专用
	CreateTime int64     `gorm:"index;autoCreateTime;not null" json:"create_time"`

	Network    string  `gorm:"not null" json:"network"`
	Address    string  `gorm:"index;not null" json:"address"`
	PrivateKey *string `json:"private_key"`

	BalanceData *JSONField `gorm:"type:json" json:"balance_data"`

	//StartLockTime uint `gorm:"not null" json:"start_lock_time"`
	EndLockTime *uint `gorm:"" json:"end_lock_time"`

	RefreshTime int64  `json:"refresh_time"`
	Priority    int64  `gorm:"default:0;not null" json:"priority"`
	Remark      string `json:"remark"`

	//UserID uuid.UUID `json:"user_id"`
	//User   User      `gorm:"foreignKey:UserID"`

	Orders    []Order    `gorm:"constraint:OnDelete:SET NULL;"`
	Transfers []Transfer `gorm:"constraint:OnDelete:SET NULL;"`
}

func (*Wallet) TableName() string {
	return "wallet"
}
func (w *Wallet) BeforeCreate(tx *gorm.DB) (err error) {
	w.ID = uuid.New()

	return
}
func (*Wallet) DefaultOrder() string {
	return "priority DESC"
}
func NewWallet(network config.Network, address string, privateKey *string, status int, priority int64) *Wallet {
	wallet := &Wallet{
		Network:    string(network),
		Address:    address,
		PrivateKey: privateKey,
		Status:     status,
		Priority:   priority,
	}
	return wallet
}
func (w *Wallet) AddBalance(currency string, price decimal.Decimal) {
	temp := make(map[string]decimal.Decimal)

	// 如果w.BalanceData为nil，则说明为null，新建一个空对象给他
	// 如果不为空，则遍历他，将他的键值保存在temp中，值转成decimal（其实不用遍历，只需关注currency即可）
	if w.BalanceData != nil {
		for key, val := range *w.BalanceData {
			strValue, ok := val.(string)
			if !ok {
				continue
			}
			decValue, err := decimal.NewFromString(strValue)
			if err != nil {
				continue
			}
			temp[key] = decValue
		}
	} else {
		w.BalanceData = &JSONField{}
	}

	// 如果temp中的currency存在，则添加，不存在，则赋值
	value, exists := temp[currency]
	if exists {
		//temp[currency] = temp[currency].Add(price)
		(*w.BalanceData)[currency] = value.Add(price).String()
	} else {
		//temp[currency] = price
		(*w.BalanceData)[currency] = price.String()
	}
}
