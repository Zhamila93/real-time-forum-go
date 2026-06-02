# Real-Time Forum (SPA)

Форум-SPA с регистрацией, постами, комментариями и приватными сообщениями в реальном времени.

## Стек

- **Backend:** Go (clean architecture: `cmd` + `internal`)
- **WebSocket:** gorilla/websocket, thread-safe Hub
- **Database:** SQLite (mattn/go-sqlite3)
- **Frontend:** чистый HTML/CSS/JS (SPA), `frontend/`

## Структура проекта

```
REAL-TIME-FORUM/
├── cmd/
│   └── server/
│       └── main.go                 # точка входа, DI и маршруты
├── internal/
│   ├── config/                     # адрес, путь к БД, статика
│   ├── domain/                     # модели предметной области
│   ├── database/                   # миграции и seed категорий
│   ├── repository/                 # SQL (параметризованные запросы)
│   ├── service/                    # бизнес-логика
│   └── transport/
│       ├── http/                   # REST API + cookies сессий
│       └── websocket/              # Hub, Client, real-time chat
├── frontend/                       # SPA (HTML/CSS/JS)
├── schema.sql                      # справочная схема
├── go.mod
└── forum.db                        # SQLite (создаётся при запуске)
```

### Слои

| Слой | Ответственность |
|------|-----------------|
| `cmd/server` | Сборка зависимостей, `ListenAndServe` |
| `transport/http` | JSON, cookies, auth middleware |
| `transport/websocket` | Hub (mutex + channels), pumps |
| `service` | Валидация, оркестрация |
| `repository` | Доступ к SQLite |
| `domain` | Типы без инфраструктуры |

## Запуск

```bash
go mod tidy
go run ./cmd/server
# → http://localhost:8080
```

`forum.db` создаётся автоматически при первом запуске.

## Возможности

- Регистрация / логин / сессии (cookie, bcrypt)
- Посты с категориями, комментарии, фильтр по категории
- Приватный чат в реальном времени (WebSocket)
- Онлайн-статус, typing, реконнект на клиенте

## WebSocket Hub

- Один goroutine `Run()` владеет картами клиентов (`register` / `unregister` / `broadcast`)
- `sync.RWMutex` для `SendToUser` и `OnlineUsersMap`
- Сообщения чата сохраняются в `ChatService` (не в Hub); Hub только рассылает
- `BroadcastForum()` рассылает `reaction_updated`, `post_created`, `comment_created` всем клиентам

## API

| Метод | URL | Описание |
|-------|-----|----------|
| POST | /api/register | Регистрация |
| POST | /api/login | Вход |
| POST | /api/logout | Выход |
| GET | /api/me | Текущий пользователь |
| GET | /api/posts | Лента (`?category=`) |
| POST | /api/posts | Создать пост |
| GET | /api/posts/{id} | Пост + комментарии |
| POST | /api/posts/{id}/reaction | Like / dislike / remove (`{"reaction":"like"}`) |
| POST | /api/comments | Комментарий |
| GET | /api/categories | Категории |
| GET | /api/users | Список пользователей |
| GET | /api/messages?with=&offset=&limit= | История чата |
| GET | /ws | WebSocket |
