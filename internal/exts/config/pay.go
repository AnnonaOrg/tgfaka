package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/shopspring/decimal"
	"github.com/umfaka/tgfaka/internal/utils/functions"
)

type Network string
type Currency string
type Fiat string

const (
	TRON Network  = "TRON"
	TRX  Currency = "TRX"
	USDT Currency = "USDT"
	CNY  Currency = "CNY"
)

type PaymentOption struct {
	Network  Network  `json:"network"`
	Currency Currency `json:"currency"`
}

var NetworkCurrencies = map[Network][]Currency{
	TRON: {TRX, USDT},
}
var Fiats = []string{string(CNY)}

type ExchangeRateStruct struct {
	ExchangeRate map[Currency]decimal.Decimal `json:"exchange_rate"`
	UpdateTime   string                       `json:"update_time"`
}

var ExchangeRateData = ExchangeRateStruct{
	ExchangeRate: make(map[Currency]decimal.Decimal),
	UpdateTime:   "",
}
var ExchangeRateLock = &sync.RWMutex{}

type DecimalUnitsStruct struct {
	TRX  decimal.Decimal
	USDT decimal.Decimal
}

var DecimalWalletUnitMap = map[Currency]decimal.Decimal{
	USDT: decimal.RequireFromString("0.0001"),
	TRX:  decimal.RequireFromString("0.0001"),
}
var DecimalWalletMaxOrderCount = 500

func GetFixedExchangeRate(currency Currency) (decimal.Decimal, error) {
	var exchangeRateDecimal decimal.Decimal

	var fixedExchangeRateMap map[Currency]decimal.Decimal
	err := json.Unmarshal([]byte(SiteConfig.FixedExchangeRate), &fixedExchangeRateMap)
	if err != nil {
		return exchangeRateDecimal, err
	}

	if fixedExchangeRateMap[currency].Equal(decimal.Zero) {
		return decimal.Zero, errors.New("获取固定汇率失败")
	}

	return fixedExchangeRateMap[currency], nil
}

// 获取汇率
func GetExchangeRate(currency Currency) (decimal.Decimal, error) {
	if currency == CNY {
		return decimal.RequireFromString("1"), nil
	}

	var exchangeRate decimal.Decimal
	if SiteConfig.EnableFixExchangeRate {
		exchangeRate, err := GetFixedExchangeRate(currency)
		if err != nil {
			return exchangeRate, err
		} else {
			return exchangeRate, nil
		}
	}

	// 先到内存取，没有再到文件取
	var value decimal.Decimal
	if !ExchangeRateData.ExchangeRate[currency].Equal(decimal.Zero) {
		value = ExchangeRateData.ExchangeRate[currency]
	} else {
		var exchangeRateDataInFile ExchangeRateStruct
		path := ExchangeRateDataPath
		fileContent, err := os.ReadFile(path)
		if err != nil {
			return exchangeRate, err
		}
		err = json.Unmarshal(fileContent, &exchangeRateDataInFile)
		if err != nil {
			return exchangeRate, err
		}

		if !exchangeRateDataInFile.ExchangeRate[currency].Equal(decimal.Zero) {
			value = exchangeRateDataInFile.ExchangeRate[currency]
		}

	}
	if value.Equal(decimal.Zero) {
		return exchangeRate, errors.New("获取失败")
	}

	exchangeRate = value

	return exchangeRate, nil
}
func SetExchangeRate(exchangeRate ExchangeRateStruct) error {
	path := ExchangeRateDataPath
	fileContent, err := json.Marshal(exchangeRate)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, fileContent, 0600)
	if err != nil {
		return err
	}

	return nil
}

//func GetExchangeRatio(baseCurrency string, quoteCurrency string) (decimal.Decimal, error) {
//	// 1 BaseCurrency 换 how much QuoteCurrency
//	// 全部换算成一种基础货币,
//
//	if !functions.SliceContainString(GetAllCurrencies(), baseCurrency) || !functions.SliceContainString(GetAllCurrencies(), quoteCurrency) {
//		return decimal.Decimal{}, errors.New("currency not exist")
//	}
//	if baseCurrency == quoteCurrency {
//		return decimal.NewFromInt(1), nil
//	}
//
//	return
//}

func GetAllPaymentMethods() []string {
	var result []string
	for network, currencies := range NetworkCurrencies {
		for _, currency := range currencies {
			result = append(result, fmt.Sprintf("%s-%s", currency, network))
		}
	}
	return result
}
func GetAvailablePaymentMethods() []string {
	var result []string
	for network, currencies := range NetworkCurrencies {
		for _, currency := range currencies {
			paymentMethodString := fmt.Sprintf("%s-%s", currency, network)
			if strings.Contains(SiteConfig.PaymentMethods, paymentMethodString) {
				result = append(result, paymentMethodString)
			}
		}
	}
	return result
}
func ParsePaymentMethod(inputPaymentMethod string) (*PaymentOption, error) {
	parts := strings.Split(inputPaymentMethod, "-")
	if len(parts) != 2 {
		return nil, errors.New("格式错误")
	}

	if !functions.SliceContainString(GetAllPaymentMethods(), inputPaymentMethod) {
		return nil, errors.New("未知方式")
	}

	return &PaymentOption{
		Currency: Currency(parts[0]),
		Network:  Network(parts[1]),
	}, nil
}

func GetCurrencies() []string {
	var allCurrencies []string

	for _, currencies := range NetworkCurrencies {
		for _, currency := range currencies {
			allCurrencies = append(allCurrencies, string(currency))
		}
	}

	allCurrencies = append(allCurrencies, Fiats...)
	return allCurrencies
}

func GetAllNetworks() []string {
	var allNetworks []string
	for network, _ := range NetworkCurrencies {
		allNetworks = append(allNetworks, string(network))
	}
	return allNetworks
}

func GetCryptoCurrencies() []string {
	var allCurrencies []string

	for _, currencies := range NetworkCurrencies {
		for _, currency := range currencies {
			allCurrencies = append(allCurrencies, string(currency))
		}
	}
	return allCurrencies
}

func ConvertCurrencyPrice(amount decimal.Decimal, fromCurrency, toCurrency Currency) (decimal.Decimal, error) {
	// 同名返回同样的价格
	if fromCurrency == toCurrency {
		return amount, nil
	}

	// 获取汇率,1源货币换多少CNY
	fromRate, err := GetExchangeRate(fromCurrency)
	if err != nil {
		return decimal.Zero, fmt.Errorf("获取汇率失败: %s, Error:%s", fromCurrency, err)
	}

	// 获取汇率,1目标货币换多少CNY
	toRate, err := GetExchangeRate(toCurrency)
	if err != nil {
		return decimal.Zero, fmt.Errorf("获取汇率失败: %s, Error:%s", toCurrency, err)
	}

	// 先把源货币转成CNY数额
	amountInBase := amount.Mul(fromRate)

	// 再从CNY转成目标货币的数额
	convertedAmount := amountInBase.Div(toRate)
	return convertedAmount, nil
}
func IsPaymentEnable(input string) bool {
	if !strings.Contains(SiteConfig.PaymentMethods, input) {
		return false
	}
	return true
}
