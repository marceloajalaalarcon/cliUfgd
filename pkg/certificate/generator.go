package certificate

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"certificado-ufgd/pkg/security"

	"github.com/nguyenthenguyen/docx"
	"github.com/skip2/go-qrcode"
)

// Função para debugar e listar todas as imagens no DOCX
func listImagesInDocx(templatePath string) {
	log.Println("=== DEBUG: Listando imagens no template ===")

	// DOCX é um arquivo ZIP, vamos inspecionar seu conteúdo
	r, err := zip.OpenReader(templatePath)
	if err != nil {
		log.Printf("Erro ao abrir DOCX como ZIP: %v", err)
		return
	}
	defer r.Close()

	for _, f := range r.File {
		// Procurar por arquivos de imagem
		if strings.Contains(f.Name, "media/") || strings.Contains(f.Name, "image") {
			log.Printf("Arquivo de mídia encontrado: %s", f.Name)
		}

		// Procurar por relacionamentos (relationships)
		if strings.Contains(f.Name, "rels") {
			log.Printf("Arquivo de relacionamento: %s", f.Name)
		}

		// Procurar no document.xml por referências de imagem
		if strings.Contains(f.Name, "document.xml") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			content, err := ioutil.ReadAll(rc)
			rc.Close()
			if err != nil {
				continue
			}

			contentStr := string(content)
			if strings.Contains(contentStr, "QRCODE") {
				log.Printf("Referência 'QRCODE' encontrada em document.xml")
			}

			// Procurar por tags de desenho/imagem
			if strings.Contains(contentStr, "drawing") || strings.Contains(contentStr, "pic:pic") {
				log.Printf("Elementos de desenho/imagem encontrados em document.xml")
			}
		}
	}
	log.Println("=== FIM DEBUG ===")
}

func GenerateCertificateFromDocx(templatePath, outputDir, fullName, cpf, eventName, certificateID string) (string, string, error) {
	// DEBUG: Listar imagens no template
	listImagesInDocx(templatePath)

	// 1. Gerar QR Code
	expiration := time.Now().Add(365 * 24 * time.Hour).Unix()
	qrDataPayload := fmt.Sprintf("%s;%s;%d", certificateID, cpf, expiration)
	signature := security.SignData([]byte(qrDataPayload))
	qrDataWithSignature := fmt.Sprintf("%s;%s", qrDataPayload, signature)

	qrImagePath := filepath.Join(os.TempDir(), "qr_temp.png")
	err := qrcode.WriteFile(qrDataWithSignature, qrcode.Medium, 256, qrImagePath)
	if err != nil {
		return "", "", fmt.Errorf("falha ao gerar QR Code: %w", err)
	}
	defer os.Remove(qrImagePath)

	// 2. Abrir template
	r, err := docx.ReadDocxFile(templatePath)
	if err != nil {
		return "", "", fmt.Errorf("falha ao ler template: %w", err)
	}
	defer r.Close()

	docxFile := r.Editable()

	// 3. Substituir texto
	docxFile.Replace("{{NOME_COMPLETO}}", fullName, -1)
	docxFile.Replace("{{CPF}}", formatCPF(cpf), -1)
	docxFile.Replace("{{EVENTO}}", eventName, -1)
	docxFile.Replace("{{ID_CERTIFICADO}}", certificateID, -1)
	docxFile.Replace("{{DATA_GERACAO}}", time.Now().Format("02 de January de 2006"), -1)

	// 4. Tentar múltiplas variações de nomes para a imagem
	possibleImageNames := []string{
		"QRCODE",     // Descrição que você colocou
		"qrcode",     // minúscula
		"image1.png", // Nome padrão do Word
		"image2.png",
		"image3.png",
		"image1", // Sem extensão
		"image2",
		"image3",
		"rId1", // IDs de relacionamento
		"rId2",
		"rId3",
		"rId4",
		"rId5",
	}

	imageReplaced := false
	for _, imageName := range possibleImageNames {
		log.Printf("Tentando substituir imagem com nome: '%s'", imageName)
		err = docxFile.ReplaceImage(imageName, qrImagePath)
		if err == nil {
			log.Printf("✅ SUCESSO! QR Code inserido substituindo imagem '%s'", imageName)
			imageReplaced = true
			break
		} else {
			log.Printf("❌ Falhou com '%s': %v", imageName, err)
		}
	}

	if !imageReplaced {
		log.Printf("⚠️ AVISO: Não foi possível substituir nenhuma imagem. Inserindo como texto.")
		// Fallback: substituir qualquer referência textual restante
		docxFile.Replace("{{QRCODE}}", fmt.Sprintf("[QR CODE: %s]", certificateID), -1)
	}

	// 5. Salvar
	outputPath := filepath.Join(outputDir, fmt.Sprintf("certificado_%s.docx", cpf))
	err = docxFile.WriteToFile(outputPath)
	if err != nil {
		return "", "", fmt.Errorf("falha ao salvar: %w", err)
	}

	// 6. Gerar hash
	fileBytes, err := ioutil.ReadFile(outputPath)
	if err != nil {
		return "", "", fmt.Errorf("falha ao ler arquivo final: %w", err)
	}
	hash := sha256.Sum256(fileBytes)
	fileHash := hex.EncodeToString(hash[:])

	log.Printf("Certificado gerado para %s em %s", fullName, outputPath)
	return outputPath, fileHash, nil
}

func formatCPF(cpf string) string {
	if len(cpf) == 11 {
		return fmt.Sprintf("%s.%s.%s-**", cpf[0:3], cpf[3:6], cpf[6:9])
	}
	return cpf
}
