package admin_handler

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopay/internal/exts/config"
	"gopay/internal/exts/db"
	"gopay/internal/models"
	"gopay/internal/services"
	"gopay/internal/utils/crypto_api"
	"gopay/internal/utils/crypto_api/tron"
	"gopay/internal/utils/functions"
	"gopay/internal/utils/restful"
	"net/http"
	"strings"
)

// validate:"required" 与binding不同,validate是在绑定后再验证，所以required验证的是零值，binding是绑定前验证，判断是否存在
func GenerateWallet(c *gin.Context) {
	var requestData struct {
		Network    string `json:"network" binding:"required,network"`
		Num        uint   `json:"num" binding:"required"`
		WalletType int    `json:"wallet_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		restful.ParamErr(c, "参数错误"+err.Error())
		return
	}

	var wallets []*models.Wallet
	for i := 0; i < int(requestData.Num); i++ {
		var account crypto_api.Account
		var wallet *models.Wallet
		var err error
		switch requestData.Network {
		case string(config.TRON):
			client := tron.New(config.SiteConfig.TronGridApiKey)
			account, err = client.Generate()
			if err != nil {
				restful.ParamErr(c, "生成钱包错误")
				return
			}
			wallet = models.NewWallet(config.TRON, account.Address, &account.PrivateKey, requestData.WalletType, 0)
		default:
			restful.ParamErr(c, "主网错误")
			return
		}
		if err != nil {
			restful.ParamErr(c, "生成错误")
			return
		}
		wallets = append(wallets, wallet)
	}

	err := services.CreateEntities[*models.Wallet](wallets)
	if err != nil {
		restful.ParamErr(c, "创建失败")
		return
	}

	restful.Ok(c)
}

func ImportWallet(c *gin.Context) {
	var requestData struct {
		WalletData *string `json:"wallet_data" binding:"required"`
		Network    *string `json:"network" binding:"required,network"`
		Priority   *int64  `json:"priority" binding:"required"`
		WalletType *int    `json:"wallet_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		restful.ParamErr(c, "参数错误"+err.Error())
		return
	}

	var network config.Network
	switch config.Network(*requestData.Network) {
	case config.TRON:
		network = config.TRON
	default:
		restful.ParamErr(c, "主网错误")
		return
	}

	var wallets []models.Wallet
	var errMsg string
	lines := strings.Split(*requestData.WalletData, "\n")
	for _, line := range lines {
		if functions.IsWhitespace(line) {
			continue
		}

		splitFunc := func(c rune) bool {
			return c == ',' || c == '\t'
		}
		parts := strings.FieldsFunc(line, splitFunc)

		var address string
		var privateKey *string
		address = parts[0]
		client := tron.New(config.SiteConfig.TronGridApiKey)
		if !client.ValidateAddress(address) {
			errMsg = errMsg + fmt.Sprintf("%s 地址格式错误\n", address)
			continue
		}
		if len(parts) > 1 {
			privateKey = &parts[1]
			if ok := client.ValidatePrivateKey(*privateKey, address); !ok {
				errMsg = errMsg + fmt.Sprintf("%s 密钥校验错误\n", address)
				continue
			}
		}

		// 检查是否重复
		var tempWallet models.Wallet
		result := db.DB.Where("address = ? and network = ?", address, network).Limit(1).Find(&tempWallet)
		if result.RowsAffected > 0 {
			errMsg = errMsg + fmt.Sprintf("%s 地址已存在\n", address)
			continue
		}

		wallet := models.NewWallet(network, address, privateKey, *requestData.WalletType, *requestData.Priority)
		wallets = append(wallets, *wallet)
	}

	if len(wallets) > 0 {
		result := db.DB.Create(&wallets)
		if result.Error != nil {
			restful.ParamErr(c, "添加失败")
			return
		}

		restful.Ok(c, fmt.Sprintf("成功添加:%d个钱包\n%s", result.RowsAffected, errMsg))
		return
	}

	restful.Ok(c, errMsg)
}

func EditWallet(c *gin.Context) {
	var requestData struct {
		ID       *uuid.UUID `json:"id" binding:"required"`
		Priority *int64     `json:"priority" `
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		restful.ParamErr(c, "参数错误")
		return
	}

	updateMap := functions.StructToMap(requestData, functions.StructToMapExcludeMode, "id")
	err := services.UpdateEntity[*models.Wallet](*requestData.ID, updateMap)
	if err != nil {
		restful.ParamErr(c, "编辑失败")
		return
	}

	restful.Ok(c, "编辑成功")
}

//func DeleteWallets(c *gin.Context) {
//	var requestData struct {
//		IDsString string `json:"ids"`
//	}
//	if err := c.ShouldBindJSON(&requestData); err != nil {
//		restful.ParamErr(c, "参数错误")
//		return
//	}
//
//	ids, err := functions.ParseIDsString(requestData.IDsString)
//	if err != nil {
//		restful.ParamErr(c, "id格式错误")
//		return
//	}
//
//	err = services.DeleteWallets(ids)
//	if err != nil {
//		restful.ParamErr(c, "删除失败")
//		return
//	}
//
//	restful.Ok(c, "删除成功")
//}
//func DeleteAllWallets(c *gin.Context) {
//	var allWalletIDs []uuid.UUID
//	db.DB.Model(&models.Wallet{}).Pluck("id", &allWalletIDs)
//
//	err := services.DeleteWallets(allWalletIDs)
//	if err != nil {
//		restful.ParamErr(c, "删除失败")
//		return
//	}
//
//	restful.Ok(c, "删除成功")
//}

func RefreshWallet(c *gin.Context) {
	var requestData struct {
		IDString uuid.UUID `json:"id"`
	}
	if err := c.ShouldBindJSON(&requestData); err != nil {
		restful.ParamErr(c, "参数错误")
		return
	}

	var toRefreshWallet models.Wallet
	result := db.DB.Where("id = ?", requestData.IDString).Find(&toRefreshWallet)
	if result.RowsAffected == 0 {
		restful.ParamErr(c, "没有该钱包")
		return
	}

	client := tron.New(config.SiteConfig.TronGridApiKey)
	balance, err := client.GetBalance(toRefreshWallet.Address)
	if err != nil {
		restful.ParamErr(c, "获取钱包余额失败")
		return
	}

	err = services.UpdateWalletBalance(toRefreshWallet.ID, balance)
	if err != nil {
		restful.ParamErr(c, "更新钱包余额失败")
		return
	}

	restful.Ok(c, "刷新成功")
}
func ExportWallets(c *gin.Context) {

	var wallets []models.Wallet
	if result := db.DB.Find(&wallets); result.Error != nil {
		restful.ParamErr(c, "查询错误")
		return
	}

	var csvBuilder strings.Builder
	writer := csv.NewWriter(&csvBuilder)

	// Write CSV header
	writer.Write([]string{"address", "private_key", "balance_data"})

	// Write data rows
	for _, wallet := range wallets {
		var privateKey string
		if wallet.PrivateKey != nil {
			privateKey = *wallet.PrivateKey
		}

		bytes, _ := json.Marshal(wallet.BalanceData)
		balanceDataString := string(bytes)

		writer.Write([]string{
			wallet.Address,
			privateKey,
			balanceDataString,
		})
	}

	writer.Flush()
	csvData := csvBuilder.String()

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=wallets.csv")
	c.Data(http.StatusOK, "text/csv", []byte(csvData))

}
