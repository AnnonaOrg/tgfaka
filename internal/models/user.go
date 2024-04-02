package models

import "github.com/google/uuid"

type User struct {
	ID uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
}

func (User) TableName() string {
	return "user"
}
