# Integration тесты

## Описание

Integration тесты проверяют работу сервиса через HTTP API с реальной базой данных PostgreSQL.

## Что тестируется

### TestFullWorkflow
Полный сценарий работы:
- Создание команды с пользователями
- Получение команды
- Создание PR с автоназначением ревьюверов
- Получение PR для ревьювера
- Переназначение ревьювера
- Merge PR
- Проверка идемпотентности merge
- Проверка запрета переназначения после merge

### TestStatistics
Работа эндпоинта статистики:
- Создание нескольких PR
- Merge одного из них
- Получение и проверка статистики

### TestDeactivateTeamUsers
Массовая деактивация пользователей:
- Создание двух команд
- Создание PR
- Деактивация всей команды
- Проверка результатов деактивации

### TestUserActivation
Управление активностью пользователей:
- Деактивация пользователя
- Активация обратно

## Запуск

Тесты требуют запущенную PostgreSQL с тестовой БД.

### Подготовка

```bash
# Создать тестовую БД (если еще не создана)
docker exec pr-reviewer-postgres psql -U app -d pr_service -c "CREATE DATABASE pr_service_test;"
```

### Запуск тестов

```bash
# Запуск integration тестов
go test -v ./tests/integration/... -count=1

# Запуск всех тестов (unit + integration)
go test -v ./... -count=1
```

## Результат

Все 4 теста должны пройти успешно:
```
PASS: TestFullWorkflow
PASS: TestStatistics
PASS: TestDeactivateTeamUsers
PASS: TestUserActivation
```

