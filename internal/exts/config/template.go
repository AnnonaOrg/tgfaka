package config

import (
	"bytes"
	"fmt"
	"gopay/internal/utils/functions"
	"text/template"
	"time"
)

var templates *template.Template

const (
	welcomeTplName                     = "welcome.tpl"
	productListTplName                 = "product_list.tpl"
	productDetailTplName               = "product_detail.tpl"
	payOrderTplName                    = "pay_order.tpl"
	payOrderBalanceTplName             = "pay_order_balance.tpl"
	orderCallbackTplName               = "order_callback.tpl"
	orderCallbackMoreTplName           = "order_callback_more.tpl"
	orderCallbackBalanceTplName        = "order_callback_balance.tpl"
	orderCallbackChatInviteLinkTplName = "order_callback_chatinvitelink.tpl"

	paidOrderListTplName       = "paid_order_list.tpl"
	contactSupportTplName      = "contact_support.tpl"
	userInfoTplName            = "user_info.tpl"
	productItemDownloadTplName = "product_item_download.tpl"

	inviteRewardsBalanceTplName = "invite_rewards_balance.tpl"

	warningTplName = "warning.tpl"
)

func timestampToDatetime(timestamp int64) string {
	t := time.Unix(timestamp, 0)                   // Converts Unix timestamp to time.Time
	return t.In(Loc).Format("2006-01-02 15:04:05") // Formats the time in a human-readable form
}

func LoadTemplates() {
	funcMap := template.FuncMap{
		"TimestampToDatetime": timestampToDatetime,
	}

	var err error
	templates, err = template.New("base").Funcs(funcMap).ParseGlob(functions.GetExecutableDir() + "/templates/*.tpl")
	if err != nil {
		panic(err)
	}
	templateNames := []string{
		welcomeTplName,
		productListTplName,
		productDetailTplName,
		payOrderTplName,
		payOrderBalanceTplName,
		orderCallbackTplName,
		orderCallbackMoreTplName,
		orderCallbackBalanceTplName,
		orderCallbackChatInviteLinkTplName,
		paidOrderListTplName,
		contactSupportTplName,
		userInfoTplName,
		productItemDownloadTplName,
		inviteRewardsBalanceTplName,
		warningTplName,
	}
	for _, name := range templateNames {
		if templates.Lookup(name) == nil {
			panic(fmt.Sprintf("Template %s not found", name))
		}
	}
}

func ExecuteTemplate(templateName string, data interface{}) string {
	var buf bytes.Buffer
	err := templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		return "execute tpl err"
	}
	return buf.String()
}

func WelcomeMsg(data interface{}) string {
	return ExecuteTemplate(welcomeTplName, data)
}
func ProductListMsg(data interface{}) string {
	return ExecuteTemplate(productListTplName, data)
}

func ProductDetailMsg(data interface{}) string {
	return ExecuteTemplate(productDetailTplName, data)
}
func PayOrderMsg(data interface{}) string {
	return ExecuteTemplate(payOrderTplName, data)
}
func PayOrderBalanceMsg(data interface{}) string {
	return ExecuteTemplate(payOrderBalanceTplName, data)
}
func OrderCallbackMsg(data interface{}) string {
	return ExecuteTemplate(orderCallbackTplName, data)
}
func OrderCallbackMoreMsg(data interface{}) string {
	return ExecuteTemplate(orderCallbackMoreTplName, data)
}
func OrderCallbackBalanceMsg(data interface{}) string {
	return ExecuteTemplate(orderCallbackBalanceTplName, data)
}
func OrderCallbackChatInviteLinkMsg(data interface{}) string {
	return ExecuteTemplate(orderCallbackChatInviteLinkTplName, data)
}
func PaidOrderListMsg(data interface{}) string {
	return ExecuteTemplate(paidOrderListTplName, data)
}
func ContactSupportMsg(data interface{}) string {
	return ExecuteTemplate(contactSupportTplName, data)
}
func UserInfoMsg(data interface{}) string {
	return ExecuteTemplate(userInfoTplName, data)
}

func ProductItemDownloadMsg(data interface{}) string {
	return ExecuteTemplate(productItemDownloadTplName, data)
}

func InviteRewardsBalanceMsg(data interface{}) string {
	return ExecuteTemplate(inviteRewardsBalanceTplName, data)
}

func WarningMsg(data interface{}) string {
	return ExecuteTemplate(warningTplName, data)
}
