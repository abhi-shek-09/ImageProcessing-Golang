package database

import (
	"database/sql"
	"os"
	"log"
	"github.com/joho/godotenv"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func ConnectDB() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_URL")
	var err error
	DB, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatalf("Database not reachable: %v", err)
	}
	log.Println("Postgres is connected successfully!")
}

func CloseDB(){
	if DB != nil {
		DB.Close()
		log.Println("Database connection closed successfully!")
	}
}