package tg_handler

import (
	"fmt"

	"github.com/umfaka/tgfaka/internal/exts/config"
	"github.com/umfaka/tgfaka/internal/models"
	"github.com/umfaka/tgfaka/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

func paginationToRows(pagination services.Pagination) [][]tgbotapi.InlineKeyboardButton {
	var rows [][]tgbotapi.InlineKeyboardButton

	for _, item := range pagination.Items {
		product := item.(models.Product)
		// buttonText := fmt.Sprintf("%s : %s 库存:%d", product.Name, product.Description, product.InStockCount)
		buttonText := "" // fmt.Sprintf("%s 库存:%d", product.Name, product.InStockCount)
		if config.IsBalanceProduct(product.ID.String()) {
			buttonText = product.Name
		} else {
			buttonText = fmt.Sprintf("%s 库存:%d", product.Name, product.InStockCount)
			if config.GetSiteConfig().EnableHiddenInStockCount {
				buttonText = fmt.Sprintf("%s", product.Name)
			} else {
				buttonText = fmt.Sprintf("%s 库存:%d", product.Name, product.InStockCount)
			}
		}

		row := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData(buttonText, ProductDetailPrefix+product.ID.String()+"_1")}
		rows = append(rows, row)
	}

	var paginationRow []tgbotapi.InlineKeyboardButton
	if pagination.Page > 1 {
		paginationRow = append(paginationRow, tgbotapi.NewInlineKeyboardButtonData("上一页", ProductListPagePrefix+fmt.Sprintf("%d", pagination.Page-1)))
	}
	if int64(pagination.Page) < pagination.TotalPage {
		paginationRow = append(paginationRow, tgbotapi.NewInlineKeyboardButtonData("下一页", ProductListPagePrefix+fmt.Sprintf("%d", pagination.Page+1)))
	}
	// row不能为空，空了发不出去
	if len(paginationRow) != 0 {
		rows = append(rows, paginationRow)
	}

	return rows
}

// 支付方式按钮
func paymentSelectRow(productID uuid.UUID, buyNum int64) []tgbotapi.InlineKeyboardButton {
	var paymentSelectRow []tgbotapi.InlineKeyboardButton

	if !config.GetSiteConfig().EnableOnlyBalancePaymentMethods || config.IsBalanceProduct(productID.String()) {
		for _, v := range config.GetAvailablePaymentMethods() {
			callbackData := fmt.Sprintf("%s%s_%s_%d", PayOrderPrefix, productID, v, buyNum)
			paymentSelectRow = append(paymentSelectRow, tgbotapi.NewInlineKeyboardButtonData(v, callbackData))
		}
	}
	if config.GetSiteConfig().EnableUserBalance && !config.IsBalanceProduct(productID.String()) {
		text := "余额支付"
		// paymentMethod := config.GetSiteConfig().BalanceCurrency
		callbackData := fmt.Sprintf("%s%s_%d", PayOrderByBalancePrefix, productID, buyNum)
		paymentSelectRow = append(paymentSelectRow,
			tgbotapi.NewInlineKeyboardButtonData(text, callbackData),
		)
	}
	return paymentSelectRow
}
func deleteMsgRow() []tgbotapi.InlineKeyboardButton {
	var paymentSelectRow []tgbotapi.InlineKeyboardButton
	paymentSelectRow = append(paymentSelectRow, tgbotapi.NewInlineKeyboardButtonData("关闭", "delete_msg"))

	return paymentSelectRow
}

func GoBackRow(callBackData string) []tgbotapi.InlineKeyboardButton {
	goBackRow := []tgbotapi.InlineKeyboardButton{tgbotapi.NewInlineKeyboardButtonData("返回", callBackData)}
	return goBackRow
}

func buyNumButtonRow(productID uuid.UUID, buyNum int64) []tgbotapi.InlineKeyboardButton {
	callbackData := ProductDetailPrefix + productID.String() //	+ "_" //+ fmt.Sprintf("%d", buyNum)

	var buyNumSelectRow []tgbotapi.InlineKeyboardButton
	buyNumSelectRow = append(buyNumSelectRow,
		tgbotapi.NewInlineKeyboardButtonData("+", callbackData+fmt.Sprintf("_%d", buyNum+1)),
	)

	buyNumSelectRow = append(buyNumSelectRow,
		tgbotapi.NewInlineKeyboardButtonData("输入数量", BuyNumPrefix+productID.String()),
	)

	if buyNum < 1 {
		buyNum = 2
	}
	buyNumSelectRow = append(buyNumSelectRow,
		tgbotapi.NewInlineKeyboardButtonData("-", callbackData+fmt.Sprintf("_%d", buyNum-1)),
	)

	return buyNumSelectRow
}

func confirmBuyNumRow(productID uuid.UUID, buyNum int64) []tgbotapi.InlineKeyboardButton {
	callBackData := ProductDetailPrefix + productID.String() + "_" + fmt.Sprintf("%d", buyNum)
	confirmRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("确认",
			callBackData,
		),
	}
	return confirmRow
}

func downloadItemRow(productID uuid.UUID) []tgbotapi.InlineKeyboardButton {
	callBackData := DownloadItemPrefix + productID.String() //+ "_" + fmt.Sprintf("%d", inStockCount)
	downloadRow := []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(
			PRODUCT_ITEM_DOWNLOAD,
			callBackData,
		),
	}
	return downloadRow
}
