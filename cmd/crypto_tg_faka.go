package main

import (
	"flag"
	"fmt"
	"gopay/internal/exts/db"
	"gopay/internal/models"
	"gopay/internal/router"
	"gopay/internal/utils/schedule"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := router.SetupRoutes()

	if err := db.DB.AutoMigrate(
		&models.Order{},
		&models.Transfer{},
		&models.Wallet{},
		&models.User{},
		&models.Product{},
		&models.ProductItem{},
		&models.BalanceHistory{},
		&models.UserBalance{},
		// &models.UserFans{},
		&models.ChatInviteLinkHistory{},
	); err != nil {
		panic(err)
	}

	go router.RunTgBot()
	schedule.StartSchedule()

	port := flag.Int("port", 8082, "Port on which the server will run")
	flag.Parse()
	host := fmt.Sprintf(":%d", *port)
	fmt.Println("运行在 " + host)

	if err := r.Run(host); err != nil {
		panic(err)
	}
}
