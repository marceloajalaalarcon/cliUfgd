// internal/domain/models.go
package domain

import "time"

// Event representa um evento acadêmico.
type Event struct {
	ID        int
	Name      string
	Organizer string
	CreatedAt time.Time
}

// Participant representa um participante de um evento.
type Participant struct {
	ID              int
	EventID         int
	FullName        string
	CPFEncrypted    string
	CertificateHash string
	GeneratedAt     time.Time
}
