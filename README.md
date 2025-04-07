# HTTP a partir de TCP

## Visão Geral
Este projeto implementa um servidor HTTP construído diretamente sobre sockets TCP. Ele foi desenvolvido para demonstrar como funciona o protocolo HTTP no nível mais baixo, manipulando manualmente as requisições e respostas HTTP.

O foco principal é o servidor HTTP, enquanto os componentes de escuta TCP e envio UDP são mantidos como utilitários de desenvolvimento que não são rastreados no controle de versão.

## Estrutura do Repositório

O projeto está organizado em vários aplicativos de linha de comando:

```
cmd/
  httpserver/    # Implementação principal do servidor HTTP
internal/
  headers/       # Manipulação de cabeçalhos HTTP
  request/       # Processamento de requisições HTTP
  respose/       # Geração de respostas HTTP
  server/        # Implementação do servidor e manipuladores
```

## Servidor HTTP

O servidor HTTP é o componente principal deste projeto. Ele implementa uma versão simplificada do protocolo HTTP/1.1, incluindo:

### Recursos

- Análise e processamento manual de requisições HTTP
- Suporte a diferentes códigos de status (200, 400, 500)
- Manipulação de cabeçalhos HTTP
- Suporte para Transfer-Encoding: chunked
- Suporte para trailers HTTP
- Proxy simples para httpbin.org
- Páginas de erro personalizadas
- Servir conteúdo estático (HTML, vídeo)

## Instalação

Para instalar e executar o servidor HTTP:

```bash
# Clone o repositório
git clone https://github.com/seuusuario/httpfromtcp.git

# Navegue até o diretório do servidor HTTP
cd httpfromtcp/cmd/httpserver

# Compile o servidor
go build

# Execute o servidor
./httpserver
```

## Uso

O servidor escuta na porta 42069 por padrão e oferece os seguintes endpoints:

- `/`: Retorna uma página de sucesso (200 OK)
- `/yourproblem`: Retorna um erro 400 Bad Request
- `/myproblem`: Retorna um erro 500 Internal Server Error
- `/video`: Serve um arquivo de vídeo (requer o arquivo vim.mp4 no diretório assets)
- `/httpbin/*`: Funciona como proxy para httpbin.org

## Implementação Técnica

Este projeto implementa manualmente o processamento de requisições HTTP e a geração de respostas, incluindo:

- Análise da linha de requisição (método, caminho, versão HTTP)
- Análise e manipulação de cabeçalhos HTTP
- Processamento de corpo de requisição
- Geração de respostas com códigos de status apropriados
- Suporte a Transfer-Encoding: chunked para respostas
- Geração de trailers HTTP
