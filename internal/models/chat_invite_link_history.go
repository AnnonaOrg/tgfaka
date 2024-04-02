package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 使用指针可以方便的置空，使用原则：必须要判断是否为空的情况
type ChatInviteLinkHistory struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;not null" json:"id"`
	CreateTime     int64     `gorm:"index;autoCreateTime;not null" json:"create_time"`
	ProductItemID  uuid.UUID `gorm:"type:uuid;primary_key;not null" json:"product_item_id"`
	ChatInviteLink string    `json:"chat_invite_link"`
}

func (*ChatInviteLinkHistory) TableName() string {
	return "chat_invite_link_history"
}
func (*ChatInviteLinkHistory) DefaultOrder() string {
	return "create_time DESC"
}
func (t *ChatInviteLinkHistory) BeforeCreate(tx *gorm.DB) (err error) {
	t.ID = uuid.New()
	// create 后会把 foreignKey 对象丢掉
	// 赋予空对象，如果是指针则赋予nil，不然会插入数据库
	// 另一种做法是再建一个field用于储存，不赋予gorm标签
	//t.Order = nil
	// t.Order = nil
	return
}
func NewChatInviteLinkHistory(productItemID uuid.UUID, chatInviteLink string) *ChatInviteLinkHistory {
	itemHistory := &ChatInviteLinkHistory{
		ProductItemID:  productItemID,
		ChatInviteLink: chatInviteLink,
	}
	return itemHistory
}
