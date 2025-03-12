FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copiar arquivos de dependências
COPY go.mod go.sum* ./

# Baixar dependências
RUN go mod download

# Copiar o código  fonte
COPY . .

# Compilar a aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -o /video-conversor-go ./cmd/app

# Imagem final 
FROM golang:1.24-alpine

# Instalar FFmpeg e dependências
RUN apk add --no-cache ffmpeg ca-certificates tzdata

# Criar diretório para uploads
RUN mkdir -p /uploads

# Copiar o binário compilado
COPY --from=builder /video-conversor-go /usr/local/bin/video-conversor-go

# Definir diretório de trabalho
WORKDIR /app

# Copiar o código fonte para permitir desenvolvimento dentro do container
COPY . .

# Expor porta
EXPOSE 8080

# Comando para executar a aplicação
# Usando tail -f /dev/null como fallback para manter o container em execução
CMD ["sh", "-c", "tail -f /dev/null"]