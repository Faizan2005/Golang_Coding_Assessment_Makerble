package models

import "database/sql"

type User struct {
	ID       string `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Email    string `json:"email" db:"email"`
	Password string `json:"-" db:"password"`
	Role     string `json:"role" db:"role"`
}

type Patient struct {
	ID        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	Age       uint   `json:"age" db:"age"`
	Gender    string `json:"gender" db:"gender"`
	Diagnosis string `json:"diagnosis" db:"diagnosis"`
	CreatedBy string `json:"created_by" db:"created_by"`
}

type Storage interface {
	AddPatient(*Patient) error
	GetPatients() ([]*Patient, error)
	GetPatientByID(string) (*Patient, error)
	UpdatePatientByID(string) error
	DeletePatientByID(string) error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) (*PostgresStore, error) {
	return &PostgresStore{
		db: db,
	}, nil
}
