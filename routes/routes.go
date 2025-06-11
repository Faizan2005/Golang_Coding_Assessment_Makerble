package routes

import (
	"github.com/Faizan2005/Golang_Coding_Assessment_Makerble/models"
	"github.com/gofiber/fiber/v2"
)

type APIServer struct {
	listenAddr string
	storage    models.Storage
}

func NewAPIServer(listenAddr string) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
	}
}

func (s *APIServer) Run() {
	app := fiber.New()

	api1 := app.Group("/patients")

	api1.Post("/", s.handleAddPatient)
	api1.Get("/", s.handleGetPatients)
	api1.Get("/:id", s.handleGetPatientByID)
	api1.Put("/:id", s.handleUpdatePatientByID)
	api1.Delete("/:id", s.handleDeletePatientByID)

	app.Listen(s.listenAddr)
}

func (s *APIServer) handleAddPatient(c *fiber.Ctx) error {

}

func (s *APIServer) handleGetPatients(c *fiber.Ctx) error {

}

func (s *APIServer) handleGetPatientByID(c *fiber.Ctx) error {

}

func (s *APIServer) handleUpdatePatientByID(c *fiber.Ctx) error {

}

func (s *APIServer) handleDeletePatientByID(c *fiber.Ctx) error {

}
