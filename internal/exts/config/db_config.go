package config

import (
	"gopkg.in/ini.v1"
)

const (
	POSTGRES = "postgres"
	SQLITE   = "sqlite"
)

type dBConfigStruct struct {
	Host     string
	Username string
	DBName   string
	Password string
	Port     uint
	DBType   string
}

func LoadDBConfig() {
	path := configBaseDir + "/db_config.ini"
	config := new(dBConfigStruct)

	cfg, err := ini.Load(path)
	if err != nil {
		panic(err)
	}
	err = cfg.MapTo(&config)
	if err != nil {
		panic(err)
	}
	DBConfig = config
}

var DBConfig *dBConfigStruct
