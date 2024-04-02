package tron

type RawTransaction struct {
	Visible    bool    `json:"visible"`
	TxID       string  `json:"txID"`
	RawData    RawData `json:"raw_data"`
	RawDataHex string  `json:"raw_data_hex"`
}

type SignedTransaction struct {
	Visible    bool     `json:"visible"`
	TxID       string   `json:"txID"`
	RawData    string   `json:"raw_data"`
	RawDataHex string   `json:"raw_data_hex"`
	Signature  []string `json:"signature,omitempty"`
}

type RawData struct {
	Contract []struct {
		Parameter struct {
			Value struct {
				Data            string `json:"data"`
				Amount          int    `json:"amount"`
				OwnerAddress    string `json:"owner_address"`
				ContractAddress string `json:"contract_address"`
				ToAddress       string `json:"to_address"`
			} `json:"value"`
			TypeURL string `json:"type_url"`
		} `json:"parameter"`
		Type string `json:"type"`
	} `json:"contract"`
	RefBlockBytes string `json:"ref_block_bytes"`
	RefBlockHash  string `json:"ref_block_hash"`
	Expiration    int64  `json:"expiration"`
	Timestamp     int64  `json:"timestamp"`
	FeeLimit      int64  `json:"fee_limit"`
}

type blockDataStruct struct {
	Block []struct {
		BlockID     string `json:"blockID"`
		BlockHeader struct {
			RawData struct {
				Number         int64  `json:"number"`
				TxTrieRoot     string `json:"txTrieRoot"`
				WitnessAddress string `json:"witness_address"`
				ParentHash     string `json:"parentHash"`
				Version        int    `json:"version"`
				Timestamp      int64  `json:"timestamp"`
			} `json:"raw_data"`
			WitnessSignature string `json:"witness_signature"`
		} `json:"block_header"`
		Transactions []struct {
			Ret []struct {
				ContractRet string `json:"contractRet"`
			} `json:"ret"`
			Signature []string `json:"signature"`
			TxID      string   `json:"txID"`
			RawData   struct {
				Contract []struct {
					Parameter struct {
						//Value struct {
						//	Amount       int    `json:"amount"`
						//	AssetName    string `json:"asset_name"`
						//	OwnerAddress string `json:"owner_address"`
						//	ToAddress    string `json:"to_address"`
						//} `json:"value"`

						Value map[string]interface{}

						TypeURL string `json:"type_url"`
					} `json:"parameter"`
					Type string `json:"type"`
				} `json:"contract"`
				RefBlockBytes string `json:"ref_block_bytes"`
				RefBlockHash  string `json:"ref_block_hash"`
				Expiration    int64  `json:"expiration"`
				Timestamp     int64  `json:"timestamp"`
			} `json:"raw_data"`
			RawDataHex string `json:"raw_data_hex"`
		} `json:"transactions"`
	} `json:"block"`
}

type TransferAssetContractValue struct {
	Amount       int64  `json:"amount"`
	AssetName    string `json:"asset_name"`
	OwnerAddress string `json:"owner_address"`
	ToAddress    string `json:"to_address"`
}

type TriggerSmartContractValue struct {
	Data            string `json:"data"`
	OwnerAddress    string `json:"owner_address"`
	ContractAddress string `json:"contract_address"`
}
