package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/umfaka/tgfaka/internal/log"

	adminTPL "github.com/umfaka/tgfaka/internal/admin_templates"
	"github.com/umfaka/tgfaka/internal/exts/cache"
	"github.com/umfaka/tgfaka/internal/exts/config"
	"github.com/umfaka/tgfaka/internal/exts/db"
	"github.com/umfaka/tgfaka/internal/exts/tg_bot"
	"github.com/umfaka/tgfaka/internal/handlers/admin_handler"
	"github.com/umfaka/tgfaka/internal/handlers/tg_handler"
	"github.com/umfaka/tgfaka/internal/models"
	"github.com/umfaka/tgfaka/internal/router/middleware"
	"github.com/umfaka/tgfaka/internal/utils/handle_defender"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func SetupRoutes() *gin.Engine {

	config.LoadAllConfig()
	db.InitAllDB()
	tg_bot.InitTGBot()
	cache.InitCache()
	log.Init()

	r := gin.Default()

	r.GET("/api/admin/token_login", admin_handler.TokenLogin)
	r.POST("/api/admin/logout", admin_handler.Logout)

	r.POST("/api/admin/info", middleware.AdminAuthMiddleware(), admin_handler.Info)
	r.POST("/api/admin/dashboard", middleware.AdminAuthMiddleware(), admin_handler.Dashboard)
	r.POST("/api/admin/dashboard_chart", middleware.AdminAuthMiddleware(), admin_handler.DashboardChart)

	r.POST("/api/admin/product", middleware.AdminAuthMiddleware(), admin_handler.FetchList[*models.Product])
	r.POST("/api/admin/create_product", middleware.AdminAuthMiddleware(), admin_handler.CreateProduct)
	r.POST("/api/admin/edit_product", middleware.AdminAuthMiddleware(), admin_handler.EditProduct)
	r.POST("/api/admin/delete_products", middleware.AdminAuthMiddleware(), admin_handler.DeleteEntities[*models.Product])

	r.POST("/api/admin/product_item", middleware.AdminAuthMiddleware(), admin_handler.FetchList[*models.ProductItem])
	r.POST("/api/admin/create_product_items", middleware.AdminAuthMiddleware(), admin_handler.CreateProductItems)
	r.POST("/api/admin/delete_product_items", middleware.AdminAuthMiddleware(), admin_handler.DeleteProductItems) //这个要更新product库存，deletebefore是按删除个数执行的，效率低
	//r.POST("/api/admin/delete_product_items", middleware.AdminAuthMiddleware(), admin_handler.DeleteEntities[*models.ProductItem])

	r.POST("/api/admin/order", middleware.AdminAuthMiddleware(), admin_handler.FetchList[*models.Order])
	r.POST("/api/admin/release_orders", middleware.AdminAuthMiddleware(), admin_handler.ReleaseOrders)

	r.POST("/api/admin/transfer", middleware.AdminAuthMiddleware(), admin_handler.FetchList[*models.Transfer])

	r.POST("/api/admin/wallet", middleware.AdminAuthMiddleware(), admin_handler.FetchList[*models.Wallet])
	r.POST("/api/admin/generate_wallet", middleware.AdminAuthMiddleware(), admin_handler.GenerateWallet)
	r.POST("/api/admin/import_wallet", middleware.AdminAuthMiddleware(), admin_handler.ImportWallet)
	r.POST("/api/admin/edit_wallet", middleware.AdminAuthMiddleware(), admin_handler.EditWallet)
	r.POST("/api/admin/delete_wallets", middleware.AdminAuthMiddleware(), admin_handler.DeleteEntities[*models.Wallet])
	r.POST("/api/admin/delete_all_wallets", middleware.AdminAuthMiddleware(), admin_handler.DeleteAllEntities[*models.Wallet])
	r.POST("/api/admin/refresh_wallet", middleware.AdminAuthMiddleware(), admin_handler.RefreshWallet)
	r.GET("/api/admin/export_wallets", middleware.AdminAuthMiddleware(), admin_handler.ExportWallets)

	r.POST("/api/admin/setting", middleware.AdminAuthMiddleware(), admin_handler.Setting)
	r.POST("/api/admin/edit_setting", middleware.AdminAuthMiddleware(), admin_handler.EditSetting)

	r.GET("/page/admin/", admin_handler.HomeHandle)
	r.StaticFS("/res/", http.FS(adminTPL.StaticSdk))

	return r
}

func RunTgBot() {
	bot := tg_bot.Bot
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		go handleUpdate(update)
	}
}

func handleUpdate(update tgbotapi.Update) {
	defer func() {
		if r := recover(); r != nil {
			msgText := fmt.Sprintf("机器人处理消息崩溃, Error: %v", r)
			handle_defender.HandlePanic(r, msgText)
		}
	}()

	if config.GetSiteConfig().EnableWhitelistBuy && update.SentFrom() != nil {
		if sender := update.SentFrom(); sender != nil {
			senderUsername := ""
			if len(sender.UserName) > 0 {
				senderUsername = "@" + sender.UserName
			}
			senderID := sender.ID
			if !config.IsWhitelist(senderID) {
				log.Debugf("白名单模式:非白名单成员: %d %s", senderID, senderUsername)
				tg_handler.Warning(update)
				return
			} else {
				log.Debugf("白名单模式:白名单成员: %d %s", senderID, senderUsername)
			}
		} else {
			log.Debugf("未识别类型消息: %+v", update)
			return
		}
	} else {
		log.Debugf("非白名单模式")
	}

	if update.Message != nil {
		// if config.GetSiteConfig().EnableWhitelistBuy && update.Message.Chat.IsPrivate() {
		// 	if !config.IsWhitelist(update.Message.Chat.ID) {
		// 		tg_handler.Warning(update)
		// 		return
		// 	}
		// }
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				tg_handler.StartCommand(update)
			case "login":
				tg_handler.LoginCommand(update)
			case "product_list":
				tg_handler.ProductList(update)
			case "paid_order":
				tg_handler.PaidOrder(update)
			case "all":
				tg_handler.SendToAllCommand(update)
			case "topup":
				tg_handler.TopUpCommand(update)
			}

			return
		}
		switch update.Message.Text {
		case tg_handler.PRODUCT_LIST:
			tg_handler.ProductList(update)
		case tg_handler.ORDER_LIST:
			tg_handler.PaidOrder(update)
		case tg_handler.CONTACT_SUPPORT:
			tg_handler.GetContactSupport(update)
		case tg_handler.OFFICIAL_GROUP:
			tg_handler.GetContactSupport(update)
		case tg_handler.USER_INFO:
			tg_handler.GetUserInfo(update)
		case tg_handler.BALANCE_PRODUCT:
			tg_handler.ProductBalanceDetail(update)
		default:
			tg_handler.GetBuyNum(update)
		}

	}
	if update.CallbackQuery != nil {
		callbackData := update.CallbackQuery.Data
		switch {
		case strings.HasPrefix(callbackData, tg_handler.ProductListPagePrefix):
			tg_handler.ProductList(update)

		case strings.HasPrefix(callbackData, tg_handler.ProductDetailPrefix):
			tg_handler.ProductDetail(update)

		case strings.HasPrefix(callbackData, tg_handler.BuyNumPrefix):
			tg_handler.BuyNumCheck(update)

		case strings.HasPrefix(callbackData, tg_handler.PayOrderPrefix):
			tg_handler.PayOrder(update)

		case strings.HasPrefix(callbackData, tg_handler.PayOrderByBalancePrefix):
			tg_handler.PayOrderByBalance(update)

		case strings.HasPrefix(callbackData, tg_handler.GetPaidOrderResultPrefix):
			tg_handler.GetPaidOrderResult(update)

		case strings.HasPrefix(callbackData, tg_handler.DownloadItemPrefix):
			tg_handler.ProductItemDownload(update)

		case strings.HasPrefix(callbackData, "delete_msg"):
			tg_handler.CallbackDeleteMsg(update)

		}

	}
}
