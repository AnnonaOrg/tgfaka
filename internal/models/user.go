package models

import "github.com/google/uuid"

type User struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
}

func (User) TableName() string {
	return "user"
}

//
//import (
//	"github.com/google/uuid"
//	"golang.org/x/crypto/bcrypt"
//	"gopay/internal/utils/functions"
//	"gorm.io/gorm"
//)
//
//type User struct {
//	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
//	Status       int      `gorm:"index;default:1" json:"status"`
//	Username     string    `gorm:"unique;index" json:"username"`
//	PasswordHash string    `json:"password_hash"`
//
//	CreateTime uint   `gorm:"autoCreateTime" json:"create_time"`
//	LoginTime  uint   `json:"login_time"`
//	CreateIP   string `json:"create_ip"`
//	LoginIP    string `json:"login_ip"`
//
//	MerchantSecret string `json:"merchant_secret"`
//
//	Points     uint `gorm:"default:0" json:"points"`
//	Plan       uint `gorm:"default:0" json:"plan"`
//	ExpireTime uint `json:"expire_time"`
//
//	BalanceData map[string]interface{} `gorm:"type:jsonb" json:"balance_data"`
//	Config      map[string]interface{} `gorm:"type:jsonb" json:"config"`
//	Meta        map[string]interface{} `gorm:"type:jsonb" json:"meta"`
//}
//
//func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
//	u.ID = uuid.New()
//	u.MerchantSecret = functions.GenerateRandomString(16)
//
//	return
//}
//
//func (u *User) SetPassword(password string) {
//	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
//	u.PasswordHash = string(hash)
//	return
//}
