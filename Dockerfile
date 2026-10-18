# --- стадия 1: сборка ---
FROM golang:1.27-alpine AS builder
WORKDIR /app

# кэшируем зависимости отдельно: слой пересобирается
# только при изменении go.mod/go.sum, а не при каждой правке кода
COPY go.mod go.sum ./
RUN go mod download

# копируем весь код и собираем статический бинарник под Linux
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# --- стадия 2: рантайм ---
FROM alpine:3.20
WORKDIR /app

# забираем из builder только готовый бинарник
COPY --from=builder /app/server ./server

EXPOSE 8080
CMD ["./server"]
