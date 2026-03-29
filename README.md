# Board - 直播客户管理系统

[![Release](https://img.shields.io/github/v/release/ctsunny/board)](https://github.com/ctsunny/board/releases/latest)
[![License](https://img.shields.io/github/license/ctsunny/board)](LICENSE)

Board 是一个轻量级的直播客户管理后台，基于 Go 单二进制 + SQLite + Vue 3 构建，无需额外依赖，一键部署。

---

## 功能特性

- **仪表盘（Dashboard）**：总览客户数量、到期预警、服务器状态等关键指标
- **客户管理（Customers）**：新增、编辑、删除、批量续期、批量删除、CSV 导出
- **地区管理（Regions）**：维护多地区分类，方便客户归档
- **线路管理（Routes）**：管理多条线路，支持与节点关联
- **服务器管理（Servers）**：录入服务器信息，支持立即 Ping 检测
- **节点管理（Nodes）**：细粒度节点维护，关联服务器与线路
- **API Token**：生成、查看、撤销 API 访问令牌，支持外部集成
- **审计日志（Audit Logs）**：记录所有关键操作，可按资源/动作筛选
- **系统设置（Settings）**：配置端口、Base Path、管理员账号、邮件通知等
- **Gmail OAuth2**：内置 Gmail OAuth2 授权流程，到期自动发送邮件通知客户
- **ICMP Ping 监控**：定时 Ping 服务器，异常时发送告警邮件
- **在线更新**：后台一键升级到最新版本

---

## 快速安装

> 需要 root 权限，支持 Ubuntu 20.04+、Debian 11+、CentOS 7+

```bash
bash <(curl -Ls https://raw.githubusercontent.com/ctsunny/board/main/install.sh)
```

安装菜单会自动显示当前服务器已安装版本和 GitHub 最新版本。执行安装时，可按提示选择是否启用域名 HTTPS 自动申请/续期证书（基于 Let's Encrypt，需提前将域名解析到服务器并放通 80/443 端口）。

安装完成后将显示：

```
╔══════════════════════════════════════════════╗
║        Board 安装完成！                       ║
╠══════════════════════════════════════════════╣
║  访问地址: http://1.2.3.4:12345/mgmt-xxxxxx/ ║
║  用户名:   admin_xxxxxx                       ║
║  密  码:   xxxxxxxxxxxx                       ║
╚══════════════════════════════════════════════╝
```

---

## 管理命令

| 命令 | 说明 |
|------|------|
| `bash install.sh install` | 安装（默认） |
| `bash install.sh uninstall` | 卸载（保留配置和数据） |
| `bash install.sh update` | 更新到最新版本 |
| `bash install.sh start` | 启动服务 |
| `bash install.sh stop` | 停止服务 |
| `bash install.sh restart` | 重启服务 |
| `bash install.sh status` | 查看服务状态 |
| `bash install.sh log` | 查看最近日志 |

---

## 更新

```bash
bash <(curl -Ls https://raw.githubusercontent.com/ctsunny/board/main/install.sh) update
```

---

## Gmail OAuth2 配置

1. 登录后进入 **系统设置 → 邮件通知**
2. 填写 Google Cloud Console 中创建的 OAuth2 Client ID 和 Client Secret
3. 点击 **授权 Gmail**，完成 Google 账号授权
4. 授权成功后，到期提醒和服务器告警邮件将自动发送

> 详细配置步骤请参考设置页面内的引导说明。

---

## 技术架构

| 组件 | 说明 |
|------|------|
| 后端 | Go 单二进制，CGO_ENABLED=0 静态编译 |
| 数据库 | SQLite（嵌入，无需额外安装） |
| 前端 | Vue 3 + Vite，编译后内嵌至二进制 |
| 部署 | systemd 服务，支持 amd64 / arm64 |

配置文件：`/etc/board/config.json`（首次启动自动生成随机端口、路径和凭据）

数据目录：`/var/lib/board/`
