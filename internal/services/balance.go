package services

import (
	"fmt"
	"gopay/internal/exts/db"
	"gopay/internal/models"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

func CreateUserBalance(
	userid int64, name, first, last string, balance int64,
	inviter, inviterCode string,
) error {
	item := models.NewUserBalance(
		userid, userid, name, name, decimal.NewFromInt(balance),
		first, last,
		inviter, inviterCode,
	)

	if err := db.DB.
		// Table("user").
		Model(&models.UserBalance{}).
		Create(item).
		Error; err != nil {
		return err
	} else {
		return nil
	}
}
func CheckUserBalance(userid int64, balance decimal.Decimal) (*models.UserBalance, bool, error) {
	var targetUser models.UserBalance
	if err := db.DB.
		Model(&models.UserBalance{}).
		Where("userid = ?", fmt.Sprintf("%d", userid)).
		First(&targetUser).
		Error; err != nil {

		return nil, false, fmt.Errorf("请充值后使用 err: %v", err)
	}
	targetBalance := balance // decimal.NewFromFloat(balance)
	if targetUser.Balance.Cmp(targetBalance) > 0 {
		return &targetUser, true, nil
	} else {
		return &targetUser, false, fmt.Errorf("余额不足: 待付金额(%v) 账户余额(%v)", targetBalance, targetUser.Balance)
	}
}

func UpdateUserBalance(userid int64, balance int64) error {
	if err := db.DB.
		// Table("user").
		Model(&models.UserBalance{}).
		Where("userid = ?", fmt.Sprintf("%d", userid)).
		Update("balance",
			gorm.Expr("balance + ?", balance),
		).
		Error; err != nil {
		return err
	} else {
		return nil
	}
}

func UpdateUserBalanceEx(userid int64, username, firstName, lastName string) error {
	if err := db.DB.
		// Table("user").
		Model(&models.UserBalance{}).
		Where("userid = ?", userid).
		Updates(
			map[string]interface{}{
				"username":     username,
				"first_name":   firstName,
				"last_name":    lastName,
				"inviter_code": fmt.Sprintf("%d", userid),
			},
		).
		Error; err != nil {
		return err
	} else {
		return nil
	}
}

func GetUserBalanceCount(userid int64) int64 {
	var count int64
	if err := db.DB.
		// Table("user").
		Model(&models.UserBalance{}).
		Where("userid = ?", fmt.Sprintf("%d", userid)).
		Count(&count).
		Error; err != nil {
		return 0
	} else {
		return count
	}
}
func GetUserBalanceCountByInviterCode(inviterCode string) int64 {
	var count int64
	if err := db.DB.Model(&models.UserBalance{}).
		Where("inviter_code = ?", inviterCode).
		Count(&count).
		Error; err != nil {
		return 0
	} else {
		return count
	}
}

func GetUserBalanceByInviterCode(inviterCode string) (*models.UserBalance, error) {
	var targetUser models.UserBalance
	if err := db.DB.
		// Table("user").
		Model(&models.UserBalance{}).
		Where("inviter_code = ?", inviterCode).
		First(&targetUser).
		Error; err != nil {
		return nil, err
	} else {
		return &targetUser, nil
	}
}

func GetUserBalanceNum(userid int64) (decimal.Decimal, error) {
	var balance decimal.Decimal
	var targetUser models.UserBalance
	if err := db.DB.
		// Table("user").
		Model(&models.UserBalance{}).
		Where("userid = ?", fmt.Sprintf("%d", userid)).
		First(&targetUser).
		Error; err != nil {
		return decimal.NewFromInt(0), err
	} else {
		balance = targetUser.Balance
		return balance.Round(2), nil
	}
}

func GetUserBalanceNumAndInviterCode(userid int64) (decimal.Decimal, string, error) {
	var balance decimal.Decimal
	var targetUser models.UserBalance
	if err := db.DB.
		// Table("user").
		Model(&models.UserBalance{}).
		Where("userid = ?", fmt.Sprintf("%d", userid)).
		First(&targetUser).
		Error; err != nil {
		return decimal.NewFromInt(0), "", err
	} else {
		balance = targetUser.Balance
		inviterCode := targetUser.InviterCode
		if len(inviterCode) == 0 {
			inviterCode = fmt.Sprintf("%d", targetUser.Userid)
		}
		return balance.Round(2), inviterCode, nil
	}
}

func GetUserBalanceAll() ([]models.UserBalance, error) {
	var list []models.UserBalance
	if err :=
		db.DB.Model(&models.UserBalance{}).
			Where("status = 0").
			Find(&list).Error; err != nil {
		return list, fmt.Errorf("UserBalance出错: %v", err)
	}
	return list, nil
}
