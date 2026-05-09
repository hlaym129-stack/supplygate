# SupplyGate

> 基于 [Sub2API](https://github.com/Wei-Shaw/sub2api) 改造的 AI API 中转与供应商接入平台。

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.26.2+-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D.svg)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED.svg)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-LGPL_v3-blue.svg)](LICENSE)

**让平台运营方管理用户、分组、计费和调度，同时允许第三方供应商提交自己的上游账号参与服务。**

</div>

## 项目定位

SupplyGate 保留了 Sub2API 的核心网关能力：用户通过平台生成的 API Key 调用 Claude、OpenAI、Gemini、Antigravity 等上游能力，平台负责鉴权、调度、计费、并发控制、用量记录和后台管理。

本项目在此基础上新增了供应商中心：供应商注册后提交主体资料，经管理员审核通过后，可以提交自有上游账号。账号必须先完成连通性预检和管理员审核，审核通过后才会进入调度池。这样平台不再只依赖运营方手工录入上游资源，而是可以形成可审核、可结算的供应商接入流程。

代码模块名和二进制/systemd 安装路径仍保留 `sub2api` 命名，这是为了兼容上游项目结构和既有二进制部署；发布镜像和 Docker Compose 部署口径统一使用 SupplyGate。

## 核心能力

- **多协议网关**：支持 Claude Messages、OpenAI Responses / Chat Completions / Images、Gemini v1beta、Antigravity 专用入口，并按分组平台自动路由。
- **上游账号池**：支持 OAuth、API Key、Setup Token、AWS Bedrock、Google Service Account 等账号类型，具备代理、TLS 指纹、限额、过载冷却、粘性会话和连接池隔离等调度能力。
- **用户与 API Key 分发**：用户注册、登录、资料管理、API Key 创建、IP 限制、用量查询、可用渠道展示。
- **分组与计费**：按分组管理平台、倍率、订阅模式、余额扣费、并发限制、模型定价和用量聚合。
- **供应商中心**：供应商资料审核、账号预检、账号审核、退回修改、结算报价、报价修订审核、供应商用量与结算视图。
- **支付与运营**：内置易支付、支付宝、微信支付、Stripe，支持订单、套餐、兑换码、优惠码、邀请返利。
- **管理后台**：用户、账号、分组、渠道定价、渠道监控、代理、公告、系统设置、备份恢复、运维监控、错误重试和系统日志。
- **登录与安全**：邮箱注册登录、TOTP、LinuxDo Connect、微信、通用 OIDC、Turnstile、CSP、响应头过滤、URL 白名单。

## 供应商流程

1. 供应商在注册页选择“供应商用户”，填写邮箱、密码和主体资料。
2. 注册后可登录并维护供应商资料；未审核通过前不能提交或测试上游账号。
3. 管理员在 `/admin/suppliers` 审核供应商主体资料。
4. 已通过的供应商在 `/supplier/accounts` 提交上游账号，提交前必须完成模型连通性预检。
5. 新账号默认不可调度，状态为待审核；管理员审核通过后才会绑定分组并进入调度池。
6. 供应商可在看板查看账号、审核状态、模型用量、平台成本口径和供应商结算口径。
7. 已通过账号支持结算报价修订；当前规则限制为每天 07:00-07:30 可提交一次，需管理员审核后生效。

当前供应商账号支持范围：

| 平台 | 账号类型 |
|------|----------|
| Anthropic / Claude | `oauth`、`setup-token`、`apikey`、`bedrock`、`service_account` |
| OpenAI / Codex | `oauth`、`apikey` |
| Gemini | `oauth`、`apikey`、`service_account` |
| Antigravity | `oauth` |

## 网关入口

客户端 API Key 推荐放在 `Authorization: Bearer <key>`，也兼容 `x-api-key`；Gemini 原生入口兼容 `x-goog-api-key`。

| 入口 | 说明 |
|------|------|
| `POST /v1/messages` | Claude Messages 兼容入口；OpenAI 分组会走 OpenAI 消息兼容链路 |
| `POST /v1/messages/count_tokens` | Claude token 计数 |
| `GET /v1/models` | 模型列表 |
| `GET /v1/usage` | 当前 API Key 用量 |
| `POST /v1/responses`、`POST /responses` | OpenAI Responses / Codex 兼容入口 |
| `GET /v1/responses`、`GET /responses` | OpenAI Responses WebSocket 入口 |
| `POST /backend-api/codex/responses` | Codex 客户端直连兼容入口 |
| `POST /v1/chat/completions`、`POST /chat/completions` | OpenAI Chat Completions 兼容入口 |
| `POST /v1/images/generations`、`POST /v1/images/edits` | OpenAI Images 兼容入口 |
| `/v1beta/models...` | Gemini 原生 API 兼容入口 |
| `/antigravity/v1/...`、`/antigravity/v1beta/...` | Antigravity 专用入口 |

通过 Nginx 反向代理并搭配 Codex CLI / 多账号粘性会话时，需要在 Nginx `http` 块中开启：

```nginx
underscores_in_headers on;
```

否则 Nginx 会丢弃 `session_id` 等带下划线的请求头，影响粘性会话。

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.26.2、Gin、Ent、Wire |
| 前端 | Vue 3.4、Vite 5、TypeScript、Pinia、TailwindCSS |
| 数据库 | PostgreSQL 15+（Docker Compose 默认 PostgreSQL 18） |
| 缓存/队列 | Redis 7+（Docker Compose 默认 Redis 8） |
| 部署 | Docker、Docker Compose、systemd、内嵌前端静态资源 |

## 快速启动

### 方式一：一键 Docker 部署

用于直接安装已发布的 SupplyGate 镜像和部署模板。

```bash
curl -sSL https://raw.githubusercontent.com/hlaym129-stack/supplygate/main/deploy/docker-deploy.sh | bash
docker compose up -d
docker compose logs -f supplygate
```

访问：`http://localhost:8080`

### 方式二：从本地源码 Docker 启动

这是运行当前改造版本最直接的方式。`docker-compose.dev.yml` 会从本仓库源码构建镜像，并自动启动 PostgreSQL、Redis 和应用。

```bash
cd deploy
cp .env.example .env

# 至少修改 POSTGRES_PASSWORD；生产环境还应固定 JWT_SECRET、TOTP_ENCRYPTION_KEY、ADMIN_PASSWORD
$EDITOR .env

docker compose -f docker-compose.dev.yml up --build -d
docker compose -f docker-compose.dev.yml logs -f supplygate
```

访问：`http://localhost:8080`

如果 `ADMIN_PASSWORD` 留空，首次自动初始化时会在容器日志中输出生成的管理员密码。

### 方式三：本地开发运行

先准备 PostgreSQL 和 Redis。可以使用本机服务，也可以只启动开发 Compose 里的数据库和缓存：

```bash
cd deploy
cp .env.example .env
$EDITOR .env
docker compose -f docker-compose.dev.yml up -d postgres redis
```

后端：

```bash
cd backend

# 可选：手动配置。也可以不创建 config.yaml，让首次启动进入 Web Setup。
cp ../deploy/config.example.yaml config.yaml
$EDITOR config.yaml

go run ./cmd/server
```

前端：

```bash
cd frontend
corepack enable
pnpm install
pnpm dev
```

开发访问：`http://localhost:3000`，Vite 默认代理到 `http://localhost:8080`。如需修改后端地址，可设置 `VITE_DEV_PROXY_TARGET`。

## 常用命令

```bash
# 构建后端与前端
make build

# 后端测试
cd backend && go test ./...

# 前端检查与测试
cd frontend
pnpm run lint:check
pnpm run typecheck
pnpm run test:run

# 构建内嵌前端的单二进制
pnpm --dir frontend run build
cd backend && CGO_ENABLED=0 go build -tags embed -trimpath -o bin/server ./cmd/server
```

## 目录结构

```text
backend/        Go 后端、Ent schema、迁移、网关与业务服务
frontend/       Vue 管理后台、用户端、供应商端
deploy/         Docker Compose、systemd、示例配置和部署脚本
docs/           支付和后台集成相关文档
assets/         README 与合作方展示素材
tools/          辅助检查脚本
```

## 配置说明

配置文件默认查找顺序为：

1. `DATA_DIR/config.yaml`
2. `/app/data/config.yaml`
3. 当前工作目录下的 `config.yaml`
4. `./config/config.yaml`
5. `/etc/sub2api/config.yaml`

首次启动时，如果数据目录下没有 `config.yaml` 和 `.installed`，服务会进入安装向导；Docker 自动部署则通过 `AUTO_SETUP=true` 和环境变量生成配置。

关键配置项：

- `server`：监听地址、端口、CORS、h2c、请求体限制。
- `database` / `redis`：PostgreSQL 和 Redis 连接信息。
- `jwt.secret`：生产环境必须固定，避免重启后登录态失效。
- `totp.encryption_key`：生产环境必须固定，避免 2FA 密钥失效。
- `run_mode`：`standard` 启用完整 SaaS 计费与余额校验；`simple` 隐藏部分 SaaS 功能并跳过余额校验。
- `gateway`：上游超时、请求体限制、连接池隔离、Codex / Sora / WebSocket 等网关行为。
- `security.url_allowlist`：限制可访问的上游、定价和 CRS 同步地址。
- `pricing`：模型价格数据来源与本地兜底文件。

完整配置示例见 [`deploy/config.example.yaml`](deploy/config.example.yaml)，Docker 环境变量示例见 [`deploy/.env.example`](deploy/.env.example)。

## 从 Sub2API 的主要改造

- 增加 `supplier_profiles`、供应商账号归属、账号审核状态和供应商结算报价修订。
- 注册流程支持供应商类型，供应商注册时即提交主体资料。
- 新增供应商端路由：`/supplier/dashboard`、`/supplier/accounts`、`/supplier/usage`、`/supplier/profile`。
- 新增管理员供应商中心：供应商资料审核、供应商账号审核、退回修改、报价审核。
- 供应商账号提交前必须通过服务端预检；审核通过前不会进入调度池。
- 用量统计增加供应商归属和结算成本口径。

## 贡献

欢迎提交 Issue 和 PR。提交前请确认：

- 不提交真实密钥、数据库配置、运行日志和本地数据。
- 后端改动至少运行相关 `go test`。
- 前端改动至少运行 `pnpm run lint:check`、`pnpm run typecheck` 和相关 Vitest。
- 涉及账号调度、计费、供应商审核、支付、权限的改动需要补充测试。

提交 PR 前请阅读并签署 [贡献者许可协议](CLA.md)。

## 联系

- 邮箱：hlaym129@gmail.com
- 微信群：扫码加入，或邮件联系拉你进群

<img src="assets/wechat-group-qr.jpeg" alt="微信群二维码" width="200">

## 致谢

本项目基于 [Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api) 改造，感谢原作者及所有贡献者。

## 许可证

[LGPL v3](LICENSE)
