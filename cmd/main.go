package main

import (
	"go-server-template/internal/routes"
	"go-server-template/pkg/db"
	"os"

	"github.com/joho/godotenv"
)

func initEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("ENV file not found")
	}
}

func initServer() {
	initEnv()
	DBString := os.Getenv("DB_CONNECTION_STRING")
	db.RunMigrations(DBString)
	routes.InitializeRouter()
}

func main() {
	initServer()
}
