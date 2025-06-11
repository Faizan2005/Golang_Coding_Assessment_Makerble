package main

import (
	"log"

	config "github.com/Faizan2005/Golang_Coding_Assessment_Makerble/config"
	models "github.com/Faizan2005/Golang_Coding_Assessment_Makerble/models"
	routes "github.com/Faizan2005/Golang_Coding_Assessment_Makerble/routes"
)

func main() {
	db, _ := config.ConnectDB()

	store, err := models.NewPostgresStore(db)
	if err != nil {
		log.Fatal(err)
	}

	listenAddr := ":3000"
	server := routes.NewAPIServer(listenAddr, store)
	server.Run()
}
