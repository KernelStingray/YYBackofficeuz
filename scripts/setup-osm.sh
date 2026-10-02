#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

DATA_OSM_DIR="$ROOT_DIR/data/osm"
DATA_PHOTON_DIR="$ROOT_DIR/data/photon"

echo "🗺️  Инициализация локального картографического стека OpenStreetMap..."

mkdir -p "$DATA_OSM_DIR"
mkdir -p "$DATA_PHOTON_DIR"

PBF_FILE="$DATA_OSM_DIR/uzbekistan-latest.osm.pbf"
OSRM_FILE="$DATA_OSM_DIR/uzbekistan-latest.osrm"

# 1. Скачивание PBF-файла карт Узбекистана
if [ ! -f "$PBF_FILE" ]; then
    echo "📥 Скачивание актуальных данных картографии Узбекистана (Geofabrik PBF)..."
    curl -L -o "$PBF_FILE" "https://download.geofabrik.de/asia/uzbekistan-latest.osm.pbf"
    echo "✅ PBF файл успешно скачан!"
else
    echo "ℹ️  Файл $PBF_FILE уже существует, пропускаем скачивание."
fi

# 2. Обработка графа OSRM
if [ ! -f "$OSRM_FILE" ]; then
    echo "⚙️  Компиляция дорожного графа для OSRM (экстракция, партиционирование и кастомизация)..."
    
    docker run -t -v "$DATA_OSM_DIR:/data" osrm/osrm-backend:latest osrm-extract -p /opt/car.lua /data/uzbekistan-latest.osm.pbf
    docker run -t -v "$DATA_OSM_DIR:/data" osrm/osrm-backend:latest osrm-partition /data/uzbekistan-latest.osrm
    docker run -t -v "$DATA_OSM_DIR:/data" osrm/osrm-backend:latest osrm-customize /data/uzbekistan-latest.osrm

    echo "✅ Сборка графа OSRM завершена!"
else
    echo "ℹ️  Граф OSRM $OSRM_FILE уже скомпилирован."
fi

echo "🎉 Подготовка карт OSM для Узбекистана завершена!"
