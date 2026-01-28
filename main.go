package main

import (
	"fmt"
	"log"

	"github.com/common-nighthawk/go-figure"
	"github.com/pandusatrianura/code-with-umam-second-meeting/api"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/config"
	"github.com/pandusatrianura/code-with-umam-second-meeting/pkg/database"
	"github.com/spf13/viper"
)

// @title Kasir API
// @version 1.0
// @host kasir-api-pandusatrianura.up.railway.app/
// @BasePath /

func main() {
	config.InitConfig()

	myFigure := figure.NewFigure("Kasir API", "rectangles", true)
	myFigure.Print()
	fmt.Println()
	fmt.Println("==========================================================")

	port := viper.GetString("PORT")

	db, err := database.InitDatabase()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	server := api.NewAPIServer(fmt.Sprintf(":%s", port), db)
	if err := server.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
