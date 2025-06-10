package main

import (
	routes "github.com/Faizan2005/Golang_Coding_Assessment_Makerble/routes"
)

func main() {
	apiServer := routes.NewAPIServer(":3000")

	apiServer.Run()
}
