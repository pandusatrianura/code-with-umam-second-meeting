package config

import (
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func InitConfig() {
	v := viper.New()

	if _, err := os.Stat(".env"); err == nil {
		v.SetConfigFile(".env")
	}

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}
