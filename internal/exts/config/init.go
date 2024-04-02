package config

import (
	"time"

	"github.com/umfaka/tgfaka/internal/utils/functions"
)

func LoadAllConfig() {
	LoadSiteConfig()
	LoadSiteSecret()
	LoadDBConfig()

	SetSiteConfig(*GetSiteConfig())

	LoadTemplates()

	Loc, _ = time.LoadLocation("Asia/Shanghai")
}

var appBaseDir = functions.GetExecutableDir()
var configBaseDir = appBaseDir + "/conf"

var SiteConfigPath = configBaseDir + "/config.ini"
var LoginUrlSessionPath = appBaseDir + "/.urlsession"
var LoginTokenSessionPath = appBaseDir + "/.tokensession"
var ExchangeRateDataPath = appBaseDir + "/.exchangerate"

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
