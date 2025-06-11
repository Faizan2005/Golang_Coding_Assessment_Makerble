package routes

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"strings"

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

	app.Post("/register", s.handleCreateUserAccount)
	app.Post("/login", s.handleLoginUserAccount)

	authGroup := app.Group("/api", auth.JWTMiddleware)

	receptionistGroup := authGroup.Group("/receptionist")
	receptionistGroup.Use(auth.RoleMiddleware("receptionist"))
	{
		receptionistGroup.Post("/patients", s.handleAddPatient)
		receptionistGroup.Get("/patients", s.handleGetPatients)
		receptionistGroup.Get("/patients/:id", s.handleGetPatientByID)
		receptionistGroup.Put("/patients/:id", s.handleUpdatePatientByID)
		receptionistGroup.Delete("/patients/:id", s.handleDeletePatientByID)
		receptionistGroup.Get("/patients/:id/export/csv", s.handleExportPatientCSV)
	}

	doctorGroup := authGroup.Group("/doctor")
	doctorGroup.Use(auth.RoleMiddleware("doctor"))
	{
		doctorGroup.Get("/patients", s.handleGetPatients)
		doctorGroup.Get("/patients/:id", s.handleGetPatientByID)
		doctorGroup.Put("/patients/:id", s.handleUpdatePatientByDoctor)
		doctorGroup.Get("/patients/:id/export/csv", s.handleExportPatientCSV)
	}

	log.Printf("Server listening on %s", s.listenAddr)
	log.Fatal(app.Listen(s.listenAddr))
}

func (s *APIServer) handleAddPatient(c *fiber.Ctx) error {
	var p models.Patient

	var tempPatient struct {
		Name      string  `json:"name"`
		Age       uint    `json:"age"`
		Gender    string  `json:"gender"`
		Diagnosis *string `json:"diagnosis"`
		CreatedBy string  `json:"created_by"`
	}

	if err := c.BodyParser(&tempPatient); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if tempPatient.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Patient name is required."})
	}
	if tempPatient.Age <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Patient age must be a positive number."})
	}
	if tempPatient.Gender == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Patient gender is required."})
	}

	if tempPatient.Diagnosis != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Receptionists cannot set patient diagnosis. Diagnosis is added by doctors."})
	}

	p.Name = tempPatient.Name
	p.Age = tempPatient.Age
	p.Gender = tempPatient.Gender
	p.Diagnosis = sql.NullString{}

	userID, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Authenticated user ID not found"})
	}
	p.CreatedBy = userID

	if err := s.storage.AddPatient(&p); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to add patient: %v", err)})
	}

	return c.Status(fiber.StatusCreated).JSON(p)
}

func (s *APIServer) handleGetPatients(c *fiber.Ctx) error {
	nameQuery := c.Query("name")

	page := c.QueryInt("page", 1)    // Default to page 1
	limit := c.QueryInt("limit", 20) // Default to 20 items per page

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	offset := (page - 1) * limit
	patients, err := s.storage.GetPatients(nameQuery, limit, offset)
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

	var tempPatientUpdate struct {
		Name      *string `json:"name"`
		Age       *uint   `json:"age"`
		Gender    *string `json:"gender"`
		Diagnosis *string `json:"diagnosis"`
		CreatedBy *string `json:"created_by"`
	}

	if err := c.BodyParser(&tempPatientUpdate); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if tempPatientUpdate.Diagnosis != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Receptionists cannot update patient diagnosis. Diagnosis can only be updated by doctors."})
	}

	existingPatient, err := s.storage.GetPatientByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to fetch existing patient for update: %v", err)})
	}

	if tempPatientUpdate.Name != nil {
		existingPatient.Name = *tempPatientUpdate.Name
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Patient name is required for update."})
	}
	if tempPatientUpdate.Age != nil {
		if *tempPatientUpdate.Age <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Patient age must be a positive number."})
		}
		existingPatient.Age = *tempPatientUpdate.Age
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Patient age is required for update."})
	}
	if tempPatientUpdate.Gender != nil {
		existingPatient.Gender = *tempPatientUpdate.Gender
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Patient gender is required for update."})
	}

	userID, ok := c.Locals("userID").(string)
	if !ok {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Authenticated user ID not found in context for update."})
	}
	existingPatient.CreatedBy = userID

	if err := s.storage.UpdatePatient(existingPatient); err != nil {
		fmt.Printf("DEBUG: Error from storage.UpdatePatient: %v\n", err)
		if strings.Contains(err.Error(), "not found for update") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Patient with provided ID not found for update."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to update patient details: %v", err)})
	}

	return c.JSON(existingPatient)
}

func (s *APIServer) handleDeletePatientByID(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.storage.DeletePatientByID(id); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(204)
}

func (s *APIServer) handleUpdatePatientByDoctor(c *fiber.Ctx) error {
	id := c.Params("id")

	var reqBody struct {
		Diagnosis string `json:"diagnosis"`
	}

	if err := c.BodyParser(&reqBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if reqBody.Diagnosis == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Diagnosis field is required for update."})
	}

	existingPatient, err := s.storage.GetPatientByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to fetch existing patient: %v", err)})
	}

	existingPatient.Diagnosis = sql.NullString{String: reqBody.Diagnosis, Valid: true}

	userID := c.Locals("userID").(string) // Doctor's ID
	if userID != "" {
		existingPatient.CreatedBy = userID // Update CreatedBy to the doctor who last modified diagnosis
	}

	if err := s.storage.UpdatePatient(existingPatient); err != nil {
		if strings.Contains(err.Error(), "not found for update") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Patient with provided ID not found for update."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to update diagnosis: %v", err)})
	}

	return c.JSON(existingPatient)
}

func (s *APIServer) handleExportPatientCSV(c *fiber.Ctx) error {
	id := c.Params("id")

	patient, err := s.storage.GetPatientByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	header := []string{"ID", "Name", "Age", "Gender", "Diagnosis", "Created By"}
	if err := writer.Write(header); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to write CSV header"})
	}

	diagnosisValue := ""
	if patient.Diagnosis.Valid {
		diagnosisValue = patient.Diagnosis.String
	}

	dataRow := []string{
		patient.ID,
		patient.Name,
		fmt.Sprintf("%d", patient.Age),
		patient.Gender,
		diagnosisValue,
		patient.CreatedBy,
	}

	if err := writer.Write(dataRow); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to write CSV data"})
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to flush CSV writer: %v", err)})
	}

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"patient_%s.csv\"", patient.ID))

	return c.Send(buf.Bytes())
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

	u.Password = "" // Clear password before sending response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user":    u,
	})
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
		"role":    dbuser.Role,
	})
}
