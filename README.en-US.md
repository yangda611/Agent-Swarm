# maliang swarm (English)

`maliang swarm` is an enterprise-oriented multi-agent orchestration desktop application built with `Go + Wails + React`.

## Key Capabilities

- Chief-brain orchestration: converts an incoming mission into roles, atomic tasks, and execution lanes
- Multi-agent coordination: supports up to 100 concurrent agents (config-driven)
- Enterprise governance gates: peer review, QA, security, human approval, and delivery sign-off
- Pixel office visualization: live desks, agent status, and handoff routes
- AI provider registry: supports OpenAI Compatible, Anthropic, Gemini, Azure OpenAI, Ollama, and Custom HTTP
- Agent traceability: shows each agent's latest action, output, handoff target, and local artifact path
- Local persistence: runtime snapshots, timeline, settings, and state in SQLite
- Artifact persistence: task outputs are saved to `data/artifacts/<run-id>/`

## Tech Stack

- Backend: Go 1.22+
- Desktop shell: Wails v2
- Frontend: React + TypeScript + Vite + ECharts
- Storage: SQLite (`modernc.org/sqlite`)

## Local Development

```bash
wails dev
```

## Build

```bash
wails build
```

## Release Automation

GitHub Actions release workflow is configured:

- Workflow file: `.github/workflows/release.yml`
- Trigger: push a `v*` tag (for example `v0.1.1`) or run manually
- Output assets:
  - Windows: `.exe`
  - macOS: `.zip` (app bundle or binary)
  - Linux: `.tar.gz` (binary package)

## Project Layout

- `internal/orchestrator`: planning, execution progression, handoff logic, trace enrichment
- `internal/storage`: SQLite schema and persistence
- `frontend/src`: dashboard, office view, task board, timeline, settings, and dialogs
- `docs/swarm_enterprise_plan.md`: planning document

## License

No OSS license is declared yet. Add a `LICENSE` file before public open-source release.
