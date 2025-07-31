// internal/service/certificate_service.go
package service

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"certificado-ufgd/internal/domain"
	"certificado-ufgd/internal/repository"
	"certificado-ufgd/pkg/certificate"
	"certificado-ufgd/pkg/security"
)

type CertificateService struct {
	repo         *repository.CertificateRepository
	storagePath  string
	templatePath string
}

// Ajuste a função NewCertificateService
func NewCertificateService(repo *repository.CertificateRepository, storagePath, templatePath string) *CertificateService {
	return &CertificateService{
		repo:         repo,
		storagePath:  storagePath,
		templatePath: templatePath,
	}
}

// CreateEventAndParticipants não muda...
func (s *CertificateService) CreateEventAndParticipants(eventName, organizer, csvPath string) error {
	eventID, err := s.repo.CreateEvent(eventName, organizer)
	if err != nil {
		return fmt.Errorf("falha ao criar evento: %w", err)
	}
	log.Printf("Evento '%s' criado com ID: %d", eventName, eventID)
	file, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("falha ao abrir arquivo CSV: %w", err)
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return fmt.Errorf("falha ao ler arquivo CSV: %w", err)
	}
	for _, record := range records[1:] {
		fullName := record[0]
		cpf := record[1]
		cpfEncrypted, err := security.EncryptAES(cpf)
		if err != nil {
			log.Printf("AVISO: Falha ao criptografar CPF para %s: %v. Pulando...", fullName, err)
			continue
		}
		participant := domain.Participant{
			EventID:      int(eventID),
			FullName:     fullName,
			CPFEncrypted: cpfEncrypted,
		}
		if err := s.repo.AddParticipant(participant); err != nil {
			log.Printf("AVISO: Falha ao adicionar participante %s: %v. Pulando...", fullName, err)
		}
	}
	log.Printf("%d participantes adicionados ao evento.", len(records)-1)
	return nil
}

// GenerateCertificatesForEvent não muda...
func (s *CertificateService) GenerateCertificatesForEvent(eventID int, eventName string) error {
	participants, err := s.repo.GetParticipantsToProcess(eventID)
	if err != nil {
		return fmt.Errorf("falha ao buscar participantes: %w", err)
	}
	if len(participants) == 0 {
		log.Println("Nenhum certificado novo para gerar neste evento.")
		return nil
	}
	log.Printf("Iniciando geração de %d certificados para o evento '%s'...", len(participants), eventName)
	var wg sync.WaitGroup
	concurrencyLimit := 10
	jobQueue := make(chan domain.Participant, len(participants))
	for i := 0; i < concurrencyLimit; i++ {
		wg.Add(1)
		go s.certificateWorker(&wg, jobQueue, eventName)
	}
	for _, p := range participants {
		jobQueue <- p
	}
	close(jobQueue)
	wg.Wait()
	log.Println("Geração de certificados concluída.")
	return nil
}

// certificateWorker é a nossa goroutine que processa a fila de certificados.
func (s *CertificateService) certificateWorker(wg *sync.WaitGroup, jobs <-chan domain.Participant, eventName string) {
	defer wg.Done()

	for p := range jobs {
		cpf, err := security.DecryptAES(p.CPFEncrypted)
		if err != nil {
			log.Printf("ERRO: Falha ao descriptografar CPF para participante ID %d: %v", p.ID, err)
			continue
		}

		certificateID := fmt.Sprintf("UFGD-%d-%d-%d", p.EventID, p.ID, time.Now().UnixNano())

		// Em vez de gerar um PDF em memória, agora geramos um .docx a partir de um template.
		_, docxHash, err := certificate.GenerateCertificateFromDocx(
			s.templatePath,
			s.storagePath,
			p.FullName,
			cpf,
			eventName,
			certificateID,
		)
		if err != nil {
			log.Printf("ERRO: Falha ao gerar DOCX para %s: %v", p.FullName, err)
			continue
		}

		// Atualiza o banco de dados com o hash do novo arquivo .docx
		if err := s.repo.UpdateCertificateHash(p.ID, docxHash); err != nil {
			log.Printf("ERRO: Falha ao atualizar DB para %s: %v", p.FullName, err)
		}
	}
}
