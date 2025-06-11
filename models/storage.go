package models

type Storage interface {
	AddPatient(*Patient) error
	GetPatients() ([]*Patient, error)
	GetPatientByID(string) (*Patient, error)
	UpdatePatientByID(string) error
	DeletePatientByID(string) error
}
