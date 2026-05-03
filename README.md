# SupplyGate

> **社区驱动的 AI API 中转平台** — 让任何人都能成为 API 供应商，共建开放的中转生态。

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8.svg)](https://golang.org/)
[![Vue](https://img.shields.io/badge/Vue-3.4+-4FC08D.svg)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791.svg)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7+-DC382D.svg)](https://redis.io/)
[![License](https://img.shields.io/badge/License-LGPL_v3-blue.svg)](LICENSE)

**🚧 早期开发阶段，欢迎贡献！**

</div>

---

## 这是什么？

SupplyGate 是一个**开放的 AI API 中转平台**，基于 [Sub2API](https://github.com/Wei-Shaw/sub2api) 改造而来。

**原版 Sub2API** 是一个优秀的 API 中转网关，但上游账号只能由平台运营方手动录入管理。

**SupplyGate 的目标**：把上游账号的录入能力**开放出去** — 任何第三方供应商都可以注册、提交自己的 API 账号，经管理员审核通过后即可接入平台对外提供服务。

## 工作流程

1. **供应商注册** — 填写邮箱密码 + 主体资料（个人或企业均可），直接登录进入供应商看板
2. **提交账号** — 供应商提交自己的 API 上游账号（如 OpenAI Key）
3. **管理员审核** — 审核通过后，该账号正式接入平台
4. **对外服务** — 平台用户即可通过该供应商的账号调用 AI API
5. **用量结算** — 供应商在自己的看板查看用量和收益

> 简单说：原版 Sub2API 只能运营方自己录入上游账号；SupplyGate 把这个能力开放出去，**人人都能当供应商**。

## 核心改造计划

| 功能 | 状态 | 说明 |
|------|------|------|
| 供应商自助注册 | 🚧 开发中 | 注册时填写主体资料，即可成为供应商 |
| 账号提交与审核 | ✅ 已就绪 | 供应商提交 API 账号，管理员审核后启用 |
| 供应商入驻页面移除 | 🚧 开发中 | 合并到注册流程，不再需要单独的入驻页面 |
| 供应商看板 | 🚧 开发中 | 供应商登录后查看自己的账号和用量 |

详见 [供应商注册流程改造计划](.trae/documents/supplier-registration-flow.md)

## 快速开始

```bash
# 后端
cd backend
cp config.example.yaml config.yaml  # 编辑配置
go run ./cmd/server/

# 前端（必须用 pnpm）
cd frontend
pnpm install
pnpm dev
```

详细环境配置见 [DEV_GUIDE.md](DEV_GUIDE.md)

## 贡献

目前项目处在早期阶段，欢迎任何形式的贡献：

- 💡 **讨论想法** — 提 Issue 聊聊你对供应商市场的想法
- 🐛 **修复 Bug** — 看看 Sub2API 原版有什么可改进的
- 🔧 **功能开发** — 挑一个上面 TODO 列表中的功能下手

提交 PR 前请阅读 [贡献者许可协议](CLA.md) 并签名。

## 联系

- 📧 **邮箱** — hlaym129@gmail.com
- 💬 **微信群** — 扫码加入（QR 码见下方），或邮件联系拉你进群

## 致谢

本项目基于 [Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api)（LGPL v3），感谢原作者及所有贡献者。

## 许可证

[LGPL v3](LICENSE)
