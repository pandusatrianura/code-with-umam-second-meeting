package database

import (
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func InitDatabase() (*DB, error) {
	username := viper.GetString("DATABASE_USER")
	password := viper.GetString("DATABASE_PASSWORD")
	host := viper.GetString("DATABASE_HOST")
	port := viper.GetInt("DATABASE_PORT")
	dbname := viper.GetString("DATABASE_NAME")
	maxLifetimeConnection := viper.GetDuration("DATABASE_MAX_LIFETIME_CONNECTION")
	maxIdleConnection := viper.GetInt("DATABASE_MAX_IDLE_CONNECTION")
	maxOpenConnection := viper.GetInt("DATABASE_MAX_OPEN_CONNECTION")

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", username, password, host, port, dbname)

	database, err := Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	err = database.DB.Ping()
	if err != nil {
		log.Fatal(err)
	}

	database.DB.SetMaxOpenConns(maxOpenConnection)
	database.DB.SetMaxIdleConns(maxIdleConnection)
	database.DB.SetConnMaxLifetime(maxLifetimeConnection)

	log.Println("Successfully connected to the database!")

	return database, nil
}
