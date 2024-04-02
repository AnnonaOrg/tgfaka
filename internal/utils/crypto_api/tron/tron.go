package tron

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcutil/base58"
	"gopay/internal/exts/cache"
	"gopay/internal/exts/config"
	my_log "gopay/internal/exts/log"
	"gopay/internal/models"
	"gopay/internal/utils/crypto_api"
	"gopay/internal/utils/functions"
	"gopay/internal/utils/requests"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"
)

var tracebackThreshold = 2000
var scheduleLastKey = "TRON_last_block_num"
var ratio = 1e6

const (
	balanceFuncSelector string = "70a08231"
	usdtContractAddress string = "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t"
)

type Tron struct {
	BaseURL           string
	TronGridApiKey    string
	Denomination      decimal.Decimal
	TRC20Denomination decimal.Decimal
	RequestsHeader    *config.Headers
}

func New(tronGridApiKey string) *Tron {
	var header *config.Headers
	if tronGridApiKey == "" {
		header = &config.Headers{"Tron-PRO-API-KEY": tronGridApiKey}
	} else {
		header = nil
	}
	return &Tron{
		BaseURL:           "https://api.trongrid.io",
		TronGridApiKey:    tronGridApiKey,
		Denomination:      decimal.NewFromInt(1000000),
		TRC20Denomination: decimal.NewFromInt(1000000),
		RequestsHeader:    header,
	}
}

func (client *Tron) Generate() (crypto_api.Account, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return crypto_api.Account{}, err
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	hexPrivateKey := hexutil.Encode(privateKeyBytes)[2:]
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return crypto_api.Account{}, fmt.Errorf("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	address = "41" + address[2:]
	addb, err := hex.DecodeString(address)
	if err != nil {
		return crypto_api.Account{}, err
	}
	hash1 := s256(s256(addb))
	secret := hash1[:4]
	addb = append(addb, secret...)
	base58Address := base58.Encode(addb)

	if err != nil {
		return crypto_api.Account{}, err
	}

	return crypto_api.Account{PrivateKey: hexPrivateKey, Address: base58Address}, nil
}

func (client *Tron) SendUSDT(wallet crypto_api.Account, toAddress string, amount decimal.Decimal) (string, error) {
	value := fmt.Sprintf("%x", amount.Mul(client.TRC20Denomination).IntPart())
	reqData := map[string]interface{}{
		"owner_address":     base58ToHex(wallet.Address),
		"contract_address":  base58ToHex(usdtContractAddress),
		"function_selector": "transfer(address,uint256)",
		"parameter":         strings.Repeat("0", 24) + base58ToHex(toAddress)[2:] + strings.Repeat("0", 64-len(value)) + value,
		"call_value":        0,
		"fee_limit":         10000000000,
	}
	response, err := requests.Post(client.BaseURL+"/wallet/triggersmartcontract", reqData, client.RequestsHeader)
	if err != nil {
		return "", err
	}

	var result struct {
		Transaction RawTransaction
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return "", err
	}

	if result.Transaction.TxID == "" {
		return "", fmt.Errorf("SendTRC20 response: %s", string(response))
	}

	pk, err := crypto.HexToECDSA(wallet.PrivateKey)
	if err != nil {
		return "", err
	}
	result.Transaction.Visible = false
	signedTransaction, err := client.signRawTransaction(&result.Transaction, pk)
	if err != nil {
		return "", err
	}

	return client.broadcastTransaction(signedTransaction)
}

func (client *Tron) getUSDTBalance(address string) (decimal.Decimal, error) {
	balance := decimal.Zero
	data := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "eth_call",
		"params": []interface{}{
			map[string]interface{}{
				"to":    "0x" + base58ToHex(usdtContractAddress)[2:],
				"value": "0x0",
				"data":  "0x" + balanceFuncSelector + strings.Repeat("0", 24) + base58ToHex(address)[2:],
			},
			"latest",
		},
	}
	response, err := requests.Post(client.BaseURL+"/jsonrpc", data, client.RequestsHeader)
	if err != nil {
		return balance, err
	}
	var result struct {
		ID      any    `json:"id"`
		JSONRPC string `json:"jsonrpc"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
		Result string `json:"result,omitempty"`
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return balance, err
	}
	if result.Result == "" {
		return balance, fmt.Errorf("GetTRC20Balance body: %s", string(response))
	}

	i := new(big.Int)
	i.SetString(result.Result[2:], 16)
	sun := decimal.NewFromBigInt(i, 0)
	balance = sun.DivRound(client.Denomination, 18)
	return balance, nil
}

func (client *Tron) SendTRX(wallet crypto_api.Account, to string, amount decimal.Decimal) (string, error) {
	var rawTransaction RawTransaction
	amount = amount.Mul(client.Denomination)

	reqData := map[string]interface{}{
		"owner_address": wallet.Address,
		"to_address":    to,
		"amount":        amount.IntPart(),
		"visible":       true,
	}
	response, err := requests.Post(client.BaseURL+"/wallet/createtransaction", reqData, client.RequestsHeader)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(response, &rawTransaction)
	if err != nil {
		return "", err
	}
	if rawTransaction.TxID == "" {
		return "", fmt.Errorf(string(response))
	}

	pk, err := crypto.HexToECDSA(wallet.PrivateKey)
	if err != nil {
		return "", err
	}

	signedTransaction, err := client.signRawTransaction(&rawTransaction, pk)
	if err != nil {
		return "", err
	}

	return client.broadcastTransaction(signedTransaction)
}

func (client *Tron) signRawTransaction(tx *RawTransaction, key *ecdsa.PrivateKey) (*SignedTransaction, error) {
	rawData, err := json.Marshal(tx.RawData)
	if err != nil {
		return &SignedTransaction{}, err
	}

	signedTransaction := &SignedTransaction{
		Visible:    tx.Visible,
		TxID:       tx.TxID,
		RawData:    string(rawData),
		RawDataHex: tx.RawDataHex,
	}
	txIDbytes, err := hex.DecodeString(tx.TxID)
	if err != nil {
		return &SignedTransaction{}, err
	}

	signature, err := crypto.Sign(txIDbytes, key)
	if err != nil {
		return &SignedTransaction{}, err
	}

	signedTransaction.Signature = append(signedTransaction.Signature, hex.EncodeToString(signature))
	return signedTransaction, nil
}

func (client *Tron) broadcastTransaction(tx *SignedTransaction) (string, error) {
	reqData := functions.StructToMap(tx, functions.StructToMapExcludeMode)
	response, err := requests.Post(client.BaseURL+"/wallet/broadcasttransaction", reqData, client.RequestsHeader)
	if err != nil {
		return "", err
	}

	var result struct {
		Result bool
		Txid   string
	}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return "", err
	}
	if !result.Result || result.Txid == "" {
		return "", fmt.Errorf("broadcastTransaction response: %s", string(response))
	}

	return result.Txid, nil
}

func (client *Tron) getTRXBalance(address string) (decimal.Decimal, error) {
	reqData := map[string]interface{}{
		"id":      1,
		"jsonrpc": "2.0",
		"method":  "eth_getBalance",
		"params":  []string{"0x" + base58ToHex(address), "latest"},
	}
	response, err := requests.Post(client.BaseURL+"/jsonrpc", reqData, client.RequestsHeader)

	if err != nil {
		return decimal.Zero, err
	}

	type Result struct {
		Jsonrpc string `json:"jsonrpc"`
		ID      any    `json:"id"`
		Error   struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
		Result string `json:"result,omitempty"`
	}
	result := Result{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		return decimal.Zero, err
	}
	if result.Error.Code != 0 || result.Error.Message != "" {
		return decimal.Zero, fmt.Errorf("%s", result.Error.Message)
	}
	if len(result.Result) < 3 {
		return decimal.Zero, fmt.Errorf("%s", "unexpected response")
	}

	if len(result.Result) <= 2 {
		return decimal.Zero, errors.New("balance value not found")
	}
	i := new(big.Int)
	i.SetString(result.Result[2:], 16)
	sun := decimal.NewFromBigInt(i, 0)
	trx := sun.DivRound(client.Denomination, 18)
	return trx, nil
}
func (client *Tron) GetBalance(address string) (map[string]decimal.Decimal, error) {
	balanceMap := make(map[string]decimal.Decimal)
	var err error

	if balanceMap[string(config.TRX)], err = client.getTRXBalance(address); err != nil {
		return balanceMap, errors.New("获取TRX余额失败")
	}

	if balanceMap[string(config.USDT)], err = client.getUSDTBalance(address); err != nil {
		return balanceMap, errors.New("获取USDT余额失败")
	}

	return balanceMap, nil
}

func (client *Tron) ValidatePrivateKey(privateKey string, address string) bool {
	result, err := privateKeyToAddress(privateKey)
	if err != nil || result != address {
		return false
	}
	return true
}
func (client *Tron) ValidateAddress(address string) bool {
	// 长度校验
	if len(address) != 34 {
		return false
	}

	// 开头校验
	if !strings.HasPrefix(address, "T") {
		return false
	}

	// base58解码校验
	decoded := base58.Decode(address)
	if len(decoded) == 0 {
		return false
	}

	// 校验码
	// Separate the payload and the checksum
	payload := decoded[:len(decoded)-4]
	checksum := decoded[len(decoded)-4:]

	// Double SHA256 hash the payload
	hash1 := sha256.Sum256(payload)
	hash2 := sha256.Sum256(hash1[:])

	// The first 4 bytes of the second hash is the checksum
	computedChecksum := hash2[:4]

	// Compare the computed checksum with the checksum from the address
	return bytes.Equal(checksum, computedChecksum)
}

func (client *Tron) getLatestBlockNum() (int64, error) {
	//url := fmt.Sprintf("%s/walletsolidity/getblock", client.BaseURL)
	url := fmt.Sprintf("%s/wallet/getblock", client.BaseURL)
	data := map[string]interface{}{
		"detail": false,
	}
	respByte, err := requests.Post(url, data, client.RequestsHeader)
	if err != nil {
		my_log.LogError(fmt.Sprintf("Request Err, Url:%s ,Error: %v", url, err))
		return 0, err
	}

	var result struct {
		BlockHeader struct {
			RawData struct {
				Number int64 `json:"number"`
			} `json:"raw_data"`
		} `json:"block_header"`
	}

	err = json.Unmarshal(respByte, &result)
	if err != nil {
		return 0, err
	}

	blockNum := result.BlockHeader.RawData.Number
	if blockNum == 0 {
		return 0, errors.New("获取区块为0")
	}

	return blockNum, nil
}
func (client *Tron) GetScheduleTransfers() ([]models.Transfer, error) {
	endBlockNum, err := client.getLatestBlockNum()
	if err != nil {
		return []models.Transfer{}, err
	}
	startBlockNum, ok := cache.ScheduleCache.Get(scheduleLastKey).(int64)
	if !ok {
		startBlockNum = endBlockNum - 50
	}
	startBlockNum = startBlockNum + 1

	// 限制最多回溯
	if endBlockNum-startBlockNum > 200 {
		startBlockNum = endBlockNum - 200
	}

	bulkSize := int64(50)
	ranges := functions.SplitRangeIntoBulkRanges(startBlockNum, endBlockNum, bulkSize)

	var transactions []models.Transfer

	for _, rangeItem := range ranges {
		bulkTransactions, err := client.getTransactionsByBlockRange(rangeItem.Start, rangeItem.End)
		if err != nil {
			return []models.Transfer{}, err
		}
		transactions = append(transactions, bulkTransactions...)
	}

	// 本轮查询完成
	cache.ScheduleCache.Set(scheduleLastKey, endBlockNum, time.Second*300)

	return transactions, nil
}

// 块的起终不返回最后一个块，即1-2只返回1
func (client *Tron) getTransactionsByBlockRange(startBlockNum int64, endBlockNum int64) ([]models.Transfer, error) {
	my_log.LogDebug(fmt.Sprintf("Tron block range: %d - %d", startBlockNum, endBlockNum))
	var transactions []models.Transfer

	if startBlockNum > endBlockNum {
		return transactions, nil
	}

	//url := fmt.Sprintf("%s/walletsolidity/getblockbylimitnext", client.BaseURL)
	url := fmt.Sprintf("%s/wallet/getblockbylimitnext", client.BaseURL)
	reqData := map[string]interface{}{"startNum": startBlockNum, "endNum": endBlockNum + 1}
	respByte, err := requests.Post(url, reqData, client.RequestsHeader)
	if err != nil {
		my_log.LogError(fmt.Sprintf("Request Err, Url:%s ,Error: %v", url, err))
		return transactions, err
	}

	var blockData blockDataStruct
	err = json.Unmarshal(respByte, &blockData)
	if err != nil {
		return transactions, err
	}

	if len(blockData.Block) == 0 {
		return transactions, errors.New("empty_block")
	}

	for _, block := range blockData.Block {
		blockTimestamp := block.BlockHeader.RawData.Timestamp
		blockTimestamp = blockTimestamp / 1000
		for _, transactionData := range block.Transactions {
			if len(transactionData.RawData.Contract) != 0 {
				contractBlock := transactionData.RawData.Contract[0]
				contractType := contractBlock.Type
				switch contractType {
				case "TriggerSmartContract":
					// 触发智能合约
					var contractValue TriggerSmartContractValue
					err := functions.MapToStruct(contractBlock.Parameter.Value, &contractValue)
					if err != nil {
						return transactions, errors.New("convert_err")
					}

					methodID, toAddress, value := decodeContractData(contractValue.Data)
					fromAddress := hexToBase58(contractValue.OwnerAddress)
					toAddress = hexToBase58("41" + toAddress)

					// 非USDT跳过
					if hexToBase58(contractValue.ContractAddress) != usdtContractAddress {
						continue
					}
					// 非转账合约跳过
					if methodID != "a9059cbb" {
						continue
					}
					amount := decimal.NewFromBigInt(value, 0).Div(decimal.NewFromFloat(ratio))

					transaction := models.NewTransfer(transactionData.TxID, config.USDT, config.TRON, fromAddress, toAddress, amount, blockTimestamp)
					transactions = append(transactions, *transaction)

				case "TransferContract":
					// TRX转账
					var contractValue TransferAssetContractValue
					err := functions.MapToStruct(contractBlock.Parameter.Value, &contractValue)
					if err != nil {
						return transactions, errors.New("convert_err")
					}
					fromAddress := hexToBase58(contractValue.OwnerAddress)
					toAddress := hexToBase58(contractValue.ToAddress)

					amount := decimal.NewFromInt(contractValue.Amount).Div(decimal.NewFromFloat(ratio))

					transaction := models.NewTransfer(transactionData.TxID, config.TRX, config.TRON, fromAddress, toAddress, amount, blockTimestamp)
					transactions = append(transactions, *transaction)

				default:
					continue
				}
			}
		}
	}

	return transactions, err
}
