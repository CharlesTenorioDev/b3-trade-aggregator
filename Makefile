# Variável auxiliar para carregar as variáveis de ambiente do arquivo .env.
# Isso garante que todas as variáveis definidas no  arquivo .env
# (como DATABASE_URL, SRV_PORT, FILE_PATH, etc.)
# estejam disponíveis para os comandos Go subsequentes.
# 'set -a' ativa o modo de exportação automática.
# '. ./.env' (ou 'source ./.env') lê e executa o arquivo .env no shell atual.
# 'set +a' desativa o modo de exportação automática.
# '|| true' garante que o Makefile não falhe se o arquivo .env não existir (útil em alguns cenários).
_LOAD_ENV := set -a && . ./.env && set +a || true

.PHONY: build run test clean docker-build docker-run deps migrate build-cli run-cli run-cli-env \
        test-coverage docker-stop docker-logs setup setup-full dev db-reset perf-test \
        cli-help cli-version cli-example cli-example-env

# Build da aplicação principal (servidor API)
build:
    go build -o bin/app cmd/app/main.go

# Build da ferramenta CLI de ingestão
build-cli:
    go build -o bin/ingest cmd/ingest/main.go

# Executa a aplicação principal (servidor API).
# Primeiro, garante que o binário 'app' está construído,
# depois carrega as variáveis de ambiente e executa o binário.
run: build
    $(info Rodando a aplicação principal...)
    $(info Carregando variáveis de ambiente do .env...)
    $(_LOAD_ENV) && ./bin/app

# Executa a ferramenta CLI de ingestão.
# Primeiro, garante que o binário 'ingest' está construído,
# depois carrega as variáveis de ambiente e executa o binário.
run-cli: build-cli
    $(info Rodando a ferramenta CLI de ingestão...)
    $(info Carregando variáveis de ambiente do .env (incluindo FILE_PATH)...)
    $(_LOAD_ENV) && ./bin/ingest

# Executa a ferramenta CLI de ingestão, definindo o FILE_PATH explicitamente
# via variável de ambiente no comando, além de carregar outras do .env.
run-cli-env: build-cli
    $(info Rodando a ferramenta CLI de ingestão com FILE_PATH explícito...)
    $(info Carregando outras variáveis de ambiente do .env...)
    $(_LOAD_ENV) && FILE_PATH=data/29-08-2025_NEGOCIOSAVISTA.txt ./bin/ingest

# Executa todos os testes unitários do projeto
test:
    go test ./...

# Executa testes com cobertura de código e gera um relatório HTML
test-coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out

# Limpa os artefatos de build (binários e arquivos de cobertura)
clean:
    rm -rf bin/
    rm -f coverage.out

# Constrói a imagem Docker da aplicação com a tag 'b3-trade-aggregator'
docker-build:
    docker build -t b3-trade-aggregator .

# Inicia os serviços definidos no docker-compose.yml em modo detached (-d).
# O docker-compose por padrão já procura e carrega o arquivo .env na mesma pasta.
docker-run:
    docker-compose up -d

# Para e remove os containers e redes criados pelo docker-compose
docker-stop:
    docker-compose down

# Exibe os logs dos serviços do Docker Compose em tempo real
docker-logs:
    docker-compose logs -f

# Executa as migrações do banco de dados.
# ATENÇÃO: Você precisa substituir os comentários pelo comando real da sua ferramenta de migração.
# Certifique-se de que sua ferramenta de migração (ex: 'migrate') esteja instalada
# e que ela consiga ler a DATABASE_URL do ambiente.
migrate:
    $(info Executando migrações do banco de dados...)
    $(info Certifique-se de que a ferramenta de migração (ex: 'migrate') está instalada.)
    $(info Variáveis de ambiente do DB serão carregadas do .env.)
    # Exemplo: $(_LOAD_ENV) migrate -path migrations -database "$(DATABASE_URL)" up
    # Ou se usar uma ferramenta Go customizada para migrações:
    # $(_LOAD_ENV) go run cmd/migrate/main.go up

# Instala as dependências do módulo Go e sincroniza o go.mod/go.sum
deps:
    go mod download
    go mod tidy

# Cria os diretórios necessários para o projeto
setup:
    mkdir -p data
    mkdir -p bin

# Configuração completa: instala dependências, cria diretórios e compila ambos os binários
setup-full: setup deps build build-cli

# Modo de desenvolvimento com hot reload (requer 'air' ou ferramenta similar).
# Instalar 'air' se não presente: go install github.com/cosmtrek/air@latest
dev:
    $(info Iniciando modo de desenvolvimento com hot reload (usando 'air')...)
    $(info Certifique-se de que 'air' está instalado.)
    $(info Variáveis de ambiente do .env serão carregadas para 'air'.)
    $(if $(shell which air), $(_LOAD_ENV) && air, $(error "air não encontrado. Por favor, instale-o com 'go install github.com/cosmtrek/air@latest'"))

# Operações de banco de dados: para, remove volumes, inicia apenas o postgres e executa migrações
db-reset: docker-stop
    $(info Reiniciando o container do PostgreSQL e aplicando migrações...)
    docker-compose up -d postgres
    sleep 5 # Pequena pausa para garantir que o banco de dados esteja totalmente iniciado
    $(MAKE) migrate # Executa as migrações após o reset do DB

# Adicione comandos para testes de performance aqui.
perf-test:
    $(info Adicione comandos de teste de performance aqui.)
    # Exemplo: ab -n 1000 -c 10 http://localhost:8080/trades/aggregated?ticker=PETR4

# Exemplos de uso da ferramenta CLI
cli-help: build-cli
    $(info Exibindo ajuda da CLI...)
    $(_LOAD_ENV) && ./bin/ingest -help

cli-version: build-cli
    $(info Exibindo versão da CLI...)
    $(_LOAD_ENV) && ./bin/ingest -version

cli-example: build-cli
    $(info Exemplo de uso da CLI com argumento de linha de comando (-file)...)
    $(_LOAD_ENV) && ./bin/ingest -file data/29-08-2025_NEGOCIOSAVISTA.txt

cli-example-env: build-cli
    $(info Exemplo de uso da CLI com FILE_PATH vindo do .env...)
    $(_LOAD_ENV) && ./bin/ingest