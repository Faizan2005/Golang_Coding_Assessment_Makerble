package main

import (
	"log"

	config "github.com/Faizan2005/Golang_Coding_Assessment_Makerble/config"
	models "github.com/Faizan2005/Golang_Coding_Assessment_Makerble/models"
	routes "github.com/Faizan2005/Golang_Coding_Assessment_Makerble/routes"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	store, err := models.NewPostgresStore(db)
	if err != nil {
		log.Fatal(err)
	}

	listenAddr := ":3000"
	server := routes.NewAPIServer(listenAddr, store)
	server.Run()
}
