# Auth Service (JWT + Refresh Tokens)

REST API-сервис на Go (Fiber v2) для аутентификации пользователей с использованием JWT access и refresh токенов.


## Используемый стек

1) Go 1.24
2) Fiber v2
3) PostgreSQL
4) Swagger

## Запуск
```bash
docker-compose -f docker-compose.yml up -d
```


### Приложение будет доступно на:

**API: http://127.0.0.1:8080**

**Swagger: http://127.0.0.1:8080/swagger/index.html**

Структура проекта
```
auth-service/
├── config/            - YML-конфигурации
├── connections/       - Подключение к PostgreSQL
├── docs/              - Swagger-документация
├── models/            - DTO и сущности
├── repositories/      - Слой доступа к данным
├── routers/           - HTTP-хендлер
├── services/          - Логика токенов и пользователей
├── webhook/           - Отправка webhook'а
├── main.go
├── Dockerfile
├── docker-compose.yml
├── go.mod / go.sum
└── README.md
```

| Метод  | Путь                  | Описание                                                             |
|--------|-----------------------|----------------------------------------------------------------------|
| GET    | `/api/tokens`         | Получить access + refresh токены                                     |
| POST   | `/api/refresh`        | Обновить пару токенов                                                |
| POST   | `/api/me`             | Получить GUID пользователя по токену                                 |
| POST   | `/api/logout`         | Удалить refresh токен (выйти из сессии)                              |
| GET    | `/api/get-users-GUID` | Получить список пользователей (GUID) - Путь сделан для проверяющего! |

**При отсутствии пользователей вызывается `/api/get-users-GUID` в `config/config.yml` можете выставить необходимое кол-во пользователей, которые будут создаваться** 

## Конфигурация (config/config.yml)
```yaml
application:
  host: localhost
  prefix: app-
  port: 8080
  name: app
postgres:
  host: db
  port: 5432
  user: postgres
  database: auth_service
  password: postgres
usr:
  count: 10 # количество пользователей
jwt:
  issuer: "www.issuer.com"
  secret_key: "super-secret"
webhook:
  url: "" # указываем необходимый адрес для отправки веб-хука

```

## Пример запроса на получение GUID пользователей
```bash
curl -X GET "http://localhost:8080/api/get-users-GUID"
```

Пример ответа:
```json
[
  {"guid": "a1b2c3d4-e5f6-7890"},
  {"guid": "b2c3d4e5-f6a7-8901"}
]
```

