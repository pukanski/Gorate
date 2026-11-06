package config

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type HTTPMethod string
type HeaderFlag []string

func (h *HeaderFlag) String() string {
	return strings.Join(*h, ", ")
}

func (h *HeaderFlag) Set(value string) error {
	*h = append(*h, value)
	return nil
}

const (
	MethodGet    HTTPMethod = http.MethodGet
	MethodPost   HTTPMethod = http.MethodPost
	MethodPut    HTTPMethod = http.MethodPut
	MethodDelete HTTPMethod = http.MethodDelete
	MethodPatch  HTTPMethod = http.MethodPatch
)

type Config struct {
	URLs          []string
	TotalRequests int
	Method        HTTPMethod
	Body          []byte
	HeadersCustom map[string]string
	Timeout       time.Duration
	Concurrency   int
	RateLimit     int
}

func ParseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("gorate", flag.ContinueOnError)

	var headers HeaderFlag

	urlFlag := fs.String("url", "", "Url ou arquivo a serem usadas")
	methodFlag := fs.String("metodo", "GET", "Método HTTP (GET, POST, PUT, DELETE, PATCH)")
	bodyFlag := fs.String("body", "", "Body da requisição")
	fs.Var(&headers, "H", "Adicionar cabeçalho HTTP (pode ser repetido)")
	tempoFlag := fs.Int64("timeout", 5000, "Tempo para timeout em milisegundos")
	//outras flags opcionais colocar depois

	//outras flags (TotalRequests, Concurrency, RateLimit)

	totalReqFlag := fs.Int("n", 0, "O total de requisições feitas para cada url")
	concurrencyFlag := fs.Int("concurrency", 10, "O número de requisições simultâneas para a URL")
	rateLimitFlag := fs.Int("rate", 10, "O número de requisições por segundo permitidas")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	urlsValidas, err := validateURLs(*urlFlag)
	if err != nil {
		return nil, err
	}

	metodoValido, err := validateMethod(*methodFlag)
	if err != nil {
		return nil, err
	}

	var body []byte

	if *bodyFlag != "" {
		if metodoValido == MethodGet || metodoValido == MethodDelete {
			return nil, fmt.Errorf("não é permitido enviar body (arquivo %s) com o metodo %s", *bodyFlag, metodoValido)
		}

		body, err = validateBody(*bodyFlag)
		if err != nil {
			return nil, err
		}

	}

	headersMap, err := validateHeaders(headers)
	if err != nil {
		return nil, err
	}

	timeout, err := validateTimeout(*tempoFlag)
	if err != nil {
		return nil, err
	}

	totalReq, err := validateTotalRequests(*totalReqFlag)
	if err != nil {
		return nil, err
	}

	concurrencyValida, err := validateConcurrency(*concurrencyFlag)
	if err != nil {
		return nil, err
	}
	if *concurrencyFlag > totalReq {
		concurrencyValida = totalReq
	}

	rateLimitValido, err := validateRateLimit(*rateLimitFlag)
	if err != nil {
		return nil, err
	}

	return &Config{
		URLs:          urlsValidas,
		TotalRequests: totalReq,
		Method:        metodoValido,
		Body:          body,
		HeadersCustom: headersMap,
		Timeout:       timeout,
		Concurrency:   concurrencyValida,
		RateLimit:     rateLimitValido,
	}, nil
}

func validateURLs(urlInput string) (urlValida []string, err error) {
	if urlInput == "" {
		err = fmt.Errorf("não foi informada nenhuma url/arquivo")
		return nil, err
	}

	if strings.HasPrefix(urlInput, "http://") || strings.HasPrefix(urlInput, "https://") {
		u, err := url.Parse(urlInput)
		if err != nil {
			return nil, fmt.Errorf("formato de url invalido: %w", err)
		}
		if u.Host == "" {
			return nil, fmt.Errorf("a url informada precisa conter um host valido")
		}
		return []string{urlInput}, nil
	}

	arquivo, err := os.Open(urlInput)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir o arquivo de urls: %w", err)
	}
	defer arquivo.Close()

	scanner := bufio.NewScanner(arquivo)
	urlsValidas := []string{}

	for scanner.Scan() {
		linha := strings.TrimSpace(scanner.Text())
		if linha == "" || strings.HasPrefix(linha, "#") {
			continue
		}

		parsedLinha, err := url.Parse(linha)
		if err != nil || parsedLinha.Scheme == "" || parsedLinha.Host == "" {
			return nil, fmt.Errorf("url invalida dentro do arquivo: %s", linha)
		}

		urlsValidas = append(urlsValidas, linha)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("erro na leitura do arquivo de urls: %w", err)
	}

	if len(urlsValidas) == 0 {
		return nil, fmt.Errorf("o arquivo de urls esta vazio ou só contem comentarios")
	}

	return urlsValidas, nil
}

func validateMethod(method string) (metodo HTTPMethod, err error) {
	metodo = HTTPMethod(strings.ToUpper(strings.TrimSpace(method)))
	switch metodo {
	case MethodGet, MethodPost, MethodPut, MethodDelete, MethodPatch:
		return metodo, nil
	default:
		err = fmt.Errorf("método HTTP inválido: %s", method)
		return "", err
	}
}

func validateBody(caminhoArquivo string) (conteudo []byte, err error) {
	if strings.TrimSpace(caminhoArquivo) == "" {
		return nil, nil
	}

	conteudo, err = os.ReadFile(caminhoArquivo)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler arquivo do body: %w", err)
	}

	return conteudo, nil
}

func validateHeaders(headersList []string) (mapa map[string]string, err error) {
	mapa = make(map[string]string)

	for _, h := range headersList {
		partes := strings.SplitN(h, ":", 2)
		if len(partes) != 2 {
			return nil, fmt.Errorf("header com formato invalido: %s", h)
		}

		chave := strings.TrimSpace(partes[0])
		valor := strings.TrimSpace(partes[1])

		if chave == "" || valor == "" {
			return nil, fmt.Errorf("header inválido: chave ou valor vazios em '%s'", h)
		}

		mapa[chave] = valor
	}

	return mapa, nil
}

func validateTimeout(tempo int64) (timeout time.Duration, err error) {
	if tempo <= 0 {
		err = fmt.Errorf("o tempo deve ser maior que zero")
		return 0, err
	}

	return (time.Duration(tempo) * time.Millisecond), nil
}

func validateTotalRequests(totalRequests int) (totalReq int, err error) {
	if totalRequests <= 0 {
		return 0, fmt.Errorf("o número de requisições deve ser maior que 0")
	}

	return totalRequests, nil
}

func validateConcurrency(concurrency int) (totalConcurrency int, err error) {
	if concurrency <= 0 {
		return 0, fmt.Errorf("o número de requisições simultâneas deve ser maior que 0")
	}

	return concurrency, nil
}

func validateRateLimit(rate int) (totalRate int, err error) {
	if rate <= 0 {
		return 0, fmt.Errorf("o número de requisições por segundo deve ser maior que 0")
	}

	return rate, nil
}
