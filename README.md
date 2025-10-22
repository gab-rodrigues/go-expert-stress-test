# Stress Test - Sistema CLI para Testes de Carga

Um sistema CLI desenvolvido em Go para realizar testes de carga em serviços web. A aplicação permite configurar o número total de requests, a concorrência e gera relatórios detalhados sobre o desempenho do serviço testado.

## Funcionalidades

- **Testes de carga concorrentes**: Utiliza goroutines e workers pattern para máxima eficiência
- **Relatórios detalhados**: Estatísticas completas incluindo tempo de execução, taxa de sucesso e distribuição de códigos HTTP
- **Containerização**: Suporte completo para Docker e Docker Compose
- **Interface CLI intuitiva**: Parâmetros simples e validação de entrada
- **Controle de timeout**: Proteção contra requests que ficam pendentes
- **Progress tracking**: Acompanhamento em tempo real do progresso dos testes

## Parâmetros CLI

| Parâmetro | Descrição | Obrigatório | Exemplo |
|-----------|-----------|-------------|---------|
| `--url` | URL do serviço a ser testado | ✅ | `--url=http://google.com` |
| `--requests` | Número total de requests | ✅ | `--requests=1000` |
| `--concurrency` | Número de chamadas simultâneas | ✅ | `--concurrency=10` |

## Arquitetura

### Estratégia de Concorrência

O sistema utiliza o **Worker Pattern** em Go para otimizar a concorrência:

- **Workers Pool**: Cria um pool de goroutines (workers) baseado no parâmetro `--concurrency`
- **Jobs Channel**: Canal buffered para distribuir trabalho entre os workers
- **Results Channel**: Canal para coletar resultados de forma thread-safe
- **Context Cancellation**: Controle graceful de cancelamento e timeouts

### Componentes Principais

1. **Config**: Estrutura para parâmetros CLI
2. **Result**: Estrutura para resultado de cada request
3. **Report**: Estrutura para relatório final
4. **worker()**: Função que processa requests HTTP
5. **runLoadTest()**: Orquestra a execução do teste
6. **printReport()**: Gera relatório formatado

## Instalação e Uso

### Opção 1: Executar com Docker (Recomendado)

```bash
# Construir a imagem
docker build -t stress-test .

# Executar teste
docker run stress-test --url=http://google.com --requests=1000 --concurrency=10
```

### Opção 2: Usar Docker Compose

```bash
# Executar com configuração padrão
docker-compose up stress-test

# Executar configuração leve
docker-compose --profile light up stress-test-light

# Executar configuração pesada
docker-compose --profile heavy up stress-test-heavy
```

### Opção 3: Executar com Go nativo

```bash
# Instalar dependências
go mod tidy

# Compilar
go build -o stress-test

# Executar
./stress-test --url=http://google.com --requests=1000 --concurrency=10
```

## Exemplos de Uso

### Teste Básico

```bash
docker run stress-test --url=https://httpbin.org/get --requests=100 --concurrency=5
```

### Teste de Alta Concorrência

```bash
docker run stress-test --url=https://jsonplaceholder.typicode.com/posts/1 --requests=5000 --concurrency=100
```

## Relatório de Saída

O sistema gera um relatório completo com as seguintes métricas:

```
==================================================
RELATÓRIO DE TESTE DE CARGA
==================================================
Tempo total de execução: 2.345s
Total de requests realizados: 1000
Requests com status 200: 950
Taxa de sucesso: 95.00%
Requests por segundo: 426.44

Distribuição de códigos de status:
  200: 950 (95.00%)
  404: 30 (3.00%)
  500: 15 (1.50%)
  Errors: 5 (0.50%)
==================================================
```