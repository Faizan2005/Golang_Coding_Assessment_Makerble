package routes

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os" // New import for environment variables
	"strings"
	"testing"
	"time" // Import time for patient ID generation

	"github.com/Faizan2005/Golang_Coding_Assessment_Makerble/models" // Assuming models is in this path
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert" // Use testify for easier assertions (optional, but good practice)
	"github.com/stretchr/testify/mock"   // Use testify/mock for mocking (optional, but good practice)
)

// --- MOCK IMPLEMENTATIONS FOR DEPENDENCIES ---

// MockStorage implements models.Storage interface
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) AddPatient(p *models.Patient) error {
	// Simulate ID generation for the mock if not already set
	if p.ID == "" {
		p.ID = fmt.Sprintf("mock-patient-%d", time.Now().UnixNano())
	}
	args := m.Called(p)
	return args.Error(0)
}

// Corrected: Now returns []*models.Patient
func (m *MockStorage) GetPatients(nameQuery string, limit, offset int) ([]*models.Patient, error) {
	args := m.Called(nameQuery, limit, offset)
	// Assert the type coming from the mock setup is []*models.Patient
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Patient), args.Error(1)
}

// Corrected: Now accepts and returns *models.Patient
func (m *MockStorage) GetPatientByID(id string) (*models.Patient, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Patient), args.Error(1)
}

func (m *MockStorage) UpdatePatient(p *models.Patient) error {
	args := m.Called(p)
	return args.Error(0)
}

func (m *MockStorage) DeletePatientByID(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockAccount implements models.Account interface
type MockAccount struct {
	mock.Mock
}

func (m *MockAccount) CreateUserAccount(u *models.User) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockAccount) LoginUserAccount(u *models.LoginUser) (*models.User, error) {
	args := m.Called(u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// --- MOCK AUTHENTICATION MIDDLEWARE ---
// These mocks simulate the behavior of your actual auth middleware
// by setting locals directly, allowing us to test route logic.

// testJWTMiddleware sets a dummy userID in context for testing.
func testJWTMiddleware(c *fiber.Ctx) error {
	// In a real scenario, this would parse and validate a JWT
	// For testing, we just set a dummy user ID.
	c.Locals("userID", "testUserID123")
	return c.Next()
}

// testRoleMiddleware checks if the role matches the expected role for testing.
func testRoleMiddleware(expectedRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// In a real scenario, this would get the role from the JWT
		// For testing, we hardcode the role based on the group.
		// Note: This simple mock implies you apply the correct mock role middleware per group.
		// A more robust mock might read a test header or use a global state.
		// For this example, we assume the test setup applies the correct middleware.
		// We'll set it directly in the context for the specific test scenario.
		role, ok := c.Locals("userRole").(string) // Assume userRole is set before this middleware
		if !ok || role != expectedRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden: Insufficient role permissions."})
		}
		return c.Next()
	}
}

// --- TEST SETUP HELPER ---

// setupTestApp creates a new Fiber app with mocked dependencies for testing.
func setupTestApp(t *testing.T) (*fiber.App, *MockStorage, *MockAccount) {
	app := fiber.New()
	mockStorage := new(MockStorage)
	mockAccount := new(MockAccount)

	server := NewAPIServer(":0", mockStorage, mockAccount) // :0 lets Fiber pick a random port

	// Register public routes
	app.Post("/register", server.handleCreateUserAccount)
	app.Post("/login", server.handleLoginUserAccount)

	// Mock authenticated group and middleware
	authGroup := app.Group("/api", testJWTMiddleware)

	// Mock receptionist group
	receptionistGroup := authGroup.Group("/receptionist")
	// For testing, set a local that `testRoleMiddleware` can check.
	// This is a simple way to simulate the role being present in the token.
	receptionistGroup.Use(func(c *fiber.Ctx) error {
		c.Locals("userRole", "receptionist")
		return c.Next()
	})
	receptionistGroup.Use(testRoleMiddleware("receptionist"))
	{
		receptionistGroup.Post("/patients", server.handleAddPatient)
		receptionistGroup.Get("/patients", server.handleGetPatients)
		receptionistGroup.Get("/patients/:id", server.handleGetPatientByID)
		receptionistGroup.Put("/patients/:id", server.handleUpdatePatientByID)
		receptionistGroup.Delete("/patients/:id", server.handleDeletePatientByID)
		receptionistGroup.Get("/patients/:id/export/csv", server.handleExportPatientCSV)
	}

	// Mock doctor group
	doctorGroup := authGroup.Group("/doctor")
	doctorGroup.Use(func(c *fiber.Ctx) error {
		c.Locals("userRole", "doctor")
		return c.Next()
	})
	doctorGroup.Use(testRoleMiddleware("doctor"))
	{
		doctorGroup.Get("/patients", server.handleGetPatients)
		doctorGroup.Get("/patients/:id", server.handleGetPatientByID)
		doctorGroup.Put("/patients/:id", server.handleUpdatePatientByDoctor)
		doctorGroup.Get("/patients/:id/export/csv", server.handleExportPatientCSV)
	}

	return app, mockStorage, mockAccount
}

// --- TEST FUNCTIONS FOR EACH ROUTE ---

func TestHandleCreateUserAccount(t *testing.T) {
	app, _, mockAccount := setupTestApp(t)

	// Expected user for creation
	newUser := &models.User{
		Email:    "test@example.com",
		Password: "password123",
		Role:     "receptionist",
	}

	// Mock the CreateUserAccount method to return success
	mockAccount.On("CreateUserAccount", mock.AnythingOfType("*models.User")).Return(nil).Once()

	// Convert user to JSON
	jsonUser, _ := json.Marshal(newUser)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(jsonUser))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Equal(t, "User registered successfully", responseBody["message"])

	// Verify that the mock method was called
	mockAccount.AssertExpectations(t)

	// Test invalid request body
	req = httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(`{"email": "bad_json"`))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleLoginUserAccount(t *testing.T) {
	app, _, mockAccount := setupTestApp(t)

	// Set a dummy JWT_SECRET for the test duration
	os.Setenv("JWT_SECRET", "test_secret_key_for_jwt")
	t.Cleanup(func() {
		os.Unsetenv("JWT_SECRET") // Unset after the test
	})

	loginUser := &models.LoginUser{
		Email:    "test@example.com",
		Password: "password123",
	}
	loggedInUser := &models.User{
		Email: "test@example.com",
		Role:  "receptionist",
	}

	// Mock the LoginUserAccount method to return a user and no error
	mockAccount.On("LoginUserAccount", mock.AnythingOfType("*models.LoginUser")).Return(loggedInUser, nil).Once()

	jsonLogin, _ := json.Marshal(loginUser)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonLogin))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var responseBody map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	assert.NoError(t, err)
	assert.Contains(t, responseBody, "token")
	assert.Equal(t, "Login successful", responseBody["message"])
	assert.Equal(t, "receptionist", responseBody["role"])

	mockAccount.AssertExpectations(t)

	// Test failed login
	mockAccount.On("LoginUserAccount", mock.AnythingOfType("*models.LoginUser")).Return(nil, fmt.Errorf("invalid credentials")).Once()
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(jsonLogin))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode) // Assuming 500 for generic login error
}

func TestHandleAddPatient(t *testing.T) {
	app, mockStorage, _ := setupTestApp(t)

	// Mock the AddPatient method to return success
	mockStorage.On("AddPatient", mock.AnythingOfType("*models.Patient")).Return(nil).Once()

	patientData := map[string]interface{}{
		"name":   "John Doe",
		"age":    25,
		"gender": "Male",
	}
	jsonPatient, _ := json.Marshal(patientData)

	req := httptest.NewRequest(http.MethodPost, "/api/receptionist/patients", bytes.NewReader(jsonPatient))
	req.Header.Set("Content-Type", "application/json")
	// The testJWTMiddleware already sets userID, and testRoleMiddleware handles the role check.

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdPatient models.Patient
	err = json.NewDecoder(resp.Body).Decode(&createdPatient)
	assert.NoError(t, err)
	assert.Equal(t, "John Doe", createdPatient.Name)
	assert.Equal(t, uint(25), createdPatient.Age)
	assert.Equal(t, "Male", createdPatient.Gender)
	assert.Equal(t, "testUserID123", createdPatient.CreatedBy) // Set by test middleware

	mockStorage.AssertExpectations(t)

	// Test invalid request body (e.g., missing name)
	invalidPatientData := map[string]interface{}{
		"age":    25,
		"gender": "Male",
	}
	jsonInvalidPatient, _ := json.Marshal(invalidPatientData)
	req = httptest.NewRequest(http.MethodPost, "/api/receptionist/patients", bytes.NewReader(jsonInvalidPatient))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Test with diagnosis (should be rejected for receptionist)
	patientWithDiagnosis := map[string]interface{}{
		"name":      "John Doe",
		"age":       25,
		"gender":    "Male",
		"diagnosis": "Flu",
	}
	jsonPatientWithDiagnosis, _ := json.Marshal(patientWithDiagnosis)
	req = httptest.NewRequest(http.MethodPost, "/api/receptionist/patients", bytes.NewReader(jsonPatientWithDiagnosis))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleGetPatients(t *testing.T) {
	app, mockStorage, _ := setupTestApp(t)

	// Corrected: mockPatients now contains pointers to models.Patient
	mockPatients := []*models.Patient{
		{ID: "p1", Name: "Alice", Age: 20, Gender: "Female"},
		{ID: "p2", Name: "Bob", Age: 30, Gender: "Male"},
	}

	// Mock GetPatients for success
	mockStorage.On("GetPatients", "", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(mockPatients, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/receptionist/patients", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// The response will still be []models.Patient after JSON unmarshalling,
	// as JSON unmarshals to value types by default if the target is a slice of values.
	var patients []models.Patient
	err = json.NewDecoder(resp.Body).Decode(&patients)
	assert.NoError(t, err)
	assert.Len(t, patients, 2)
	assert.Equal(t, "Alice", patients[0].Name)

	mockStorage.AssertExpectations(t)

	// Test with query parameters
	mockStorage.On("GetPatients", "Alice", mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return([]*models.Patient{mockPatients[0]}, nil).Once()
	req = httptest.NewRequest(http.MethodGet, "/api/receptionist/patients?name=Alice&page=1&limit=10", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	err = json.NewDecoder(resp.Body).Decode(&patients)
	assert.NoError(t, err)
	assert.Len(t, patients, 1)
	assert.Equal(t, "Alice", patients[0].Name)

	mockStorage.AssertExpectations(t)
}

func TestHandleGetPatientByID(t *testing.T) {
	app, mockStorage, _ := setupTestApp(t)

	patientID := "existing-patient-id"
	mockPatient := &models.Patient{ID: patientID, Name: "Charlie", Age: 40, Gender: "Male"}

	// Mock GetPatientByID for success
	mockStorage.On("GetPatientByID", patientID).Return(mockPatient, nil).Once()

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/receptionist/patients/%s", patientID), nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var patient models.Patient
	err = json.NewDecoder(resp.Body).Decode(&patient)
	assert.NoError(t, err)
	assert.Equal(t, patientID, patient.ID)
	assert.Equal(t, "Charlie", patient.Name)

	mockStorage.AssertExpectations(t)

	// Test patient not found
	mockStorage.On("GetPatientByID", "non-existent-id").Return(nil, fmt.Errorf("Patient details not found")).Once()
	req = httptest.NewRequest(http.MethodGet, "/api/receptionist/patients/non-existent-id", nil)
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandleUpdatePatientByIDReceptionist(t *testing.T) {
	app, mockStorage, _ := setupTestApp(t)

	patientID := "update-patient-id"
	originalPatient := &models.Patient{ID: patientID, Name: "Old Name", Age: 50, Gender: "Female", CreatedBy: "testUserID123"}
	updatedData := map[string]interface{}{
		"name":   "New Name",
		"age":    51,
		"gender": "Male",
	}
	jsonUpdate, _ := json.Marshal(updatedData)

	// Mock GetPatientByID to return the original patient
	mockStorage.On("GetPatientByID", patientID).Return(originalPatient, nil).Once()
	// Mock UpdatePatient to return success
	mockStorage.On("UpdatePatient", mock.AnythingOfType("*models.Patient")).Return(nil).Once()

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/receptionist/patients/%s", patientID), bytes.NewReader(jsonUpdate))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var resultPatient models.Patient
	err = json.NewDecoder(resp.Body).Decode(&resultPatient)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", resultPatient.Name)
	assert.Equal(t, uint(51), resultPatient.Age)
	assert.Equal(t, "Male", resultPatient.Gender)
	assert.Equal(t, "testUserID123", resultPatient.CreatedBy) // Updated by current user

	mockStorage.AssertExpectations(t)

	// Test update with diagnosis field (should fail for receptionist)
	diagnosisUpdateData := map[string]interface{}{
		"name":      "New Name",
		"age":       51,
		"gender":    "Male",
		"diagnosis": "Some Diagnosis", // Receptionist cannot set this
	}
	jsonDiagnosisUpdate, _ := json.Marshal(diagnosisUpdateData)
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/receptionist/patients/%s", patientID), bytes.NewReader(jsonDiagnosisUpdate))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleDeletePatientByID(t *testing.T) {
	app, mockStorage, _ := setupTestApp(t)

	patientID := "delete-patient-id"

	// Mock DeletePatientByID for success
	mockStorage.On("DeletePatientByID", patientID).Return(nil).Once()

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/receptionist/patients/%s", patientID), nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode) // 204 No Content

	mockStorage.AssertExpectations(t)

	// Test delete failure (e.g., patient not found)
	mockStorage.On("DeletePatientByID", "non-existent-id").Return(fmt.Errorf("Patient with ID not found")).Once()
	req = httptest.NewRequest(http.MethodDelete, "/api/receptionist/patients/non-existent-id", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode) // Assuming 500 for storage error
}

func TestHandleUpdatePatientByDoctor(t *testing.T) {
	app, mockStorage, _ := setupTestApp(t)

	patientID := "doctor-update-patient-id"
	originalPatient := &models.Patient{ID: patientID, Name: "Old Name", Age: 50, Gender: "Female", Diagnosis: sql.NullString{String: "No Diagnosis", Valid: true}, CreatedBy: "oldUserID"}
	doctorDiagnosis := map[string]interface{}{
		"diagnosis": "New Diagnosis: Flu, prescribe rest.",
	}
	jsonDiagnosis, _ := json.Marshal(doctorDiagnosis)

	// Mock GetPatientByID
	mockStorage.On("GetPatientByID", patientID).Return(originalPatient, nil).Once()
	// Mock UpdatePatient
	mockStorage.On("UpdatePatient", mock.AnythingOfType("*models.Patient")).Return(nil).Once()

	// Need to ensure the doctor group middleware is active.
	// The setupTestApp already correctly applies testRoleMiddleware("doctor") to /api/doctor group.
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/doctor/patients/%s", patientID), bytes.NewReader(jsonDiagnosis))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var resultPatient models.Patient
	err = json.NewDecoder(resp.Body).Decode(&resultPatient)
	assert.NoError(t, err)
	assert.True(t, resultPatient.Diagnosis.Valid)
	assert.Equal(t, "New Diagnosis: Flu, prescribe rest.", resultPatient.Diagnosis.String)
	assert.Equal(t, "testUserID123", resultPatient.CreatedBy) // Updated by current doctor ID

	mockStorage.AssertExpectations(t)

	// Test doctor trying to update other fields (should not affect them)
	doctorUpdateOtherFields := map[string]interface{}{
		"name":      "Changed by Doctor?", // This should be ignored by handler
		"diagnosis": "Diagnosis by Doctor 2",
	}
	jsonDoctorUpdateOtherFields, _ := json.Marshal(doctorUpdateOtherFields)

	// Mock GetPatientByID (resetting originalPatient for the next mock)
	mockStorage.On("GetPatientByID", patientID).Return(originalPatient, nil).Once()
	// Mock UpdatePatient
	mockStorage.On("UpdatePatient", mock.AnythingOfType("*models.Patient")).Return(nil).Once()

	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/doctor/patients/%s", patientID), bytes.NewReader(jsonDoctorUpdateOtherFields))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&resultPatient)
	assert.NoError(t, err)
	assert.Equal(t, "Diagnosis by Doctor 2", resultPatient.Diagnosis.String)
	// Assert that Name was not changed (it remains "Old Name" as it was not updated in handler)
	assert.Equal(t, "Old Name", resultPatient.Name) // Should retain original name
	mockStorage.AssertExpectations(t)

	// Test missing diagnosis field
	invalidDoctorUpdate := map[string]interface{}{
		"some_other_field": "value",
	}
	jsonInvalidDoctorUpdate, _ := json.Marshal(invalidDoctorUpdate)
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/doctor/patients/%s", patientID), bytes.NewReader(jsonInvalidDoctorUpdate))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandleExportPatientCSV(t *testing.T) {
	app, mockStorage, _ := setupTestApp(t)

	patientID := "csv-patient-id"
	mockPatient := &models.Patient{
		ID:        patientID,
		Name:      "CSV Export User",
		Age:       60,
		Gender:    "Female",
		Diagnosis: sql.NullString{String: "Chronic cough", Valid: true},
		CreatedBy: "creator123",
	}

	// Mock GetPatientByID for success
	mockStorage.On("GetPatientByID", patientID).Return(mockPatient, nil).Once()

	// Test as Receptionist
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/receptionist/patients/%s/export/csv", patientID), nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/csv", resp.Header.Get("Content-Type"))
	assert.Contains(t, resp.Header.Get("Content-Disposition"), fmt.Sprintf("attachment; filename=\"patient_%s.csv\"", patientID))

	csvContent, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	reader := csv.NewReader(bytes.NewReader(csvContent))
	records, err := reader.ReadAll()
	assert.NoError(t, err)
	assert.Len(t, records, 2) // Header + 1 data row
	assert.Equal(t, []string{"ID", "Name", "Age", "Gender", "Diagnosis", "Created By"}, records[0])
	assert.Equal(t, []string{patientID, "CSV Export User", "60", "Female", "Chronic cough", "creator123"}, records[1])

	mockStorage.AssertExpectations(t)

	// Test as Doctor (should have access to this route)
	mockStorage.On("GetPatientByID", patientID).Return(mockPatient, nil).Once() // Re-mock for doctor test
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/doctor/patients/%s/export/csv", patientID), nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/csv", resp.Header.Get("Content-Type"))
	mockStorage.AssertExpectations(t)

	// Test patient not found
	mockStorage.On("GetPatientByID", "non-existent-csv-id").Return(nil, fmt.Errorf("Patient details not found")).Once()
	req = httptest.NewRequest(http.MethodGet, "/api/receptionist/patients/non-existent-csv-id/export/csv", nil)
	resp, err = app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// --- Test Cases for Role-Based Access Control (Brief) ---

func TestRoleMiddlewareAccess(t *testing.T) {
	app, _, _ := setupTestApp(t) // Get the shared app and mocks

	// --- Scenario 1: Attempt doctor access to receptionist-only route (POST /patients) ---
	// Expect 405 Method Not Allowed because /api/doctor/patients path exists for GET/PUT/CSV,
	// but POST is not defined for it.
	t.Run("DoctorAttemptsReceptionistPost", func(t *testing.T) {
		patientData := map[string]interface{}{
			"name":   "Forbidden Patient",
			"age":    30,
			"gender": "Male",
		}
		jsonPatient, _ := json.Marshal(patientData)

		req := httptest.NewRequest(http.MethodPost, "/api/doctor/patients", bytes.NewReader(jsonPatient))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode) // Changed from 404 to 405
	})

	// --- Scenario 2: Attempt receptionist access to doctor-only update (PUT /patients/:id with diagnosis) ---
	t.Run("ReceptionistAttemptsDoctorDiagnosisUpdate", func(t *testing.T) {
		patientID := "any-patient-id"
		doctorDiagnosis := map[string]interface{}{
			"diagnosis": "Attempted diagnosis by receptionist.",
		}
		jsonDiagnosis, _ := json.Marshal(doctorDiagnosis)

		// We do NOT mock GetPatientByID here because the handler returns early
		// if a receptionist attempts to set a diagnosis.
		// mockStorage.On("GetPatientByID", patientID).Return(mockPatient, nil).Once() // REMOVED THIS LINE

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/receptionist/patients/%s", patientID), bytes.NewReader(jsonDiagnosis))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req) // Use the main 'app' instance
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		var responseBody map[string]string
		json.NewDecoder(resp.Body).Decode(&responseBody)
		assert.Equal(t, "Receptionists cannot update patient diagnosis. Diagnosis can only be updated by doctors.", responseBody["error"])

		// No mock expectation to assert for GetPatientByID in this specific error path.
		// mockStorage.AssertExpectations(t) // This line will now only assert other mocks if any, or pass if none.
	})
}
