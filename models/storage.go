package models

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       string `json:"id" db:"id"`
	Name     string `json:"name" db:"name"`
	Email    string `json:"email" db:"email"`
	Password string `json:"-" db:"password"`
	Role     string `json:"role" db:"role"`
}

type LoginUser struct {
	Email    string `json:"email" db:"email"`
	Password string `json:"-" db:"password"`
}

type Patient struct {
	ID        string         `json:"id" db:"id"`
	Name      string         `json:"name" db:"name"`
	Age       uint           `json:"age" db:"age"`
	Gender    string         `json:"gender" db:"gender"`
	Diagnosis sql.NullString `json:"diagnosis" db:"diagnosis"`
	CreatedBy string         `json:"created_by" db:"created_by"`
}

type Storage interface {
	AddPatient(*Patient) error
	GetPatients(string, int, int) ([]*Patient, error)
	GetTotalPatientsCount(string) (int, error)
	GetPatientByID(string) (*Patient, error)
	UpdatePatient(*Patient) error
	DeletePatientByID(string) error
}

type Account interface {
	CreateUserAccount(*User) error
	LoginUserAccount(*LoginUser) (*User, error)
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
	name, age, gender, created_by
	)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	err := s.db.QueryRow(query, p.Name, p.Age, p.Gender, p.CreatedBy).Scan(&p.ID)
	if err != nil {
		return fmt.Errorf("error inserting patient details: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetPatients(nameQuery string, limit, offset int) ([]*Patient, error) {
	query := `SELECT id, name, age, gender, diagnosis, created_by FROM patients`
	args := []interface{}{}
	paramCount := 1

	if nameQuery != "" {
		query += fmt.Sprintf(" WHERE name ILIKE '%%' || $%d || '%%'", paramCount)
		args = append(args, nameQuery)
		paramCount++
	}

	query += fmt.Sprintf(" ORDER BY name ASC, id ASC LIMIT $%d OFFSET $%d", paramCount, paramCount+1)
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error fetching patients details: %w", err)
	}
	defer rows.Close()

	var Patients []*Patient
	for rows.Next() {
		var p Patient
		err := rows.Scan(&p.ID, &p.Name, &p.Age, &p.Gender, &p.Diagnosis, &p.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("error scanning patient row: %w", err)
		}
		Patients = append(Patients, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error after scanning rows: %w", err)
	}
	return Patients, nil
}

func (s *PostgresStore) GetTotalPatientsCount(nameQuery string) (int, error) {
	query := `SELECT COUNT(*) FROM patients`
	args := []interface{}{}
	paramCount := 1

	if nameQuery != "" {
		query += fmt.Sprintf(" WHERE name ILIKE '%%' || $%d || '%%'", paramCount)
		args = append(args, nameQuery)
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting total patient count: %w", err)
	}
	return count, nil
}

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

func (s *PostgresStore) UpdatePatient(p *Patient) error {
	query := `UPDATE patients SET name=$1, age=$2, gender=$3, diagnosis=$4, created_by=$5 WHERE id=$6`

	res, err := s.db.Exec(query, p.Name, p.Age, p.Gender, p.Diagnosis, p.CreatedBy, p.ID)
	if err != nil {
		return fmt.Errorf("error updating patient details: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("patient with ID %s not found for update", p.ID)
	}
	return nil
}

func (s *PostgresStore) DeletePatientByID(id string) error {
	query := `DELETE FROM patients WHERE id=$1`
	res, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting patient: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("patient with ID %s not found for deletion", id)
	}
	return nil
}

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
		return err
	}

	return nil
}

func (s *PostgresStore) LoginUserAccount(u *LoginUser) (*User, error) {
	var dbuser User

	query := `SELECT id, name, email, password, role FROM users WHERE email=$1`

	err := s.db.QueryRow(query, u.Email).Scan(&dbuser.ID, &dbuser.Name, &dbuser.Email, &dbuser.Password, &dbuser.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, err
	}

	if !checkPassword(dbuser.Password, u.Password) {
		return nil, fmt.Errorf("invalid email or password")
	}

	return &dbuser, nil
}

func hashPassword(u *User) ([]byte, error) {
	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return password, nil
}

func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
