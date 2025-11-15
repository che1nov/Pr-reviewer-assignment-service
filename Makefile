.PHONY: help build run test test-unit test-integration test-all lint clean docker-build docker-up docker-down docker-logs migrate load-test

help:
	@echo "Доступные команды:"
	@echo "  make build              - Собрать бинарник"
	@echo "  make run                - Запустить сервис локально"
	@echo "  make test               - Запустить все тесты"
	@echo "  make test-unit          - Запустить unit тесты"
	@echo "  make test-integration   - Запустить integration тесты"
	@echo "  make lint               - Запустить линтер"
	@echo "  make clean              - Очистить артефакты сборки"
	@echo "  make docker-build       - Собрать Docker образ"
	@echo "  make docker-up          - Запустить через docker-compose"
	@echo "  make docker-down        - Остановить docker-compose"
	@echo "  make docker-logs        - Посмотреть логи"
	@echo "  make load-test          - Запустить нагрузочное тестирование"

build:
	@echo "Сборка проекта..."
	go build -o bin/service ./cmd/api
	@echo "Готово: bin/service"

run: build
	@echo "Запуск сервиса..."
	./bin/service

test: test-unit test-integration
	@echo "Все тесты пройдены"

test-unit:
	@echo "Запуск unit тестов..."
	go test -v -count=1 ./internal/...

test-integration:
	@echo "Запуск integration тестов..."
	@echo "Требуется запущенная PostgreSQL (make docker-up)"
	go test -v -count=1 ./tests/integration/...

test-all:
	@echo "Запуск всех тестов..."
	go test -v -count=1 ./...

lint:
	@echo "Запуск линтера..."
	golangci-lint run --timeout 5m

clean:
	@echo "Очистка..."
	rm -rf bin/
	rm -f load_test_results.json
	go clean
	@echo "Готово"

docker-build:
	@echo "Сборка Docker образа..."
	docker-compose build

docker-up:
	@echo "Запуск через docker-compose..."
	docker-compose up -d
	@echo "Сервис доступен на http://localhost:8080"

docker-down:
	@echo "Остановка docker-compose..."
	docker-compose down

docker-logs:
	docker-compose logs -f app

migrate:
	@echo "Миграции применяются автоматически при запуске"

load-test: docker-up
	@echo "Ожидание запуска сервиса..."
	@sleep 5
	@echo "Запуск нагрузочного тестирования..."
	k6 run load_test.js
	@echo "Результаты в load_test_results.json"

