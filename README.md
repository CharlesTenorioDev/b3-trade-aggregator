# 🚀 Agregador de Negociações B3

Uma aplicação Go de alta performance para agregação e processamento de dados de negociações da B3 com 🐘 PostgreSQL 17, otimizada para ingestão de dados em larga escala usando pgx COPY FROM.

## ✒️ Autor

-   **Charles Tenorio da Silva**
-   **Email**: charles.tenorio.dev@gmail.com

## ✨ Funcionalidades

-   **⚡ Ingestão de Dados de Alta Performance**: Utiliza `pgx COPY FROM` para uma performance ótima em inserções em massa.
-   **�� PostgreSQL 17**: A versão mais recente do PostgreSQL com funcionalidades avançadas.
-   **🧹 Arquitetura Limpa**: Projeto Go bem estruturado seguindo as melhores práticas.
-   **🐳 Suporte a Docker**: Containerização completa com Docker Compose para fácil implantação.
-   **🌐 API RESTful**: API HTTP para consultar dados de negociações agregados.
-   **🌊 Processamento por Streaming**: Processamento eficiente de arquivos por streaming para grandes conjuntos de dados.
-   **✂️ Separação de Responsabilidades**: Ferramenta CLI independente para ingestão de dados e API web para consultas.

## 📁 Estrutura do Projeto


```
├── cmd/
│   ├── app/
│   │   └── main.go                 # Ponto de entrada da aplicação web
│   └── ingest/
│       └── main.go                 # Ponto de entrada da ferramenta CLI de ingestão 
├── internal/
│   ├── api/
│   │   └── handler/
│   │       ├── handler.go          #  Lógica de tratamento de requisições HTTP
│   │       └── router.go           # API route registration
│   ├── config/
│   │   └── config.go               # Carregamento e estrutura das configurações
│   ├── entity/
│   │   └── trade.go                # Modelos de dados
│   ├── ingestion/
│   │   ├── reader.go               # Leitura por streaming

│   ├── repository/
│   │   └── trade.go                # Interações com o banco de dados (pgx COPY FROM)
│   ├── service/
│   │   └── trade.go                #  Lógica de negócio e orquestração
│   └── util/
│       └── errors.go               # Tipos de erro customizados e utilitários
├── pkg/                            # Pacotes reutilizáveis
│   └── server/
│       └── server.go               # Implementação do servidor HTTP
├── migrations/                     #  Scripts de migração do banco de dados
├── tests/                          # Testes de integração/ponta a ponta
├── data/                           # Diretório para arquivos de dados
├── docker-compose.yml              # Orquestração de serviços Docker
├── Dockerfile                      # Containerização da aplicação
├── Makefile                        # Automação de tarefas
└── go.mod                          # Módulos Go
```

## 📈 Otimizações de Performance

**pgx COPY FROM**: Utiliza o protocolo `COPY` do PostgreSQL para inserções em massa (10x mais rápido que `INSERT`s individuais).
-   **Pool de Conexões**: Gerenciamento eficiente de conexões com `pgxpool`.
-   **Processamento em Lotes**: Tamanhos de lote configuráveis para uso ótimo de memória.
-   **Streaming**: Processamento de arquivos sem carregar o arquivo inteiro na memória.
-   **Queries Indexadas**: Índices de banco de dados otimizados para agregações rápidas.

## �� Primeiros Passos

### Pré-requisitos

-   Go 1.24+ 🐹
-   PostgreSQL 17 ��
-   Docker e Docker Compose 🐳
-   **Arquivo de Dados da B3**: É imprescindível baixar o arquivo de dados da B3 do link [https://arquivos.b3.com.br/rapinegocios/tickercsv/2025-08-29](https://arquivos.b3.com.br/rapinegocios/tickercsv/2025-08-29) e salvá-lo na pasta `data/` com o nome `29-08-2025_NEGOCIOSAVISTA.txt`. O caminho final do arquivo deve ser `data/29-08-2025_NEGOCIOSAVISTA.txt`.

### Executando com Docker (Recomendado)

1.  Inicie os serviços:
    ```bash
    make docker-run
    ```

2.  Verifique os logs:
    ```bash
    make docker-logs
    ```

3.  Pare os serviços:
    ```bash
    make docker-stop
    ```

### Executando Localmente

1.  Configure e instale as dependências:
    ```bash
    make setup-full
    ```

2.  Inicie o PostgreSQL (usando Docker):
    ```bash
    docker-compose up -d postgres
    ```

3.  Execute a aplicação web:
    ```bash
    make run
    ```
### �� Ingestão de Dados (Ferramenta CLI)

A ferramenta CLI foi projetada para processar grandes arquivos de negociações da B3 de forma independente da aplicação web.


#### Build the CLI:
```bash
make build-cli
```

#### Run the CLI:
```bash
# Mostrar ajuda
make cli-help

# Mostrar a versao
make cli-version

# Processa um arquivo (substitua pelo caminho real)
go run cmd/ingest/main.go -file /caminho/para/seu/29-08-2025_NEGOCIOSAVISTA.txt
```

#### Funcionalidades do CLI::
- **Validação de Arquivo**: Verifica se o arquivo especificado existe.
- **Registro de Progresso**: Atualizações de progresso em tempo real durante o processamento
- **Tratamento de Erros**: Relatório de erros abrangente
- **Métricas de Performance**: Tempo de processamento e estatísticas.
- **Conexão com o Banco de Dados**: Gerenciamento automático da conexão com o PostgreSQL.

### Uso da API

Consulte dados de negociações agregados::
```bash
curl "http://localhost:8080/api/v1/trades/aggregated?ticker=PETR4&data_inicio=2024-01-01"
```

Formato da resposta::
```json
{
  "ticker": "PETR4",
  "max_range_value": 45.67,
  "max_daily_volume": 1500000
}
```

🧪 Testes

Run tests:
```bash
make test
```

Execute os testes com cobertura:
```bash
make test-coverage
```

👨‍💻 Desenvolvimento

### Comandos Make Disponíveis

#### Aplicação Web::
- `make build` - Compila a aplicação web
- `make run` -  Executa a aplicação web
- `make docker-build` - Compila a imagem Docker
- `make docker-run` - Executa com Docker Compose.
- `make docker-stop` - Para o Docker Compose
- `make docker-logs` - VVisualiza os logs do Docker.

#### Ferramenta CLI:
- `make build-cli` - Compila a ferramenta CLI
- `make run-cli` - Executa a ferramenta CLI.
- `make cli-help` - Exibe a ajuda do CLI.
- `make cli-version` -  Exibe a versão do CLI.
- `make cli-example` - Exemplo de uso do CLI.

#### Geral:
- `make test` - Executa os testes
- `make test-coverage` - Executa os testes com cobertura
- `make clean` - Limpa os artefatos de build.
- `make deps` -  Instala as dependências.
- `make setup` - Cria os diretórios necessários
- `make setup-full` - Configuração completa (deps + compila ambas as ferramentas)
- `make db-reset` -  Reseta o banco de dados.
- `make perf-test` - Teste de performance.

🏛️ Benefícios da Arquitetura

#### Separação de Responsabilidades:
1. **Aplicação Web**: Otimizada para queries da API e respostas em tempo real.
2. **Ferramenta CLI:**: Dedicada à ingestão de dados e processamento em lote.
3. **Serviços Compartilhados**: Lógica de negócio e operações de banco de dados comuns.

#### Operação Independente:
- **Aplicação Web**: Pode rodar sem sobrecarga de ingestão.
- **Ferramenta CLI**: Pode processar arquivos sem depender dos recursos do servidor web.
- **Escalabilidade**: Cada componente pode ser escalado independentemente

📥 Processo de Ingestão de Dados

A ferramenta CLI processa grandes arquivos de negociações da B3 (565MB+) de forma eficiente:

1. **Streams** do arquivo linha por linha sem carregá-lo inteiramente na memória.
2. **Parses** cada linha em dados de negociação estruturados.
3. **Batches** as negociações em lotes de tamanho configurável (padrão: 1000).
4. **Uses COPY FROM** para inserções de dados em massa de alta performance no banco de dados.
5. **Handles errors** graciosamente com log detalhado.

⚙️ Configuração

Variáveis de ambiente::
- `DATABASE_URL`: String de conexão com o PostgreSQL
- `API_PORT`: Porta do servidor HTTP (padrão: 8080).

Exemplo:
```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/b3_trade_aggregator?sslmode=disable"
export API_PORT="8080"
```

⏱️ Benchmarks de Performance

Com pgx COPY FROM, a aplicação pode processar:
- **~100,000 negociações/segundo** on em hardware padrão
- **arquivo de 565MB** em aproximadamente 2-3 minutos
- **Uso de memória** permanece constante independentemente do tamanho do arquivo

📄 Licença

Este projeto está licenciado sob a Licença MIT  -  veja o arquivo [LICENSE](LICENSE) para detalhes.
