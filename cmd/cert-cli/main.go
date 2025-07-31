// cmd/cert-cli/main.go
package main

import (
	"fmt"
	"log"
	"os"

	"certificado-ufgd/internal/repository"
	"certificado-ufgd/internal/service"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

const dbPath = "./storage/db/ufgd.db"
const certificatesPath = "./storage/certificates"
const templatePath = "./templates/template_certificado.docx"

func main() {
	// Garante que os diretórios de armazenamento existam
	_ = os.MkdirAll("./storage/db", os.ModePerm)
	_ = os.MkdirAll("./storage/certificates", os.ModePerm)

	// Inicializa o repositório
	repo, err := repository.NewCertificateRepository(dbPath)
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}

	// Garante que as tabelas do banco de dados existam
	if err := repo.InitDB(); err != nil {
		log.Fatalf("Erro ao inicializar as tabelas do banco: %v", err)
	}

	// Inicializa o serviço
	certService := service.NewCertificateService(repo, certificatesPath, templatePath)

	// Configuração da CLI com Cobra
	var rootCmd = &cobra.Command{Use: "cert-cli"}

	var eventName, organizer, csvPath string
	var createEventCmd = &cobra.Command{
		Use:   "create-event",
		Short: "Cria um novo evento e cadastra participantes via CSV.",
		Run: func(cmd *cobra.Command, args []string) {
			if eventName == "" || organizer == "" || csvPath == "" {
				log.Fatal("Os parâmetros --name, --organizer e --csv são obrigatórios.")
			}
			if err := certService.CreateEventAndParticipants(eventName, organizer, csvPath); err != nil {
				log.Fatalf("Erro ao processar evento e participantes: %v", err)
			}
			fmt.Println("Evento e participantes criados com sucesso!")
		},
	}
	createEventCmd.Flags().StringVarP(&eventName, "name", "n", "", "Nome do evento (obrigatório)")
	createEventCmd.Flags().StringVarP(&organizer, "organizer", "o", "", "Organizador do evento (obrigatório)")
	createEventCmd.Flags().StringVarP(&csvPath, "csv", "c", "", "Caminho para o arquivo CSV de participantes (obrigatório)")

	var eventID int
	var generateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Gera os certificados para um evento.",
		Run: func(cmd *cobra.Command, args []string) {
			if eventID == 0 || eventName == "" {
				log.Fatal("Os parâmetros --id e --name do evento são obrigatórios.")
			}
			if err := certService.GenerateCertificatesForEvent(eventID, eventName); err != nil {
				log.Fatalf("Erro ao gerar certificados: %v", err)
			}
		},
	}
	generateCmd.Flags().IntVarP(&eventID, "id", "i", 0, "ID do evento (obrigatório)")
	generateCmd.Flags().StringVarP(&eventName, "name", "n", "", "Nome do evento (obrigatório)")

	rootCmd.AddCommand(createEventCmd, generateCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
