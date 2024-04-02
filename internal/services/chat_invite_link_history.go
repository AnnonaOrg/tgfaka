package services

import (
	"gopay/internal/exts/db"
	"gopay/internal/models"
)

func CreateChatInviteLinkHistory(chatInviteLinkHistory *models.ChatInviteLinkHistory) error {
	result := db.DB.Create(&chatInviteLinkHistory)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func GetChatInviteLinkHistoryCount(productItemUUID string) int64 {
	var count int64
	if err := db.DB.Model(&models.ChatInviteLinkHistory{}).
		Where("product_item_id = ?", productItemUUID).
		Count(&count).
		Error; err != nil {
		return 0
	} else {
		return count
	}
}

func GetChatInviteLink(productItemUUID string) (string, error) {
	retText := ""
	var targetItem models.ChatInviteLinkHistory
	if err := db.DB.
		// Table("user").
		Model(&models.ChatInviteLinkHistory{}).
		Where("product_item_id = ?", productItemUUID).
		First(&targetItem).
		Error; err != nil {
		return "", err
	} else {
		retText = targetItem.ChatInviteLink
		return retText, nil
	}
}
