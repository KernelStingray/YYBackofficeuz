#!/bin/bash

echo "🚀 Запуск Cluster B разделен на множество мелких этапов (максимум по 3 сервиса) для защиты ОС..."

echo ""
echo "=== ЭТАП 1.1: Основные БД, Кэш и Vault (5 сервисов) ==="
docker compose -f docker-compose.group_saturn.yml up -d ${BUILD_FLAG} vault vault_init group_saturn rabbitmq redis
echo "⏳ Ожидание (15 секунд)..."
sleep 15

echo ""
echo "=== ЭТАП 1.2: Хранилище Minio (2 сервиса) ==="
docker compose -f docker-compose.group_saturn.yml up -d ${BUILD_FLAG} minio createbuckets
echo "⏳ Ожидание (10 секунд)..."
sleep 10

echo ""
echo "=== ЭТАП 2.1: Обсервабилити часть 1 (3 сервиса) ==="
docker compose -f docker-compose.group_saturn.yml up -d ${BUILD_FLAG} prometheus grafana loki
echo "⏳ Ожидание (15 секунд)..."
sleep 15

echo ""
echo "=== ЭТАП 2.2: Обсервабилити часть 2 (2 сервиса) ==="
docker compose -f docker-compose.group_saturn.yml up -d ${BUILD_FLAG} promtail jaeger
echo "⏳ Ожидание (10 секунд)..."
sleep 10

echo ""
echo "=== ЭТАП 3.1: Микросервисы часть 1 (3 сервиса) ==="
docker compose -f docker-compose.group_saturn.yml up -d ${BUILD_FLAG} parcel_service telematics_service billing_service
echo "⏳ Ожидание сборки и запуска (20 секунд)..."
sleep 20

echo ""
echo "=== ЭТАП 3.2: Гео-сервисы и Микросервисы часть 2 (5 сервисов) ==="
docker compose -f docker-compose.group_saturn.yml up -d ${BUILD_FLAG} osrm photon ride_service notification_service routing_service
echo "⏳ Ожидание сборки и запуска (20 секунд)..."
sleep 20

echo ""
echo "=== ЭТАП 4.1: Микросервисы часть 3 (3 сервиса) ==="
docker compose -f docker-compose.group_saturn.yml up -d ${BUILD_FLAG} helpdesk_service analytics_service antifraud_service
echo "⏳ Ожидание сборки и запуска (20 секунд)..."
sleep 20

echo ""
echo "=== ЭТАП 4.2: Микросервисы часть 4 и Админ-панель (4 сервиса) ==="
docker compose -f docker-compose.group_saturn.yml up -d ${BUILD_FLAG} backoffice_service media_service rating_service controlling_frontend
echo "⏳ Ожидание сборки и запуска (20 секунд)..."
sleep 20

echo ""
echo "=== ЭТАП 5: API Gateway, Веб-клиент YYWeb, Бекапы и Cloudflare (4 сервиса) ==="
docker compose -f docker-compose.group_saturn.yml up -d ${BUILD_FLAG} api_gateway backup_cron_b cloudflared
echo "⏳ Финальное ожидание (15 секунд)..."
sleep 15

echo ""
echo "✅ Все этапы запуска Cluster B успешно завершены! Нагрузка была распределена максимально бережно."
