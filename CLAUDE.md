# SupplyGate 项目说明书

## 我是谁

- Founder：hlaym129
- 邮箱：hlaym129@gmail.com
- 微信群：见 README 底部二维码
- 项目定位：社区驱动的 AI API 中转平台，让任何人都能成为 API 供应商

## 项目来源

- Fork 自 [Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api)
- 许可证：LGPL v3（必须保留）
- GitHub 仓库：https://github.com/hlaym129-stack/supplygate
- 本地路径：`/Users/ywfw/项目/中转/中转2.0`

## 核心目标

把 Sub2API 从「运营方自己录入上游账号」改造成「**开放的供应商市场**」：

1. 供应商自助注册（个人或企业均可）→ 直接登录供应商看板
2. 供应商提交自己的 API 账号 → 管理员审核通过后接入平台
3. 移除原版独立的"供应商入驻"页面，合并到注册流程
4. 供应商可查看用量和收益

详见：`.trae/documents/supplier-registration-flow.md`

## 技术栈

- 后端：Go 1.25+, Ent ORM + Gin
- 前端：Vue 3.4+, **pnpm**（不是 npm）
- 数据库：PostgreSQL 16 + Redis
- 详细环境配置见 `DEV_GUIDE.md`

## 同步上游

上游仓库：https://github.com/Wei-Shaw/sub2api.git

```bash
# 首次添加 remote
git remote add upstream https://github.com/Wei-Shaw/sub2api.git

# 每次同步
git fetch upstream
git merge upstream/main
```

## 不用做的事

- 不要创建独立的 README_CN.md / README_JA.md 等语言版本（保持简洁）
- 不要删除 LICENSE 文件
- 不要在 commit 中提交 OAuth 密钥等敏感信息

## 当前状态

- 仓库已公开
- README 已写好
- 代码尚未开始改动（还跟上游一样）
- 下一步：按 supplier-registration-flow.md 的计划开始改前端和后端
