package services

// import (
// 	"fmt"
// 	"gopay/internal/exts/db"
// 	"gopay/internal/models"
// )

// func CreateUserFans(item *models.UserFans) error {
// 	result := db.DB.Create(&item)
// 	if result.Error != nil {
// 		return result.Error
// 	}
// 	return nil
// }

// func GetUserFansCount(item *models.UserFans) int64 {
// 	var count int64
// 	if err := db.DB.Model(&models.UserFans{}).
// 		Where("userid = ?", item.Userid).
// 		Count(&count).
// 		Error; err != nil {
// 		return 0
// 	} else {
// 		return count
// 	}
// }
// func GetUserFansCountByUserID(userid int64) int64 {
// 	var count int64
// 	if err := db.DB.Model(&models.UserFans{}).
// 		Where("userid = ?", userid).
// 		Count(&count).
// 		Error; err != nil {
// 		return 0
// 	} else {
// 		return count
// 	}
// }
// func GetUserFansCountByInviterCode(inviterCode string) int64 {
// 	var count int64
// 	if err := db.DB.Model(&models.UserFans{}).
// 		Where("inviter_code = ?", inviterCode).
// 		Count(&count).
// 		Error; err != nil {
// 		return 0
// 	} else {
// 		return count
// 	}
// }

// func GetUserFansAll() ([]models.UserFans, error) {
// 	var list []models.UserFans
// 	if err :=
// 		db.DB.Model(&models.UserFans{}).
// 			Where("status = 0").
// 			Find(&list).Error; err != nil {
// 		return list, fmt.Errorf("UserFans出错: %v", err)
// 	}
// 	return list, nil
// }
