package config

import (
	"gopay/internal/utils/functions"
	"time"
)

func LoadAllConfig() {
	LoadSiteConfig()
	LoadSiteSecret()
	LoadDBConfig()

	SetSiteConfig(*GetSiteConfig())

	LoadTemplates()

	Loc, _ = time.LoadLocation("Asia/Shanghai")
}

var configBaseDir = functions.GetExecutableDir() + "/.env"

var SiteConfigPath = configBaseDir + "/config.ini"
var LoginUrlSessionPath = configBaseDir + "/.urlsession"
var LoginTokenSessionPath = configBaseDir + "/.tokensession"
var ExchangeRateDataPath = configBaseDir + "/.exchangerate"

type Headers map[string]string

type Proxy struct {
	EnableProxy bool   `desc:"是否开启网络代理"`
	Protocol    string `desc:"协议"`
	Host        string `desc:"域名"`
	Port        uint   `desc:"端口"`
}
type Body struct {
	Value map[string]interface{}
}

var Loc *time.Location
