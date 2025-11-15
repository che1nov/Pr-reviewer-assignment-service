#!/bin/bash

BASE_URL="http://localhost:8080"
ADMIN_TOKEN="admin-secret"
USER_TOKEN="user-secret"

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "PR Reviewer Assignment Service"
echo ""

# Функция для красивого вывода
test_endpoint() {
    local name=$1
    local expected_code=$2
    local response
    
    echo -e "${YELLOW}=== $name ===${NC}"
    response=$(eval "$3")
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | head -n-1)
    
    echo "$body" | jq '.' 2>/dev/null || echo "$body"
    
    if [ "$http_code" = "$expected_code" ]; then
        echo -e "${GREEN} Статус: $http_code (ожидалось: $expected_code)${NC}"
    else
        echo -e "${RED} Статус: $http_code (ожидалось: $expected_code)${NC}"
    fi
    echo ""
}

echo "Тест 1: Создание команды 'backend' с 5 участниками"
test_endpoint "POST /team/add (backend)" "201" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/team/add \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"team_name\": \"backend\",
    \"members\": [
      {\"user_id\": \"u1\", \"username\": \"Alice\", \"is_active\": true},
      {\"user_id\": \"u2\", \"username\": \"Bob\", \"is_active\": true},
      {\"user_id\": \"u3\", \"username\": \"Charlie\", \"is_active\": true},
      {\"user_id\": \"u4\", \"username\": \"David\", \"is_active\": true},
      {\"user_id\": \"u5\", \"username\": \"Eve\", \"is_active\": true}
    ]
  }"'

echo "Тест 2: Создание команды 'frontend' с 3 участниками"
test_endpoint "POST /team/add (frontend)" "201" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/team/add \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"team_name\": \"frontend\",
    \"members\": [
      {\"user_id\": \"u6\", \"username\": \"Frank\", \"is_active\": true},
      {\"user_id\": \"u7\", \"username\": \"Grace\", \"is_active\": true},
      {\"user_id\": \"u8\", \"username\": \"Henry\", \"is_active\": true}
    ]
  }"'

echo "Тест 3: Попытка создать дубликат команды (ошибка)"
test_endpoint "POST /team/add (дубликат)" "400" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/team/add \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"team_name\": \"backend\",
    \"members\": []
  }"'

echo "Тест 4: Получение команды 'backend'"
test_endpoint "GET /team/get (backend)" "200" \
'curl -s -w "\n%{http_code}" -X GET "$BASE_URL/team/get?team_name=backend" \
  -H "Authorization: Bearer $USER_TOKEN"'

echo "Тест 5: Получение несуществующей команды (ошибка)"
test_endpoint "GET /team/get (не существует)" "404" \
'curl -s -w "\n%{http_code}" -X GET "$BASE_URL/team/get?team_name=nonexistent" \
  -H "Authorization: Bearer $USER_TOKEN"'

echo "Тест 6: Создание PR от Alice (u1)"
test_endpoint "POST /pullRequest/create (pr-1001)" "201" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/pullRequest/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"pull_request_id\": \"pr-1001\",
    \"pull_request_name\": \"Add authentication feature\",
    \"author_id\": \"u1\"
  }"'

echo "Тест 7: Создание PR от Bob (u2)"
test_endpoint "POST /pullRequest/create (pr-1002)" "201" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/pullRequest/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"pull_request_id\": \"pr-1002\",
    \"pull_request_name\": \"Fix login bug\",
    \"author_id\": \"u2\"
  }"'

echo "Тест 8: Создание PR от Frank (u6) из команды frontend"
test_endpoint "POST /pullRequest/create (pr-2001)" "201" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/pullRequest/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"pull_request_id\": \"pr-2001\",
    \"pull_request_name\": \"Add new UI component\",
    \"author_id\": \"u6\"
  }"'

echo "Тест 9: Попытка создать дубликат PR (ошибка)"
test_endpoint "POST /pullRequest/create (дубликат)" "409" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/pullRequest/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"pull_request_id\": \"pr-1001\",
    \"pull_request_name\": \"Duplicate\",
    \"author_id\": \"u1\"
  }"'

echo "Тест 10: Получение PR для ревьювера Bob (u2)"
test_endpoint "GET /users/getReview (u2)" "200" \
'curl -s -w "\n%{http_code}" -X GET "$BASE_URL/users/getReview?user_id=u2" \
  -H "Authorization: Bearer $USER_TOKEN"'

echo "Тест 11: Получение PR для ревьювера Charlie (u3)"
test_endpoint "GET /users/getReview (u3)" "200" \
'curl -s -w "\n%{http_code}" -X GET "$BASE_URL/users/getReview?user_id=u3" \
  -H "Authorization: Bearer $USER_TOKEN"'

echo "Тест 12: Создание команды 'devops' с 4 участниками для теста переназначения"
test_endpoint "POST /team/add (devops)" "201" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/team/add \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"team_name\": \"devops\",
    \"members\": [
      {\"user_id\": \"u9\", \"username\": \"Ivan\", \"is_active\": true},
      {\"user_id\": \"u10\", \"username\": \"Julia\", \"is_active\": true},
      {\"user_id\": \"u11\", \"username\": \"Kevin\", \"is_active\": true},
      {\"user_id\": \"u12\", \"username\": \"Laura\", \"is_active\": true}
    ]
  }"'

echo "Тест 12a: Создание PR от Ivan (u9) для теста переназначения"
test_endpoint "POST /pullRequest/create (pr-3001)" "201" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/pullRequest/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"pull_request_id\": \"pr-3001\",
    \"pull_request_name\": \"Add CI/CD pipeline\",
    \"author_id\": \"u9\"
  }"'

echo "Тест 12b: Переназначение ревьювера в pr-3001"
test_endpoint "POST /pullRequest/reassign" "200" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"pull_request_id\": \"pr-3001\",
    \"old_user_id\": \"u10\"
  }"'

echo "Тест 13: Деактивация пользователя Bob (u2)"
test_endpoint "POST /users/setIsActive (деактивация)" "200" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/users/setIsActive \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"user_id\": \"u2\",
    \"is_active\": false
  }"'

echo "Тест 14: Активация пользователя Bob (u2)"
test_endpoint "POST /users/setIsActive (активация)" "200" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/users/setIsActive \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"user_id\": \"u2\",
    \"is_active\": true
  }"'

echo "Тест 15: Merge PR pr-1001"
test_endpoint "POST /pullRequest/merge (первый раз)" "200" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/pullRequest/merge \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"pull_request_id\": \"pr-1001\"
  }"'

echo "Тест 16: Повторный merge PR pr-1001 (идемпотентность)"
test_endpoint "POST /pullRequest/merge (повторный)" "200" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/pullRequest/merge \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"pull_request_id\": \"pr-1001\"
  }"'

echo "Тест 17: Попытка переназначить ревьювера на merged PR (ошибка)"
test_endpoint "POST /pullRequest/reassign (на merged PR)" "409" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"pull_request_id\": \"pr-1001\",
    \"old_user_id\": \"u3\"
  }"'

echo "Тест 18: Попытка merge несуществующего PR (ошибка)"
test_endpoint "POST /pullRequest/merge (не существует)" "404" \
'curl -s -w "\n%{http_code}" -X POST $BASE_URL/pullRequest/merge \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d "{
    \"pull_request_id\": \"pr-9999\"
  }"'

echo "Тест 19: Запрос без токена (ошибка)"
test_endpoint "GET /team/get (без токена)" "401" \
'curl -s -w "\n%{http_code}" -X GET "$BASE_URL/team/get?team_name=backend"'

echo "Тест 20: Запрос с неверным токеном (ошибка)"
test_endpoint "GET /team/get (неверный токен)" "401" \
'curl -s -w "\n%{http_code}" -X GET "$BASE_URL/team/get?team_name=backend" \
  -H "Authorization: Bearer wrong-token"'

echo ""

echo "Тестирование завершено!"
