#!/usr/bin/env bash
# ==============================================================================
# YOLYOLAKAY ENTERPRISE SERVER HARDENING & SECURITY SETUP SCRIPT
# Standard: CIS Linux Benchmark Level 2 / PCI-DSS Readiness
# Target OS: Ubuntu 22.04 / 24.04 LTS & Debian 12
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

if [[ $EUID -ne 0 ]]; then
   log_error "Этот скрипт должен быть запущен с правами root (sudo)."
   exit 1
fi

log_info "=================================================================="
log_info "  ЗАПУСК ПОДГОТОВКИ И ХАРДНИНГА СЕРВЕРА YOLYOLAKAY (ENTERPRISE)   "
log_info "=================================================================="

# ------------------------------------------------------------------------------
# 1. Обновление системы и установка критических пакетов
# ------------------------------------------------------------------------------
log_info "1/7. Обновление пакетов и установка инструментов безопасности..."
export DEBIAN_FRONTEND=noninteractive
apt-get update -q
apt-get upgrade -y -q

apt-get install -y -q \
    curl wget git jq ufw fail2ban auditd audispd-plugins \
    aide rkhunter unattended-upgrades apt-listchanges \
    htop iotop iftop ncdu net-tools ca-certificates gnupg lsb-release \
    wireguard wireguard-tools

log_success "Пакеты успешно установлены."

# ------------------------------------------------------------------------------
# 2. Установка и настройка Docker & Docker Compose (Official Repo)
# ------------------------------------------------------------------------------
log_info "2/7. Проверка и установка официального Docker Engine..."
if ! command -v docker &> /dev/null; then
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg --yes
    chmod a+r /etc/apt/keyrings/docker.gpg

    echo \
      "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
      tee /etc/apt/sources.list.d/docker.list > /dev/null

    apt-get update -q
    apt-get install -y -q docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
fi

# Настройка безопасной конфигурации Docker daemon
mkdir -p /etc/docker
cat << 'EOF' > /etc/docker/daemon.json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "50m",
    "max-file": "3"
  },
  "no-new-privileges": true,
  "userland-proxy": false,
  "live-restore": true,
  "storage-driver": "overlay2"
}
EOF

systemctl restart docker
systemctl enable docker
log_success "Docker настроен с защищенными параметрами и ротацией логов."

# ------------------------------------------------------------------------------
# 3. Настройка ядра Linux (Sysctl Security & High-Performance Hardening)
# ------------------------------------------------------------------------------
log_info "3/7. Применение настроек безопасности ядра Linux (sysctl)..."
cat << 'EOF' > /etc/sysctl.d/99-yolyolakay-security.conf
# === Защита от Network & DDoS атак ===
net.ipv4.tcp_syncookies = 1
net.ipv4.tcp_max_syn_backlog = 8192
net.ipv4.tcp_synack_retries = 2
net.ipv4.tcp_syn_retries = 2

# === Защита от IP Spoofing и MITM ===
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1
net.ipv4.conf.all.accept_source_route = 0
net.ipv4.conf.default.accept_source_route = 0
net.ipv4.conf.all.accept_redirects = 0
net.ipv4.conf.default.accept_redirects = 0
net.ipv4.conf.all.secure_redirects = 0
net.ipv4.conf.default.secure_redirects = 0
net.ipv4.conf.all.send_redirects = 0
net.ipv4.conf.default.send_redirects = 0
net.ipv4.icmp_echo_ignore_broadcasts = 1
net.ipv4.icmp_ignore_bogus_error_responses = 1

# === Включение алгоритма TCP BBR и оптимизация сокетов ===
net.core.default_qdisc = fq
net.ipv4.tcp_congestion_control = bbr
net.core.somaxconn = 65535
net.ipv4.tcp_max_tw_buckets = 2000000
net.ipv4.tcp_tw_reuse = 1
net.ipv4.tcp_fin_timeout = 15

# === Системные лимиты дескрипторов файлов ===
fs.file-max = 2097152
fs.inotify.max_user_watches = 524288
fs.inotify.max_user_instances = 512
vm.max_map_count = 262144

# === Ограничения дампа памяти для безопасности ===
fs.suid_dumpable = 0
kernel.randomize_va_space = 2
EOF

sysctl --system > /dev/null 2>&1
log_success "Параметры ядра применены (TCP BBR, SYN-flood protection, Anti-Spoofing)."

# ------------------------------------------------------------------------------
# 4. Настройка лимитов открытых файлов (Limits.conf)
# ------------------------------------------------------------------------------
log_info "4/7. Настройка лимитов ulimit (nofile, nproc)..."
cat << 'EOF' >> /etc/security/limits.conf
* soft nofile 65535
* hard nofile 65535
* soft nproc 65535
* hard nproc 65535
root soft nofile 65535
root hard nofile 65535
EOF

# ------------------------------------------------------------------------------
# 5. Настройка сетевого экрана (UFW Firewall - Default Deny)
# ------------------------------------------------------------------------------
log_info "5/7. Настройка UFW Firewall (Эшелонированная фильтрация)..."
ufw --force reset > /dev/null 2>&1
ufw default deny incoming
ufw default allow outgoing

# Разрешаем только необходимые порты
ufw allow 22/tcp comment 'SSH Management'
ufw allow 80/tcp comment 'HTTP (Cloudflare/Certbot)'
ufw allow 443/tcp comment 'HTTPS (Cloudflare SSL)'
ufw allow 51820/udp comment 'WireGuard VPN'

# Полностью блокируем публичный доступ к базам и сервисам на случай прямого биндинга
ufw deny 5432 comment 'Block PostgreSQL from WAN'
ufw deny 6379 comment 'Block Redis from WAN'
ufw deny 5672 comment 'Block RabbitMQ from WAN'
ufw deny 15672 comment 'Block RabbitMQ Management from WAN'
ufw deny 8081:8094/tcp comment 'Block Internal Microservices from WAN'

ufw --force enable
log_success "UFW активирован: все внутренние порты (БД, Redis, сервисы) закрыты от внешнего мира."

# ------------------------------------------------------------------------------
# 6. Настройка Fail2Ban (Защита от брутфорса)
# ------------------------------------------------------------------------------
log_info "6/7. Настройка Fail2Ban..."
cat << 'EOF' > /etc/fail2ban/jail.local
[DEFAULT]
bantime  = 1h
findtime = 10m
maxretry = 3
banaction = ufw
ignoreip = 127.0.0.1/8 ::1

[sshd]
enabled = true
port    = ssh
logpath = %(sshd_log)s
backend = %(default_backend)s
maxretry = 3
EOF

systemctl restart fail2ban
systemctl enable fail2ban
log_success "Fail2Ban запущен (авто-бан за 3 неудачные попытки входа)."

# ------------------------------------------------------------------------------
# 7. Настройка Auditd & Авто-обновлений безопасности
# ------------------------------------------------------------------------------
log_info "7/7. Настройка подсистемы аудита безопасности (Auditd) и Unattended Upgrades..."
cat << 'EOF' > /etc/audit/rules.d/yolyolakay-audit.rules
# Аудит модификаций учетных записей и прав
-w /etc/passwd -p wa -k identity
-w /etc/shadow -p wa -k identity
-w /etc/group -p wa -k identity
-w /etc/gshadow -p wa -k identity
-w /etc/sudoers -p wa -k sudoers
-w /etc/sudoers.d/ -p wa -k sudoers

# Аудит сетевых конфигов
-w /etc/hosts -p wa -k network_mod
-w /etc/network/ -p wa -k network_mod

# Аудит привилегированных команд
-a always,exit -F arch=b64 -S execve -F euid=0 -k root_action
EOF

systemctl restart auditd || true

# Включение автоматической установки патчей безопасности
cat << 'EOF' > /etc/apt/apt.conf.d/20auto-upgrades
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "7";
EOF

log_info "=================================================================="
log_success "ХАРДНИНГ СЕРВЕРА УСПЕШНО ЗАВЕРШЕН!"
log_info "Статус UFW:"
ufw status verbose
log_info "=================================================================="
