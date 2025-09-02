# Variável auxiliar para carregar as variáveis de ambiente do arquivo .env.
# Isso garante que todas as variáveis definidas no seu arquivo .env
# (como DATABASE_URL, SRV_PORT, FILE_PATH, etc.)
# estejam disponíveis para os comandos Go subsequentes.
# 'set -a' ativa o modo de exportação automática.
# '. ./.env' (ou 'source ./.env') lê e executa o arquivo .env no shell atual.
# 'set +a' desativa o modo de exportação automática.
# '|| true' garante que o Makefile não falhe se o arquivo .env não existir (útil em alguns cenários).
_LOAD_ENV := set -a && . ./.env && set +a || true

# Define alvos "phony" (falsos), que não correspondem a nomes de arquivos.
# Isso garante que o make sempre execute a receita, mesmo que exista um arquivo com o mesmo nome.
.PHONY: build run test clean docker-build docker-run deps migrate build-cli run-cli run-cli-env \
	test-coverage docker-stop docker-logs setup setup-full dev db-reset perf-test \
	cli-help cli-version cli-example cli-example-env docker-build-web docker-build-cli \
	docker-run-web docker-run-cli run-manual

# Build da aplicação principal (servidor API).
# Compila o executável principal da API e o coloca em 'bin/app'.
build:
	go build -o bin/app cmd/app/main.go

# Build da ferramenta CLI de ingestão.
# Compila o executável do CLI e o coloca em 'bin/ingest'.
build-cli:
	go build -o bin/ingest cmd/ingest/main.go

# Executa a aplicação principal (servidor API).
# Primeiro, garante que o binário 'app' está construído,
# depois carrega as variáveis de ambiente do .env e executa o binário.
run: build
	$(info Rodando a aplicação principal...)
	$(info Carregando variáveis de ambiente do .env...)
	$(_LOAD_ENV) && ./bin/app

# Executa a ferramenta CLI de ingestão.
# Primeiro, garante que o binário 'ingest' está construído,
# depois carrega as variáveis de ambiente do .env e executa o binário.
run-cli: build-cli
	$(info Rodando a ferramenta CLI de ingestão...)
	$(info Carregando variáveis de ambiente do .env (incluindo FILE_PATH)...)
	$(_LOAD_ENV) && SRV_DB_SSL_MODE=disable ./bin/ingest

# Executa a ferramenta CLI de ingestão, definindo o FILE_PATH explicitamente
# via variável de ambiente no comando, além de carregar outras do .env.
# Útil para testar com um arquivo específico sem alterar o .env.
run-cli-env: build-cli
	$(info Rodando a ferramenta CLI de ingestão com FILE_PATH explícito...)
	$(info Carregando outras variáveis de ambiente do .env...)
	$(_LOAD_ENV) && FILE_PATH=data/29-08-2025_NEGOCIOSAVISTA.txt ./bin/ingest

# Executa todos os testes unitários do projeto.
test:
	go test ./...

# Executa testes com cobertura de código e gera um relatório HTML de cobertura.
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Limpa os artefatos de build (binários compilados e arquivos de cobertura).
clean:
	rm -rf bin/
	rm -f coverage.out

# Constrói a imagem Docker da aplicação web com a tag 'b3-trade-aggregator:web-app'.
docker-build-web:
	docker build -t b3-trade-aggregator:web-app -f Dockerfile .

# Constrói a imagem Docker da aplicação CLI com a tag 'b3-trade-aggregator:cli-app'.
docker-build-cli:
	docker build -t b3-trade-aggregator:cli-app -f DockerfileCli .

# Constrói ambas as imagens Docker (web-app e cli-app).
docker-build: docker-build-web docker-build-cli

# Inicia os serviços definidos no docker-compose.yml em modo detached (-d).
# O docker compose por padrão já procura e carrega o arquivo .env na mesma pasta do compose.
docker-run:
	docker compose up -d

# Inicia apenas o serviço web-app.
docker-run-web:
	docker compose up -d web-app

# Inicia apenas o serviço cli-app.
docker-run-cli:
	docker compose up cli-app

# Para e remove os containers e redes criados pelo docker-compose.
docker-stop:
	docker compose down

# Exibe os logs dos serviços do Docker Compose em tempo real.
docker-logs:
	docker compose logs -f

# Executa as migrações do banco de dados.
# ATENÇÃO: É um placeholder. Você precisa instalar uma ferramenta de migração (ex: 'migrate' ou 'goose')
# e substituir o comando de exemplo pelo comando real da sua ferramenta.
# Certifique-se de que sua ferramenta de migração consiga ler a DATABASE_URL do ambiente.
migrate:
	$(info Executando migrações do banco de dados...)
	$(info Certifique-se de que a ferramenta de migração (ex: 'migrate') está instalada.)
	$(info Variáveis de ambiente do DB serão carregadas do .env.)
	# Exemplo usando 'migrate':
	# $(_LOAD_ENV) migrate -path migrations -database "$(DATABASE_URL)" up
	# Ou se usar uma ferramenta Go customizada para migrações (ex: em 'cmd/migrate/main.go'):
	# $(_LOAD_ENV) go run cmd/migrate/main.go up

# Instala as dependências do módulo Go e sincroniza o go.mod/go.sum.
deps:
	go mod download
	go mod tidy

# Cria os diretórios necessários para o projeto, como 'data' e 'bin'.
setup:
	mkdir -p data
	mkdir -p bin

# Configuração completa: instala dependências, cria diretórios e compila ambos os binários.
setup-full: setup deps build build-cli

# Modo de desenvolvimento com hot reload (recarregamento automático ao salvar).
# Requer a ferramenta 'air'. Instale-a se não presente: go install github.com/cosmtrek/air@latest
dev:
	$(info Iniciando modo de desenvolvimento com hot reload (usando 'air')...)
	$(info Certifique-se de que 'air' está instalado.)
	$(info Variáveis de ambiente do .env serão carregadas para 'air'.)
	# Verifica se 'air' está no PATH; se não, mostra um erro.
	@which air > /dev/null || (echo "Erro: 'air' não encontrado. Instale com: go install github.com/cosmtrek/air@latest" && exit 1)
	$(_LOAD_ENV) && air

# Reseta o banco de dados (para desenvolvimento).
# ATENÇÃO: Este comando DESTRÓI todos os dados do banco!
# Use apenas em ambiente de desenvolvimento.
db-reset:
	$(info ATENÇÃO: Resetando banco de dados...)
	$(info Todos os dados serão perdidos!)
	@read -p "Tem certeza? Digite 'yes' para confirmar: " confirm && [ "$$confirm" = "yes" ] || exit 1
	$(_LOAD_ENV) && docker compose down -v && docker compose up -d postgres
	$(info Aguardando PostgreSQL inicializar...)
	sleep 10
	$(_LOAD_ENV) && make migrate

# Teste de performance da ingestão.
# Executa a ingestão com métricas de tempo e memória.
perf-test: build-cli
	$(info Executando teste de performance da ingestão...)
	$(info Arquivo: data/29-08-2025_NEGOCIOSAVISTA.txt)
	$(_LOAD_ENV) && /usr/bin/time -v ./bin/ingest -file data/29-08-2025_NEGOCIOSAVISTA.txt

# Comandos auxiliares para o CLI
# Exibe a ajuda do CLI
cli-help: build-cli
	$(_LOAD_ENV) && ./bin/ingest -help

# Exibe a versão do CLI
cli-version: build-cli
	$(info Exibindo versão do CLI...)
	$(_LOAD_ENV) && ./bin/ingest -version

# Exemplo de uso do CLI com arquivo específico
cli-example: build-cli
	$(info Exemplo de uso do CLI com arquivo específico...)
	$(_LOAD_ENV) && ./bin/ingest -file data/29-08-2025_NEGOCIOSAVISTA.txt

# Exemplo de uso do CLI com variável de ambiente
cli-example-env: build-cli
	$(info Exemplo de uso do CLI com variável de ambiente...)
	$(_LOAD_ENV) && ./bin/ingest

# Executa o script manual para rodar aplicações sem Docker
# Permite escolher entre web, CLI ou ambos
run-manual:
	$(info Executando script manual para rodar aplicações sem Docker...)
	$(info O script irá verificar pré-requisitos, construir aplicações e permitir escolha.)
	./run_manual.sh