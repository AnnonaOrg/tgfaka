package tg_handler

import (
	"gopay/internal/exts/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func welcomeStartFooterRows() [][]tgbotapi.KeyboardButton {
	var rows [][]tgbotapi.KeyboardButton

	productListMsg := productListMsgFooterRow()
	if len(productListMsg) > 0 {
		rows = append(rows, productListMsg)
	}

	welcomeStart := welcomeStartMsgFooterRow()
	if len(welcomeStart) > 0 {
		rows = append(rows, welcomeStart)
	}

	return rows
}

func welcomeStartMsgFooterRow() []tgbotapi.KeyboardButton {
	officialCustomerService := config.GetSiteConfig().ContactSupport
	officialGroup := config.GetSiteConfig().OfficialGroup

	var welcomeStartRow []tgbotapi.KeyboardButton
	welcomeStartRow = append(welcomeStartRow,
		tgbotapi.NewKeyboardButton(USER_INFO),
	)

	if config.GetSiteConfig().EnableUserBalance && len(config.GetSiteConfig().BalanceProductUUID) > 0 {
		welcomeStartRow = append(welcomeStartRow,
			tgbotapi.NewKeyboardButton(BALANCE_PRODUCT),
		)
	}

	if len(officialCustomerService) > 0 || len(officialGroup) > 0 {
		welcomeStartRow = append(welcomeStartRow,
			tgbotapi.NewKeyboardButton(CONTACT_SUPPORT),
		)
	}

	return welcomeStartRow
}
func productListMsgFooterRow() []tgbotapi.KeyboardButton {
	return []tgbotapi.KeyboardButton{
		tgbotapi.NewKeyboardButton(PRODUCT_LIST),
		tgbotapi.NewKeyboardButton(ORDER_LIST),
	}
}
