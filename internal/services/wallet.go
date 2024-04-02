package services

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gopay/internal/exts/config"
	"gopay/internal/exts/db"
	"gopay/internal/models"
	"gopay/internal/utils/functions"
	"gorm.io/gorm"
	"strings"
	"time"
)

func GetWalletOrder() string {
	return "priority DESC,create_time DESC"
}

func GetFreeDecimalWallet(network string, currency config.Currency, orderPrice decimal.Decimal) (*models.Wallet, error) {
	var freeWallet *models.Wallet
	// 小数点尾数钱包
	// 获取单位步长
	decimalWalletUnit := config.DecimalWalletUnitMap[currency]
	if decimalWalletUnit.Equal(decimal.Zero) {
		return nil, errors.New("获取单位步长失败")
	}

	maxOrderCount := config.DecimalWalletMaxOrderCount
	maxPriceIncrement := decimalWalletUnit.Mul(decimal.NewFromInt(int64(maxOrderCount)))
	minPrice := orderPrice                        //闭区间
	maxPrice := orderPrice.Add(maxPriceIncrement) //开区间

	// 订单、钱包join查询，1.订单通过对应钱包地址和钱包关联 2.订单为待支付 3.订单的支付价格大于真实价格小于真实价格加上最大尾数，按钱包id分组，获取小于最大订单数的钱包
	// having 设置每个钱包对应的订单数限制
	result := db.DB.Raw(`
						SELECT wallet.* FROM wallet
						LEFT JOIN "order" ON wallet.id = "order".wallet_id AND
							 "order".status = 0 AND
							 "order".price >= ? AND
							 "order".price < ? AND
							 "order".network = ? AND
							 "order".currency = ?
						WHERE wallet.status = 2
						GROUP BY wallet.id
						HAVING COUNT("order".id) < ?
						ORDER BY wallet.priority DESC;
                  `, minPrice, maxPrice, network, currency, maxOrderCount).Order(GetWalletOrder()).Scan(&freeWallet)
	if result.Error != nil {
		return nil, errors.New("请求空闲钱包错误")
	} else if result.RowsAffected == 0 {
		return nil, errors.New("无空闲钱包2")
	}

	return freeWallet, nil

}

// 获取小数点尾数空闲钱包指定金额的可用金额
func GetFreeDecimalWalletPrice(network string, currency config.Currency, freeWalletID uuid.UUID, orderPrice decimal.Decimal) (*decimal.Decimal, error) {
	decimalWalletUnit := config.DecimalWalletUnitMap[currency]
	if decimalWalletUnit.Equal(decimal.Zero) {
		return nil, errors.New("获取单位步长失败")
	}

	maxOrderCount := config.DecimalWalletMaxOrderCount
	maxPriceIncrement := decimalWalletUnit.Mul(decimal.NewFromInt(int64(maxOrderCount)))
	minPrice := orderPrice                        //闭区间
	maxPrice := orderPrice.Add(maxPriceIncrement) //开区间

	var relatedOrders []models.Order
	result := db.DB.Where(` status = 0 AND
                        price >= ? AND
                        price < ? AND
                        network = ? AND
                        currency = ? AND
                        wallet_id = ?`, minPrice, maxPrice, network, currency, freeWalletID).Find(&relatedOrders)
	if result.Error != nil {
		return nil, errors.New("请求空闲钱包的订单错误")
	}

	// 获取所有已存在的价格的切片
	var existPrices []decimal.Decimal
	for _, relatedOrder := range relatedOrders {
		existPrices = append(existPrices, relatedOrder.Price)
	}

	var orderFinalPrice *decimal.Decimal
	for i := 0; i < 100; i++ {
		priceIncrement := decimalWalletUnit.Mul(decimal.NewFromInt(int64(i)))
		priceToCheck := orderPrice.Add(priceIncrement)
		if functions.SliceContainDecimal(existPrices, priceToCheck) {
			continue
		}
		orderFinalPrice = &priceToCheck
		break
	}

	if orderFinalPrice == nil {
		return nil, errors.New("从订单中获取最终价格失败")
	}

	return orderFinalPrice, nil
}

//
//func DeleteWallets(ids []uuid.UUID) error {
//	tx := db.DB.Begin()
//	defer tx.Rollback()
//
//	// gorm删除如果带外键自动置空
//	//// 注释订单钱包地址
//	//if err := tx.Model(models.Order{}).Where("wallet_id in ?", ids).Updates(map[string]interface{}{
//	//	"wallet_id":   gorm.Expr("NULL"),
//	//	"wallet_type": gorm.Expr("NULL"),
//	//}).Error; err != nil {
//	//	return err
//	//}
//	//
//	//if err := tx.Model(models.Transfer{}).Where("wallet_id in ?", ids).Updates(map[string]interface{}{
//	//	"wallet_id": gorm.Expr("NULL"),
//	//}).Error; err != nil {
//	//	return err
//	//}
//
//	if err := tx.Where("id in ?", ids).Delete(models.Wallet{}).Error; err != nil {
//		return err
//	}
//
//	err := tx.Commit().Error
//	if err != nil {
//		return err
//	}
//	return nil
//}

func UpdateWalletBalance(walletID uuid.UUID, balanceMap map[string]decimal.Decimal) error {
	jsonByte, err := json.Marshal(balanceMap)
	if err != nil {
		return err
	}
	jsonString := string(jsonByte)

	if result := db.DB.Model(&models.Wallet{}).Where("id=?", walletID).Updates(map[string]interface{}{
		"balance_data": jsonString,
		"refresh_time": time.Now().Unix(),
	}); result.Error != nil {
		return errors.New("更新钱包余额失败")
	} else if result.RowsAffected == 0 {
		return errors.New("更新0行")
	}
	return nil
}
func ClearExpireWallet() error {
	// 解锁钱包,只需更新status为0的钱包
	if result := db.DB.Model(&models.Wallet{}).Where("status = 0 and end_lock_time < ? ", time.Now().Unix()).Updates(map[string]interface{}{
		"status":        1,
		"end_lock_time": gorm.Expr("NULL"),
	}); result.Error != nil {
		return errors.New("解锁钱包失败")
	}
	return nil
}

// 在tx事务执行db.DB.Exec()，不知道为什么会被阻塞住
func UpdateWalletBalanceFromTransfers(wallets []models.Wallet, transfers []models.Transfer) error {
	if len(wallets) == 0 || len(transfers) == 0 {
		return nil
	}

	//// 找出钱包地址切片，并去重
	//tempMap := make(map[string]struct{})
	//var walletAddresses []string
	//for _, transfer := range transfers {
	//	var walletAddress string
	//	if transfer.Price.GreaterThan(decimal.Zero) {
	//		walletAddress = transfer.ToAddress
	//	} else {
	//		walletAddress = transfer.FromAddress
	//	}
	//	// 利用map的键来进行去重
	//	if _, exists := tempMap[walletAddress]; !exists {
	//		tempMap[walletAddress] = struct{}{}
	//		walletAddresses = append(walletAddresses, walletAddress)
	//	}
	//}

	for _, transfer := range transfers {
		for i, wallet := range wallets {
			if wallet.Address != transfer.FromAddress && wallet.Address != transfer.ToAddress {
				continue
			}
			wallets[i].AddBalance(transfer.Currency, transfer.Price)
		}
	}

	//addressUpdates := make(map[string]string)
	//for _, wallet := range wallets {
	//	// 转jsonfield成字符串
	//	balanceDataBytes, err := json.Marshal(wallet.BalanceData)
	//	if err != nil {
	//		continue
	//	}
	//	addressUpdates[wallet.Address] = string(balanceDataBytes)
	//}
	addressUpdates := make(map[string]interface{})
	for _, wallet := range wallets {
		addressUpdates[wallet.Address] = wallet.BalanceData
	}

	var sb strings.Builder
	sb.WriteString("UPDATE wallet SET balance_data = CASE ")

	var inClauseArgs []interface{}
	for address, _ := range addressUpdates {
		// pgsql设置为json了
		if config.DBConfig.DBType == config.POSTGRES {
			sb.WriteString("WHEN address = ? THEN ?::json ")
		} else {
			sb.WriteString("WHEN address = ? THEN ? ")
		}
		inClauseArgs = append(inClauseArgs, address)
	}

	sb.WriteString("END WHERE address IN (?")
	sb.WriteString(strings.Repeat(", ?", len(inClauseArgs)-1))
	sb.WriteString(")")

	args := make([]interface{}, 0, len(addressUpdates)*2)
	for address, value := range addressUpdates {
		args = append(args, address, value)
	}
	args = append(args, inClauseArgs...)

	query := sb.String()

	db.DB.Exec(query, args...)
	return nil
}
