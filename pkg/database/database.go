package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

func InitDatabase() {
	username := viper.GetString("DATABASE_USER")
	password := viper.GetString("DATABASE_PASSWORD")
	host := viper.GetString("DATABASE_HOST")
	port := viper.GetInt("DATABASE_PORT")
	dbname := viper.GetString("DATABASE_NAME")
	maxLifetimeConnection := viper.GetDuration("DATABASE_MAX_LIFETIME_CONNECTION")
	maxIdleConnection := viper.GetInt("DATABASE_MAX_IDLE_CONNECTION")
	maxOpenConnection := viper.GetInt("DATABASE_MAX_OPEN_CONNECTION")

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", username, password, host, port, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(maxOpenConnection)
	db.SetMaxIdleConns(maxIdleConnection)
	db.SetConnMaxLifetime(maxLifetimeConnection)

	log.Println("Successfully connected to the database!")
}
