package schedule

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gopay/internal/exts/config"
	"gopay/internal/exts/db"
	my_log "gopay/internal/exts/log"
	"gopay/internal/models"
	"gopay/internal/services"
	"gopay/internal/utils/crypto_api/tron"
	"gopay/internal/utils/functions"
	"gopay/internal/utils/handle_defender"
	"gopay/internal/utils/requests"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

func startCheckTransaction(network string) {
	var err error

	defer func() {
		if r := recover(); r != nil {
			msgText := fmt.Sprintf("监听任务崩溃, Network: %s", network)
			handle_defender.HandlePanic(r, msgText)
		}
		if err != nil {
			msgText := fmt.Sprintf("监听任务出错, Network: %s", network)
			handle_defender.HandleError(err, msgText)
		}
	}()

	my_log.LogInfo(fmt.Sprintf("开始获取%s交易", network))
	startTimestamp := functions.GetCurrentSecondTimestampFloat()
	defer func() {
		my_log.LogInfo(fmt.Sprintf("结束获取%s交易, 用时:%.3fs", network, functions.GetCurrentSecondTimestampFloat()-startTimestamp))
	}()

	var onChainTransfers []models.Transfer
	switch network {
	case "TRON":
		client := tron.New(config.SiteConfig.TronGridApiKey)
		onChainTransfers, err = client.GetScheduleTransfers()
		if err != nil {
			err = errors.New(fmt.Sprintf("获取最新区块失败, Error: %v", err))
			my_log.LogWarn(err.Error())
		}
	//case "POLYGON":
	default:
		return
	}

	if len(onChainTransfers) == 0 {
		return
	}

	// 所有交易的钱包地址
	var onChainWalletAddresses []string
	for _, transfersOnChain := range onChainTransfers {
		onChainWalletAddresses = append(onChainWalletAddresses, transfersOnChain.FromAddress, transfersOnChain.ToAddress)
	}

	// 更新余额,不要使用tx的事务，因为会rollback
	// 也不能单独使用db.DB.Exec，因为同时操作wallet，数据库会死锁
	// defer也不能放在defer tx.Rollback()的后面，否则会在tx未提交的过程中执行
	// 也不能直接defer function1(var),这样获取到的参数是当时的变量
	var relatedWalletsForUpdateBalance []models.Wallet
	var toInsertTransfersForUpdateBalance []models.Transfer
	defer func() {
		services.UpdateWalletBalanceFromTransfers(relatedWalletsForUpdateBalance, toInsertTransfersForUpdateBalance)
	}()

	tx := db.DB.Begin()
	defer tx.Rollback()

	// 找出与交易相关的钱包
	var relatedWallets []models.Wallet
	result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("address in ?", onChainWalletAddresses).Find(&relatedWallets)
	if result.RowsAffected == 0 {
		return
	}

	var toInsertTransfers []models.Transfer
	for _, wallet := range relatedWallets {
		for _, onChainTransfer := range onChainTransfers {
			if wallet.Address == onChainTransfer.FromAddress {
				transferTemp := onChainTransfer
				transferTemp.Price = transferTemp.Price.Neg()
				transferTemp.WalletObj = wallet
				transferTemp.WalletID = wallet.ID
				toInsertTransfers = append(toInsertTransfers, transferTemp)
			}
			if wallet.Address == onChainTransfer.ToAddress {
				transferTemp := onChainTransfer
				transferTemp.WalletObj = wallet
				transferTemp.WalletID = wallet.ID
				toInsertTransfers = append(toInsertTransfers, transferTemp)
			}
		}
	}

	// 添加交易到数据库,因为API的查询问题可能会取到重复的交易记录，如果交易存在则从toInsertTransfers删除
	// 一次查询减少消耗
	// transfer表很大，如果该步很慢可以优化，首先查any，如果有重复，然后再in
	var toInsertTransactionID []string
	for _, toInsertTransfer := range toInsertTransfers {
		toInsertTransactionID = append(toInsertTransactionID, toInsertTransfer.TransactionID)
	}

	//从数据库中取到重复的transfer id 列表
	var duplicatedTransfers []models.Transfer
	db.DB.Where("transaction_id in ?", toInsertTransactionID).Find(&duplicatedTransfers)
	var duplicatedTransactionIDs []string
	for _, duplicatedTransfer := range duplicatedTransfers {
		duplicatedTransactionIDs = append(duplicatedTransactionIDs, duplicatedTransfer.TransactionID)
	}

	if len(duplicatedTransactionIDs) != 0 {
		var removedDuplicatedTransfers []models.Transfer
		for _, toInsertTransfer := range toInsertTransfers {
			if !functions.SliceContainString(duplicatedTransactionIDs, toInsertTransfer.TransactionID) {
				removedDuplicatedTransfers = append(removedDuplicatedTransfers, toInsertTransfer)
			}
		}
		toInsertTransfers = removedDuplicatedTransfers
	}

	// 给更新余额的函数传参
	relatedWalletsForUpdateBalance = relatedWallets
	toInsertTransfersForUpdateBalance = toInsertTransfers

	// 处理交易前,如果没有transfer，直接返回
	if len(toInsertTransfers) == 0 {
		return
	}

	// 寻找订单，并绑定在transfer上
	for i, toInsertTransfer := range toInsertTransfers {
		if !toInsertTransfer.Price.GreaterThan(decimal.Zero) {
			continue
		}

		// 即使订单状态已经超时，但是最后两分钟的时候有入账，超时之后检测出该笔交易 （假设说明：订单1超时后，订单2立刻发起，两个订单对应同一个钱包，订单1的最后两分钟有入账，并且超时后才检测出来，该笔交易归属订单的查询条件查询结果仍然是订单1，因为交易时间是在订单1的时间段内）
		// 交易时间在订单时间段内，订单状态为超时或未付款，交易网络和货币匹配订单，订单绑定的钱包和交易的钱包一致
		query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("create_time < ? AND end_time > ? AND (status = ? OR status = ?) AND wallet_id = ? AND network=? AND currency = ?",
			toInsertTransfer.CreateTime, toInsertTransfer.CreateTime, 0, -1, toInsertTransfer.WalletID, toInsertTransfer.Network, toInsertTransfer.Currency).Session(&gorm.Session{})

		// gorm Find First 函数寻找目标,没找到会赋予0值而不是nil指针
		//	绑定订单到交易上
		var relatedOrder *models.Order
		var result *gorm.DB
		if toInsertTransfer.WalletObj.Status == 2 {
			// 小数点尾数
			result = query.Where("price=?", toInsertTransfer.Price).Find(&relatedOrder)
		} else {
			// 任意金额
			result = query.Find(&relatedOrder)
		}

		// 查询有没有对应订单，如果有则挂载订单对象到transfer上以便后续操作
		if result.RowsAffected > 0 {
			toInsertTransfers[i].OrderObj = relatedOrder
			toInsertTransfers[i].OrderID = &relatedOrder.ID
		}

		// 记载transfer分类，如果钱包类型为0(即任意金额钱包占用)则为1，如果为1(小数点钱包)则为2
		if toInsertTransfers[i].OrderObj != nil {
			if toInsertTransfers[i].WalletObj.Status == 0 {
				toInsertTransfers[i].Cate = 1
			} else if toInsertTransfers[i].WalletObj.Status == 2 {
				toInsertTransfers[i].Cate = 2
			}
		}

		// log动账
		//my_log.LogInfo(fmt.Sprintf("钱包收入动账, 金额:%s %s 地址: %s 到 %s", toInsertTransfer.Price, toInsertTransfer.Currency, toInsertTransfer.FromAddress, toInsertTransfer.ToAddress))
	}

	// 删除没有挂载OrderObj的transfer
	var toInsertTransfersTemp []models.Transfer
	for _, toInsertTransfer := range toInsertTransfers {
		if toInsertTransfer.OrderObj == nil {
			continue
		} else {
			toInsertTransfersTemp = append(toInsertTransfersTemp, toInsertTransfer)
		}
	}
	toInsertTransfers = toInsertTransfersTemp

	if len(toInsertTransfers) == 0 {
		return
	}

	//tx := db.DB.Begin()
	//defer tx.Rollback()

	// 要先在事务中创建,不然后面查已付价格查不到
	tx.Create(&toInsertTransfers)

	// 更新订单信息，更新状态，结束时间，如果完成则更新状态
	// 如果这里有多个transfer对应一个订单,则会走多次,实际上走一次就够了,因为获取已付金额函数是基于transfer获取的,transfer在签名的事务就已经更新完毕了,浪费性能但是概率小,无伤大雅
	for _, transfer := range toInsertTransfers {
		if transfer.OrderObj == nil {
			continue
		}
		// 处理判断订单是否完成
		if transfer.WalletObj.Status == 2 {
			//// 小数点尾数,只要绑定上了就是完成
			transfer.OrderObj.Status = 1
			transfer.OrderObj.PaidPrice = transfer.Price
			transfer.OrderObj.EndTime = transfer.CreateTime
			transfer.OrderObj.PriceIDForLock = nil
			tx.Model(&models.Order{}).Where("id=?", transfer.OrderObj.ID).Updates(map[string]interface{}{
				"status":            1,
				"paid_price":        transfer.Price,
				"end_time":          transfer.CreateTime,
				"price_id_for_lock": gorm.Expr("NULL"),
			})

		} else {
			// 任意金额
			// 这里其实只用走一次

			orderPaidPrice := services.OrderPaidPrice(*transfer.OrderObj, tx)
			// 如果已付金额大于等于订单价格,订单完成,解锁钱包,小于则更新订单已付金额
			if orderPaidPrice.GreaterThanOrEqual(transfer.OrderObj.Price) {
				// 解锁钱包，只有未超时订单才解锁钱包，因为超时订单也能完成，防止区块链延迟
				// 要放在订单状态更新前判断
				if transfer.OrderObj.Status == 0 {
					transfer.WalletObj.Status = 1
					tx.Model(&models.Wallet{}).Where("id=?", transfer.WalletObj.ID).Updates(map[string]interface{}{
						"status": 1,
					})
				}

				// 更新订单完成
				transfer.OrderObj.Status = 1
				transfer.OrderObj.EndTime = transfer.CreateTime
				transfer.OrderObj.PaidPrice = orderPaidPrice
				tx.Model(&models.Order{}).Where("id=?", transfer.OrderObj.ID).Updates(map[string]interface{}{
					"status":     1,
					"end_time":   transfer.CreateTime,
					"paid_price": orderPaidPrice,
				})

			} else {
				// 更新订单已付金额
				transfer.OrderObj.PaidPrice = orderPaidPrice
				tx.Model(&models.Order{}).Where("id=?", transfer.OrderObj.ID).Updates(map[string]interface{}{
					"paid_price": orderPaidPrice,
				})
			}

		}
	}

	// 更新钱包余额

	// 回调
	// 获取状态为已支付的订单id,去重并回调(1.因为有可能是固定金额没有一次付清，所以要判断是否已完成,2.因为有可能一次监听包含多个多同一订单的付款所以要去重)
	// 标记商品项目为已支付
	var successOrderIDs []uuid.UUID
	// 获取已完成订单ID列表
	for _, toInsertTransfer := range toInsertTransfers {
		if toInsertTransfer.OrderObj == nil || toInsertTransfer.OrderObj.Status != 1 {
			continue
		}
		successOrderIDs = append(successOrderIDs, toInsertTransfer.OrderObj.ID)
	}

	//更新商品项目状态
	tx.Model(models.ProductItem{}).Where("order_id in ?", successOrderIDs).Updates(map[string]interface{}{
		"status":        -1,
		"end_lock_time": gorm.Expr("NULL"),
	})

	err = tx.Commit().Error
	if err != nil {
		return
	}

	// 发送信息
	services.OrderCallbackMultiple(successOrderIDs)

}

func startUpdateExchangeRate() {
	var err error
	defer func() {
		if r := recover(); r != nil {
			msgText := fmt.Sprintf("更新汇率崩溃")
			handle_defender.HandlePanic(r, msgText)
		}
		if err != nil {
			msgText := fmt.Sprintf("更新汇率出错")
			handle_defender.HandleError(err, msgText)
		}
	}()
	my_log.LogInfo("开始更新汇率")
	defer my_log.LogInfo("结束更新汇率")

	exchangeRate := config.ExchangeRateData
	baseCurrency := config.CNY
	for _, targetCurrency := range config.GetCryptoCurrencies() {
		url := fmt.Sprintf("https://www.okx.com/v3/c2c/otc-ticker/quotedPrice?side=buy&quoteCurrency=%s&baseCurrency=%s", baseCurrency, targetCurrency)
		respByte, err := requests.Get(url)
		if err != nil {
			err = errors.New(fmt.Sprintf("Request Err, Url:%s ,Error: %v", url, err))
			my_log.LogWarn(err.Error())
			return
		}
		var result struct {
			Code int `json:"code"`
			Data []struct {
				BestOption  bool
				DepositName string
				Payment     string
				Price       decimal.Decimal
			}
			DetailMsg    string
			ErrorCode    string
			ErrorMessage string
			Msg          string
		}
		err = json.Unmarshal(respByte, &result)
		if err != nil {
			err = errors.New(fmt.Sprintf("json解析错误"))
			my_log.LogWarn(err.Error())
			return
		}

		if len(result.Data) == 0 {
			err = errors.New(fmt.Sprintf("json格式错误"))
			my_log.LogWarn(err.Error())
			return
		}

		if !result.Data[0].Price.GreaterThan(decimal.NewFromInt(0)) {
			my_log.LogWarn(fmt.Sprintf("汇率数值错误"))
			return
		}
		exchangeRate.ExchangeRate[config.Currency(targetCurrency)] = result.Data[0].Price
	}
	exchangeRate.UpdateTime = time.Now().In(config.Loc).Format("2006-01-02 15:04")

	// 内存
	config.ExchangeRateData = exchangeRate

	// 文件
	err = config.SetExchangeRate(exchangeRate)
	if err != nil {
		err = errors.New(fmt.Sprintf("设置汇率失败"))
		my_log.LogWarn(err.Error())
		return
	}

}

func clearExpire() {
	var err error
	defer func() {
		if r := recover(); r != nil {
			msgText := fmt.Sprintf("清理过期订单崩溃")
			handle_defender.HandlePanic(r, msgText)
		}
		if err != nil {
			msgText := fmt.Sprintf("清理过期订单出错")
			handle_defender.HandleError(err, msgText)
		}
	}()
	my_log.LogInfo("开始清理过期")
	defer my_log.LogInfo("结束清理过期")

	err = services.ClearExpireOrder()
	if err != nil {
		err = errors.New(fmt.Sprintf("清理过期订单DB错误, Error: %v", err))
		my_log.LogWarn(err.Error())
	}

	err = services.ClearExpireWallet()
	if err != nil {
		err = errors.New(fmt.Sprintf("清理过期钱包DB错误, Error: %v", err))
		my_log.LogWarn(err.Error())
	}

	err = services.ClearExpireProductItem()
	if err != nil {
		err = errors.New(fmt.Sprintf("清理过期商品项目DB错误, Error: %v", err))
		my_log.LogWarn(err.Error())
	}

}
