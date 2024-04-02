package crypto_api

import (
	"github.com/shopspring/decimal"
	"github.com/umfaka/tgfaka/internal/models"
)

type Provider interface {
	Generate() (Account, error)
	GetBalance(string) (map[string]decimal.Decimal, error)
	ValidatePrivateKey(string, string) bool
	ValidateAddress(string) bool
	GetScheduleTransfers() []models.Transfer
}

type Account struct {
	Address    string
	PrivateKey string
}

type Token string
type Coin string
