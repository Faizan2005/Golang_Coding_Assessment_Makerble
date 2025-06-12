package models

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user account in the system.
type User struct {
	ID       string `json:"id" db:"id"`       // Unique identifier for the user.
	Name     string `json:"name" db:"name"`   // Name of the user.
	Email    string `json:"email" db:"email"` // Email address of the user (used for login).
	Password string `json:"-" db:"password"`  // Hashed password, excluded from JSON output.
	Role     string `json:"role" db:"role"`   // Role of the user (e.g., "receptionist", "doctor").
}

// LoginUser represents the data structure for user login requests.
type LoginUser struct {
	Email    string `json:"email" db:"email"` // Email provided for login.
	Password string `json:"-" db:"password"`  // Password provided for login, excluded from JSON output.
}

// Patient represents a patient record in the system.
type Patient struct {
	ID        string         `json:"id" db:"id"`                 // Unique identifier for the patient.
	Name      string         `json:"name" db:"name"`             // Name of the patient.
	Age       uint           `json:"age" db:"age"`               // Age of the patient.
	Gender    string         `json:"gender" db:"gender"`         // Gender of the patient.
	Diagnosis sql.NullString `json:"diagnosis" db:"diagnosis"`   // Patient's diagnosis, can be null.
	CreatedBy string         `json:"created_by" db:"created_by"` // User ID of who created/last updated the patient.
}

// Storage defines the interface for patient data persistence operations.
type Storage interface {
	AddPatient(*Patient) error
	GetPatients(nameQuery string, limit, offset int) ([]*Patient, error)
	GetPatientByID(id string) (*Patient, error)
	UpdatePatient(*Patient) error
	DeletePatientByID(id string) error
}

// Account defines the interface for user account management operations.
type Account interface {
	CreateUserAccount(*User) error
	LoginUserAccount(*LoginUser) (*User, error)
}

// PostgresStore implements the Storage interface for PostgreSQL database.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new PostgresStore instance.
func NewPostgresStore(db *sql.DB) (*PostgresStore, error) {
	return &PostgresStore{
		db: db,
	}, nil
}

// AddPatient inserts a new patient record into the database.
func (s *PostgresStore) AddPatient(p *Patient) error {
	query := `INSERT INTO patients (
		name, age, gender, created_by
	)
	VALUES ($1, $2, $3, $4)
	RETURNING id` // RETURNING id ensures the generated ID is populated back into p.ID

	err := s.db.QueryRow(query, p.Name, p.Age, p.Gender, p.CreatedBy).Scan(&p.ID)
	if err != nil {
		return fmt.Errorf("error inserting patient details: %w", err)
	}
	return nil
}

// GetPatients retrieves a list of patients from the database.
// It supports filtering by name (case-insensitive partial match) and pagination.
func (s *PostgresStore) GetPatients(nameQuery string, limit, offset int) ([]*Patient, error) {
	query := `SELECT id, name, age, gender, diagnosis, created_by FROM patients`
	args := []interface{}{}
	paramCount := 1

	// Add name filtering if nameQuery is provided
	if nameQuery != "" {
		query += fmt.Sprintf(" WHERE name ILIKE '%%' || $%d || '%%'", paramCount)
		args = append(args, nameQuery)
		paramCount++
	}

	// Add ordering and pagination
	query += fmt.Sprintf(" ORDER BY name ASC, id ASC LIMIT $%d OFFSET $%d", paramCount, paramCount+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error fetching patients details: %w", err)
	}
	defer rows.Close()

	var patients []*Patient
	for rows.Next() {
		var p Patient
		err := rows.Scan(&p.ID, &p.Name, &p.Age, &p.Gender, &p.Diagnosis, &p.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("error scanning patient row: %w", err)
		}
		patients = append(patients, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after scanning rows: %w", err)
	}
	return patients, nil
}

// GetPatientByID retrieves a single patient record by their unique ID.
func (s *PostgresStore) GetPatientByID(id string) (*Patient, error) {
	query := `SELECT id, name, age, gender, diagnosis, created_by FROM patients WHERE id=$1`

	var p Patient

	err := s.db.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Age, &p.Gender, &p.Diagnosis, &p.CreatedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("patient with ID %s not found", id)
		}
		return nil, fmt.Errorf("error fetching patient details by ID: %w", err)
	}
	return &p, nil
}

// UpdatePatient updates an existing patient record in the database.
func (s *PostgresStore) UpdatePatient(p *Patient) error {
	query := `UPDATE patients SET name=$1, age=$2, gender=$3, diagnosis=$4, created_by=$5 WHERE id=$6`

	res, err := s.db.Exec(query, p.Name, p.Age, p.Gender, p.Diagnosis, p.CreatedBy, p.ID)
	if err != nil {
		return fmt.Errorf("error updating patient details: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected during update: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("patient with ID %s not found for update", p.ID)
	}
	return nil
}

// DeletePatientByID deletes a patient record from the database by their unique ID.
func (s *PostgresStore) DeletePatientByID(id string) error {
	query := `DELETE FROM patients WHERE id=$1`
	res, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting patient: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected during deletion: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("patient with ID %s not found for deletion", id)
	}
	return nil
}

// CreateUserAccount inserts a new user account into the database after hashing the password.
func (s *PostgresStore) CreateUserAccount(u *User) error {
	hashedPassword, err := hashPassword(u)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `INSERT INTO users (name, email, password, role)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	err = s.db.QueryRow(query, u.Name, u.Email, hashedPassword, u.Role).Scan(&u.ID)
	if err != nil {
		return fmt.Errorf("error creating user account: %w", err)
	}

	return nil
}

// LoginUserAccount authenticates a user by checking their email and password.
func (s *PostgresStore) LoginUserAccount(u *LoginUser) (*User, error) {
	var dbuser User

	query := `SELECT id, name, email, password, role FROM users WHERE email=$1`

	err := s.db.QueryRow(query, u.Email).Scan(&dbuser.ID, &dbuser.Name, &dbuser.Email, &dbuser.Password, &dbuser.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid email or password") // Generic error for security
		}
		return nil, fmt.Errorf("error retrieving user for login: %w", err)
	}

	// Compare the provided password with the hashed password from the database
	if !checkPassword(dbuser.Password, u.Password) {
		return nil, fmt.Errorf("invalid email or password") // Generic error for security
	}

	return &dbuser, nil
}

// hashPassword generates a bcrypt hash of the user's password.
func hashPassword(u *User) ([]byte, error) {
	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypt hashing failed: %w", err)
	}
	return password, nil
}

// checkPassword compares a plaintext password with a bcrypt hashed password.
func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
