package schedule

import "time"

func StartSchedule() {
	go CheckTransactionSchedule()
	go UpdateExchangeRateSchedule()
	go ClearExpireSchedule()
}

var checkTransactionInterval = time.Second * 30
var updateExchangeRateInterval = time.Second * 600
var clearExpireInterval = time.Second * 35

func ClearExpireSchedule() {
	for {
		clearExpire()
		time.Sleep(clearExpireInterval)
	}
}

func CheckTransactionSchedule() {
	for {
		//TRON
		startCheckTransaction("TRON")
		//POLYGON
		//startCheckTransaction("POLYGON")
		time.Sleep(checkTransactionInterval)
	}
}

func UpdateExchangeRateSchedule() {
	for {
		startUpdateExchangeRate()
		time.Sleep(updateExchangeRateInterval)
	}
}
