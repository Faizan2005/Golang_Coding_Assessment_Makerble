package main

import (
	"log"
	"os"

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

	listenPort := os.Getenv("PORT")
	if listenPort == "" {
		listenPort = "3000"
	}
	listenAddr := ":" + listenPort

	log.Printf("Application starting on %s", listenAddr)

	db, err := config.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	defer db.Close()

	store, err := models.NewPostgresStore(db)
	if err != nil {
		log.Fatal(err)
	}

	server := routes.NewAPIServer(listenAddr, store, store)
	server.Run()
}
