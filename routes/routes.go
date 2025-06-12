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

// APIServer represents the HTTP server for the application.
type APIServer struct {
	listenAddr string
	storage    models.Storage
	account    models.Account
}

// NewAPIServer creates a new APIServer instance.
func NewAPIServer(listenAddr string, storage models.Storage, account models.Account) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		storage:    storage,
		account:    account,
	}
}

// Run starts the Fiber API server and registers the routes.
func (s *APIServer) Run() {
	app := fiber.New()

	// Public routes for user registration and login
	app.Post("/register", s.handleCreateUserAccount)
	app.Post("/login", s.handleLoginUserAccount)

	// API group protected by JWT authentication middleware
	authGroup := app.Group("/api", auth.JWTMiddleware)

	// Receptionist-specific routes
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

	// Doctor-specific routes
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

// handleAddPatient handles the addition of a new patient by a receptionist.
func (s *APIServer) handleAddPatient(c *fiber.Ctx) error {
	var p models.Patient

	// Temporary struct to parse incoming request body, allowing for diagnosis to be optional.
	var tempPatient struct {
		Name      string  `json:"name"`
		Age       uint    `json:"age"`
		Gender    string  `json:"gender"`
		Diagnosis *string `json:"diagnosis"` // Ptr to allow checking if it was provided
	}

	if err := c.BodyParser(&tempPatient); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Basic validation for required fields
	if tempPatient.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Patient name is required."})
	}
	if tempPatient.Age <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Patient age must be a positive number."})
	}
	if tempPatient.Gender == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Patient gender is required."})
	}

	// Receptionists cannot set diagnosis
	if tempPatient.Diagnosis != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Receptionists cannot set patient diagnosis. Diagnosis is added by doctors."})
	}

	p.Name = tempPatient.Name
	p.Age = tempPatient.Age
	p.Gender = tempPatient.Gender
	p.Diagnosis = sql.NullString{} // Initialize diagnosis as null

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

// handleGetPatients retrieves a list of patients, with optional filtering by name and pagination.
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(patients)
}

// handleGetPatientByID retrieves a single patient's details by their ID.
func (s *APIServer) handleGetPatientByID(c *fiber.Ctx) error {
	id := c.Params("id")

	patient, err := s.storage.GetPatientByID(id)
	if err != nil {
		// More specific error handling for "not found" cases
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Patient details not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(patient)
}

// handleUpdatePatientByID handles updating patient details by a receptionist.
// Receptionists cannot update the diagnosis field.
func (s *APIServer) handleUpdatePatientByID(c *fiber.Ctx) error {
	id := c.Params("id")

	// Temporary struct for partial updates, using pointers for optional fields
	var tempPatientUpdate struct {
		Name      *string `json:"name"`
		Age       *uint   `json:"age"`
		Gender    *string `json:"gender"`
		Diagnosis *string `json:"diagnosis"`
	}

	if err := c.BodyParser(&tempPatientUpdate); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Prevent receptionists from updating diagnosis
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

	// Update fields if provided in the request body
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
	existingPatient.CreatedBy = userID // Update the `CreatedBy` field to the user performing the update

	if err := s.storage.UpdatePatient(existingPatient); err != nil {
		if strings.Contains(err.Error(), "not found for update") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Patient with provided ID not found for update."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to update patient details: %v", err)})
	}

	return c.JSON(existingPatient)
}

// handleDeletePatientByID handles the deletion of a patient by their ID.
func (s *APIServer) handleDeletePatientByID(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := s.storage.DeletePatientByID(id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.SendStatus(fiber.StatusNoContent) // 204 No Content for successful deletion
}

// handleUpdatePatientByDoctor allows a doctor to update a patient's diagnosis.
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

	userID, ok := c.Locals("userID").(string) // Get the doctor's ID
	if ok && userID != "" {
		existingPatient.CreatedBy = userID // Update `CreatedBy` to the doctor who last modified diagnosis
	}

	if err := s.storage.UpdatePatient(existingPatient); err != nil {
		if strings.Contains(err.Error(), "not found for update") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Patient with provided ID not found for update."})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to update diagnosis: %v", err)})
	}

	return c.JSON(existingPatient)
}

// handleExportPatientCSV exports a patient's details as a CSV file.
func (s *APIServer) handleExportPatientCSV(c *fiber.Ctx) error {
	id := c.Params("id")

	patient, err := s.storage.GetPatientByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to retrieve patient for CSV export: %v", err)})
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV header
	header := []string{"ID", "Name", "Age", "Gender", "Diagnosis", "Created By"}
	if err := writer.Write(header); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to write CSV header"})
	}

	// Handle NullString for Diagnosis
	diagnosisValue := ""
	if patient.Diagnosis.Valid {
		diagnosisValue = patient.Diagnosis.String
	}

	// Write patient data row
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

	// Set appropriate headers for CSV download
	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"patient_%s.csv\"", patient.ID))

	return c.Send(buf.Bytes())
}

// handleCreateUserAccount handles the registration of a new user account.
func (s *APIServer) handleCreateUserAccount(c *fiber.Ctx) error {
	var u models.User

	if err := c.BodyParser(&u); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	err := s.account.CreateUserAccount(&u)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	u.Password = "" // Clear password before sending response for security
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user":    u,
	})
}

// handleLoginUserAccount handles user login and generates a JWT token upon successful authentication.
func (s *APIServer) handleLoginUserAccount(c *fiber.Ctx) error {
	var user models.LoginUser

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	dbuser, err := s.account.LoginUserAccount(&user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	tokenString, err := auth.GenerateToken(dbuser)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"token":   "Bearer " + tokenString, // Prefix token with "Bearer " for common usage
		"message": "Login successful",
		"role":    dbuser.Role,
	})
}
