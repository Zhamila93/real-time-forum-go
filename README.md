# Real-Time Forum (SPA)

Форум-SPA с регистрацией, постами, комментариями и приватными сообщениями в реальном времени.

## Стек

- **Backend:** Go (стандартная библиотека + Gorilla WebSocket, bcrypt, google/uuid)
- **Database:** SQLite (mattn/go-sqlite3)
- **Frontend:** чистый HTML/CSS/JS (без фреймворков), один `index.html` (Single Page Application)
- **Realtime:** WebSocket

## Структура

```
forum/
├── main.go
├── go.mod
├── backend/
│   ├── database/database.go      # схема БД и инициализация
│   ├── handlers/
│   │   ├── handler.go            # общая структура Handler
│   │   ├── auth.go               # /api/register, /api/login, /api/logout, /api/me
│   │   ├── auth_helpers.go       # сессии (cookie)
│   │   ├── posts.go              # /api/posts, /api/posts/{id}, /api/comments, /api/categories
│   │   ├── chat.go               # /api/users, /api/messages
│   │   └── ws.go                 # /ws — апгрейд в WebSocket
│   ├── models/models.go
│   └── websocket/
│       ├── hub.go                # хаб для всех подключений
│       └── client.go             # ReadPump/WritePump
└── frontend/
    ├── index.html                # один-единственный HTML
    ├── css/styles.css
    └── js/
        ├── utils.js              # throttle, debounce, escapeHTML, formatDate, State
        ├── api.js                # обёртка над fetch
        ├── ws.js                 # WebSocket-клиент с реконнектом
        ├── auth.js               # формы логина и регистрации
        ├── posts.js              # лента, новый пост, просмотр поста
        ├── chat.js               # приватные сообщения, throttle на скролле
        └── app.js                # роутер, точка входа
```

## Запуск

```bash
# 1. Установить зависимости
go mod tidy

# 2. Запустить
go run .

# Сервер откроется на http://localhost:8080
```

Файл `forum.db` (SQLite) будет создан автоматически при первом запуске.

## Возможности

### Регистрация и логин
- Регистрация: nickname, age, gender, first name, last name, email, password
- Логин по **никнейму или email** + пароль
- Сессии через HTTP-only cookie + uuid, хранятся в БД
- Логаут доступен с любой страницы

### Посты и комментарии
- Создание поста с одной или несколькими **категориями**
- Лента — отображение всех постов
- Фильтр по категориям
- Комментарии видны только при открытии поста

### Приватные сообщения (real-time через WebSocket)
- Сайдбар с пользователями всегда виден
- Сначала идут собеседники по дате последнего сообщения (как в Discord)
- Дальше — все остальные по алфавиту
- Зелёная точка — онлайн
- Красный значок — есть непрочитанное сообщение
- Сообщения подгружаются по **10 штук**, при скролле вверх подгружаются ещё 10
- На событие скролла навешан **throttle (300 мс)** — без спама
- На ввод текста — **debounce (500 мс)** (используется для индикатора "печатает")
- Каждое сообщение содержит дату и имя отправителя
- Сообщения приходят в реальном времени без обновления страницы

## Важные детали

- **Single Page Application:** `main.go` отдаёт `index.html` только на `/`, всё остальное — JSON-API + WebSocket. Все переключения экранов делаются на клиенте.
- **Сессии:** в БД хранятся в таблице `sessions` с `expires_at`. Срок жизни — 7 дней.
- **CSRF/безопасность:** SameSite=Lax для cookie, HttpOnly. Все пароли — через `bcrypt`.
- **XSS:** все вставки от пользователей идут через `escapeHTML`.
- **WebSocket-реконнект:** экспоненциальная задержка (1с → 2с → 4с → … до 15с).
- **Множественные вкладки:** один пользователь может подключиться с нескольких устройств — Hub отслеживает все соединения, сообщение приходит на все вкладки.

## Эндпоинты

| Метод | URL                | Описание                            |
|-------|--------------------|-------------------------------------|
| POST  | /api/register      | Регистрация                         |
| POST  | /api/login         | Вход                                |
| POST  | /api/logout        | Выход                               |
| GET   | /api/me            | Информация о текущем пользователе   |
| GET   | /api/posts         | Лента (?category=Tech опционально)  |
| POST  | /api/posts         | Создать пост                        |
| GET   | /api/posts/{id}    | Один пост + комментарии             |
| POST  | /api/comments      | Создать комментарий                 |
| GET   | /api/categories    | Список категорий                    |
| GET   | /api/users         | Список пользователей (отсортирован) |
| GET   | /api/messages?with=ID&offset=N&limit=10 | История переписки      |
| GET   | /ws                | WebSocket                           |
