# 🚀 Agregador de Negociações B3

Uma aplicação Go de alta performance para agregação e processamento de dados de negociações da B3 com 🐘 PostgreSQL.

## ✨ Visão Geral

Este projeto é super prático e oferece duas formas de interagir com os dados da B3:

1.  **Aplicação Web**: Uma API para consultar dados já processados.
2.  **Ferramenta CLI**: Uma ferramenta para importar arquivos de negociações para o banco de dados.

Ambas funcionam a partir do mesmo código e são fáceis de rodar, seja com Docker ou diretamente na sua máquina.

---

## ⚠️ **Primeiro Passo Crucial para AMBOS os Métodos!** ⚠️

Antes de qualquer coisa, você precisa do arquivo de dados da B3:

*   **Baixe o arquivo**: Acesse [https://arquivos.b3.com.br/rapinegocios/tickercsv/2025-08-29](https://arquivos.b3.com.br/rapinegocios/tickercsv/2025-08-29).
*   **Salve na pasta `data/`**: Renomeie o arquivo baixado para `29-08-2025_NEGOCIOSAVISTA.txt` e coloque-o dentro da pasta `data/` do projeto.
    *   **Caminho final**: `data/29-08-2025_NEGOCIOSAVISTA.txt`

---

## �� **Como Rodar a Aplicação (Escolha seu Método!)**

### Opção 1: Com Docker (Recomendado! É o mais fácil!) ��

Use esta opção se você quer tudo funcionando rapidinho, sem instalar Go ou PostgreSQL na sua máquina.

1.  **Pré-requisitos**:
    *   Instale **Docker** e **Docker Compose** (se ainda não tiver).
2.  **Inicie tudo**: Na pasta principal do projeto, digite:
    ```bash
    make docker-run
    ```
    *   **O que acontece?** Isso vai:
        *   Construir as aplicações.
        *   **Iniciar o PostgreSQL**.
        *   Ligar a Aplicação Web.
        *   **Importar os dados da B3 automaticamente** usando a ferramenta CLI. Ela faz a ingestão e depois se desliga sozinha!
3.  **Verifique (opcional)**: Para ver o que está acontecendo:
    ```bash
    make docker-logs
    ```
4.  **Para parar**: Quando quiser desligar tudo:
    ```bash
    make docker-stop
    ```

### Opção 2: Localmente (Com o Script Interativo!) ��

Ideal se você prefere rodar o código diretamente na sua máquina. O script `run_manual.sh` cuida de quase tudo!

1.  **Pré-requisitos**:
    *   Instale **Go 1.24+** 🐹.
    *   Instale **Docker** e **Docker Compose** (o script usa o Docker Compose para iniciar o PostgreSQL, se precisar).
2.  **Execute o script**: Na pasta principal do projeto, digite:
    ```bash
    make run-manual
    ```
    *   **O que acontece?** O script vai:
        *   Verificar e configurar seu ambiente (incluindo o arquivo `.env`).
        *   **Iniciar o PostgreSQL** (se necessário).
        *   Construir as aplicações Go.
        *   Apresentar um **menu** para você escolher o que quer rodar (Aplicação Web, Ferramenta CLI ou Ambas!).

---

## 🚀 **Usando a Aplicação**

Depois de iniciar a aplicação com um dos métodos acima:

### 1. **Importação dos Dados da B3 (Ingestão)**

É fundamental que os dados da B3 estejam no banco para a API funcionar!

*   **Se você usou `make docker-run`**:
    *   **A importação já foi feita automaticamente!** Não precisa fazer mais nada.

*   **Se você usou o script (`make run-manual`)**:
    *   No **menu** do script, escolha a opção para rodar a **"CLI Application (Data Ingestion)"** ou a opção **"Both (CLI first, then Web)"**. O script fará a importação para você.

*   **Se quiser rodar a CLI manualmente (após `make build-cli`)**:
    ```bash
    # Para importar o arquivo que você baixou
    make cli-example

    # Ou, para um arquivo diferente
    ./bin/ingest -file /caminho/para/seu/outro_arquivo.txt
    ```

### 2. **Consultando Dados via API (Web)**

Com os dados importados, a Aplicação Web já está funcionando em `http://localhost:8080`.

**Exemplo de Consulta**:
Use o `curl` (ou seu navegador) para testar:

```bash
curl "http://localhost:8080/api/v1/trades/aggregated?ticker=PETR4&data_inicio=2024-01-01"

Formato da Resposta (Exemplo):
{
  "ticker": "PETR4",
  "max_range_value": 45.67,
  "max_daily_volume": 1500000
}

## ⚙️ Configuração Extra (Para Curiosos!)

A aplicação usa variáveis de ambiente que podem ser configuradas no arquivo `.env` na raiz do projeto (o script `run_manual.sh` já cuida disso, copiando do `local-env.txt` se o `.env` não existir).

Aqui está um exemplo completo do que você pode ter no seu arquivo `.env`:

```dotenv
APP_NAME=BACKEND-APLICATION
SRV_PORT=8080
SRV_MODE=DEVELOPER

# Configurações do Banco de Dados PostgreSQL
SRV_DB_HOST=localhost
SRV_DB_NAME=b3_trade_aggregator
SRV_DB_USER=postgres
SRV_DB_PASS=postgres
SRV_DB_PORT=5432
SRV_DB_SSL_MODE=require # Ou 'disable' para desenvolvimento local mais fácil

# Alternativa: Use DATABASE_URL ao invés das variáveis individuais acima
# DATABASE_URL=postgres://postgres:postgres@localhost:5432/b3_trade_aggregator?sslmode=disable

# Configuração da API Web
API_PORT=8080

# Caminho padrão para o arquivo de dados da B3 para a ferramenta CLI
FILE_PATH=data/29-08-2025_NEGOCIOSAVISTA.txt

# Configurações de Log
LOG_LEVEL=info
LOG_OUTPUT=stdout

## 👨‍💻 Para Desenvolvedores (Se Quiser se Aprofundar)

Aqui estão alguns comandos `make` úteis se você for explorar o código:

*   `make build` e `make build-cli`: Compilam as aplicações individualmente.
*   `make test` e `make test-coverage`: Para rodar os testes.
*   `make dev`: Inicia a aplicação web com recarregamento automático (requer `air`).
*   `make db-reset`: **CUIDADO!** Reseta o banco de dados e apaga TUDO. Use só em desenvolvimento.

## ✒️ Autor e Licença

*   **Autor**: Charles Tenorio da Silva (<charles.tenorio.dev@gmail.com>)
*   **Licença**: Este projeto está licenciado sob a Licença MIT. Veja o arquivo [LICENSE](LICENSE) para mais detalhes.