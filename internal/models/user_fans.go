package models

// import (
// 	"github.com/google/uuid"

// 	"gorm.io/gorm"
// )

// // fans
// type UserFans struct {
// 	ID         uuid.UUID `gorm:"type:uuid;primary_key;not null" json:"id"`
// 	Userid     int64     `gorm:"index;not null" json:"userid"`
// 	Status     int       `gorm:"default:0;not null" json:"status"`
// 	CreateTime int64     `gorm:"index;autoCreateTime;not null" json:"create_time"`

// 	Username  string `gorm:"" json:"username"`
// 	FirstName string `gorm:"" json:"first_name"`
// 	LastName  string `gorm:"" json:"last_name"`
// 	// 邀请者
// 	Inviter int64 `json:"inviter" gorm:"column:inviter;"`
// 	// 邀请码
// 	InviterCode string `json:"inviter_code" gorm:"column:inviter_code;"`
// }

// func (*UserFans) TableName() string {
// 	return "user_fans"
// }
// func (t *UserFans) BeforeCreate(tx *gorm.DB) (err error) {
// 	t.ID = uuid.New()

// 	return
// }
// func (*UserFans) DefaultOrder() string {
// 	return "create_time DESC"
// }
// func NewUserFans(userid int64, username, firstName, lastName string, inviter int64, inviterCode string) *UserFans {
// 	item := &UserFans{
// 		Userid:      userid,
// 		Username:    username,
// 		FirstName:   firstName,
// 		LastName:    lastName,
// 		Inviter:     inviter,
// 		InviterCode: inviterCode,
// 	}
// 	return item
// }
