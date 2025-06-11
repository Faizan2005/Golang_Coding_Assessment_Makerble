package routes

import (
	auth "github.com/Faizan2005/Golang_Coding_Assessment_Makerble/auth"
	"github.com/Faizan2005/Golang_Coding_Assessment_Makerble/models"
	"github.com/gofiber/fiber/v2"
)

type APIServer struct {
	listenAddr string
	storage    models.Storage
	account    models.Account
}

func NewAPIServer(listenAddr string, storage models.Storage, account models.Account) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		storage:    storage,
		account:    account,
	}
}

func (s *APIServer) Run() {
	app := fiber.New()

	api1 := app.Group("/patients", auth.JWTMiddleware)

	api1.Post("/", s.handleAddPatient)
	api1.Get("/", s.handleGetPatients)
	api1.Get("/:id", s.handleGetPatientByID)
	api1.Put("/:id", s.handleUpdatePatientByID)
	api1.Delete("/:id", s.handleDeletePatientByID)

	app.Post("/register", s.handleCreateUserAccount)
	app.Post("/login", s.handleLoginUserAccount)

	app.Listen(s.listenAddr)
}

func (s *APIServer) handleAddPatient(c *fiber.Ctx) error {
	var p models.Patient

	if err := c.BodyParser(&p); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if err := s.storage.AddPatient(&p); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(p)
}

func (s *APIServer) handleGetPatients(c *fiber.Ctx) error {
	patients, err := s.storage.GetPatients()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(patients)
}

func (s *APIServer) handleGetPatientByID(c *fiber.Ctx) error {
	id := c.Params("id")

	patient, err := s.storage.GetPatientByID(id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Patient details not found"})
	}

	return c.JSON(patient)
}

func (s *APIServer) handleUpdatePatientByID(c *fiber.Ctx) error {
	id := c.Params("id")

	var p models.Patient

	if err := c.BodyParser(&p); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	p.ID = id

	if err := s.storage.UpdatePatient(&p); err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Patient details not found"})
	}

	return c.JSON(p)
}

func (s *APIServer) handleDeletePatientByID(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.storage.DeletePatientByID(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(204)
}

func (s *APIServer) handleCreateUserAccount(c *fiber.Ctx) error {
	var u models.User

	if err := c.BodyParser(&u); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	err := s.account.CreateUserAccount(&u)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(u)
}

func (s *APIServer) handleLoginUserAccount(c *fiber.Ctx) error {
	var user models.LoginUser

	if err := c.BodyParser(&user); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	dbuser, err := s.account.LoginUserAccount(&user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	TokenString, err := auth.GenerateToken(dbuser)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"token":   "Bearer " + TokenString,
		"message": "Login successful",
	})

}
