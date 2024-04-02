package services

import (
	"fmt"
	"gopay/internal/exts/config"
	"gopay/internal/utils/functions"
	"os"
	"time"
)

func GetAdminLoginUrlSession() string {
	path := config.LoginUrlSessionPath
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}
func SetAdminLoginUrlSession(expireTime time.Duration) string {
	path := config.LoginUrlSessionPath
	token, _ := GenerateAdminLoginUrlToken(expireTime)
	os.WriteFile(path, []byte(token), 0644)
	return token
}
func ClearAdminLoginUrlSession() {
	path := config.LoginUrlSessionPath
	os.WriteFile(path, []byte(""), 0644)
}

func GetAdminLoginTokenSession() string {
	path := config.LoginTokenSessionPath
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}
func SetAdminLoginTokenSession(expireTime time.Duration) string {
	adminToken, _ := GenerateAdminLoginToken(expireTime)
	path := config.LoginTokenSessionPath
	os.WriteFile(path, []byte(adminToken), 0644)
	return adminToken
}
func ClearAdminLoginTokenSession() {
	path := config.LoginTokenSessionPath
	os.WriteFile(path, []byte(""), 0644)
}

func GenerateAdminLoginUrlToken(expireTime time.Duration) (string, error) {
	data := map[string]interface{}{
		"loginUsername": fmt.Sprintf("%d", config.GetSiteConfig().AdminTGID),
	}
	token, err := functions.GenerateToken(data, config.SiteSecret, expireTime)
	if err != nil {
		return "", err
	}
	return token, nil
}

func GenerateAdminLoginToken(expireTime time.Duration) (string, error) {
	data := map[string]interface{}{
		"role": "admin",
	}
	token, err := functions.GenerateToken(data, config.SiteSecret, expireTime)
	if err != nil {
		return "", err
	}
	return token, nil
}
