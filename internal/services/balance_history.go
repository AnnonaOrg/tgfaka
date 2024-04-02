package services

import (
	"gopay/internal/exts/db"
	"gopay/internal/models"
)

func CreateBalanceHistory(balanceHistory *models.BalanceHistory) error {
	result := db.DB.Create(&balanceHistory)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func GetBalanceHistoryCount(balanceHistory *models.BalanceHistory) int64 {
	var count int64
	if err := db.DB.Model(&models.BalanceHistory{}).
		Where("info_hash = ?", balanceHistory.InfoHash).
		Count(&count).
		Error; err != nil {
		return 0
	} else {
		return count
	}
}
