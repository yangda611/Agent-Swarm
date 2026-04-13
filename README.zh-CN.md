# maliang swarm（简体中文）

`maliang swarm` 是一个面向企业流程的多智能体编排桌面应用，基于 `Go + Wails + React` 开发。

## 核心能力

- 主脑编排：任务接入后自动规划角色、拆解原子任务、安排执行链路
- 多智能体协作：支持最多 100 个 agent 并发编排（按运行配置）
- 企业门禁：内置同评、QA、安全、人工审批、交付签收等治理节点
- 可视化办公室：以像素办公室场景展示 agent 工位、状态和资料交接
- 接口注册中心：支持多种 AI API 格式（OpenAI Compatible / Anthropic / Gemini / Azure OpenAI / Ollama / Custom HTTP）
- 运行可追踪：agent 可查看最近动作、返回内容、交接对象与本地产物路径
- 本地持久化：运行快照、任务状态、时间线与配置保存在 SQLite
- 产物落盘：任务输出自动写入 `data/artifacts/<run-id>/`

## 技术栈

- 后端：Go 1.22+
- 桌面容器：Wails v2
- 前端：React + TypeScript + Vite + ECharts
- 存储：SQLite（`modernc.org/sqlite`）

## 本地开发

```bash
wails dev
```

## 构建

```bash
wails build
```

## Release 自动化（支持三平台）

仓库已配置 GitHub Actions 发布流程：

- 文件：`.github/workflows/release.yml`
- 触发方式：推送 `v*` 标签（如 `v0.1.1`）或手动触发
- 发布产物：
  - Windows：`.exe`
  - macOS：`.zip`（`.app` 或二进制）
  - Linux：`.tar.gz`（二进制包）

## 项目结构

- `internal/orchestrator`：主脑编排、任务推进、agent 交接与追踪
- `internal/storage`：SQLite 持久化与迁移
- `frontend/src`：控制台、办公室视图、任务板、时间线、设置与弹窗
- `docs/swarm_enterprise_plan.md`：规划文档

## 许可

项目当前未声明开源许可证，如需开源发布建议补充 `LICENSE` 文件。
