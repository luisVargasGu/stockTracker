package db

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

func DbConnect(log *zap.Logger) *sqlx.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	psqlconn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	db, err := sqlx.Open("postgres", psqlconn)
	if err != nil {
		log.Fatal("Error connecting to the database:", zap.Error(err))
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error pinging the database:", zap.Error(err))
	}

	log.Info("Successfully connected!")
	return db
}
