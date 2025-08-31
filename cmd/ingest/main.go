package main

import (
	"context"
	"flag" // Para parsing de argumentos de linha de comando
	"fmt"
	"log"         // Para logs de sistema
	"os"          // Para operações de sistema de arquivos e variáveis de ambiente
	"runtime"     // Para controle do uso de CPU
	"sync/atomic" // Para operações atômicas seguras em concorrência
	"time"        // Para manipulação de tempo

	"github.com/jackc/pgx/v5/pgxpool" // Driver de pool de conexões para PostgreSQL (pgx)

	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/config"     // Configurações da aplicação
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/ingestion"  // Lógica de ingestão de dados
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/repository" // Camada de acesso a dados
	"github.com/CharlesTenorioDev/b3-trade-aggregator/internal/service"    // Camada de lógica de negócio
)

// Variáveis de versão e commit, geralmente preenchidas no momento do build.
var (
	VERSION = "0.1.0-dev"
	COMMIT  = "ABCDEFG-dev"
)

// ProgressTracker rastreia as estatísticas de processamento.
type ProgressTracker struct {
	recordsProcessed int64     // Contador de registros processados (atômico para concorrência)
	startTime        time.Time // Tempo de início do processamento
	lastUpdate       time.Time // Último tempo de atualização do progresso na tela
}

// Increment incrementa o contador de registros processados de forma segura em concorrência.
func (pt *ProgressTracker) Increment() {
	atomic.AddInt64(&pt.recordsProcessed, 1)
}

// GetCount retorna o número total de registros processados de forma segura.
func (pt *ProgressTracker) GetCount() int64 {
	return atomic.LoadInt64(&pt.recordsProcessed)
}

// GetElapsed retorna o tempo decorrido desde o início do processamento.
func (pt *ProgressTracker) GetElapsed() time.Duration {
	return time.Since(pt.startTime)
}

// GetRate calcula a taxa de processamento (registros por segundo).
func (pt *ProgressTracker) GetRate() float64 {
	elapsed := pt.GetElapsed().Seconds()
	if elapsed > 0 {
		return float64(pt.GetCount()) / elapsed
	}
	return 0
}

// PrintProgress imprime o progresso atual na mesma linha do terminal.
// Atualiza a cada 2 segundos para evitar spam excessivo no console.
func (pt *ProgressTracker) PrintProgress() {
	now := time.Now()
	if now.Sub(pt.lastUpdate) >= 2*time.Second { // Atualiza a cada 2 segundos
		elapsed := pt.GetElapsed()
		rate := pt.GetRate()
		// \r move o cursor para o início da linha, sobrescrevendo o conteúdo anterior.
		fmt.Printf("\r🔄 Processando... %d registros | %.1f registros/seg | %v decorrido",
			pt.GetCount(), rate, elapsed.Round(time.Second))
		pt.lastUpdate = now
	}
}

// main é o ponto de entrada da aplicação CLI.
func main() {
	// Define o número máximo de núcleos de CPU a serem usados.
	// 6 núcleos é um bom ponto de partida para balancear CPU e I/O.
	runtime.GOMAXPROCS(6)
	log.Printf("🔧 Núcleos de CPU limitados a: %d", runtime.GOMAXPROCS(0))

	// Define e parseia as flags de linha de comando.
	var (
		filePath    = flag.String("file", "", "Caminho para o arquivo de dados da B3 (obrigatório, ou defina a variável de ambiente FILE_PATH)")
		showVersion = flag.Bool("version", false, "Exibe informações da versão")
		showHelp    = flag.Bool("help", false, "Exibe informações de ajuda")
	)
	flag.Parse() // Executa o parsing das flags

	// Se a flag --version foi solicitada, exibe a versão e encerra.
	if *showVersion {
		fmt.Printf("CLI B3 Trade Aggregator v%s (%s)\n", VERSION, COMMIT)
		os.Exit(0)
	}

	// Tenta obter o caminho do arquivo da flag -file ou da variável de ambiente FILE_PATH.
	actualFilePath := *filePath
	if actualFilePath == "" {
		actualFilePath = os.Getenv("FILE_PATH")
	}

	// Se a flag --help foi solicitada ou nenhum arquivo foi especificado, exibe a ajuda e encerra.
	if *showHelp || actualFilePath == "" {
		fmt.Println("B3 Trade Aggregator - CLI de Ingestão de Dados")
		fmt.Println("Uso: go run cmd/ingest/main.go -file <caminho_do_arquivo>")
		fmt.Println("   ou: go run cmd/ingest/main.go (com a variável de ambiente FILE_PATH definida)")
		fmt.Println("\nFlags:")
		flag.PrintDefaults() // Imprime as flags padrão definidas
		fmt.Println("\nVariáveis de Ambiente:")
		fmt.Println("  FILE_PATH    Caminho para o arquivo de dados de negociações da B3")
		fmt.Println("\nExemplos:")
		fmt.Println("  go run cmd/ingest/main.go -file data/29-08-2025_NEGOCIOSAVISTA.txt")
		fmt.Println("  FILE_PATH=data/29-08-2025_NEGOCIOSAVISTA.txt go run cmd/ingest/main.go")
		os.Exit(0)
	}

	// Valida se o arquivo especificado existe.
	if _, err := os.Stat(actualFilePath); os.IsNotExist(err) {
		log.Fatalf("Arquivo não encontrado: %s", actualFilePath)
	}

	// Obtém o tamanho do arquivo para cálculo de progresso e estatísticas.
	fileInfo, err := os.Stat(actualFilePath)
	if err != nil {
		log.Fatalf("Erro ao obter informações do arquivo: %v", err)
	}
	fileSizeMB := float64(fileInfo.Size()) / (1024 * 1024) // Converte bytes para MB

	log.Printf("Iniciando CLI B3 Trade Aggregator v%s", VERSION)
	log.Printf("Processando arquivo: %s (%.1f MB)", actualFilePath, fileSizeMB)

	// Carrega as configurações da aplicação (incluindo DATABASE_URL).
	cfg := config.LoadConfig()

	// Inicializa a conexão com o banco de dados PostgreSQL usando pgxpool.
	log.Println("Conectando ao PostgreSQL...")
	// context.Background() é usado para o contexto inicial da criação do pool.
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Falha ao criar o pool de conexões: %v", err)
	}
	defer pool.Close() // Garante que o pool de conexões será fechado ao final da main

	// Testa a conexão com o banco de dados.
	if err = pool.Ping(context.Background()); err != nil {
		log.Fatalf("Falha ao pingar o banco de dados: %v", err)
	}
	log.Println("✅ Conexão com PostgreSQL estabelecida com sucesso!")

	// Inicializa o rastreador de progresso.
	progressTracker := &ProgressTracker{
		startTime:  time.Now(),
		lastUpdate: time.Now(),
	}

	// Inicializa as dependências do serviço de ingestão.
	tradeReader := ingestion.NewTradeStreamReader()
	// Passa o pool de conexões (pgxpool.Pool) para o repositório.
	// O repositório deve ser adaptado para usar pgxpool.
	tradeRepo := repository.NewPostgresTradeRepository(pool)
	tradeService := service.NewTradeService(tradeReader, tradeRepo)

	// Inicia o processo de ingestão de dados.
	log.Println("🚀 Iniciando o processo de ingestão de dados...")
	fmt.Println("�� Monitoramento de progresso ativado...")

	// Cria um contexto com timeout para a operação de ingestão.
	// Se a ingestão demorar mais de 14 minutos, o contexto será cancelado.
	ingestionCtx, cancel := context.WithTimeout(context.Background(), 14*time.Minute)
	defer cancel() // Garante que o cancelamento do contexto será chamado.

	// Inicia uma goroutine para monitorar e imprimir o progresso periodicamente.
	go func() {
		ticker := time.NewTicker(2 * time.Second) // Ticker para disparar a cada 2 segundos
		defer ticker.Stop()                       // Garante que o ticker será parado
		for {
			select {
			case <-ingestionCtx.Done(): // Se o contexto for cancelado, a goroutine termina.
				return
			case <-ticker.C: // No tick do temporizador, imprime o progresso.
				progressTracker.PrintProgress()
			}
		}
	}()

	// Chama o método de ingestão do serviço, passando o contexto e o rastreador de progresso.
	// Assumindo que o tradeService agora tem um método ProcessIngestionWithProgress.
	if err := tradeService.ProcessIngestionWithProgress(ingestionCtx, actualFilePath, progressTracker); err != nil {
		fmt.Println() // Limpa a linha de progresso antes de logar o erro
		log.Fatalf("❌ Falha na ingestão: %v", err)
	}

	// Limpa a linha de progresso e imprime as estatísticas finais.
	fmt.Println() // Limpa a linha de progresso final no terminal
	fmt.Println("📈 Estatísticas de Processamento:")
	fmt.Printf("   📁 Arquivo: %s\n", actualFilePath)
	fmt.Printf("   📏 Tamanho: %.1f MB\n", fileSizeMB)
	fmt.Printf("   📊 Registros Processados: %d\n", progressTracker.GetCount())
	fmt.Printf("   ⏱️  Tempo Total: %v\n", progressTracker.GetElapsed().Round(time.Second))
	fmt.Printf("   🚀 Taxa Média: %.1f registros/seg\n", progressTracker.GetRate())
	// Cálculo da velocidade de processamento em MB/seg (tempo total deve ser maior que zero para evitar divisão por zero)
	processingSpeed := 0.0
	if progressTracker.GetElapsed().Seconds() > 0 {
		processingSpeed = fileSizeMB / progressTracker.GetElapsed().Seconds()
	}
	fmt.Printf("   📈 Velocidade de Processamento: %.1f MB/seg\n", processingSpeed)

	// Exibe o tempo total em minutos para melhor legibilidade.
	totalMinutes := progressTracker.GetElapsed().Minutes()
	fmt.Printf("   ⏰ Tempo Total: %.2f minutos\n", totalMinutes)

	fmt.Println("✅ Ingestão concluída com sucesso!")
	fmt.Println("🎉 Processamento de dados finalizado!")
}
