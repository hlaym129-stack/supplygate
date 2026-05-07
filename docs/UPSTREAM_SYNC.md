# 跟进 sub2api 上游更新

SupplyGate 不是 GitHub 上的 `Wei-Shaw/sub2api` fork，当前仓库也没有可直接 merge 的共同 Git 祖先。同步上游时不要直接把 `upstream/main` 合入 `main`，而是先生成隔离分支和候选补丁。

## 一次性配置

本地只读上游远程应为：

```bash
git remote add upstream https://github.com/Wei-Shaw/sub2api.git
git remote set-url --push upstream DISABLED
```

同步脚本每次运行都会校验并修正这个配置，并使用 `--no-tags` 拉取上游分支，避免把 sub2api 的版本标签带入 SupplyGate。

## 同步流程

从干净工作区运行：

```bash
.github/scripts/sync-sub2api.sh
```

脚本会：

- 从 `origin/main` 创建 `codex/sync-sub2api-YYYYMMDD` 分支。
- 读取 `.github/upstream-sub2api.env` 中记录的上游基线。
- 生成从该基线到 `upstream/main` 的候选补丁。
- 排除 SupplyGate 自有文件：`README.md`、`CLA.md`、版本文件、发布工作流和关键项目链接页面。
- 应用补丁后运行保护检查。
- 更新 `.github/upstream-sub2api.env`，表示本次 PR 接受后新的上游基线。

只查看候选变更、不改工作区：

```bash
.github/scripts/sync-sub2api.sh --dry-run
```

应用候选补丁后立即跑完整检查：

```bash
.github/scripts/sync-sub2api.sh --run-tests
```

## 合并前检查

同步分支必须先人工检查 diff：

```bash
git diff --stat origin/main
git diff origin/main
.github/scripts/check-sub2api-sync-guard.sh origin/main
.github/scripts/run-sub2api-sync-checks.sh
```

PR 分支名必须使用 `codex/sync-sub2api-` 前缀。GitHub Actions 会对这类 PR 自动运行 `Upstream Sync Guard`，如果上游补丁改动了 SupplyGate 自有文件会直接失败。

## 发布规则

同步 PR 合并到 `main` 后，不会自动发布新版本。确认项目正常后，再按 SupplyGate 自己的版本节奏创建标签，例如 `v0.0.2`、`v0.0.3`。
