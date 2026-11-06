package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestParseFlags(t *testing.T) {
	tmpDir := t.TempDir()

	arquivosURLsValidos := filepath.Join(tmpDir, "urls.txt")
	conteudoURLs := []byte("https://google.com\n\n# Isso é um comentário\nhttps://g1.globo.com/\n")
	err := os.WriteFile(arquivosURLsValidos, conteudoURLs, 0644)
	if err != nil {
		t.Fatalf("falha ao criar arquivo de urls temporário: %v", err)
	}

	arquivoVazio := filepath.Join(tmpDir, "vazio.txt")
	os.WriteFile(arquivoVazio, []byte("\n# comentário\n"), 0644)

	bodyValido := filepath.Join(tmpDir, "payload.json")
	conteudoBody := []byte(`{"nome": "docker"}`)
	os.WriteFile(bodyValido, conteudoBody, 0644)

	arquivoInexistente := filepath.Join(tmpDir, "naoExiste.txt")

	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		wantConfig *Config
	}{

		//RNF05: Validação das Flags Básicas
		{
			name:    "Falha se nenhuma URL for passada",
			args:    []string{"-n", "10"},
			wantErr: true,
		},
		{
			name:    "Falha se TotalRequests (n) for zero ou negativo",
			args:    []string{"-url", "https://google.com", "-n", "0"},
			wantErr: true,
		},
		{
			name:    "Falha se a URL for inválida",
			args:    []string{"-url", "http:// url errada", "-n", "10"},
			wantErr: true,
		},
		{
			name:    "Falha se o método HTTP for inválido",
			args:    []string{"-url", "https://google.com", "-n", "10", "-metodo", "INVALIDO"},
			wantErr: true,
		},
		{
			name:    "Falha se enviar body com método GET",
			args:    []string{"-url", "https://google.com", "-n", "10", "-metodo", "GET", "-body", "payload.json"},
			wantErr: true,
		},
		{
			name:    "Deve ajustar Concurrency para ser igual ao TotalRequests se for maior",
			args:    []string{"-url", "https://google.com", "-n", "5", "-concurrency", "20"},
			wantErr: false,
			wantConfig: &Config{
				URLs:          []string{"https://google.com"},
				TotalRequests: 5,
				Concurrency:   5,
				Method:        MethodGet,
				HeadersCustom: map[string]string{},
				Timeout:       5000 * time.Millisecond,
				RateLimit:     10,
			},
		},

		// RNF03: Parsing dos Headers Customizados (-H)
		{
			name:    "Deve parsear corretamente um header customizado",
			args:    []string{"-url", "https://google.com", "-n", "10", "-H", "Authorization: Bearer token123"},
			wantErr: false,
			wantConfig: &Config{
				URLs:          []string{"https://google.com"},
				TotalRequests: 10,
				Concurrency:   10,
				Method:        MethodGet,
				HeadersCustom: map[string]string{
					"Authorization": "Bearer token123",
				},
				Timeout:   5000 * time.Millisecond,
				RateLimit: 10,
			},
		},
		{
			name: "Deve parsear múltiplos headers customizados",
			args: []string{
				"-url", "https://google.com",
				"-n", "10",
				"-H", "Accept: application/json",
				"-H", "X-Custom-Header: MeuValor",
			},
			wantErr: false,
			wantConfig: &Config{
				URLs:          []string{"https://google.com"},
				TotalRequests: 10,
				Concurrency:   10,
				Method:        MethodGet,
				HeadersCustom: map[string]string{
					"Accept":          "application/json",
					"X-Custom-Header": "MeuValor",
				},
				Timeout:   5000 * time.Millisecond,
				RateLimit: 10,
			},
		},
		{
			name:    "Falha se o header não tiver os dois pontos ':'",
			args:    []string{"-url", "https://google.com", "-n", "10", "-H", "HeaderSemDoisPontos"},
			wantErr: true,
		},
		{
			name:    "Falha se o header tiver chave ou valor vazios",
			args:    []string{"-url", "https://google.com", "-n", "10", "-H", "Authorization: "},
			wantErr: true,
		},

		// RNF: Validação de Arquivos (t.TempDir)
		{
			name:    "Deve ler um arquivo de URLs com sucesso ignorando comentários",
			args:    []string{"-url", arquivosURLsValidos, "-n", "10"},
			wantErr: false,
			wantConfig: &Config{
				URLs:          []string{"https://google.com", "https://g1.globo.com/"},
				TotalRequests: 10,
				Concurrency:   10,
				Method:        MethodGet,
				HeadersCustom: map[string]string{},
				Timeout:       5000 * time.Millisecond,
				RateLimit:     10,
			},
		},
		{
			name:    "Falha se o arquivo de URLs não existir",
			args:    []string{"-url", arquivoInexistente, "-n", "10"},
			wantErr: true,
		},
		{
			name:    "Falha se o arquivo de URLs estiver vazio ou só tiver comentários",
			args:    []string{"-url", arquivoVazio, "-n", "10"},
			wantErr: true,
		},
		{
			name:    "Deve ler o arquivo de Body com sucesso num método POST",
			args:    []string{"-url", "https://google.com", "-n", "10", "-metodo", "POST", "-body", bodyValido},
			wantErr: false,
			wantConfig: &Config{
				URLs:          []string{"https://google.com"},
				TotalRequests: 10,
				Concurrency:   10,
				Method:        MethodPost,
				Body:          []byte(`{"nome": "docker"}`),
				HeadersCustom: map[string]string{},
				Timeout:       5000 * time.Millisecond,
				RateLimit:     10,
			},
		},
		{
			name:    "Falha se o arquivo de Body não existir",
			args:    []string{"-url", "https://google.com", "-n", "10", "-metodo", "POST", "-body", arquivoInexistente},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfig, err := ParseFlags(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFlags() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.wantConfig != nil {
				if !reflect.DeepEqual(gotConfig, tt.wantConfig) {
					t.Errorf("ParseFlags() = %+v, want %+v", gotConfig, tt.wantConfig)
				}
			}
		})
	}
}
