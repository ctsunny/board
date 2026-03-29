# AI 知识库 | board

> 自动维护。记录每次修改的目的、思路与关键技术，供后续任务参考。
> 规则：只追加，不删改历史。

## 项目概况
- **类型**：Go 单文件部署的直播资源与客户管理后台
- **技术栈**：Go、Gin、GORM、SQLite、Vue 3、Element Plus、Vite
- **核心功能**：客户管理、直播地区/线路/服务器/节点维护、审计日志、系统设置
- **特别约定**：前端产物位于 `web/dist` 并被 Go 嵌入；历史知识库仅追加不改写

---

### [2026-03-29] 资源界面整合

**目的**：把直播地区、直播线路、服务器、节点融合到一个统一界面里录入和管理。

**思路**：保留后端原接口不动，新建统一前端页，用标签页和概览卡整合四类资源，同时让旧路由跳到对应标签减少兼容风险。

**关键技术**：新增 `LiveResources.vue` 复用现有 CRUD API；用 query `tab` 同步标签状态；侧边栏改成单入口，旧地址通过路由 redirect 保持可用。

**遗留/风险**：统一页逻辑较集中，后续若继续扩展可再拆分可复用子组件；前端构建后需同步更新 `web/dist`。

**涉及文件**：`ai.md` `web/src/views/LiveResources.vue` `web/src/router/index.ts` `web/src/components/Layout.vue`

---
### [2026-03-29] Gmail配置修复

**目的**：修复 Gmail 通知提醒在系统设置中保存配置失败、状态回显异常的问题。

**思路**：延续前后端最小改动原则，不改设置页交互，改后端同时兼容前端已发送的平铺字段，并补齐读取接口返回字段。

**关键技术**：`UpdateSettings` 同时支持 `gmail_client_id` 等平铺字段与 `gmail` 嵌套对象；`GetSettings` 增补 `gmail_client_id`、`gmail_admin_email`、`gmail_configured` 供前端直接回显。

**遗留/风险**：当前“已配置”仍主要表示 OAuth 已完成；若后续想区分“已保存凭据”和“已完成授权”，需再拆分状态字段。

**涉及文件**：`ai.md` `internal/api/handlers.go` `internal/api/handlers_settings_test.go`
