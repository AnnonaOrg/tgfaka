package config

import (
	"gopay/internal/utils/functions"
	"os"
)

var SiteSecret string

func LoadSiteSecret() {
	secretFilePath := configBaseDir + "/.secret"
	if _, err := os.Stat(secretFilePath); os.IsNotExist(err) {
		secret := functions.GenerateRandomString(32)
		err = os.WriteFile(secretFilePath, []byte(secret), 0600)
		if err != nil {
			panic(err)
		}
	}

	secret, err := os.ReadFile(secretFilePath)
	if err != nil {
		panic(err)
	}
	SiteSecret = string(secret)
}
