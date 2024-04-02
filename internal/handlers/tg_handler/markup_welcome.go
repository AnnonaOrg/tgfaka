package tg_handler

import (
	"gopay/internal/exts/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func welcomeStartRows() [][]tgbotapi.InlineKeyboardButton {
	var rows [][]tgbotapi.InlineKeyboardButton
	welcomeStart := welcomeStartMsgRow()
	if len(welcomeStart) > 0 {
		rows = append(rows, welcomeStart)
	}
	productListMsg := productListMsgRow()
	if len(productListMsg) > 0 {
		rows = append(rows, productListMsg)
	}
	deleteMsg := deleteMsgRow()
	if len(deleteMsg) > 0 {
		rows = append(rows, deleteMsg)
	}
	return rows
}

func welcomeStartMsgRow() []tgbotapi.InlineKeyboardButton {
	officialCustomerService := config.GetSiteConfig().ContactSupport
	officialGroup := config.GetSiteConfig().OfficialGroup

	var welcomeStartRow []tgbotapi.InlineKeyboardButton
	if len(officialCustomerService) > 0 {
		welcomeStartRow = append(welcomeStartRow,
			tgbotapi.NewInlineKeyboardButtonURL(CONTACT_SUPPORT,
				officialCustomerService,
			),
		)
	}
	if len(officialGroup) > 0 {
		welcomeStartRow = append(welcomeStartRow,
			tgbotapi.NewInlineKeyboardButtonURL(OFFICIAL_GROUP,
				officialGroup,
			),
		)
	}
	return welcomeStartRow
}
func productListMsgRow() []tgbotapi.InlineKeyboardButton {
	return []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData(PRODUCT_LIST, ProductListPagePrefix+"start"),
	}
}
