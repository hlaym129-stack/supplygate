# AI Worklog

本文件用于记录 AI 或自动化编码工具对项目做出的实际修改。它不是变更的权威来源，权威内容仍以 Git diff / commit 为准；这里记录的是改动意图、影响范围、验证结果和遗留事项，帮助下一位接手者快速进入上下文。

记录规则：

- 每次 AI 完成实际文件修改后，在顶部追加一条记录。
- 不粘贴大段代码或完整文件内容，只写摘要、关键判断和验证结果。
- 记录未完成事项、未运行测试的原因、以及工作区中与本次无关但仍存在的脏文件。
- 涉及敏感信息时只描述类型，不写真实值。

## 2026-05-05 - README 与 AI 协作入口整理

- 需求：根据项目当前内容重写主 README，并降低后续 AI 被上游旧文档、历史命名和本地运行产物干扰的概率。
- 改动：
  - 重写 `README.md`，明确 SupplyGate 的定位、供应商流程、网关入口、启动方式和配置说明。
  - 删除根目录 `README_CN.md`、`README_JA.md`，避免继续传播上游 Sub2API 原版说明。
  - 新增 `AGENTS.md` 作为唯一权威 AI 协作说明。
  - 新增 `CLAUDE.md`、`GEMINI.md`、`.github/copilot-instructions.md`，均只跳转到 `AGENTS.md`。
  - 更新 `.gitignore`，允许提交 AI 入口文档，并忽略 `.runlogs/`、`backend/.runlogs/`、`frontend/.runlogs/`、`.data/`、`.test-data/` 等本地运行数据。
  - 删除已跟踪的 `.runlogs/backend.pid`、`.runlogs/frontend.pid`。
- 关键判断：
  - 保留代码、包路径、Docker 镜像示例和部分默认站点名中的 `sub2api` 命名，避免破坏上游兼容结构。
  - `AGENTS.md` 作为 AI 协作单一事实来源，其他 AI 工具入口只做跳转，避免多份上下文分叉。
  - 子目录 README 属于模块说明，保留。
- 验证：
  - 检查 `README.md` 内部链接目标存在。
  - 检查 `AGENTS.md`、`CLAUDE.md`、`GEMINI.md`、`.github/copilot-instructions.md` 未被 `.gitignore` 忽略。
  - 检查运行日志目录会被 `.gitignore` 忽略。
- 未运行测试：
  - 本次只改文档、AI 入口和忽略规则，没有改业务代码。
- 遗留：
  - 工作区仍有未跟踪图片 `assets/partners/logos/6F3938A4-D8EE-4E69-8587-3A848FBE5973.jpeg`，本次未处理。
