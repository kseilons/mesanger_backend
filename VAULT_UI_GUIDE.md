# 🎨 Руководство по работе с Vault UI

## 🚀 Запуск Vault с UI

```bash
# Запустить Vault
docker-compose up -d vault

# Проверить статус
docker-compose ps vault
```

## 🌐 Доступ к интерфейсу

1. **Откройте браузер**: http://localhost:8200/ui
2. **Введите токен**: `messenger-token`
3. **Нажмите "Sign In"**

## 📝 Создание секретов через UI

### 1. Включение KV Secrets Engine

1. Перейдите в **Secrets** → **Enable new engine**
2. Выберите **KV**
3. Введите путь: `secret`
4. Версия: **Version 2**
5. Нажмите **Enable Engine**

### 2. Создание секретов

#### Секрет для базы данных:
1. Перейдите в **Secrets** → **secret**
2. Нажмите **Create secret**
3. **Path**: `database/password`
4. **Key**: `value`
5. **Value**: `messenger_password`
6. Нажмите **Save**

#### Секрет для Redis:
1. **Path**: `redis/password`
2. **Key**: `value`
3. **Value**: `redis_password`
4. Нажмите **Save**

#### Секрет для JWT:
1. **Path**: `jwt/secret`
2. **Key**: `value`
3. **Value**: `your-super-secret-jwt-key-change-this-in-production`
4. Нажмите **Save**

## 🔍 Просмотр секретов

1. Перейдите в **Secrets** → **secret**
2. Выберите нужный секрет
3. Нажмите **View** для просмотра
4. Нажмите **Edit** для изменения

## 🗂️ Структура секретов

```
secret/
├── database/
│   └── password
├── redis/
│   └── password
└── jwt/
    └── secret
```

## 🛠️ Дополнительные возможности UI

### 1. Просмотр логов
- **Tools** → **Logs** - просмотр логов Vault

### 2. Мониторинг
- **Tools** → **Metrics** - метрики производительности

### 3. Политики доступа
- **Access** → **Policies** - управление политиками безопасности

### 4. Аутентификация
- **Access** → **Auth Methods** - настройка методов аутентификации

## 🔧 Полезные команды для проверки

```bash
# Проверить статус Vault
docker exec vault vault status

# Посмотреть все секреты
docker exec vault vault kv list secret/

# Проверить конкретный секрет
docker exec vault vault kv get secret/database/password
```
