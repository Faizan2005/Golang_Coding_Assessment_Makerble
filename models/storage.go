package models

import (
	"database/sql"
	"fmt"
)

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
	UpdatePatient(string) error
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

func (s *PostgresStore) AddPatient(p *Patient) error {

	query := `INSERT INTO patients (
	name, age, gender, diagnosis, created_by
	)
	VALUES ($1, $2, $3, $4, $5) 
	RETURNING id`

	err := s.db.QueryRow(query, p.Name, p.Age, p.Gender, p.Diagnosis, p.CreatedBy).Scan(&p.ID)
	if err != nil {
		return fmt.Errorf("error inserting patient details: %v", err)
	}
	return nil
}

func (s *PostgresStore) GetPatients() ([]*Patient, error) {

	query := `SELECT id, name, age, gender, diagnosis, created_by FROM patients`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error fetching patients details: %v", err)
	}

	defer rows.Close()

	var Patients []*Patient

	for rows.Next() {
		var p Patient

		err := rows.Scan(&p.ID, &p.Name, &p.Age, &p.Gender, &p.Diagnosis, &p.CreatedBy)
		if err != nil {
			return nil, err
		}

		Patients = append(Patients, &p)
	}
	return Patients, nil
}

func (s *PostgresStore) GetPatientByID(id string) (*Patient, error) {
	query := `SELECT id, name, age, gender, diagnosis, created_by FROM patients WHERE id=$1`

	var p Patient

	err := s.db.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Age, &p.Gender, &p.Diagnosis, &p.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("error fetching patient details: %v", err)
	}

	return &p, nil
}

func (s *PostgresStore) UpdatePatient(p *Patient) error {
	query := `UPDATE patients SET name=$1, age=$2, gender=$3, diagnosis=$4, created_by=$5`
	_, err := s.db.Exec(query, p.Name, p.Age, p.Gender, p.Diagnosis, p.CreatedBy)
	return err
}

func (s *PostgresStore) DeletePatientByID(id string) error {
	query := `DELETE FROM patients WHERE id=$1`

	_, err := s.db.Exec(query, id)
	return err
}
