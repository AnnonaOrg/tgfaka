package services

import (
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/umfaka/tgfaka/internal/exts/config"
	"github.com/umfaka/tgfaka/internal/exts/db"
	"github.com/umfaka/tgfaka/internal/exts/tg_bot"
	"github.com/umfaka/tgfaka/internal/log"
	"github.com/umfaka/tgfaka/internal/models"
	"github.com/umfaka/tgfaka/internal/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 普通商品订单
func CreateOrder(targetCurrency config.Currency, targetNetwork string, product models.Product, tgChatID int64, tgUsername string, buyNum int64) (*models.Order, error) {
	if config.IsBalanceProduct(product.ID.String()) {
		return CreateOrderEx(
			targetCurrency, targetNetwork, product, tgChatID, tgUsername, buyNum,
		)
	}
	isInternalBuy := config.IsInternal(tgChatID)

	// 基础金额,需换算
	baseCurrency := product.Currency
	baseCurrencyPrice := product.Price.Mul(decimal.NewFromInt(buyNum))

	targetPrice, err := config.ConvertCurrencyPrice(baseCurrencyPrice, config.Currency(baseCurrency), targetCurrency)
	if err != nil {
		return nil, errors.New("获取汇率失败")
	}
	orderNote := ""
	targetPrice, orderNote = ApplyDiscount(targetPrice)
	targetPrice, orderNote = ApplyDiscountNum(buyNum, targetPrice)

	// 精度截断，当精度大于小数尾数步长，则截断，否则保留精度
	targetPrice = targetPrice.Round(-config.DecimalWalletUnitMap[targetCurrency].Exponent())

	// 使用tx.begin应该慎重，需要显式commit或rollback，不然会导致会话过多塞满数据库
	tx := db.DB.Begin()
	defer tx.Rollback()

	// 获取空闲钱包,分为1.任意金额钱包 2.小数点尾数钱包,
	// 任意金额钱包要锁,绑定订单后状态会从1变成0
	// 并获取最后的订单价格
	var orderFinalPrice *decimal.Decimal
	var priceIDForLock *string

	var freeWallet *models.Wallet
	walletType := config.SiteConfig.WalletType
	if walletType == 1 {
		orderFinalPrice = &targetPrice
		// 获取钱包并上锁，因为这个钱包状态需要更改的
		if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("status=? and network=?", 1, targetNetwork).Order(GetWalletOrder()).Find(&freeWallet); result.Error != nil {
			return nil, errors.New("获取钱包出错")
		} else if result.RowsAffected == 0 {
			return nil, errors.New("无空闲钱包1")
		}
		// 修改钱包状态为锁定
		if result := tx.Model(&models.Wallet{}).Where("id=?", freeWallet.ID).Update("status", 0); result.Error != nil {
			return nil, err
		} else if result.RowsAffected == 0 {
			return nil, errors.New("钱包状态出错")
		}

	} else if walletType == 2 {
		freeWallet, err = GetFreeDecimalWallet(targetNetwork, targetCurrency, targetPrice)
		if err != nil {
			return nil, err
		}
		// 获取最终的订单价格
		orderFinalPrice, err = GetFreeDecimalWalletPrice(targetNetwork, targetCurrency, freeWallet.ID, targetPrice)
		if err != nil {
			return nil, errors.New("获取最终订单价格失败: " + err.Error())
		}
		// 无法从函数获得指针，需要变量中转
		temp := fmt.Sprintf("%s-%s-%s-%s", freeWallet.Address, targetNetwork, targetCurrency, orderFinalPrice)
		if !isInternalBuy {
			priceIDForLock = &temp
		}

	} else {
		return nil, errors.New("钱包类型设置错误")
	}

	// 商品库存,获取一个空闲商品项目,并锁定
	// var productItem models.ProductItem
	productItemList := make([]*models.ProductItem, 0)
	if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("product_id=? and status=1", product.ID).Limit(int(buyNum)).Find(&productItemList); result.Error != nil {
		return nil, err
	} else if result.RowsAffected == 0 {
		return nil, errors.New("商品无库存")
	}
	productItemIDList := make([]uuid.UUID, 0)
	for _, v := range productItemList {
		productItemIDList = append(productItemIDList, v.ID)
	}

	end_time := time.Now().Unix() + int64(config.SiteConfig.OrderExpireDuration.Seconds())

	// 创建订单
	order := models.NewOrder(end_time, string(targetCurrency), targetNetwork, *orderFinalPrice, priceIDForLock, baseCurrency, baseCurrencyPrice, freeWallet.ID, freeWallet.Address, walletType, product.ID, tgChatID, tgUsername)
	order.BuyNum = buyNum
	order.Note = orderNote

	productItemStatus := 0
	if isInternalBuy {
		order.Status = 1
		end_time = time.Now().Unix() + 5
		order.EndTime = end_time
		order.WalletType = 0
		productItemStatus = -1
	}
	tx.Create(order)

	// 更新项目为待支付,设置解锁时间,并绑定到订单上,要在订单创建的事务之后
	// if result := tx.Model(&models.ProductItem{}).Where("id=?", productItem.ID).Updates(map[string]interface{}{
	if result := tx.Model(&models.ProductItem{}).Where("id IN ?", productItemIDList).Updates(map[string]interface{}{
		"status":        productItemStatus,
		"order_id":      order.ID,
		"end_lock_time": end_time,
	}); result.Error != nil {
		return nil, err
	} else if result.RowsAffected == 0 {
		return nil, errors.New("商品项目更新失败")
	}

	// 设置钱包解锁时间
	if result := tx.Model(&models.Wallet{}).Where("id=?", freeWallet.ID).Updates(map[string]interface{}{
		"end_lock_time": end_time,
	}); result.Error != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("提交失败, " + err.Error())
	}

	// 更新库存信息
	UpdateProductInStockCount([]uuid.UUID{product.ID})
	return order, nil
}

// 充值到余额的订单
func CreateOrderEx(targetCurrency config.Currency, targetNetwork string, product models.Product, tgChatID int64, tgUsername string, buyNum int64) (*models.Order, error) {
	isInternalBuy := config.IsInternal(tgChatID)
	balanceNum := buyNum

	// 基础金额,需换算
	baseCurrency := product.Currency
	baseCurrencyPrice := product.Price.Mul(decimal.NewFromInt(buyNum))

	targetPrice, err := config.ConvertCurrencyPrice(baseCurrencyPrice, config.Currency(baseCurrency), targetCurrency)
	if err != nil {
		return nil, errors.New("获取汇率失败")
	}
	orderNote := ""
	targetPrice, orderNote = ApplyDiscount(targetPrice)
	targetPrice, orderNote = ApplyDiscountNum(buyNum, targetPrice)
	// 精度截断，当精度大于小数尾数步长，则截断，否则保留精度
	targetPrice = targetPrice.Round(-config.DecimalWalletUnitMap[targetCurrency].Exponent())

	// 使用tx.begin应该慎重，需要显式commit或rollback，不然会导致会话过多塞满数据库
	tx := db.DB.Begin()
	defer tx.Rollback()

	// 获取空闲钱包,分为1.任意金额钱包 2.小数点尾数钱包,
	// 任意金额钱包要锁,绑定订单后状态会从1变成0
	// 并获取最后的订单价格
	var orderFinalPrice *decimal.Decimal
	var priceIDForLock *string

	var freeWallet *models.Wallet
	walletType := config.SiteConfig.WalletType
	if walletType == 1 {
		orderFinalPrice = &targetPrice
		// 获取钱包并上锁，因为这个钱包状态需要更改的
		if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("status=? and network=?", 1, targetNetwork).Order(GetWalletOrder()).Find(&freeWallet); result.Error != nil {
			return nil, errors.New("获取钱包出错")
		} else if result.RowsAffected == 0 {
			return nil, errors.New("无空闲钱包1")
		}
		// 修改钱包状态为锁定
		if result := tx.Model(&models.Wallet{}).Where("id=?", freeWallet.ID).Update("status", 0); result.Error != nil {
			return nil, err
		} else if result.RowsAffected == 0 {
			return nil, errors.New("钱包状态出错")
		}

	} else if walletType == 2 {
		freeWallet, err = GetFreeDecimalWallet(targetNetwork, targetCurrency, targetPrice)
		if err != nil {
			return nil, err
		}
		// 获取最终的订单价格
		orderFinalPrice, err = GetFreeDecimalWalletPrice(targetNetwork, targetCurrency, freeWallet.ID, targetPrice)
		if err != nil {
			return nil, errors.New("获取最终订单价格失败: " + err.Error())
		}
		// 无法从函数获得指针，需要变量中转
		temp := fmt.Sprintf("%s-%s-%s-%s", freeWallet.Address, targetNetwork, targetCurrency, orderFinalPrice)
		if !isInternalBuy {
			priceIDForLock = &temp
		}

	} else {
		return nil, errors.New("钱包类型设置错误")
	}

	// // 商品库存,获取一个空闲商品项目,并锁定
	// // var productItem models.ProductItem
	// productItemList := make([]*models.ProductItem, 0)
	// if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("product_id=? and status=1", product.ID).Limit(int(buyNum)).Find(&productItemList); result.Error != nil {
	// 	return nil, err
	// } else if result.RowsAffected == 0 {
	// 	return nil, errors.New("商品无库存")
	// }
	// productItemIDList := make([]uuid.UUID, 0)
	// for _, v := range productItemList {
	// 	productItemIDList = append(productItemIDList, v.ID)
	// }

	end_time := time.Now().Unix() + int64(config.SiteConfig.OrderExpireDuration.Seconds())

	// 创建订单
	order := models.NewOrder(
		end_time, string(targetCurrency), targetNetwork, *orderFinalPrice, priceIDForLock, baseCurrency, baseCurrencyPrice,
		freeWallet.ID, freeWallet.Address, walletType, product.ID, tgChatID, tgUsername,
	)
	order.BuyNum = balanceNum
	order.Note = orderNote
	if isInternalBuy {
		order.Status = 1
		end_time = time.Now().Unix() + 5
		order.EndTime = end_time
		order.WalletType = 0
	}
	tx.Create(order)

	// // 更新项目为待支付,设置解锁时间,并绑定到订单上,要在订单创建的事务之后
	// // if result := tx.Model(&models.ProductItem{}).Where("id=?", productItem.ID).Updates(map[string]interface{}{
	// if result := tx.Model(&models.ProductItem{}).Where("id IN ?", productItemIDList).Updates(map[string]interface{}{
	// 	"status":        productItemStatus,
	// 	"order_id":      order.ID,
	// 	"end_lock_time": end_time,
	// }); result.Error != nil {
	// 	return nil, err
	// } else if result.RowsAffected == 0 {
	// 	return nil, errors.New("商品项目更新失败")
	// }

	// 设置钱包解锁时间
	if result := tx.Model(&models.Wallet{}).Where("id=?", freeWallet.ID).Updates(map[string]interface{}{
		"end_lock_time": end_time,
	}); result.Error != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("提交失败, " + err.Error())
	}

	// // 更新库存信息
	// UpdateProductInStockCount([]uuid.UUID{product.ID})
	return order, nil
}

// 普通商品订单 通过余额支付
func CreateOrderByBalance(targetCurrency config.Currency, targetNetwork string, product models.Product, tgChatID int64, tgUsername string, buyNum int64) (*models.Order, error) {
	isInternalBuy := config.IsInternal(tgChatID)
	// targetCurrency := config.GetSiteConfig().BalanceCurrency
	// targetNetwork := "TRON"
	// 基础金额,需换算
	baseCurrency := product.Currency
	baseCurrencyPrice := product.Price.Mul(decimal.NewFromInt(buyNum))

	targetPrice, err := config.ConvertCurrencyPrice(baseCurrencyPrice, config.Currency(baseCurrency), targetCurrency)
	if err != nil {
		return nil, errors.New("获取汇率失败")
	}
	orderNote := ""
	targetPrice, orderNote = ApplyDiscount(targetPrice)
	targetPrice, orderNote = ApplyDiscountNum(buyNum, targetPrice)
	// 精度截断，当精度大于小数尾数步长，则截断，否则保留精度
	targetPrice = targetPrice.Round(-config.DecimalWalletUnitMap[targetCurrency].Exponent())
	// 核查用户余额是否可足额付款
	userBalance, _, err := CheckUserBalance(tgChatID, targetPrice)
	if err != nil {
		log.Errorf("CheckUserBalance(%d,%v): %v", tgChatID, targetPrice, err)
		return nil, fmt.Errorf("CheckUserBalance(%d,%v): %v", tgChatID, targetPrice, err)
	}

	// // 精度截断，当精度大于小数尾数步长，则截断，否则保留精度
	// targetPrice = targetPrice.Round(-config.DecimalWalletUnitMap[targetCurrency].Exponent())

	// 使用tx.begin应该慎重，需要显式commit或rollback，不然会导致会话过多塞满数据库
	tx := db.DB.Begin()
	defer tx.Rollback()

	// 获取空闲钱包,分为1.任意金额钱包 2.小数点尾数钱包,
	// 任意金额钱包要锁,绑定订单后状态会从1变成0
	// 并获取最后的订单价格
	var orderFinalPrice *decimal.Decimal
	orderFinalPrice = &targetPrice
	var priceIDForLock *string

	var freeWallet *models.Wallet
	walletType := config.SiteConfig.WalletType

	freeWallet, err = GetFreeDecimalWallet(targetNetwork, targetCurrency, targetPrice)
	if err != nil {
		log.Errorf("GetFreeDecimalWallet(%v,%v,%v): %v", targetNetwork, targetCurrency, targetPrice, err)

		return nil, fmt.Errorf("GetFreeDecimalWallet(%v,%v,%v): %v", targetNetwork, targetCurrency, targetPrice, err)
	}

	// var freeWallet *models.Wallet
	walletType = 100 //config.SiteConfig.WalletType + 1

	// 商品库存,获取一个空闲商品项目,并锁定
	// var productItem models.ProductItem
	productItemList := make([]*models.ProductItem, 0)
	if result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("product_id=? and status=1", product.ID).Limit(int(buyNum)).Find(&productItemList); result.Error != nil {
		return nil, fmt.Errorf("锁定商品(%v)库存出错: %v", product.ID, err)
	} else if result.RowsAffected == 0 {
		return nil, errors.New("商品无库存")
	}
	productItemIDList := make([]uuid.UUID, 0)
	for _, v := range productItemList {
		productItemIDList = append(productItemIDList, v.ID)
	}

	end_time := time.Now().Unix() + int64(10) //+ int64(config.SiteConfig.OrderExpireDuration.Seconds())

	// 创建订单
	productItemStatus := 0
	order := models.NewOrder(
		end_time, string(baseCurrency), "余额", *orderFinalPrice, priceIDForLock, baseCurrency, baseCurrencyPrice,
		freeWallet.ID, freeWallet.Address, walletType, product.ID, tgChatID, tgUsername,
	)
	order.BuyNum = buyNum
	order.Note = orderNote
	order.Status = 1
	productItemStatus = -1

	if isInternalBuy {
		order.Status = 1
		end_time = time.Now().Unix() + 10
		order.EndTime = end_time
		order.WalletType = 0
		productItemStatus = -1
	}

	tx.Create(order)

	// 更新项目为待支付,设置解锁时间,并绑定到订单上,要在订单创建的事务之后
	// if result := tx.Model(&models.ProductItem{}).Where("id=?", productItem.ID).Updates(map[string]interface{}{
	if result := tx.Model(&models.ProductItem{}).Where("id IN ?", productItemIDList).Updates(map[string]interface{}{
		"status":        productItemStatus,
		"order_id":      order.ID,
		"end_lock_time": end_time,
	}); result.Error != nil {
		return nil, fmt.Errorf("设置订单(%s)解锁时间出错: %v", order.ID, err)
	} else if result.RowsAffected == 0 {
		return nil, errors.New("商品项目更新失败")
	}

	newBalance := userBalance.Balance.Sub(*orderFinalPrice)
	// if err := db.DB.
	// 	// Table("user").
	// 	Model(&models.UserBalance{}).
	// 	Where("userid = ?", userid).
	// 	Update("balance", newBalance).
	// 	Error; err != nil {
	// 	return err
	// } else {
	// 	return nil
	// }
	// 更新用户余额
	if result := tx.Model(&models.UserBalance{}).Where("userid=?", tgChatID).Updates(map[string]interface{}{
		"balance": newBalance,
	}); result.Error != nil {
		return nil, fmt.Errorf("更新用户(%d)余额出错: %v", tgChatID, err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("提交失败, " + err.Error())
	}

	// 更新库存信息
	UpdateProductInStockCount([]uuid.UUID{product.ID})
	return order, nil
}

func ClearExpireOrder() error {

	// 设置订单过期
	if result := db.DB.Model(&models.Order{}).Where("status = 0 and end_time < ?", time.Now().Unix()).Updates(map[string]interface{}{
		"status":            -1,
		"end_time":          time.Now().Unix(),
		"price_id_for_lock": gorm.Expr("NULL"),
	}); result.Error != nil {
		return errors.New("设置订单过期失败")
	}

	return nil
}

// 强行关闭订单
func ReleaseOrders(toReleaseOrderIDsInput []uuid.UUID) error {
	// 如果不判斷訂單已過期會導致後面的解鎖項目出問題，商品項目售出會解鎖重新出售，錢包會無故解鎖
	var toReleaseOrders []models.Order
	if result := db.DB.Where("status = 0 and id in ?", toReleaseOrderIDsInput).Find(&toReleaseOrders); result.Error != nil {
		return errors.New("查询订单失败")
	}
	var toReleaseOrderIDs []uuid.UUID
	for _, toReleaseOrders := range toReleaseOrders {
		toReleaseOrderIDs = append(toReleaseOrderIDs, toReleaseOrders.ID)
	}
	if len(toReleaseOrderIDs) == 0 {
		return nil
	}

	tx := db.DB.Begin()
	defer tx.Rollback()
	// 设置订单失效
	if result := db.DB.Model(&models.Order{}).Where("id in ?", toReleaseOrderIDs).Updates(map[string]interface{}{
		"status":            -2,
		"end_time":          time.Now().Unix(),
		"price_id_for_lock": gorm.Expr("NULL"),
	}); result.Error != nil {
		return errors.New("设置订单关闭失败")
	}

	// 解锁钱包,只需更新status为0的钱包
	var toReleaseWalletIDs []uuid.UUID
	for _, toReleaseOrder := range toReleaseOrders {
		toReleaseWalletIDs = append(toReleaseWalletIDs, toReleaseOrder.WalletID)
	}

	if result := tx.Model(&models.Wallet{}).Where("status = 0 and id in ?", toReleaseWalletIDs).Updates(map[string]interface{}{
		"status": 1,
	}); result.Error != nil {
		return errors.New("解锁钱包失败")
	}

	// 解锁商品项目,取消绑定订单
	if result := tx.Model(&models.ProductItem{}).Where("order_id in ?", toReleaseOrderIDs).Updates(map[string]interface{}{
		"status":        1,
		"order_id":      gorm.Expr("NULL"),
		"end_lock_time": gorm.Expr("NULL"),
	}); result.Error != nil {
		return errors.New("解锁商品项目失败")
	}
	if err := tx.Commit().Error; err != nil {
		return errors.New("清理过期订单提交失败, " + err.Error())
	}

	// 更新商品库存
	var productIDs []uuid.UUID
	for _, expiredOrder := range toReleaseOrders {
		productIDs = append(productIDs, expiredOrder.ProductID)
	}
	UpdateProductInStockCount(productIDs)

	// 删消息
	for _, expiredOrder := range toReleaseOrders {
		deleteConfig := tgbotapi.DeleteMessageConfig{
			ChatID:    expiredOrder.TGChatID,
			MessageID: int(expiredOrder.TGMsgID),
		}
		tg_bot.Bot.Request(deleteConfig)
	}

	return nil
}

func OrderPaidPrice(order models.Order, options ...interface{}) decimal.Decimal {
	// 查已支付，没有更新，只是借用tx事务进行查询，因此不用commit
	tx := db.DB
	for _, value := range options {
		if opt, ok := value.(*gorm.DB); ok {
			tx = opt
		}
	}

	var result decimal.Decimal
	var transfers []models.Transfer
	tx.Where("order_id = ?", order.ID).Find(&transfers)

	for _, transfer := range transfers {
		result = result.Add(transfer.Price)
	}

	return result
}

func OrderCallbackMultiple(successOrderIDs []uuid.UUID) error {
	var successOrders []models.Order
	db.DB.Preload("Product").Preload("ProductItem").Where("id in ?", successOrderIDs).Find(&successOrders)
	for _, successOrder := range successOrders {
		// 发消息
		SendOrderCallBack(successOrder.TGChatID, int(successOrder.TGMsgID), successOrder, successOrder.Product, successOrder.ProductItem)
	}

	return nil
}
func SetOrderTGMsgID(orderID uuid.UUID, tgMsgID int64) {
	db.DB.Model(&models.Order{}).Where("id = ?", orderID).Update("tg_msg_id", tgMsgID)
}

func GetOrderIncomeByTimestampRange(startTimestamp int64, endTimestamp int64) (decimal.Decimal, error) {
	var orders []models.Order
	filterParams := make(map[string]interface{})
	filterParams["timestamp_range"] = fmt.Sprintf("%d,%d", startTimestamp, endTimestamp)
	query := models.ApplyFilters(db.DB, filterParams).Where("status=1")
	if result := query.Find(&orders); result.Error != nil {
		return decimal.Decimal{}, errors.New("获取订单错误")
	}

	orderPriceSum := decimal.Zero
	for _, order := range orders {
		convertedPrice, err := config.ConvertCurrencyPrice(order.Price, config.Currency(order.Currency), config.CNY)
		if err != nil {
			return decimal.Decimal{}, errors.New("获取汇率失败")
		}
		orderPriceSum = orderPriceSum.Add(convertedPrice)
	}

	return orderPriceSum, nil
}
func GetPaidOrdersByCustomer(tgChatID int64) ([]models.Order, error) {
	var orders []models.Order
	if result := db.DB.Preload("Product").Preload("ProductItem").Where("status = 1 and tg_chat_id = ?", tgChatID).Order("create_time desc").Limit(10).Find(&orders); result.Error != nil {
		return orders, errors.New("获取订单错误")
	}

	return orders, nil
}
func GetPaidOrderByCustomerByID(orderID uuid.UUID) (models.Order, error) {
	var orders models.Order
	if result := db.DB.Preload("Product").Preload("ProductItem").Where("id = ?", orderID).Find(&orders); result.Error != nil {
		return orders, errors.New("获取订单错误")
	} else if result.RowsAffected == 0 {
		return orders, errors.New("没有该订单")
	}
	return orders, nil
}
func SendOrderCallBack(chatID int64, toDeleteMsgID int, order models.Order, product models.Product, productItemList []models.ProductItem) {
	if len(productItemList) > 0 {
		itemContent := productItemList[0].Content
		switch {
		case checkFileProductItem(itemContent):
			SendOrderCallBackFileProduct(chatID, toDeleteMsgID, order, product, productItemList)
			return
		case checkChatInviteProductItem(itemContent):
			SendOrderCallBackChatInviteLinkProduct(chatID, toDeleteMsgID, order, product, productItemList)
			return
		}
	}
	if len(productItemList) > 10 {
		SendOrderCallBackMore(chatID, toDeleteMsgID, order, product, productItemList)
		return
	}
	buyNum := len(productItemList)
	if buyNum == 0 {
		buyNum = int(order.BuyNum)
	}
	msgText := config.OrderCallbackMsg(map[string]interface{}{
		"Order":   order,
		"Product": product,
		"BuyNum":  buyNum,
		// "ProductItem": productItem,
		"ProductItemList": productItemList,
	})
	//newMsg := tgbotapi.NewEditMessageText(chatID, msgID, msgText)
	msg := tgbotapi.NewMessage(chatID, msgText)
	msg.DisableWebPagePreview = true
	msg.ParseMode = tgbotapi.ModeHTML
	var retMsgID int
	if ret, err := tg_bot.Bot.Send(msg); err != nil {
	} else {
		retMsgID = ret.MessageID
	}

	if toDeleteMsgID != 0 {
		tg_bot.DeleteMsg(chatID, toDeleteMsgID)
	}

	SendOrderWithBalanceHistoryCallBack(chatID, retMsgID, order)
}

func SendOrderCallBackMore(chatID int64, toDeleteMsgID int, order models.Order, product models.Product, productItemList []models.ProductItem) {
	msgText := config.OrderCallbackMoreMsg(map[string]interface{}{
		"Order":   order,
		"Product": product,
		"BuyNum":  len(productItemList),
		// "ProductItem": productItem,
		// "ProductItemList": productItemList,
	})

	var productItems []string
	for _, v := range productItemList {
		productItems = append(productItems, v.Content)
	}

	var retMsgID int
	if len(productItems) > 0 {
		productItemsStr := strings.Join(productItems, "\n")
		// productItemsStr = product.Name + "\n" + productItemsStr
		fileBody := []byte(productItemsStr)
		fileBytes := tgbotapi.FileBytes{Name: "all_" + order.ID.String() + ".txt", Bytes: fileBody}
		newMsg := tgbotapi.NewDocument(chatID, fileBytes)
		newMsg.Caption = msgText
		newMsg.ParseMode = tgbotapi.ModeHTML
		if result, err := tg_bot.Bot.Send(newMsg); err != nil {
			log.Errorf("Bot.Send(%+v):%v", newMsg, err)
		} else {
			log.Errorf("result: %+v", result)
		}
	} else {
		msgText = msgText + "\n" + "未找到商品项目:请联系客服处理"
		newMsg := tgbotapi.NewMessage(chatID, msgText)
		newMsg.DisableWebPagePreview = true
		newMsg.ParseMode = tgbotapi.ModeHTML
		if result, err := tg_bot.Bot.Send(newMsg); err != nil {
			log.Errorf("Bot.Send(%+v):%v", newMsg, err)
		} else {
			log.Debugf("result: %+v", result)
			retMsgID = result.MessageID
		}
	}

	// 删除原消息
	if toDeleteMsgID != 0 {
		tg_bot.DeleteMsg(chatID, toDeleteMsgID)
	}
	SendOrderWithBalanceHistoryCallBack(chatID, retMsgID, order)
}

// 余额充值回调
func SendOrderWithBalanceHistoryCallBack(chatID int64, toDeleteMsgID int, order models.Order) {
	if !config.IsBalanceProduct(order.ProductID.String()) {
		return
	}
	balanceHistory := models.NewBalanceHistory(
		order.Price,
		order.BaseCurrency,
		order.BaseCurrencyPrice,
		&order.ID,
		chatID,
		order.TGUsername,
	)
	balanceHistory.BalanceNum = order.BuyNum
	if count := GetBalanceHistoryCount(balanceHistory); count > 0 {
		return
	}
	var balanceResult string
	var balanceNum int64
	if err := CreateBalanceHistory(balanceHistory); err != nil {
		// 添加兑换记录失败
		balanceResult = "添加充值记录失败: " + err.Error()
	} else {
		balanceResult = "添加充值记录成功"
		// 兑换积分 product.Price.Mul(decimal.NewFromInt(buyNum))
		// balanceNum = float64(order.BalanceNum)
		balanceNum = order.BuyNum
		if count := GetUserBalanceCount(chatID); count > 0 {
			if err := UpdateUserBalance(chatID, balanceNum); err != nil {
				balanceResult = "充值失败: " + err.Error()
			} else {
				balanceResult = "充值成功"
			}
		} else {
			if err := CreateUserBalance(chatID, order.TGUsername, "", "", balanceNum, "", fmt.Sprintf("%d", chatID)); err != nil {
				balanceResult = "新增充值失败: " + err.Error()
			} else {
				balanceResult = "新增充值成功"
			}
		}
	}

	msgText := config.OrderCallbackBalanceMsg(map[string]interface{}{
		"Order":         order,
		"Balance":       balanceNum,
		"BalanceResult": balanceResult,
	})
	newMsg := tgbotapi.NewMessage(chatID, msgText)
	tg_bot.Bot.Send(newMsg)

	// 删除原消息
	if toDeleteMsgID != 0 {
		tg_bot.DeleteMsg(chatID, toDeleteMsgID)
	}
}

// Apply discount
// 应用折扣 满金额折扣
func ApplyDiscount(orgPrice decimal.Decimal) (decimal.Decimal, string) {
	finalPrice := orgPrice
	note := ""
	if !config.GetSiteConfig().EnableDiscount {
		return finalPrice, ""
	}
	// 解析折扣规则
	discountRules := config.GetSiteConfig().DiscountRules
	discountRuleList := strings.Split(discountRules, "|")
	discountMap := make(map[int]float64)
	var discountS []int //从小到大放置
	for _, v := range discountRuleList {
		item := v
		itemS := strings.Split(item, ":")
		if len(itemS) != 2 {
			continue
		} else {
			k, err := strconv.Atoi(itemS[0])
			if err != nil || k <= 0 {
				continue
			}
			v, err := strconv.ParseFloat(itemS[1], 64)
			if err != nil || v <= 0.1 {
				continue
			}
			discountMap[k] = v
			discountS = append(discountS, k)
		}
	}

	// 排序 discountS
	sort.Ints(discountS)
	for _, v := range discountS {
		vc := decimal.NewFromInt(int64(v))
		if orgPrice.Cmp(vc) >= 0 {
			discount := decimal.NewFromFloat(discountMap[v])
			finalPrice = orgPrice.Mul(discount).Div(decimal.NewFromFloat(10.0))
			note = fmt.Sprintf("本次订单已参与 [满%v,%v折] 优惠活动", v, discountMap[v])
			continue
		}
	}
	return finalPrice.Round(2), note
}

// 应用折扣 满n件折扣
func ApplyDiscountNum(buyNum int64, orgPrice decimal.Decimal) (decimal.Decimal, string) {
	finalPrice := orgPrice
	note := ""
	if !config.GetSiteConfig().EnableDiscountNum {
		return finalPrice, ""
	}
	// 解析折扣规则
	discountRules := config.GetSiteConfig().DiscountNumRules
	discountRuleList := strings.Split(discountRules, "|")
	discountMap := make(map[int]float64)
	var discountS []int //从小到大放置
	for _, v := range discountRuleList {
		item := v
		itemS := strings.Split(item, ":")
		if len(itemS) != 2 {
			continue
		} else {
			k, err := strconv.Atoi(itemS[0]) //strconv.ParseInt(itemS[0], 10, 64)
			if err != nil || k <= 0 {
				continue
			}
			v, err := strconv.ParseFloat(itemS[1], 64)
			if err != nil || v <= 0.1 {
				continue
			}
			discountMap[k] = v
			discountS = append(discountS, k)
		}
	}

	// 排序 discountS
	sort.Ints(discountS)
	for _, v := range discountS {
		vc := v
		if buyNum >= int64(vc) {
			discount := decimal.NewFromFloat(discountMap[v])
			finalPrice = orgPrice.Mul(discount).Div(decimal.NewFromFloat(10.0))
			note = fmt.Sprintf("本次订单已参与 [满%v件,%v折] 优惠活动", v, discountMap[v])
			continue
		}
	}
	return finalPrice.Round(2), note
}

func SendOrderCallBackFileProduct(chatID int64, toDeleteMsgID int, order models.Order, product models.Product, productItemList []models.ProductItem) {
	msgText := config.OrderCallbackMoreMsg(map[string]interface{}{
		"Order":   order,
		"Product": product,
		"BuyNum":  len(productItemList),
		// "ProductItem": productItem,
		// "ProductItemList": productItemList,
	})
	fileProductDir := config.GetSiteConfig().FileProductDir
	if len(fileProductDir) == 0 {
		fileProductDir = "upload"
	}
	var productItems []string
	for _, v := range productItemList {
		vContent := v.Content
		item := strings.TrimPrefix(vContent, FILE_PRODUCT_ITEM_Prefix)
		item = filepath.Join(".", fileProductDir, item)
		productItems = append(productItems, item)
		log.Debugf("productItem file: %s", item)
	}
	var fileBody []byte
	var err error
	if len(productItems) > 0 {
		fileBody, err = utils.ZipFilesToByte(productItems)
		if err != nil {
			msgText = msgText + "\n" + fmt.Sprintf("文件压缩打包出错: %v", err)
			log.Errorf("productItems(%+v):%v", productItems, err)
		} else {
			log.Errorf("len(%+v):%d", productItems, len(fileBody))
		}
	}

	// var retMsgID int
	if len(fileBody) > 0 {
		fileBytes := tgbotapi.FileBytes{Name: "file_" + order.ID.String() + ".zip", Bytes: fileBody}
		newMsg := tgbotapi.NewDocument(chatID, fileBytes)
		newMsg.Caption = msgText
		newMsg.ParseMode = tgbotapi.ModeHTML
		if result, err := tg_bot.Bot.Send(newMsg); err != nil {
			log.Errorf("Bot.Send(%+v):%v", newMsg, err)
		} else {
			log.Debugf("result: %+v", result)
		}
	} else {
		msgText = msgText + "\n" + "请联系客服处理"
		newMsg := tgbotapi.NewMessage(chatID, msgText)
		newMsg.DisableWebPagePreview = true
		newMsg.ParseMode = tgbotapi.ModeHTML
		if result, err := tg_bot.Bot.Send(newMsg); err != nil {
			log.Errorf("Bot.Send(%+v):%v", newMsg, err)
		} else {
			log.Debugf("result: %+v", result)
			// retMsgID = result.MessageID
		}
	}

	// 删除原消息
	if toDeleteMsgID != 0 {
		tg_bot.DeleteMsg(chatID, toDeleteMsgID)
	}
	// SendOrderWithBalanceHistoryCallBack(chatID, retMsgID, order)
}

func SendOrderCallBackChatInviteLinkProduct(chatID int64, toDeleteMsgID int, order models.Order, product models.Product, productItemList []models.ProductItem) {

	var productItems []string
	for _, v := range productItemList {
		if GetChatInviteLinkHistoryCount(v.ID.String()) > 0 {
			item, err := GetChatInviteLink(v.ID.String())
			if err != nil {
				log.Errorf("GetChatInviteLink(%s): %v", v.ID.String(), err)
				continue
			}
			productItems = append(productItems, item)
			log.Debugf("productItem file: %s", item)
		} else {
			vContent := v.Content
			itemStr := strings.TrimPrefix(vContent, CHATINVITELINK_PRODUCT_ITEM_Prefix)
			item, err := createInvite(itemStr, 1)
			if err != nil {
				log.Errorf("createInvite(%s): %v", vContent, err)
				continue
			}
			productItems = append(productItems, item)
			chatInviteLinkHistory := models.NewChatInviteLinkHistory(v.ID, item)
			if err := CreateChatInviteLinkHistory(chatInviteLinkHistory); err != nil {
				log.Errorf("CreateChatInviteLinkHistory(%+v): %v", chatInviteLinkHistory, err)
			}
			log.Debugf("productItem file: %s", item)
		}

	}

	msgText := config.OrderCallbackChatInviteLinkMsg(map[string]interface{}{
		"Order":   order,
		"Product": product,
		"BuyNum":  len(productItemList),
		// "ChatInviterLink": chatInviterLink,//<a href="{{.ChatInviterLink}}">一次性邀请入群地址</a>
		"ProductItemList": productItems,
	})

	newMsg := tgbotapi.NewMessage(chatID, msgText)
	newMsg.DisableWebPagePreview = true
	newMsg.ParseMode = tgbotapi.ModeHTML
	if result, err := tg_bot.Bot.Send(newMsg); err != nil {
		log.Errorf("Bot.Send(%+v):%v", newMsg, err)
	} else {
		log.Debugf("result: %+v", result)
	}

	// 删除原消息
	if toDeleteMsgID != 0 {
		tg_bot.DeleteMsg(chatID, toDeleteMsgID)
	}
}

// true 文件商品，false 非文件商品 或未开启文件商品
func checkFileProductItem(productItemContent string) bool {
	if !config.GetSiteConfig().EnableFileProduct {
		return false
	}
	if strings.HasPrefix(productItemContent, FILE_PRODUCT_ITEM_Prefix) {
		return true
	}
	return false
}

// true 会话邀请商品，false 非会话邀请商品 或未开启会话邀请商品
func checkChatInviteProductItem(productItemContent string) bool {
	if !config.GetSiteConfig().EnableChatInviteProduct {
		return false
	}
	if strings.HasPrefix(productItemContent, CHATINVITELINK_PRODUCT_ITEM_Prefix) {
		return true
	}
	return false
}
