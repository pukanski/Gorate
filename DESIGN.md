# DESIGN.md — gorate

## Objetivo
Ferramenta CLI de load testing, escrita em Go, que dispara requisições HTTP concorrentes contra uma ou mais URLs e gera um relatório de latência com percentis (p50/p95/p99), taxa de erros, requisições por segundo, latência miníma/máxima e o total de requisições.

## Escopo
### Incluído
- Múltiplas URLs
- Rate limiting + concorrência
- Relatório com percentis
- Métodos suportados: GET/POST/PUT/DELETE/PATCH
- Headers Customizados
- Suporte a Body atráves de arquivos

### Fora de escopo (e por quê)
- JSON inline devido a complexidade no manuseio dele dentro do CLI

## Inputs (flags)
| Flag | Tipo | Obrigatória? | Descrição |
|------|------|--------------|-----------|
| URLs | String/File | sim | O endereço (inline) ou arquivo (file) com os endereços da(s) URL(s) alvo(s)|
| Total de Requisições | int | sim | O número de requisições desejadas |
| Método | String | não | Método HTTP que ira ser usado para as requisições, com valor default GET |
| Body | File | não | O body da requisição selecionada |
| Headers Customizados | String | não | Os headers da requisição |
| Timeout | int | não | Tempo em ms para o timeout da requisição |
| Concorrência | int | não | O número de requisições simultâneas para a URL, com valor default de 10 |
| Rate limit | int | não | O número de requisições por segundo permitidas, com valor default de 10 |


## Outputs
Tabela com o relatório contendo 

| Coluna |Nome Completo |Descrição |
|------|-----------|-----------|
|URL|Caminho ou Endpoint|O endpoint testado|
|RPS|Requisições por Segundo|Indica quantas requisições por segundo o sistema conseguiu processar durante o teste|
|Erros|Taxa de Erro|A porcentagem de requisições que falharam em relação ao total|
|4xx/5xx|Status Codes 4xx/5xx|O número de respostas com código 4xx ou 5xx|
|Timeouts|Timeouts|O número de timeouts das requisições |
|P50|Percentil 50|O tempo que 50% das requisições levaram para ser respondidas. Representa a experiência típica dos usuários|
|P95|Percentil 95|O tempo que cobre 95% das requisições. Indica o comportamento para quase todos os usuários|
|P99|Percentil 99|O tempo que cobre 99% das requisições. Mostra o pior cenário para os usuários|
|Min/Max| Miníma/Máxima|O menor e o maior tempo de resposta para aquela URL. Mostra os extremos de velocidade|
|Total|Total de Requisições|Quantidade de requisições disparadas e processadas para aquela URL ao longo de todo testes|

## Critérios de sucesso
- Teste manual verificando o tempo de execução de 10 requisições para um servidor de teste que atrasa cada resposta em 1 segundo, para garantir a concorrência das requisições

### Funcionais
| Código | Descrição |
|------|------|
|RF01|O sistema deve disparar requisições HTTP contra uma ou mais URLs simultaneamente|
|RF02|O usuário pode mandar um arquivo contendo o body da requisição|
|RF03|O usuário pode mandar headers customizados para a requisição|
|RF04|O usuário pode configurar um timeout em milisegundos (ms) para a requisição|
|RF05|O usuário pode configurar o rate limit para controlar a quantidade de requisições por segundo|
|RF06|O usuário pode configurar o número de requisições simultâneas para a URL(s) alvo(s)|
|RF07|O sistema deve mostrar um relatório final mostrando os seguintes dados: urls, rps, taxa de erro, p50, p95, p99, min/max e total de requisições|

### Robustez
- A ferramenta não deve travar (panic) se o servidor cair ou recusar conexão no meio do teste, os erros de rede devem ser capturados e contados, não derrubar o programa
- Timeouts contados seperademento de erros HTTP (4xx/5xx) no relatório final
- Se o arquivo de URLs ou o arquivo de body não existir/for inválido, a ferramenta deve mostrar uma mensagem de erro legível antes de começar a disparar requisições.
- Caso o usuário interrompa com Ctrl+C no meio do teste deve exibir um relatório parcial.

### Não-funcionais
| Código | Descrição |
|------|------|
|RNF01|README com exemplos de uso|
|RNF02|Testes unitários cobrindo o cálculo de percentis|
|RNF03|Testes unitários cobrindo o parsing dos headers customizados|
|RNF04|Testes unitários cobrindo o módulo de estatística e métricas|
|RNF05|Testes unitários cobrindo a validação das flags|
|RNF06|Testes unitários cobrindo o rate limiting|
