#!/usr/bin/env bash
# ==============================================================================
# YOLYOLAKAY ENTERPRISE WIREGUARD VPN MESH SETUP
# Connects: yy-europe (161.97.167.143 / 10.8.0.1) <---> yy-uz (189.74.97.243 / 10.8.0.2)
# ==============================================================================

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

HOST_EU="yy-europe"
HOST_UZ="yy-uz"
IP_EU="161.97.167.143"
IP_UZ="189.74.97.243"
WG_PORT="4500"

log_info "=================================================================="
log_info " НАСТРОЙКА ЗАШИФРОВАННОГО WIREGUARD ТУННЕЛЯ МЕЖДУ СЕРВЕРАМИ      "
log_info " Europe ($IP_EU: 10.8.0.1) <===> Uzbekistan ($IP_UZ: 10.8.0.2)  "
log_info "=================================================================="

# 1. Проверка SSH доступности
log_info "1/6. Проверка связи с обоими серверами по SSH..."
ssh -o ConnectTimeout=5 "$HOST_EU" "echo 'SSH Europe OK'" >/dev/null || { log_error "Не удалось подключиться к $HOST_EU"; exit 1; }
ssh -o ConnectTimeout=5 "$HOST_UZ" "echo 'SSH Uzbekistan OK'" >/dev/null || { log_error "Не удалось подключиться к $HOST_UZ"; exit 1; }
log_success "Связь с обоими серверами установлена."

# 2. Генерация ключей на европейском сервере
log_info "2/6. Генерация ключей WireGuard на сервере Европа ($HOST_EU)..."
ssh "$HOST_EU" "mkdir -p /etc/wireguard && chmod 700 /etc/wireguard && \
  if [ ! -f /etc/wireguard/private.key ]; then \
    wg genkey | tee /etc/wireguard/private.key | wg pubkey > /etc/wireguard/public.key; \
  fi"
EU_PUBKEY=$(ssh "$HOST_EU" "cat /etc/wireguard/public.key")
EU_PRIVKEY=$(ssh "$HOST_EU" "cat /etc/wireguard/private.key")
log_success "Ключи Europe готовы. Публичный ключ: $EU_PUBKEY"

# 3. Генерация ключей на сервере в Узбекистане
log_info "3/6. Генерация ключей WireGuard на сервере Узбекистан ($HOST_UZ)..."
ssh "$HOST_UZ" "mkdir -p /etc/wireguard && chmod 700 /etc/wireguard && \
  if [ ! -f /etc/wireguard/private.key ]; then \
    wg genkey | tee /etc/wireguard/private.key | wg pubkey > /etc/wireguard/public.key; \
  fi"
UZ_PUBKEY=$(ssh "$HOST_UZ" "cat /etc/wireguard/public.key")
UZ_PRIVKEY=$(ssh "$HOST_UZ" "cat /etc/wireguard/private.key")
log_success "Ключи Uzbekistan готовы. Публичный ключ: $UZ_PUBKEY"

# 4. Настройка wg0.conf на Европейском сервере (10.8.0.1)
log_info "4/6. Настройка конфигурации /etc/wireguard/wg0.conf на $HOST_EU..."
ssh "$HOST_EU" "cat << 'EOF' > /etc/wireguard/wg0.conf
[Interface]
Address = 10.8.0.1/24
ListenPort = $WG_PORT
PrivateKey = $EU_PRIVKEY
PostUp = iptables -t nat -A POSTROUTING -o %i -j MASQUERADE; iptables -A FORWARD -i %i -j ACCEPT; iptables -A FORWARD -o %i -j ACCEPT
PostDown = iptables -t nat -D POSTROUTING -o %i -j MASQUERADE; iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT
SaveConfig = false

[Peer]
PublicKey = $UZ_PUBKEY
AllowedIPs = 10.8.0.0/24
Endpoint = $IP_UZ:$WG_PORT
PersistentKeepalive = 25
EOF
chmod 600 /etc/wireguard/wg0.conf
ufw allow in on wg0 comment 'Allow private VPN subnet' || true
ufw allow $WG_PORT/udp comment 'WireGuard port' || true
systemctl enable --now wg-quick@wg0
systemctl restart wg-quick@wg0
"

# 5. Настройка wg0.conf на сервере в Узбекистане (10.8.0.2)
log_info "5/6. Настройка конфигурации /etc/wireguard/wg0.conf на $HOST_UZ..."
ssh "$HOST_UZ" "cat << 'EOF' > /etc/wireguard/wg0.conf
[Interface]
Address = 10.8.0.2/24
ListenPort = $WG_PORT
PrivateKey = $UZ_PRIVKEY
PostUp = iptables -t nat -A POSTROUTING -o %i -j MASQUERADE; iptables -A FORWARD -i %i -j ACCEPT; iptables -A FORWARD -o %i -j ACCEPT
PostDown = iptables -t nat -D POSTROUTING -o %i -j MASQUERADE; iptables -D FORWARD -i %i -j ACCEPT; iptables -D FORWARD -o %i -j ACCEPT
SaveConfig = false

[Peer]
PublicKey = $EU_PUBKEY
AllowedIPs = 10.8.0.0/24
Endpoint = $IP_EU:$WG_PORT
PersistentKeepalive = 25
EOF
chmod 600 /etc/wireguard/wg0.conf
ufw allow in on wg0 comment 'Allow private VPN subnet' || true
ufw allow $WG_PORT/udp comment 'WireGuard port' || true
systemctl enable --now wg-quick@wg0
systemctl restart wg-quick@wg0
"

# 6. Проверка пинга по защищенному туннелю
log_info "6/6. Проверка связи по шифрованному туннелю (10.8.0.1 <==> 10.8.0.2)..."
sleep 2

if ssh "$HOST_EU" "ping -c 3 10.8.0.2" >/dev/null; then
    log_success "Пинг 10.8.0.1 -> 10.8.0.2 (Европа -> Узбекистан) УСПЕШЕН!"
else
    log_warn "Пинг пока не прошел, подождем установления соединения..."
    sleep 3
    ssh "$HOST_EU" "ping -c 3 10.8.0.2"
fi

# 7. Установка автоматического сторожевого таймера (Auto-Heal Watchdog)
log_info "7/7. Установка службы самовосстановления (WireGuard Watchdog)..."
ssh "$HOST_EU" "cat << 'EOF' > /usr/local/bin/wireguard-watchdog.sh
#!/usr/bin/env bash
TARGET=\"10.8.0.2\"
if ! ping -c 2 -W 3 \"\$TARGET\" > /dev/null 2>&1; then
    echo \"\$(date -u +'%Y-%m-%dT%H:%M:%SZ') [WARN] WireGuard ping to \$TARGET failed. Restarting wg-quick@wg0...\" >> /var/log/wireguard-watchdog.log
    systemctl restart wg-quick@wg0
fi
EOF
chmod +x /usr/local/bin/wireguard-watchdog.sh
echo '*/2 * * * * root /usr/local/bin/wireguard-watchdog.sh' > /etc/cron.d/wireguard-watchdog
chmod 644 /etc/cron.d/wireguard-watchdog
"

ssh "$HOST_UZ" "cat << 'EOF' > /usr/local/bin/wireguard-watchdog.sh
#!/usr/bin/env bash
TARGET=\"10.8.0.1\"
if ! ping -c 2 -W 3 \"\$TARGET\" > /dev/null 2>&1; then
    echo \"\$(date -u +'%Y-%m-%dT%H:%M:%SZ') [WARN] WireGuard ping to \$TARGET failed. Restarting wg-quick@wg0...\" >> /var/log/wireguard-watchdog.log
    systemctl restart wg-quick@wg0
fi
EOF
chmod +x /usr/local/bin/wireguard-watchdog.sh
echo '*/2 * * * * root /usr/local/bin/wireguard-watchdog.sh' > /etc/cron.d/wireguard-watchdog
chmod 644 /etc/cron.d/wireguard-watchdog
"
log_success "Watchdog установлен на обоих серверах (проверка каждые 2 минуты)."

log_info "=================================================================="
log_success "WIREGUARD VPN МЕЖДУ СЕРВЕРАМИ УСПЕШНО РАБОТАЕТ!"
log_info "Внутренние приватные адреса:"
log_info "  • Сервер Узбекистан (Кластер A - Персональные данные): 10.8.0.2"
log_info "  • Сервер Европа (Кластер B - Вычисления и роутинг):    10.8.0.1"
log_info "=================================================================="
