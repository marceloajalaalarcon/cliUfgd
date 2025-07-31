// internal/repository/sqlite_repo.go
package repository

import (
	"certificado-ufgd/internal/domain"
	"database/sql"
	"log"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

type CertificateRepository struct {
	db      *sql.DB
	dbMutex sync.Mutex
}

func NewCertificateRepository(dbPath string) (*CertificateRepository, error) {
	// 3. Habilitamos o modo WAL na string de conexão para melhor performance de concorrência.
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	log.Println("Conexão com o banco de dados SQLite estabelecida.")
	return &CertificateRepository{db: db}, nil
}

// InitDB cria as tabelas necessárias se elas não existirem.
func (r *CertificateRepository) InitDB() error {
	query := `
    CREATE TABLE IF NOT EXISTS events (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        organizer TEXT NOT NULL,
        created_at DATETIME NOT NULL
    );
    CREATE TABLE IF NOT EXISTS participants (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        event_id INTEGER NOT NULL,
        full_name TEXT NOT NULL,
        cpf_encrypted TEXT NOT NULL,
        certificate_hash TEXT,
        generated_at DATETIME,
        FOREIGN KEY (event_id) REFERENCES events(id)
    );
    `
	_, err := r.db.Exec(query)
	return err
}

// CreateEvent cria um novo evento.
func (r *CertificateRepository) CreateEvent(name, organizer string) (int64, error) {
	res, err := r.db.Exec("INSERT INTO events (name, organizer, created_at) VALUES (?, ?, ?)", name, organizer, time.Now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// AddParticipant adiciona um novo participante a um evento.
func (r *CertificateRepository) AddParticipant(p domain.Participant) error {
	_, err := r.db.Exec("INSERT INTO participants (event_id, full_name, cpf_encrypted) VALUES (?, ?, ?)",
		p.EventID, p.FullName, p.CPFEncrypted)
	return err
}

// GetParticipantsToProcess busca participantes que ainda não têm certificado.
func (r *CertificateRepository) GetParticipantsToProcess(eventID int) ([]domain.Participant, error) {
	rows, err := r.db.Query("SELECT id, event_id, full_name, cpf_encrypted FROM participants WHERE event_id = ? AND certificate_hash IS NULL", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []domain.Participant
	for rows.Next() {
		var p domain.Participant
		if err := rows.Scan(&p.ID, &p.EventID, &p.FullName, &p.CPFEncrypted); err != nil {
			return nil, err
		}
		participants = append(participants, p)
	}
	return participants, nil
}

// UpdateCertificateHash atualiza os dados do certificado de um participante de forma segura.
func (r *CertificateRepository) UpdateCertificateHash(participantID int, hash string) error {
	// 4. Usamos o Mutex para garantir que apenas uma goroutine escreva no banco por vez.
	r.dbMutex.Lock()
	defer r.dbMutex.Unlock()

	_, err := r.db.Exec("UPDATE participants SET certificate_hash = ?, generated_at = ? WHERE id = ?",
		hash, time.Now(), participantID)
	return err
}
