#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

GITHUB_REPO="ctsunny/board"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/board"
DATA_DIR="/var/lib/board"
SERVICE_NAME="board"
BINARY_NAME="board"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
VERSION_FILE="${DATA_DIR}/installed_version"

print_info()    { echo -e "${GREEN}[INFO]${NC} $1"; }
print_warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }
print_error()   { echo -e "${RED}[ERROR]${NC} $1"; }
print_banner()  { echo -e "${BLUE}$1${NC}"; }

check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root."
        exit 1
    fi
}

detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS_ID="${ID}"
        OS_VERSION="${VERSION_ID}"
    else
        print_error "Cannot detect OS. /etc/os-release not found."
        exit 1
    fi

    case "${OS_ID}" in
        ubuntu)
            VER_MAJOR=$(echo "${OS_VERSION}" | cut -d. -f1)
            if [[ "${VER_MAJOR}" -lt 20 ]]; then
                print_error "Ubuntu 20.04+ is required (detected ${OS_VERSION})."
                exit 1
            fi
            PKG_MANAGER="apt-get"
            ;;
        debian)
            VER_MAJOR=$(echo "${OS_VERSION}" | cut -d. -f1)
            if [[ "${VER_MAJOR}" -lt 11 ]]; then
                print_error "Debian 11+ is required (detected ${OS_VERSION})."
                exit 1
            fi
            PKG_MANAGER="apt-get"
            ;;
        centos|rhel|rocky|almalinux)
            VER_MAJOR=$(echo "${OS_VERSION}" | cut -d. -f1)
            if [[ "${VER_MAJOR}" -lt 7 ]]; then
                print_error "CentOS/RHEL 7+ is required (detected ${OS_VERSION})."
                exit 1
            fi
            PKG_MANAGER="yum"
            ;;
        *)
            print_warn "Unsupported OS: ${OS_ID}. Proceeding anyway..."
            PKG_MANAGER="apt-get"
            ;;
    esac
    print_info "Detected OS: ${OS_ID} ${OS_VERSION}"
}

detect_arch() {
    ARCH=$(uname -m)
    case "${ARCH}" in
        x86_64)  ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        *)
            print_error "Unsupported architecture: ${ARCH}. Only amd64 and arm64 are supported."
            exit 1
            ;;
    esac
    print_info "Detected architecture: ${ARCH}"
}

get_latest_version() {
    print_info "Fetching latest release version..."
    LATEST_VERSION=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" \
        | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    if [[ -z "${LATEST_VERSION}" ]]; then
        print_error "Failed to fetch latest version from GitHub."
        exit 1
    fi
    print_info "Latest version: ${LATEST_VERSION}"
}

get_latest_version_quiet() {
    LATEST_VERSION=$(curl -fsSL --max-time 10 "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" 2>/dev/null \
        | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' || true)
}

get_installed_version() {
    INSTALLED_VERSION="未安装"

    if [[ -x "${INSTALL_DIR}/${BINARY_NAME}" ]]; then
        local version_output
        version_output=$("${INSTALL_DIR}/${BINARY_NAME}" --version 2>/dev/null || true)
        if [[ -n "${version_output}" ]]; then
            INSTALLED_VERSION=$(printf '%s\n' "${version_output}" | head -1)
            return
        fi
    fi

    if [[ -f "${VERSION_FILE}" ]]; then
        INSTALLED_VERSION=$(head -1 "${VERSION_FILE}")
    fi
}

download_binary() {
    local version="${1}"
    local arch="${2}"
    local dest="${3}"
    local binary_name="${BINARY_NAME}-linux-${arch}"
    local url="https://github.com/${GITHUB_REPO}/releases/download/${version}/${binary_name}"

    print_info "Downloading ${binary_name} from ${url}..."
    curl -fsSL -o "${dest}" "${url}"
    chmod +x "${dest}"
}

ensure_deps() {
    local deps=("curl" "systemctl")
    for dep in "${deps[@]}"; do
        if ! command -v "${dep}" &>/dev/null; then
            print_warn "${dep} not found, attempting to install..."
            if [[ "${PKG_MANAGER}" == "apt-get" ]]; then
                apt-get update -qq && apt-get install -y -qq "${dep}"
            else
                yum install -y -q "${dep}"
            fi
        fi
    done
}

save_installed_version() {
    local version="${1}"
    mkdir -p "${DATA_DIR}"
    printf '%s\n' "${version}" > "${VERSION_FILE}"
}

read_config_value() {
    local key="${1}"
    local config_file="${CONFIG_DIR}/config.json"
    if [[ -f "${config_file}" ]]; then
        grep -o "\"${key}\"[[:space:]]*:[[:space:]]*[^,}]*" "${config_file}" \
            | head -1 | sed -E 's/.*:[[:space:]]*//' | tr -d '"' | tr -d ' '
    fi
}

show_access_info() {
    local config_file="${CONFIG_DIR}/config.json"
    if [[ ! -f "${config_file}" ]]; then
        print_warn "Config file not found at ${config_file}. Start the service first."
        return
    fi

    local port base_path admin_user admin_pass domain
    port=$(read_config_value "port")
    base_path=$(read_config_value "base_path")
    admin_user=$(read_config_value "admin_user")
    admin_pass=$(read_config_value "admin_password")
    domain=$(read_config_value "domain")

    # Get public IP
    local ip
    ip=$(curl -fsSL --max-time 5 https://api.ipify.org 2>/dev/null || \
         curl -fsSL --max-time 5 https://ifconfig.me 2>/dev/null || \
         hostname -I | awk '{print $1}')

    local access_path="${base_path}"
    if [[ -z "${access_path}" ]]; then
        access_path="/"
    elif [[ "${access_path}" != */ ]]; then
        access_path="${access_path}/"
    fi

    local access_url
    if [[ -n "${domain}" ]]; then
        access_url="https://${domain}${access_path}"
    else
        access_url="http://${ip}:${port}${access_path}"
    fi

    # Box is 58 display columns wide (56 inner). CJK chars each occupy 2 display
    # columns, so spacing around them is adjusted accordingly.
    print_banner "╔════════════════════════════════════════════════════════╗"
    print_banner "║  Board 安装完成！                                      ║"
    print_banner "╠════════════════════════════════════════════════════════╣"
    printf "${BLUE}║${NC}  访问地址: %-44s${BLUE}║${NC}\n" "${access_url}"
    printf "${BLUE}║${NC}  用户名:   %-44s${BLUE}║${NC}\n" "${admin_user}"
    printf "${BLUE}║${NC}  密  码:   %-44s${BLUE}║${NC}\n" "${admin_pass}"
    print_banner "╚════════════════════════════════════════════════════════╝"
}

create_service() {
    cat > "${SERVICE_FILE}" <<EOF
[Unit]
Description=Board - Live-stream Customer Management
After=network.target

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/${BINARY_NAME} --config ${CONFIG_DIR}/config.json
Restart=on-failure
RestartSec=5s
LimitNOFILE=65536
WorkingDirectory=${DATA_DIR}

[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload
    print_info "Systemd service created at ${SERVICE_FILE}"
}

escape_sed_replacement() {
    printf '%s' "${1}" | sed -e 's/[&|]/\\&/g'
}

update_config_string() {
    local key="${1}"
    local value="${2}"
    local config_file="${CONFIG_DIR}/config.json"
    local escaped_value

    if [[ ! -f "${config_file}" ]]; then
        print_error "Config file not found at ${config_file}"
        return 1
    fi

    escaped_value=$(escape_sed_replacement "${value}")
    sed -i -E "s|(\"${key}\"[[:space:]]*:[[:space:]]*\")[^\"]*(\")|\1${escaped_value}\2|" "${config_file}"
}

configure_tls() {
    local enable_tls domain cert_dir
    cert_dir="${DATA_DIR}/certs"

    echo ""
    read -rp "是否启用 HTTPS 自动申请/续期证书？[y/N]: " enable_tls
    case "${enable_tls}" in
        y|Y|yes|YES)
            while true; do
                read -rp "请输入已解析到本机的域名: " domain
                if [[ -n "${domain}" ]]; then
                    break
                fi
                print_warn "域名不能为空。"
            done

            mkdir -p "${cert_dir}"
            update_config_string "domain" "${domain}"
            update_config_string "cert_dir" "${cert_dir}"

            print_warn "请确保域名已解析到本机，并已放通 80/443 端口。"
            systemctl restart "${SERVICE_NAME}"
            sleep 2

            if systemctl is-active --quiet "${SERVICE_NAME}"; then
                print_info "HTTPS 自动证书已启用，程序会自动申请并续期证书。"
            else
                print_error "启用 HTTPS 后服务启动失败。请检查日志: journalctl -u ${SERVICE_NAME} -n 50"
                exit 1
            fi
            ;;
        *)
            ;;
    esac
}

install() {
    check_root
    detect_os
    detect_arch
    ensure_deps
    get_latest_version

    print_info "Installing Board ${LATEST_VERSION}..."

    # Create directories
    mkdir -p "${CONFIG_DIR}" "${DATA_DIR}"

    # Download binary
    download_binary "${LATEST_VERSION}" "${ARCH}" "${INSTALL_DIR}/${BINARY_NAME}"

    # Create systemd service
    create_service

    # Enable and start service (binary generates config on first run)
    systemctl enable "${SERVICE_NAME}"
    systemctl start "${SERVICE_NAME}"

    # Wait for config to be generated
    print_info "Waiting for service to initialize..."
    sleep 3

    if systemctl is-active --quiet "${SERVICE_NAME}"; then
        print_info "Service started successfully."
        configure_tls
        save_installed_version "${LATEST_VERSION}"
        show_access_info
    else
        print_error "Service failed to start. Check logs with: journalctl -u ${SERVICE_NAME} -n 50"
        exit 1
    fi
}

remove_installation() {
    systemctl stop "${SERVICE_NAME}" 2>/dev/null || true
    systemctl disable "${SERVICE_NAME}" 2>/dev/null || true
    rm -f "${SERVICE_FILE}"
    systemctl daemon-reload

    rm -f "${INSTALL_DIR}/${BINARY_NAME}"
}

uninstall() {
    check_root
    print_warn "Uninstalling Board..."

    remove_installation
    print_warn "Config and data directories (${CONFIG_DIR}, ${DATA_DIR}) are preserved."
    print_warn "Remove them manually if needed: rm -rf ${CONFIG_DIR} ${DATA_DIR}"
    print_info "Board uninstalled."
}

purge() {
    check_root
    print_warn "This will completely remove Board, including all config and historical data."
    read -rp "Type YES to continue: " confirm
    if [[ "${confirm}" != "YES" ]]; then
        print_info "Cancelled."
        return
    fi

    remove_installation
    rm -rf "${CONFIG_DIR}" "${DATA_DIR}"
    print_info "Board and all data have been removed."
}

update() {
    check_root
    detect_arch
    get_latest_version

    print_info "Updating Board to ${LATEST_VERSION}..."

    systemctl stop "${SERVICE_NAME}" 2>/dev/null || true

    download_binary "${LATEST_VERSION}" "${ARCH}" "${INSTALL_DIR}/${BINARY_NAME}"

    systemctl start "${SERVICE_NAME}"

    sleep 2
    if systemctl is-active --quiet "${SERVICE_NAME}"; then
        save_installed_version "${LATEST_VERSION}"
        print_info "Board updated to ${LATEST_VERSION} and restarted successfully."
    else
        print_error "Service failed to start after update. Check logs: journalctl -u ${SERVICE_NAME} -n 50"
        exit 1
    fi
}

start_service() {
    check_root
    systemctl start "${SERVICE_NAME}"
    print_info "Board started."
}

stop_service() {
    check_root
    systemctl stop "${SERVICE_NAME}"
    print_info "Board stopped."
}

restart_service() {
    check_root
    systemctl restart "${SERVICE_NAME}"
    print_info "Board restarted."
}

show_status() {
    systemctl status "${SERVICE_NAME}" --no-pager
}

show_log() {
    journalctl -u "${SERVICE_NAME}" -n 100 --no-pager
}

show_help() {
    echo -e "Usage: $0 ${GREEN}[command]${NC}"
    echo ""
    echo "Commands:"
    printf "  ${GREEN}%-12s${NC} %s\n" "install"   "Install Board"
    printf "  ${GREEN}%-12s${NC} %s\n" "update"    "Update to latest version"
    printf "  ${GREEN}%-12s${NC} %s\n" "uninstall" "Uninstall Board"
    printf "  ${GREEN}%-12s${NC} %s\n" "purge"     "Completely uninstall Board and remove all data"
    printf "  ${GREEN}%-12s${NC} %s\n" "start"     "Start service"
    printf "  ${GREEN}%-12s${NC} %s\n" "stop"      "Stop service"
    printf "  ${GREEN}%-12s${NC} %s\n" "restart"   "Restart service"
    printf "  ${GREEN}%-12s${NC} %s\n" "status"    "Show service status"
    printf "  ${GREEN}%-12s${NC} %s\n" "log"       "Show recent logs"
}

show_menu() {
    get_installed_version
    get_latest_version_quiet
    if [[ -z "${LATEST_VERSION}" ]]; then
        LATEST_VERSION="获取失败"
    fi

    echo ""
    print_banner "╔════════════════════════════════════════════════════════╗"
    print_banner "║            Board 管理脚本                             ║"
    print_banner "╠════════════════════════════════════════════════════════╣"
    printf "  当前已安装版本: ${GREEN}%s${NC}\n" "${INSTALLED_VERSION}"
    printf "  最新可用版本:   ${GREEN}%s${NC}\n" "${LATEST_VERSION}"
    print_banner "╟────────────────────────────────────────────────────────╢"
    printf "  ${GREEN}1)${NC} 安装 Board\n"
    printf "  ${GREEN}2)${NC} 更新 Board\n"
    printf "  ${GREEN}3)${NC} 卸载 Board\n"
    printf "  ${GREEN}4)${NC} 彻底卸载（包含所有数据）\n"
    printf "  ${GREEN}5)${NC} 启动服务\n"
    printf "  ${GREEN}6)${NC} 停止服务\n"
    printf "  ${GREEN}7)${NC} 重启服务\n"
    printf "  ${GREEN}8)${NC} 查看状态\n"
    printf "  ${GREEN}9)${NC} 查看日志\n"
    printf "  ${GREEN}0)${NC} 退出\n"
    print_banner "╚════════════════════════════════════════════════════════╝"
    echo ""
    read -rp "请选择操作 [0-9]: " choice
    case "${choice}" in
        1) install ;;
        2) update ;;
        3) uninstall ;;
        4) purge ;;
        5) start_service ;;
        6) stop_service ;;
        7) restart_service ;;
        8) show_status ;;
        9) show_log ;;
        0) exit 0 ;;
        *) print_error "无效选项: ${choice}"; show_menu ;;
    esac
}

# Main entry
case "${1:-}" in
    install)    install ;;
    uninstall)  uninstall ;;
    purge)      purge ;;
    update)     update ;;
    start)      start_service ;;
    stop)       stop_service ;;
    restart)    restart_service ;;
    status)     show_status ;;
    log)        show_log ;;
    help)       show_help ;;
    "")         show_menu ;;
    *)          show_help ;;
esac
