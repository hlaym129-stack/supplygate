# AI 协作说明

这份文档给所有接手本项目的 AI 和自动化编码工具阅读。目标是快速建立一致上下文，避免被历史命名、运行产物和上游文档干扰。

## 先读顺序

1. `README.md`：项目当前定位、功能范围、启动方式和部署说明。
2. `AGENTS.md`：当前这份 AI 协作约定。
3. `AI_WORKLOG.md`：最近 AI 修改记录、验证结果和遗留事项。
4. `frontend/src/router/index.ts`：前端页面结构、角色路由和导航守卫。
5. `backend/internal/server/router.go` 与 `backend/internal/server/routes/*.go`：后端 API 与网关入口。
6. 供应商相关改动优先看：
   - `backend/internal/service/supplier.go`
   - `backend/internal/handler/supplier_handler.go`
   - `backend/internal/handler/admin/supplier_handler.go`
   - `frontend/src/api/supplier.ts`
   - `frontend/src/views/supplier/`
   - `frontend/src/views/admin/SuppliersView.vue`

## 项目事实

- 当前项目名是 `SupplyGate`，定位是基于 Sub2API 改造的 AI API 中转与供应商接入平台。
- Go module、Docker 镜像示例、默认站点名、部分注释和包路径仍保留 `sub2api` / `github.com/Wei-Shaw/sub2api`。这是兼容上游结构，不要当作错误批量改名。
- 根目录只保留 `README.md` 作为主 README。`README_CN.md`、`README_JA.md` 已删除，避免继续传播上游旧说明。
- 子目录 README 是模块说明，例如 `deploy/README.md`、`frontend/src/router/README.md`，不要因为根 README 收敛就删除它们。
- `CLAUDE.md`、`GEMINI.md`、`.github/copilot-instructions.md` 只是入口跳转文件；本文件是唯一权威 AI 协作说明。
- `AI_WORKLOG.md` 记录每次 AI 实际修改后的摘要、验证和遗留事项；它不是代码事实来源，代码事实仍以 Git diff / commit 为准。

## 主要代码地图

- `backend/`：Go 后端、Ent schema、数据库迁移、业务服务、网关转发。
- `frontend/`：Vue 3 前端，包含用户端、管理员后台和供应商中心。
- `deploy/`：Docker Compose、systemd、部署脚本和完整配置示例。
- `docs/`：支付和后台集成文档。
- `backend/internal/service/`：核心业务服务。网关、计费、账号调度、供应商流程都在这里。
- `backend/internal/server/routes/`：HTTP 路由注册。先看这里再找 handler。
- `backend/ent/schema/`：数据库 schema 源文件；`backend/ent/` 大量文件为生成代码。
- `frontend/src/views/`：页面级组件。
- `frontend/src/api/`：前端 API client。
- `frontend/src/stores/`：Pinia 状态。
- `frontend/src/i18n/locales/`：中英文界面文案。

## 供应商流程注意点

- 供应商注册时提交主体资料，资料状态为 `pending`。
- 未审核通过的供应商可以登录和维护资料，但不能测试或提交上游账号。
- 供应商账号提交前必须通过服务端预检，拿到 `test_token` 后才能创建或重新提交账号。
- 供应商账号默认 `schedulable=false` 且 `approval_status=pending`，管理员审核通过后才进入调度池。
- 供应商结算报价只影响供应商结算口径，不等同于用户扣费模型。
- 已审核账号的报价修订当前限制为每天 07:00-07:30 提交一次，并需管理员审核生效。

## 不要被这些内容干扰

- `.runlogs/`、`backend/.runlogs/`、`frontend/.runlogs/`、`.data/`、`.test-data/`、`backend/data/` 是本地运行产物或测试数据，不属于产品逻辑。
- `backend/config.yaml`、`deploy/config.yaml`、`.env`、`.installed` 可能包含本地配置或敏感信息，不要提交真实值。
- `backend/ent/` 下大部分是生成代码。修改数据库结构时先改 `backend/ent/schema/` 和迁移，再按项目方式重新生成。
- 不要为了“统一品牌”批量替换 `Sub2API`、`sub2api`、包路径或镜像名，除非任务明确要求完整改名并处理所有兼容影响。
- 不要依据根目录已删除的多语言 README 或上游 Sub2API 在线文档来覆盖本项目当前定位。

## 常用验证命令

后端：

```bash
cd backend
go test ./...
```

前端：

```bash
cd frontend
pnpm run lint:check
pnpm run typecheck
pnpm run test:run
```

全量构建：

```bash
make build
```

本地 Docker 源码启动：

```bash
cd deploy
docker compose -f docker-compose.dev.yml up --build -d
```

## 改动原则

- 先定位路由、handler、service、repository 的调用链，再改代码。
- 保持改动范围小；不要顺手重构无关模块。
- 涉及账号调度、计费、供应商审核、支付、鉴权、迁移的改动必须补测试或说明无法测试的原因。
- 如果工作区已有与任务无关的本地改动，保留它们，不要回滚。
- 每次完成实际文件修改后，在 `AI_WORKLOG.md` 顶部追加一条记录，说明需求、改动、关键判断、验证和遗留。
- 最终回复要明确说明改了哪些文件、跑了哪些验证、还有哪些本地脏文件不是本次改动。
