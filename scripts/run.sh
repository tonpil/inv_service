set -a
source ./config.env
set +a

# Запуск сервиса
go run ./cmd/doc_service/main.go