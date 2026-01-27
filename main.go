package main

import (
	"log"

	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/config"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/database"
	"github.com/spf13/viper"
)

func main() {
	config.InitConfig()
	database.InitDatabase()

	port := viper.GetString("PORT")
	log.Println("Server started successfully on port:", port)
}
