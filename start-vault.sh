#!/bin/bash

echo "🚀 Запуск Vault с UI..."

# Остановить существующие контейнеры
echo "🛑 Останавливаем существующие контейнеры..."
docker-compose down

# Запустить Vault
echo "▶️  Запускаем Vault..."
docker-compose up -d vault

# Ждем запуска
echo "⏳ Ждем запуска Vault..."
sleep 10

# Проверяем статус
echo "🔍 Проверяем статус Vault..."
docker-compose ps vault

echo ""
echo "✅ Vault запущен!"
echo ""
echo "🌐 Доступ к UI:"
echo "   URL: http://localhost:8200/ui"
echo "   Token: messenger-token"
echo ""
echo "🔧 Полезные команды:"
echo "   docker-compose logs vault          # Посмотреть логи"
echo "   docker exec vault vault status     # Статус Vault"
echo "   docker exec vault vault kv list secret/  # Список секретов"
echo ""
echo "📖 Подробная инструкция: VAULT_UI_GUIDE.md"
