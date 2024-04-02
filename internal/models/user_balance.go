package models

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// 用户余额
type UserBalance struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;not null" json:"id"`
	Userid int64     `gorm:"index;not null" json:"userid"`
	// Username   string    `gorm:"index;not null" json:"username"`
	Status     int   `gorm:"default:0;not null" json:"status"`
	CreateTime int64 `gorm:"index;autoCreateTime;not null" json:"create_time"`

	Balance decimal.Decimal `gorm:"not null" json:"balance"`

	TGUsername string `gorm:"index;" json:"tg_username"`
	TGChatID   int64  `gorm:"index;not null" json:"tg_chat_id"`

	Username  string `gorm:"column:username;" json:"username"`
	FirstName string `gorm:"column:first_name;" json:"first_name"`
	LastName  string `gorm:"column:last_name;" json:"last_name"`
	// 邀请者
	Inviter string `json:"inviter" gorm:"column:inviter;"`
	// 邀请码
	InviterCode string `json:"inviter_code" gorm:"column:inviter_code;"`
}

func (*UserBalance) TableName() string {
	return "user_balance"
}
func (t *UserBalance) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()

	return
}
func (*UserBalance) DefaultOrder() string {
	return "create_time DESC"
}
func NewUserBalance(
	userid, tGChatID int64, username, tGUsername string, userBalance decimal.Decimal,
	firstName, lastName string,
	inviter, inviterCode string,
) *UserBalance {
	balance := &UserBalance{
		Userid: userid,
		// Username:   username,
		Balance:    userBalance,
		TGUsername: tGUsername,
		TGChatID:   tGChatID,

		Username:    username,
		FirstName:   firstName,
		LastName:    lastName,
		Inviter:     inviter,
		InviterCode: inviterCode,
	}
	return balance
}
