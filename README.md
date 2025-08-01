# 🎓 cliUfgd - Certificados UFGD

Este projeto é uma ferramenta de linha de comando desenvolvida em Go para facilitar a **geração de certificados de participação** em eventos organizados pela **Faculdade de Ciências Exatas e Tecnologia da UFGD (FACS/UFGD)**.

## ✨ Funcionalidades

* Criar eventos acadêmicos com base em arquivos CSV.
* Registrar automaticamente os participantes.
* Gerar certificados personalizados em DOCX para cada participante.

## 📦 Requisitos

* [Go](https://go.dev/dl/) 1.20 ou superior

## 🚀 Como usar

### 1. Clonar o repositório

```bash
git clone https://github.com/marceloajalaalarcon/cliUfgd.git
cd cliUfgd
```
### 1.2 Baixar as dependências

```bash
go mod tidy
```

### 2. Criar um novo evento

```bash
go run cmd/cert-cli/main.go create-event --name "Semana da Computação 2025" --organizer "FACS/UFGD" --csv "participantes.csv"
```

Esse comando:

* Cria o banco de dados para o evento.
* Registra os participantes a partir do arquivo `participantes.csv`.

### 3. Gerar certificados

```bash
go run cmd/cert-cli/main.go generate --id 1 --name "Semana da Computação 2025"
```

Esse comando:

* Gera os certificados dos participantes do evento com ID `1`.
* Os arquivos são salvos em uma pasta de saída (storage/certificates).

## 📂 Estrutura esperada do CSV

```csv
nome_completo,cpf
Joao da Silva,11122233344
Maria Oliveira,55566677788
...
```

## 🧪 Testes e Desenvolvimento

* O código está modularizado usando o `DDD`.
* O banco de dados usado SQLite (podendo trocar para qualquer outro, basta configurar).
* Os certificados são gerados como DOCX utilizando bibliotecas de manipulação de documentos.


## 🤛‍♂️ Autor

Desenvolvido por **Marcelo Ajala Alarcon**
[@marceloajalaalarcon](https://github.com/marceloajalaalarcon)
