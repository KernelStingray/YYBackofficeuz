# Пошаговое руководство по настройке и запуску Enterprise-инфраструктуры YYBackend

Данное руководство предназначено для подготовки двух серверов (Сервер 1: Европа `161.97.167.143` и Сервер 2: Узбекистан `189.74.97.243`) к продуктовой эксплуатации.

---

## ШАГ 1. Применение скрипта Enterprise Hardening на серверах

Скрипт [`scripts/enterprise_hardening.sh`](file:///home/stingray/Документы/HEXTEC%20Group/Projects/YYBackend/scripts/enterprise_hardening.sh) выполняет полную настройку ОС:
* Устанавливает UFW Firewall и закрывает ВСЕ порты, кроме 22, 80, 443;
* Настраивает Fail2Ban (бан за 3 неудачные попытки входа);
* Применяет параметры безопасности ядра Linux (sysctl: TCP BBR, защита от SYN-flood и спуфинга);
* Устанавливает и настраивает Docker с защищенными политиками и ротацией логов;
* Включает автоматические обновления безопасности (`unattended-upgrades`).

### Запуск на Сервере 1 (Европа `161.97.167.143`):
С вашего локального компьютера скопируйте и запустите скрипт:
```bash
scp scripts/enterprise_hardening.sh root@161.97.167.143:/root/
ssh root@161.97.167.143 "bash /root/enterprise_hardening.sh"
```

### Запуск на Сервере 2 (Узбекистан `189.74.97.243`):
```bash
scp scripts/enterprise_hardening.sh root@189.74.97.243:/root/
ssh root@189.74.97.243 "bash /root/enterprise_hardening.sh"
```

---

## ШАГ 2. Копирование проекта на основной сервер (Сервер 1)

На Сервере 1 создайте рабочую директорию проекта:
```bash
ssh root@161.97.167.143 "mkdir -p /opt/yolyolakay"
```

С локального компьютера синхронизируйте репозиторий:
```bash
rsync -avz --exclude='.git' --exclude='node_modules' --exclude='tmp' ./ root@161.97.167.143:/opt/yolyolakay/
```

---

## ШАГ 3. Настройка переменных окружения `.env`

На Сервере 1 перейдите в каталог проекта и создайте боевой `.env`:
```bash
ssh root@161.97.167.143
cd /opt/yolyolakay
cp configs/.env.example .env
chmod 600 .env
```

Отредактируйте `.env`, указав надежные пароли для PostgreSQL, RabbitMQ, Redis и ключи Click/Payme:
```bash
nano .env
```

---

## ШАГ 4. Настройка Nginx и SSL-сертификатов (Cloudflare)

1. Скопируйте конфигурацию Nginx:
```bash
cp /opt/yolyolakay/deployments/nginx/cloudflare_ips.conf /etc/nginx/conf.d/
cp /opt/yolyolakay/deployments/nginx/enterprise_yolyolakay.conf /etc/nginx/sites-available/yolyolakay
ln -sf /etc/nginx/sites-available/yolyolakay /etc/nginx/sites-enabled/
```

2. Выпустите SSL-сертификат через Cloudflare Origin CA или Let's Encrypt Certbot:
```bash
apt-get install -y certbot python3-certbot-nginx
certbot --nginx -d api.yolyolakay.uz -d api.yolyolakay.com
```

3. Проверьте конфигурацию и перезапустите Nginx:
```bash
nginx -t
systemctl reload nginx
```

---

## ШАГ 5. Инициализация баз данных и запуск микросервисов

Запустите контейнеры баз данных и инфраструктуры (PostgreSQL, Redis, RabbitMQ):
```bash
cd /opt/yolyolakay
docker compose -f docker-compose.group_jupiter.yml up -d
```

Примените миграции баз данных:
```bash
bash scripts/migrate.sh
```

Запустите все 15 микросервисов платформы:
```bash
docker compose -f docker-compose.group_saturn.yml up -d
```

---

## ШАГ 6. Настройка автоматического резервного копирования (Cron)

Настройте ежедневный запуск шифрованного бэкапа в 03:00 ночи:
```bash
(crontab -l 2>/dev/null; echo "0 3 * * * /opt/yolyolakay/scripts/prod_backup.sh >> /var/log/yolyolakay_backup.log 2>&1") | crontab -
```
