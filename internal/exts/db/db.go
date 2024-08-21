package db

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/umfaka/tgfaka/internal/exts/config"
	"github.com/umfaka/tgfaka/internal/utils/functions"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	var gormConfig = gorm.Config{}

	if config.SiteConfig.EnableDBDebug {
		gormConfig.Logger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second, // Slow SQL threshold; set it to a low value like 1ns if you want to log all queries
				LogLevel:      logger.Info, // Log level; set to `Info` to log all queries
				Colorful:      true,        // Enable color
			},
		)
	}

	var db *gorm.DB
	var err error

	switch config.DBConfig.DBType {
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai", config.DBConfig.Host, config.DBConfig.Username, config.DBConfig.Password, config.DBConfig.DBName, config.DBConfig.Port)
		db, err = gorm.Open(postgres.Open(dsn), &gormConfig)
	case "sqlite":
		dsn := functions.GetExecutableDir() + "/conf" + "/.db"
		db, err = gorm.Open(sqlite.Open(dsn), &gormConfig)
	default:
		panic("dbname_err")
	}

	if err != nil {
		panic(err)
	}
	return db
}

func InitAllDB() {
	DB = InitDB()
}
