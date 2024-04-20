package tg_handler

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/umfaka/tgfaka/internal/constvar"

	"github.com/umfaka/tgfaka/internal/exts/config"
	"github.com/umfaka/tgfaka/internal/exts/db"
	"github.com/umfaka/tgfaka/internal/exts/tg_bot"
	"github.com/umfaka/tgfaka/internal/log"
	"github.com/umfaka/tgfaka/internal/models"
	"github.com/umfaka/tgfaka/internal/services"
	"github.com/umfaka/tgfaka/internal/utils/functions"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

// 修改消息，文本为空则忽视，markup为空则新发一个消息
var ProductListPagePrefix = "p_l_p_"
var ProductDetailPrefix = "p_d_"
var PayOrderPrefix = "p_o_"           // 支付方式
var PayOrderByBalancePrefix = "p_ob_" // 支付方式 余额支付
var GetPaidOrderResultPrefix = "g_p_o_r_"
var BuyNumPrefix = "b_n_"       //购买数量
var DownloadItemPrefix = "d_i_" //下载未售出列表

// 开始指令
func StartCommand(update tgbotapi.Update) {
	if !update.Message.Chat.IsPrivate() {
		return
	}

	msgText := config.WelcomeMsg(map[string]interface{}{})
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
	// 将键盘添加到消息中
	if config.GetSiteConfig().EnableTGBotFooterMenu {
		rows := welcomeStartFooterRows()
		if len(rows) > 0 {
			keyboard := tgbotapi.NewReplyKeyboard(rows...)
			msg.ReplyMarkup = keyboard
		} else {
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(productListMsgFooterRow())
		}
	} else {
		rows := welcomeStartRows()
		if len(rows) > 0 {
			keyboard := tgbotapi.NewInlineKeyboardMarkup(rows...)
			msg.ReplyMarkup = keyboard
		} else {
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(productListMsgRow(), deleteMsgRow())
		}
	}

	tg_bot.Bot.Send(msg)
	go func() {
		userid := update.Message.Chat.ID
		username := update.Message.Chat.UserName
		firstName := update.Message.Chat.FirstName
		lastName := update.Message.Chat.LastName
		var startInviter int64
		startInviteCode := ""

		if services.GetUserBalanceCount(update.Message.Chat.ID) > 0 {
			if err := services.UpdateUserBalanceEx(userid, username, firstName, lastName); err != nil {
				log.Errorf("services.UpdateUserBalanceEx(%d,%s): %v", userid, username)
			}
			return
		}

		args := update.Message.CommandArguments()
		if config.GetSiteConfig().EnableInviteRewards && len(args) > 0 {
			inviteRewardsBalanceNum := config.GetSiteConfig().InviteRewardsBalanceNum
			parameters := strings.Fields(args)
			for _, v := range parameters {
				if len(v) > 0 {
					if vi, err := strconv.ParseInt(v, 10, 64); err != nil {
					} else {
						startInviteCode = v
						// 检查邀请码有效性
						if services.GetUserBalanceCountByInviterCode(startInviteCode) > 0 {
							// 获取邀请者信息
							if uTmp, err := services.GetUserBalanceByInviterCode(startInviteCode); err != nil {
								log.Errorf("services.GetUserBalanceByInviterCode(%s): %v", vi, err)
							} else {
								startInviter = uTmp.Userid
								if err := services.UpdateUserBalance(uTmp.Userid, inviteRewardsBalanceNum); err != nil {
									log.Errorf("services.UpdateUserBalance(%d,%d): %v", uTmp.Userid, inviteRewardsBalanceNum, err)
								} else {
									msgText := config.InviteRewardsBalanceMsg(map[string]interface{}{
										"InviteRewardsBalanceNum": inviteRewardsBalanceNum,
										"Invitees":                fmt.Sprintf("%s %s", firstName, lastName),
										"BalanceCurrency":         config.GetSiteConfig().BalanceCurrency,
									})
									msg := tgbotapi.NewMessage(uTmp.Userid, msgText)
									tg_bot.Bot.Send(msg)

								}
							}
							break
						}
					}
				}
			}
		}

		if err := services.CreateUserBalance(
			userid, username, firstName, lastName, 0,
			fmt.Sprintf("%d", startInviter), fmt.Sprintf("%d", userid),
		); err != nil {
			log.Errorf("CreateUserBalance(%d,%s): %v", userid, username, err)
		}
	}()
}

// 登录指令
func LoginCommand(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	expireTime := time.Second * 60

	if chatID != config.GetSiteConfig().AdminTGID {
		return
	}

	token := services.SetAdminLoginUrlSession(expireTime)
	loginUrl := fmt.Sprintf("%s/api/admin/token_login?token=%s", config.GetSiteConfig().Host, token)

	text := fmt.Sprintf("<a href=\"%s\">一次性登录地址(60秒内有效)</a>", loginUrl)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	msg.DisableWebPagePreview = true
	sentMsg, err := tg_bot.Bot.Send(msg)
	if err != nil {
		//log.Println("Error sending message:", err)
	}

	//清除login session
	services.ClearAdminLoginTokenSession()

	//定时删除
	go func(ChatID int64, messageID int) {
		time.Sleep(expireTime)
		tg_bot.DeleteMsg(chatID, messageID)
	}(chatID, sentMsg.MessageID)

}

// 登录指令
func SendToAllCommand(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	if chatID != config.GetSiteConfig().AdminTGID {
		return
	}

	msgText := update.Message.CommandArguments()

	go func() {
		userList, err := services.GetUserBalanceAll()
		if err != nil {
			log.Errorf("GetUserBalanceAll(): %v", err)
			return
		}
		retMsgID := 0
		retMsg, err := tg_bot.Bot.Send(
			tgbotapi.NewMessage(chatID,
				fmt.Sprintf("准备发送以下消息内容(%d):\n%s", len(userList), msgText),
			),
		)
		if err != nil {
			log.Errorf("发送反馈结果失败: %v", err)
		} else {
			retMsgID = retMsg.MessageID
		}
		retText := ""
		for _, v := range userList {
			vc := v
			msg := tgbotapi.NewMessage(vc.Userid, msgText)
			if _, err := tg_bot.Bot.Send(msg); err != nil {
				retText = retText + "\n" +
					fmt.Sprintf(
						"发送给 ID(%d) @%s %s %s 的消息失败: %v",
						vc.Userid, vc.Username, vc.FirstName, vc.LastName, err,
					)
			} else {
				retText = retText + "\n" + fmt.Sprintf("ID(%d) 发送成功", vc.Userid)
			}
			time.Sleep(3 * time.Second)
		}

		if retMsgID != 0 && len(retText) > 0 {
			// time.Sleep(5 * time.Second)
			expireTime := time.Second * 60
			retMsg := tgbotapi.NewEditMessageText(chatID, retMsgID, retText)
			if finalMsg, err := tg_bot.Bot.Send(retMsg); err != nil {
				log.Errorf("发送反馈结果失败: %v", err)
				return
			} else {
				//定时删除
				go func(ChatID int64, messageID int) {
					time.Sleep(expireTime)
					tg_bot.DeleteMsg(chatID, messageID)
				}(chatID, finalMsg.MessageID)
			}
		}
	}()

}

// 登录指令
func TopUpCommand(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	if chatID != config.GetSiteConfig().AdminTGID {
		return
	}

	argList := strings.Fields(update.Message.CommandArguments())
	if len(argList) != 2 {
		newMsg := tgbotapi.NewMessage(chatID, "失败,参考指令格式 /topup 用户id 金额")
		tg_bot.Bot.Send(newMsg)
		return
	}
	retMsg := ""
	userID, err := strconv.ParseInt(argList[0], 10, 64)
	if err != nil {
		retMsg = fmt.Sprintf("失败,解析用户id失败: %v", argList[0])
		newMsg := tgbotapi.NewMessage(chatID, retMsg)
		tg_bot.Bot.Send(newMsg)
		return
	}
	balance, err := strconv.ParseInt(argList[1], 10, 64)
	if err != nil {
		retMsg = fmt.Sprintf("失败,解析充值金额失败: %v", argList[1])
		newMsg := tgbotapi.NewMessage(chatID, retMsg)
		tg_bot.Bot.Send(newMsg)
		return
	}
	if err := services.UpdateUserBalance(userID, balance); err != nil {
		retMsg = fmt.Sprintf("失败,充值失败: %v", err)
		newMsg := tgbotapi.NewMessage(chatID, retMsg)
		tg_bot.Bot.Send(newMsg)
		return
	} else {
		retMsg = fmt.Sprintf("充值成功: 用户(%d) 金额(%d) ", userID, balance)
		newMsg := tgbotapi.NewMessage(chatID, retMsg)
		tg_bot.Bot.Send(newMsg)
		return
	}
}

// 商品列表
func ProductList(update tgbotapi.Update) {
	currentPage := 1
	if update.CallbackQuery != nil {
		callbackData := update.CallbackQuery.Data
		value, err := strconv.Atoi(strings.TrimPrefix(callbackData, ProductListPagePrefix))
		if err == nil {
			currentPage = value
		}
	}

	pagination := services.Pagination{Limit: 10, Page: currentPage}
	err := services.GetProductsByCustomer(&pagination)
	if err != nil {
		tg_bot.Bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "获取失败"))
	}

	var products []models.Product
	for _, item := range pagination.Items {
		products = append(products, item.(models.Product))
	}

	msgText := config.ProductListMsg(map[string]interface{}{})
	rows := append(paginationToRows(pagination), deleteMsgRow())
	replyMarkup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	if update.Message != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		msg.ReplyMarkup = replyMarkup
		tg_bot.Bot.Send(msg)
	} else {
		msg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, msgText)
		replyMarkup := tgbotapi.NewInlineKeyboardMarkup(rows...)
		msg.ReplyMarkup = &replyMarkup
		tg_bot.Bot.Send(msg)
	}
}

// 商品详情
func ProductDetail(update tgbotapi.Update) {
	senderChatID := update.CallbackQuery.Message.Chat.ID
	callbackData := update.CallbackQuery.Data
	// productID, err := uuid.Parse(strings.TrimPrefix(callbackData, ProductDetailPrefix))
	// if err != nil {
	// 	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "id错误")
	// 	tg_bot.Bot.Request(callback)
	// 	return
	// }
	log.Debug("callbackData: " + callbackData)
	value := strings.TrimPrefix(callbackData, ProductDetailPrefix)

	parts := strings.Split(value, "_")
	if len(parts) != 2 {
		log.Debug("callbackData(" + value + ") len(parts): " + strconv.Itoa(len(parts)))
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "参数长度错误")
		tg_bot.Bot.Request(callback)
		return
	}
	productIDString := parts[0]
	buyNumOptionString := parts[1]

	productID, err := uuid.Parse(productIDString)
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "商品ID错误")
		tg_bot.Bot.Request(callback)
		return
	}

	// buyNum, err := strconv.Atoi(buyNumOptionString)
	buyNum, err := strconv.ParseInt(buyNumOptionString, 10, 64)
	if err != nil || buyNum <= 0 {
		buyNum = 1
	}

	product, err := services.GetProductByIDByCustomer(productID)
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "商品不存在")
		tg_bot.Bot.Request(callback)
		return
	}
	if buyNum > product.InStockCount && !config.IsBalanceProduct(product.ID.String()) {
		buyNum = product.InStockCount - 1
	}
	if config.IsBalanceProduct(product.ID.String()) {
		product.InStockCount = 888
	} else {
		if buyNum > product.InStockCount {
			buyNum = product.InStockCount - 1
		}
	}

	msgText := config.ProductDetailMsg(map[string]interface{}{
		"Product": product,
		"BuyNum":  buyNum,
	})
	newMsg := tgbotapi.NewEditMessageText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, msgText)

	goBackRow := GoBackRow(ProductListPagePrefix + "1")
	paymentRow := paymentSelectRow(product.ID, buyNum)
	if len(paymentRow) == 0 {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "没有设置支付方式")
		tg_bot.Bot.Request(callback)
		return
	}
	closeRow := deleteMsgRow()
	buyNumRow := buyNumButtonRow(productID, buyNum)
	markupPtr := tgbotapi.NewInlineKeyboardMarkup(buyNumRow, paymentRow, goBackRow, closeRow)
	if isInternalBuy := config.IsInternal(senderChatID); isInternalBuy && config.GetSiteConfig().EnableInternalDownloadItems {
		downloadRow := downloadItemRow(productID)
		markupPtr = tgbotapi.NewInlineKeyboardMarkup(buyNumRow, paymentRow, downloadRow, goBackRow, closeRow)
	}
	newMsg.ReplyMarkup = &markupPtr
	tg_bot.Bot.Send(newMsg)
}

// 商品详情 余额充值
func ProductBalanceDetail(update tgbotapi.Update) {
	senderChatID := update.Message.Chat.ID
	productIDString := config.GetSiteConfig().BalanceProductUUID
	if len(productIDString) == 0 {
		log.Errorf("商品ID(%s)未设置", productIDString)
		return
	}

	productID, err := uuid.Parse(productIDString)
	if err != nil {
		log.Errorf("商品ID(%s)错误", productIDString)
		return
	}
	var buyNum int64
	buyNum = 1

	product, err := services.GetProductByIDByCustomer(productID)
	if err != nil {
		log.Errorf("商品(%s)不存在", productIDString)
		return
	}
	product.InStockCount = 888

	msgText := config.ProductDetailMsg(map[string]interface{}{
		"Product": product,
		"BuyNum":  buyNum,
	})
	// newMsg := tgbotapi.NewEditMessageText(update.Message.Chat.ID, update.Message.MessageID, msgText)
	newMsg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
	goBackRow := GoBackRow(ProductListPagePrefix + "1")
	paymentRow := paymentSelectRow(product.ID, buyNum)
	if len(paymentRow) == 0 {
		log.Error("没有设置支付方式")
		return
	}
	closeRow := deleteMsgRow()
	buyNumRow := buyNumButtonRow(productID, buyNum)
	markupPtr := tgbotapi.NewInlineKeyboardMarkup(buyNumRow, paymentRow, goBackRow, closeRow)
	if isInternalBuy := config.IsInternal(senderChatID); isInternalBuy && config.GetSiteConfig().EnableInternalDownloadItems {
		downloadRow := downloadItemRow(productID)
		markupPtr = tgbotapi.NewInlineKeyboardMarkup(buyNumRow, paymentRow, downloadRow, goBackRow, closeRow)
	}
	newMsg.ReplyMarkup = &markupPtr
	tg_bot.Bot.Send(newMsg)
}

// 关闭按钮
func CallbackDeleteMsg(update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	messageID := update.CallbackQuery.Message.MessageID
	tg_bot.DeleteMsg(chatID, messageID)
}

// 支付 生成订单
func PayOrder(update tgbotapi.Update) {
	senderChatID := update.CallbackQuery.Message.Chat.ID
	senderMsgID := update.CallbackQuery.Message.MessageID
	senderUsername := update.CallbackQuery.From.UserName

	callbackData := update.CallbackQuery.Data
	value := strings.TrimPrefix(callbackData, PayOrderPrefix)
	parts := strings.Split(value, "_")
	if len(parts) != 3 {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "参数长度错误")
		tg_bot.Bot.Request(callback)
		return
	}

	productIDString := parts[0]
	paymentOptionString := parts[1]
	buyNumOptionString := parts[2]

	buyNum, err := strconv.ParseInt(buyNumOptionString, 10, 64)
	if err != nil || buyNum <= 0 {
		buyNum = 1
	}

	if !config.IsPaymentEnable(paymentOptionString) {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "支付方式不存在")
		tg_bot.Bot.Request(callback)
		return
	}
	var paymentOption *config.PaymentOption
	paymentOption, err = config.ParsePaymentMethod(paymentOptionString)
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "支付方式错误")
		tg_bot.Bot.Request(callback)
		return
	}
	productID, err := uuid.Parse(productIDString)
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "商品ID错误")
		tg_bot.Bot.Request(callback)
		return
	}
	product, err := services.GetProductByIDByCustomer(productID)
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "商品不存在")
		tg_bot.Bot.Request(callback)
		return
	}

	// 一個用戶只能有一個訂單，由於放在前面釋放，先釋放后创建
	var toReleaseOrderIDs []uuid.UUID
	if result := db.DB.Model(&models.Order{}).Where("status = 0 and tg_chat_id = ?", senderChatID).Pluck("id", &toReleaseOrderIDs); result.RowsAffected > 0 {
		services.ReleaseOrders(toReleaseOrderIDs)
	}

	// 创建订单
	order, err := services.CreateOrder(
		paymentOption.Currency, string(paymentOption.Network), product, senderChatID, senderUsername, buyNum,
	)
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, err.Error())
		tg_bot.Bot.Request(callback)
		return
	}

	// 生成图片
	qrImageBytes, err := functions.GenerateQrCodeBytes(order.WalletAddress)
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, err.Error())
		tg_bot.Bot.Request(callback)
		return
	}

	photoFileBytes := tgbotapi.FileBytes{Name: "qr.png", Bytes: qrImageBytes}
	photoMsg := tgbotapi.NewPhoto(senderChatID, photoFileBytes)
	// photoMsg.Caption = config.PayOrderMsg(map[string]interface{}{
	// 	"Order": order,
	// })
	msgText := ""
	orderNoteTitle := ""
	if len(order.Note) > 0 {
		orderNoteTitle = "提示:"
	}
	msgText = config.PayOrderMsg(map[string]interface{}{
		"Order":          order,
		"OrderNoteTitle": orderNoteTitle,
		"OrderNote":      order.Note,
	})
	photoMsg.Caption = msgText

	photoMsg.ParseMode = tgbotapi.ModeHTML

	result, _ := tg_bot.Bot.Send(photoMsg)

	// 删除原消息
	tg_bot.DeleteMsg(senderChatID, senderMsgID)

	// 给订单设置msgID用于删除
	services.SetOrderTGMsgID(order.ID, int64(result.MessageID))

	if isInternalBuy := config.IsInternal(senderChatID); isInternalBuy && order.Status == 1 {
		go func() {
			newMsg := tgbotapi.NewMessage(senderChatID, "内部购买:无需付款")
			tg_bot.Bot.Send(newMsg)
			time.Sleep(5 * time.Second)
			orderID := order.ID
			orderTmp, err := services.GetPaidOrderByCustomerByID(orderID)
			if err != nil {
				msg := tgbotapi.NewMessage(senderChatID, err.Error())
				tg_bot.Bot.Send(msg)
				return
			}
			// services.OrderCallbackMultiple([]uuid.UUID{order.ID})
			services.SendOrderCallBack(senderChatID, senderMsgID, orderTmp, orderTmp.Product, orderTmp.ProductItem)
		}()
	}
}

// 支付 生成订单 余额支付
func PayOrderByBalance(update tgbotapi.Update) {
	senderChatID := update.CallbackQuery.Message.Chat.ID
	senderMsgID := update.CallbackQuery.Message.MessageID
	senderUsername := update.CallbackQuery.From.UserName

	callbackData := update.CallbackQuery.Data
	value := strings.TrimPrefix(callbackData, PayOrderByBalancePrefix)
	parts := strings.Split(value, "_")
	if len(parts) != 2 {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "参数长度错误")
		tg_bot.Bot.Request(callback)
		return
	}

	productIDString := parts[0]
	buyNumOptionString := parts[1]

	buyNum, err := strconv.ParseInt(buyNumOptionString, 10, 64)
	if err != nil || buyNum <= 0 {
		buyNum = 1
	}

	productID, err := uuid.Parse(productIDString)
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "商品ID错误")
		tg_bot.Bot.Request(callback)
		return
	}
	product, err := services.GetProductByIDByCustomer(productID)
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "商品不存在")
		tg_bot.Bot.Request(callback)
		return
	}

	paymentOptionString := config.GetSiteConfig().BalanceCurrency
	if !config.IsPaymentEnable(paymentOptionString) {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "支付方式不存在")
		tg_bot.Bot.Request(callback)
		return
	}
	var paymentOption *config.PaymentOption
	paymentOption, err = config.ParsePaymentMethod(paymentOptionString)
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "支付方式错误")
		tg_bot.Bot.Request(callback)
		return
	}

	// 一個用戶只能有一個訂單，由於放在前面釋放，先釋放后创建
	var toReleaseOrderIDs []uuid.UUID
	if result := db.DB.Model(&models.Order{}).Where("status = 0 and tg_chat_id = ?", senderChatID).Pluck("id", &toReleaseOrderIDs); result.RowsAffected > 0 {
		services.ReleaseOrders(toReleaseOrderIDs)
	}

	// 创建订单
	order, err := services.CreateOrderByBalance(
		paymentOption.Currency, string(paymentOption.Network),
		product, senderChatID, senderUsername, buyNum,
	)
	if err != nil {
		log.Errorf(
			"services.CreateOrderByBalance(%+v,%d,%s,%d): %v",
			product.ID, senderChatID, senderUsername, buyNum, err,
		)
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, err.Error())
		tg_bot.Bot.Request(callback)
		return
	}
	msgText := ""
	orderNoteTitle := ""
	if len(order.Note) > 0 {
		orderNoteTitle = "提示:"
	}
	msgText = config.PayOrderBalanceMsg(map[string]interface{}{
		"Order":          order,
		"OrderNoteTitle": orderNoteTitle,
		"OrderNote":      order.Note,
	})
	newMsg := tgbotapi.NewMessage(senderChatID, msgText)
	result, _ := tg_bot.Bot.Send(newMsg)

	// 删除原消息
	tg_bot.DeleteMsg(senderChatID, senderMsgID)

	// 给订单设置msgID用于删除
	services.SetOrderTGMsgID(order.ID, int64(result.MessageID))

	if isInternalBuy := config.IsInternal(senderChatID); isInternalBuy && order.Status == 1 {
		go func() {
			newMsg := tgbotapi.NewMessage(senderChatID, "内部购买:无需付款")
			tg_bot.Bot.Send(newMsg)
			time.Sleep(5 * time.Second)
			orderID := order.ID
			orderTmp, err := services.GetPaidOrderByCustomerByID(orderID)
			if err != nil {
				msg := tgbotapi.NewMessage(senderChatID, err.Error())
				tg_bot.Bot.Send(msg)
				return
			}
			// services.OrderCallbackMultiple([]uuid.UUID{order.ID})
			services.SendOrderCallBack(senderChatID, senderMsgID, orderTmp, orderTmp.Product, orderTmp.ProductItem)
		}()
	} else if order.Status == 1 {
		go func() {
			newMsg := tgbotapi.NewMessage(senderChatID, "余额支付:付款完成")
			tg_bot.Bot.Send(newMsg)
			time.Sleep(5 * time.Second)
			orderID := order.ID
			orderTmp, err := services.GetPaidOrderByCustomerByID(orderID)
			if err != nil {
				msg := tgbotapi.NewMessage(senderChatID, err.Error())
				tg_bot.Bot.Send(msg)
				return
			}
			// services.OrderCallbackMultiple([]uuid.UUID{order.ID})
			services.SendOrderCallBack(senderChatID, senderMsgID, orderTmp, orderTmp.Product, orderTmp.ProductItem)
		}()
	}
}

// 订单列表
func PaidOrder(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	paidOrders, err := services.GetPaidOrdersByCustomer(chatID)
	if err != nil {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "获取订单错误")
		tg_bot.Bot.Send(msg)
		return
	}
	if len(paidOrders) == 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "没有已付订单")
		tg_bot.Bot.Send(msg)
		return
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, paidOrder := range paidOrders {
		buttonText := fmt.Sprintf("%s %s %s%s ", time.Unix(paidOrder.CreateTime, 0).In(config.Loc).Format("2006-01-02"), paidOrder.Product.Name, paidOrder.Price, paidOrder.Currency)
		row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(buttonText, GetPaidOrderResultPrefix+paidOrder.ID.String())}
		rows = append(rows, row)
	}
	msgText := config.PaidOrderListMsg(map[string]interface{}{})
	rows = append(rows, deleteMsgRow())
	replyMarkup := tgbotapi.NewInlineKeyboardMarkup(rows...)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
	msg.ReplyMarkup = replyMarkup
	tg_bot.Bot.Send(msg)
}

// 订单详情
func GetPaidOrderResult(update tgbotapi.Update) {
	callbackData := update.CallbackQuery.Data
	senderChatID := update.CallbackQuery.Message.Chat.ID
	senderMsgID := update.CallbackQuery.Message.MessageID

	orderID, err := uuid.Parse(strings.TrimPrefix(callbackData, GetPaidOrderResultPrefix))
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "id错误")
		tg_bot.Bot.Request(callback)
		return
	}
	order, err := services.GetPaidOrderByCustomerByID(orderID)
	if err != nil {
		msg := tgbotapi.NewMessage(senderChatID, err.Error())
		tg_bot.Bot.Send(msg)
		return
	}

	services.SendOrderCallBack(senderChatID, senderMsgID, order, order.Product, order.ProductItem)
}

// 联系客服
func GetContactSupport(update tgbotapi.Update) {
	msgText := config.ContactSupportMsg(map[string]interface{}{
		"ContactSupport": config.GetSiteConfig().ContactSupport,
		"OfficialGroup":  config.GetSiteConfig().OfficialGroup,
	})

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)

	tg_bot.Bot.Send(msg)
}

func RemoveKeyboard(update tgbotapi.Update) {
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	tg_bot.Bot.Send(msg)
}

// 用户中心
func GetUserInfo(update tgbotapi.Update) {
	var userID, username string
	if update.SentFrom() != nil {
		userID = fmt.Sprintf("%d", update.SentFrom().ID)
		username = fmt.Sprintf("%s", update.SentFrom().UserName)
	}
	msgText := ""
	if config.GetSiteConfig().EnableUserBalance {
		userBalanceStr := ""
		userInviterCode := ""
		if uB, uI, err := services.GetUserBalanceNumAndInviterCode(update.SentFrom().ID); err != nil {
			userBalanceStr = "0 " + config.GetSiteConfig().BalanceCurrency
		} else {
			userBalanceStr = fmt.Sprintf("%v %s", uB, config.GetSiteConfig().BalanceCurrency)
			userInviterCode = uI
		}

		msgText = config.UserInfoMsg(map[string]interface{}{
			"UserID":       userID,
			"Username":     username,
			"BalanceTitle": "零钱余额: ",
			"Balance":      userBalanceStr,
			"InviterTitle": "邀请链接: ",
			"InviterLink":  fmt.Sprintf("https://t.me/%s?start=%s", tg_bot.Bot.Self.UserName, userInviterCode),
		})
	} else {
		msgText = config.UserInfoMsg(map[string]interface{}{
			"UserID":   userID,
			"Username": username,
		})
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
	msg.DisableWebPagePreview = true

	tg_bot.Bot.Send(msg)
}

var syncMap sync.Map

// 输入购买数量 按钮相应
func BuyNumCheck(update tgbotapi.Update) {
	senderChatID := update.CallbackQuery.Message.Chat.ID
	callbackData := update.CallbackQuery.Data
	productID, err := uuid.Parse(strings.TrimPrefix(callbackData, BuyNumPrefix))
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "id错误")
		tg_bot.Bot.Request(callback)
		return
	}
	syncMap.Store(senderChatID, productID)

	newMsg := tgbotapi.NewMessage(senderChatID, "请输入购买数量:")
	tg_bot.Bot.Send(newMsg)
}

// 接收数字消息 获取 输入购买数量
func GetBuyNum(update tgbotapi.Update) {
	senderChatID := update.Message.Chat.ID
	// 处理文本消息
	newMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

	var productID uuid.UUID
	if v, ok := syncMap.Load(senderChatID); ok {
		// 将值转换为UUID
		productID = v.(uuid.UUID)
		// syncMap.Delete(senderChatID)
	} else {
		// 忽略非法输入
		return
	}
	// 检查是否是数字
	// num, err := strconv.Atoi(update.Message.Text)
	buyNum, err := strconv.ParseInt(update.Message.Text, 10, 64)
	if err != nil {
		newMsg.Text = "请输入购买数量(数字):"
	} else {
		syncMap.Delete(senderChatID)
		newMsg.Text = "购买数量: " + fmt.Sprintf("%d", buyNum)
		newMsg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			confirmBuyNumRow(productID, buyNum),
		)
	}
	tg_bot.Bot.Send(newMsg)
}

// 下载按钮响应
func ProductItemDownload(update tgbotapi.Update) {
	senderChatID := update.CallbackQuery.Message.Chat.ID
	senderMsgID := update.CallbackQuery.Message.MessageID
	callbackData := update.CallbackQuery.Data

	productID, err := uuid.Parse(strings.TrimPrefix(callbackData, DownloadItemPrefix))
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "商品ID错误")
		tg_bot.Bot.Request(callback)
		return
	}

	product, err := services.GetProductByIDByCustomer(productID)
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "商品不存在")
		tg_bot.Bot.Request(callback)
		return
	}
	productItemList, err := services.GetProductItemListByProductIDByCustomer(product.ID)
	if err != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "商品项目不存在")
		tg_bot.Bot.Request(callback)
		return
	}

	var productItems []string
	for _, v := range productItemList {
		productItems = append(productItems, v.Content)
	}
	msgText := config.ProductItemDownloadMsg(map[string]interface{}{
		"Product": product,
	})
	if len(productItems) > 0 {
		productItemsStr := strings.Join(productItems, "\n")
		productItemsStr = product.Name + "\n" + productItemsStr
		log.Debug("productItemsStr: " + productItemsStr)
		fileBody := []byte(productItemsStr)
		fileBytes := tgbotapi.FileBytes{Name: "all.txt", Bytes: fileBody}
		newMsg := tgbotapi.NewDocument(senderChatID, fileBytes)
		newMsg.Caption = msgText
		newMsg.ParseMode = tgbotapi.ModeHTML
		if result, err := tg_bot.Bot.Send(newMsg); err != nil {
			log.Errorf("Bot.Send(%+v):%v", newMsg, err)
		} else {
			log.Debugf("result: %+v", result)
		}
	} else {
		msgText = msgText + "\n" + "未找到商品项目"
		newMsg := tgbotapi.NewMessage(senderChatID, msgText)
		newMsg.ParseMode = tgbotapi.ModeHTML
		if result, err := tg_bot.Bot.Send(newMsg); err != nil {
			log.Errorf("Bot.Send(%+v):%v", newMsg, err)
		} else {
			log.Debugf("result: %+v", result)
		}
	}

	// 删除原消息
	tg_bot.DeleteMsg(senderChatID, senderMsgID)
}

// 白名单购买 警告消息
func Warning(update tgbotapi.Update) {
	var userID, username string
	if update.SentFrom() != nil {
		userID = fmt.Sprintf("%d", update.SentFrom().ID)
		username = fmt.Sprintf("%s", update.SentFrom().UserName)
	}

	msgText := config.WarningMsg(map[string]interface{}{
		"UserID":   userID,
		"Username": username,
	})

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
	msg.DisableWebPagePreview = true

	tg_bot.Bot.Send(msg)
}

// 白名单购买 警告消息
func Version(update tgbotapi.Update) {
	if update.SentFrom() != nil {
		senderID := update.SentFrom().ID
		msgText := constvar.APPVersion()
		msg := tgbotapi.NewMessage(senderID, msgText)
		tg_bot.Bot.Send(msg)
	}
}
