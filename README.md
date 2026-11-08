# gorate — HTTP Load Testing CLI


O **gorate** é uma ferramenta de linha de comando (CLI) escrita em Go para testes de carga e benchmark de serviços HTTP. Ele dispara requisições concorrentes com controle de taxa (rate limiting) contra uma ou múltiplas URLs simultaneamente, gerando relatórios estatísticos com distribuição de latência por percentis (`P50`, `P95`, `P99`), vazão (`RPS`), classificação de erros (`4xx/5xx` vs `Timeouts`) e suporte a interrupção segura (graceful shutdown).

---

## Destaques Técnicos e Arquitetura

Este projeto foi construído utilizando a biblioteca padrão do go (Zero Dependências Externas), aplicando padrões de engenharia voltados para performance, concorrência segura e testabilidade:

* **Worker Pool Pattern (`goroutines` + `channels`):** Controle do número de trabalhadores simultâneos por URL, evitando exaustão de memória e de descritores de arquivo do sistema operacional.
* **Rate Limiting Preciso (`time.Ticker`):** Controle da taxa de disparos por segundo antes do enfileiramento no canal de trabalho.
* **Otimização de Rede (TCP Keep-Alive):** Reaproveitamento de conexões HTTP através do ajuste dinâmico no `MaxIdleConnsPerHost` junto a concorrência configurada e descarte eficiente de payloads (`io.Discard`).
* **Graceful Shutdown (`signal.NotifyContext`):** Captura de sinais do sistema operacional (`SIGINT` / `Ctrl+C` e `SIGTERM`), interrompendo novas requisições imediatamente e gerando o relatório parcial exato das requisições concluídas até o momento da interrupção.
* **Fail-Fast & Alta Testabilidade:** Validação prévia de parâmetros, arquivos e cabeçalhos antes da alocação de recursos de rede, com suíte de testes unitários com o padrão Table-Driven Tests.

---

## Estrutura do Projeto

O repositório segue o padrão de layout da comunidade Go (`Standard Go Project Layout`), isolando as regras de negócio no diretório `internal/`:

```text
gorate/
├── cmd/
│   └── gorate/
│       └── main.go               # Entrypoint do aplicativo e gerenciamento de sinais
├── internal/
│   ├── config/
│   │   ├── config.go             # Parsing de flags (FlagSet), leitura de arquivos e validação Fail-Fast
│   │   └── config_test.go        # Testes unitários de configuração e parsing de headers
│   ├── display/
│   │   ├── display.go            # Formatação tabular alinhada no terminal (text/tabwriter)
│   │   └── display_test.go       # Testes de renderização de saída
│   ├── metrics/
│   │   ├── metrics.go            # Cálculos estatísticos (RPS, Taxa de Erro, Min/Max e Percentis)
│   │   └── metrics_test.go       # Testes matemáticos e de agregação de relatórios
│   └── runner/
│       ├── runner.go             # Motor concorrente HTTP, Worker Pool e Rate Limiter
│       └── runner_test.go        # Testes de integração local com net/http/httptest
├── DESIGN.md                     # Especificação técnica, escopo e requisitos do sistema
└── go.mod                        # Definição do módulo Go
```

---

## Instalação e Compilação

### Pré-requisitos
* **Go 1.21** ou superior instalado.

### Compilando o binário
Clone o repositório e compile o executável na raiz do projeto:

```bash
git clone https://github.com/pukanski/Gorate.git
cd Gorate

# Compila o binário "gorate"
go build -o gorate ./cmd/gorate
```

Ele também pode ser executar diretamente sem gerar o binário usando `go run ./cmd/gorate [flags]`.

---

## Configuração (Flags Disponíveis)

| Flag | Tipo | Obrigatória? | Default | Descrição |
| :--- | :--- | :---: | :---: | :--- |
| `-url` | `string` | **Sim** | `""` | URL alvo única (ex: `https://api.com`) ou caminho para um arquivo `.txt` contendo múltiplas URLs (uma por linha). |
| `-n` | `int` | **Sim** | `0` | Total de requisições a serem disparadas para cada URL. |
| `-metodo` | `string` | Não | `"GET"` | Método HTTP utilizado nas requisições (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`). |
| `-concurrency` | `int` | Não | `10` | Número de trabalhadores (requisições simultâneas) por URL. |
| `-rate` | `int` | Não | `10` | Limite máximo de requisições por segundo (RPS) permitidas por URL. |
| `-timeout` | `int` | Não | `5000` | Tempo limite (*timeout*) de cada requisição individual em **milissegundos (ms)**. |
| `-H` | `string` | Não | — | Cabeçalho HTTP customizado no formato `"Chave: Valor"`. Pode ser repetido múltiplas vezes no mesmo comando. |
| `-body` | `string` | Não | `""` | Caminho para um arquivo contendo o corpo (*payload*) da requisição (permitido para `POST`, `PUT` e `PATCH`). |

---

## 💻 Exemplos de Uso

### 1. Teste de carga básico (GET)
Dispara **50 requisições** contra um único endpoint utilizando **10 conexões simultâneas** e limite de **25 requisições por segundo**:

```bash
./gorate -url https://httpbin.org/get -n 50 -concurrency 10 -rate 25
```

**Exemplo de Saída:**
```text
Iniciando load test em 1 URL(s)...
Concorrência: 10 | Total de Requisições por URL: 50
Pressione Ctrl+C para interromper e ver o relatório parcial.

URL                       RPS     Erros   4xx/5xx   Timeouts   P50     P95     P99     Min/Max         Total
https://httpbin.org/get   23.34   0.0%    0         0          141ms   448ms   529ms   141ms / 529ms   50
```

---

### 2. Requisição POST com Body via arquivo e múltiplos Headers (`-H`)
Envia um arquivo `payload.json` utilizando o método `POST` com cabeçalhos de autenticação e tipo de conteúdo:

```bash
./gorate \
  -url https://httpbin.org/post \
  -n 10 \
  -metodo POST \
  -body ./payload.json \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer tokenSecreto" \
  -rate 2 \
  -timeout 3000
```

**Exemplo de Saída:**
```text
Iniciando load test em 1 URL(s)...
Concorrência: 10 | Total de Requisições por URL: 10
Pressione Ctrl+C para interromper e ver o relatório parcial.

URL                        RPS    Erros   4xx/5xx   Timeouts   P50     P95     P99     Min/Max         Total
https://httpbin.org/post   1.94   0.0%    0         0          141ms   469ms   469ms   141ms / 469ms   10
```

---

### 3. Testando múltiplas URLs simultaneamente via arquivo
Crie um arquivo `urls.txt` (linhas em branco e comentários iniciados com `#` são ignorados automaticamente):

```text
# Endpoints de Produção
https://httpbin.org/status/200
https://httpbin.org/status/500
https://httpbin.org/delay/2
```

Execute o **gorate** apontando para o arquivo com um timeout curto (`1000ms`) para observar a separação entre sucesso, erros `5xx` e `Timeouts`:

```bash
./gorate -url urls.txt -n 10 -concurrency 5 -rate 10 -timeout 1000
```

**Exemplo de Saída:**
```text
Iniciando load test em 3 URL(s)...
Concorrência: 5 | Total de Requisições por URL: 10
Pressione Ctrl+C para interromper e ver o relatório parcial.

URL                              RPS    Erros    4xx/5xx   Timeouts   P50     P95     P99     Min/Max         Total
https://httpbin.org/status/200   8.76   0.0%     0         0          141ms   467ms   467ms   140ms / 467ms   10
https://httpbin.org/status/500   8.76   100.0%   10        0          141ms   467ms   467ms   140ms / 467ms   10
https://httpbin.org/delay/2      4.00   100.0%   0         10         0s      0s      0s      0s / 0s         10
```

---

## Métricas do Relatório

* **`URL`**: O endpoint testado
* **`RPS` (Requests Per Second)**: Vazão real alcançada pelo sistema durante a janela de execução daquela URL
* **`Erros`**: Porcentagem total de requisições que falharam (soma de erros de rede, timeouts e respostas HTTP `4xx`/`5xx`)
* **`4xx/5xx`**: Quantidade de respostas recebidas com código de status HTTP de erro do cliente (`400-499`) ou do servidor (`500-599`)
* **`Timeouts`**: Quantidade de requisições abortadas por excederem o tempo limite configurado na flag `-timeout`
* **`P50` (Mediana)**: Tempo máximo de resposta para 50% das requisições bem-sucedidas
* **`P95`**: Tempo máximo de resposta que cobre 95% das requisições
* **`P99`**: Tempo máximo de resposta que cobre 99% das requisições
* **`Min/Max`**: O menor e o maior tempo de resposta registrados nos testes
* **`Total`**: Total de requisições para aquela URL

---

## Executando os Testes Automatizados

O projeto conta com testes unitários cobrindo validação de flags, parsing de headers, cálculos estatísticos/percentis, formatação de saída e controle de concorrência/rate limiting.

Para rodar toda a suíte de testes com verificação de cobertura (*coverage*) e detector de condições de corrida (*race detector*):

```bash
go test -v -race -cover ./...
```

Saída esperada:
```text
        github.com/pukanski/Gorate/cmd/gorate           coverage: 0.0% of statements
ok      github.com/pukanski/Gorate/internal/config      1.025s  coverage: 89.2% of statements
ok      github.com/pukanski/Gorate/internal/display     1.021s  coverage: 100.0% of statements
ok      github.com/pukanski/Gorate/internal/metrics     1.020s  coverage: 100.0% of statements
ok      github.com/pukanski/Gorate/internal/runner      1.681s  coverage: 90.3% of statements
```