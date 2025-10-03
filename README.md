# 🚀 Messenger Backend

Современный мессенджер на Go с WebSocket, PostgreSQL, Redis, Kafka и Docker.

## ✨ Особенности

- **WebSocket** - Реальное время для сообщений и уведомлений
- **PostgreSQL** - Надежное хранение данных
- **Redis** - Быстрое кеширование и управление сессиями
- **Kafka** - Асинхронные уведомления между сервисами
- **Docker** - Простое развертывание
- **Vault** - Безопасное хранение секретов
- **Миграции БД** - Автоматическое управление схемой

## 🏗️ Архитектура

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   WebSocket     │    │   HTTP API      │    │     Kafka       │
│     Hub         │    │   (Gin)         │    │   Producer      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Services      │
                    │   (Business     │
                    │    Logic)       │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │  Repositories   │
                    │   (Data Access) │
                    └─────────────────┘
                                 │
         ┌───────────────────────┼───────────────────────┐
         │                       │                       │
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │     Redis       │    │     Vault       │
│   (Database)    │    │    (Cache)      │    │   (Secrets)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📋 Функциональность

### 💬 Сообщения
- Отправка личных сообщений
- Групповые чаты
- Каналы (публичные/приватные)
- Ответы на сообщения (threads)
- Редактирование и удаление сообщений
- Реакции на сообщения (эмодзи)

### 👥 Пользователи и группы
- Управление пользователями
- Создание групп и каналов
- Роли участников (owner, admin, moderator, member)
- Поиск пользователей

### 🔔 Уведомления
- Уведомления о новых сообщениях
- Статус "печатает"
- Онлайн статус пользователей
- Kafka события для интеграции с другими сервисами

### 📁 Файлы
- Загрузка файлов и изображений
- Поддержка различных типов файлов
- Миниатюры изображений

## 🚀 Быстрый старт

### Требования
- Docker и Docker Compose
- Go 1.21+ (для разработки)

### 1. Клонирование репозитория
```bash
git clone <repository-url>
cd messenger-backend
```

### 2. Настройка конфигурации
```bash
# Скопируйте template конфигурации
cp config/config.template.yaml config/config.yaml

# Отредактируйте конфигурацию под ваши нужды
nano config/config.yaml
```

### 3. Запуск с Docker Compose
```bash
# Запуск всех сервисов
docker-compose up -d

# Просмотр логов
docker-compose logs -f messenger-backend
```

### 4. Проверка работоспособности
```bash
# Health check
curl http://localhost/api/v1/health

# WebSocket подключение
wscat -c ws://localhost/ws
```

## 🔧 Конфигурация

### Переменные окружения

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `SERVER_PORT` | Порт HTTP сервера | `8080` |
| `DB_HOST` | Хост PostgreSQL | `postgres` |
| `DB_PORT` | Порт PostgreSQL | `5432` |
| `REDIS_HOST` | Хост Redis | `redis` |
| `REDIS_PORT` | Порт Redis | `6379` |
| `KAFKA_BROKERS` | Kafka brokers | `kafka:29092` |
| `VAULT_ADDR` | Vault адрес | `http://vault:8200` |

### Флаги функций

| Флаг | Описание |
|------|----------|
| `WEBSOCKET_ENABLED` | Включить WebSocket |
| `KAFKA_ENABLED` | Включить Kafka |
| `FILE_UPLOAD_ENABLED` | Включить загрузку файлов |
| `RATE_LIMIT_ENABLED` | Включить rate limiting |
| `DEBUG_ENABLED` | Режим отладки |

## 📚 API Документация

### WebSocket API

#### Подключение
```javascript
const ws = new WebSocket('ws://localhost/ws');

// Присоединение к комнате
ws.send(JSON.stringify({
  type: 'join_room',
  data: { room_id: 'group-123' }
}));

// Отправка сообщения о печати
ws.send(JSON.stringify({
  type: 'typing',
  data: { room_id: 'group-123', channel_id: 'channel-456' }
}));
```

#### События WebSocket
- `new_message` - Новое сообщение
- `edit_message` - Сообщение отредактировано
- `delete_message` - Сообщение удалено
- `new_reaction` - Добавлена реакция
- `remove_reaction` - Удалена реакция
- `user_typing` - Пользователь печатает
- `user_online` - Пользователь онлайн
- `user_offline` - Пользователь офлайн

### HTTP API

#### Пользователи
```bash
# Создать пользователя
POST /api/v1/users
{
  "username": "john_doe",
  "email": "john@example.com",
  "display_name": "John Doe"
}

# Получить пользователя
GET /api/v1/users/{id}

# Поиск пользователей
GET /api/v1/users?q=john&limit=20&offset=0
```

#### Сообщения
```bash
# Отправить сообщение
POST /api/v1/messages
{
  "group_id": "group-123",
  "channel_id": "channel-456",
  "content": "Hello, world!",
  "message_type": "text"
}

# Получить сообщения группы
GET /api/v1/messages/group/{group_id}?limit=50&offset=0

# Добавить реакцию
POST /api/v1/messages/{message_id}/reactions
{
  "emoji": "👍"
}
```

## 🗄️ База данных

### Миграции
```bash
# Применить миграции (автоматически при запуске)
docker-compose exec messenger-backend ./migrate up

# Откатить миграции
docker-compose exec messenger-backend ./migrate down
```

### Схема БД

#### Основные таблицы
- `users` - Пользователи
- `groups` - Группы и каналы
- `messages` - Сообщения
- `message_reactions` - Реакции на сообщения
- `notifications` - Уведомления

#### Связи
- `group_members` - Участники групп
- `channel_members` - Участники каналов
- `message_attachments` - Вложения к сообщениям
- `message_reads` - Статус прочтения сообщений

## 🔐 Безопасность

### Vault интеграция
- Секреты БД и Redis хранятся в Vault
- Автоматическая ротация ключей
- Шифрование чувствительных данных

### JWT токены
```bash
# TODO: Реализовать JWT аутентификацию
```

### CORS
- Настроен для всех доменов (в разработке)
- В продакшене настроить конкретные домены

## 📊 Мониторинг

### Логирование
- Структурированные логи (JSON)
- Уровни: DEBUG, INFO, WARN, ERROR
- Интеграция с внешними системами мониторинга

### Метрики
```bash
# TODO: Добавить Prometheus метрики
```

## 🚀 Развертывание

### Production
```bash
# Создать production конфигурацию
cp config/config.template.yaml config/config.prod.yaml

# Запустить с production конфигурацией
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### Масштабирование
- Горизонтальное масштабирование WebSocket хаба
- Шардинг базы данных
- Кластеризация Redis
- Kafka партиционирование

## 🧪 Тестирование

```bash
# Запуск тестов
go test ./...

# Тесты с покрытием
go test -cover ./...

# Интеграционные тесты
go test -tags=integration ./...
```

## 🤝 Разработка

### Структура проекта
```
internal/
├── api/          # HTTP API handlers
├── cache/        # Redis кеширование
├── config/       # Конфигурация
├── kafka/        # Kafka интеграция
├── models/       # Модели данных
├── repository/   # Репозитории БД
├── service/      # Бизнес-логика
└── websocket/    # WebSocket логика
```

### Добавление новой функции
1. Создать модель в `internal/models/`
2. Добавить миграцию БД
3. Создать репозиторий в `internal/repository/`
4. Реализовать сервис в `internal/service/`
5. Добавить HTTP handlers в `internal/api/handlers/`
6. Интегрировать с WebSocket и Kafka

## 📝 TODO

- [ ] JWT аутентификация
- [ ] Rate limiting
- [ ] Файловое хранилище (S3)
- [ ] Push уведомления
- [ ] Видео/голосовые звонки
- [ ] Шифрование сообщений
- [ ] Боты и интеграции
- [ ] Аналитика и метрики
- [ ] Автоматические тесты
- [ ] CI/CD pipeline

## 📄 Лицензия

MIT License

## 👥 Команда

- Backend: Go, Gin, WebSocket
- Database: PostgreSQL
- Cache: Redis
- Message Queue: Kafka
- Secrets: Vault
- Containerization: Docker

---

**Примечание**: Это MVP версия мессенджера. Для продакшена необходимо добавить аутентификацию, безопасность, тестирование и мониторинг.
