# maliang swarm（日本語）

`maliang swarm` は、`Go + Wails + React` で構築された企業向けマルチエージェント編成デスクトップアプリです。

## 主な機能

- 主脳オーケストレーション：入力されたミッションを役割・原子タスク・実行レーンへ自動変換
- マルチエージェント協調：設定により最大 100 エージェント並列に対応
- 企業ガバナンスゲート：同僚レビュー、QA、セキュリティ、人手承認、納品承認を標準装備
- ピクセルオフィス可視化：デスク配置、エージェント状態、ハンドオフ経路を可視化
- AI API レジストリ：OpenAI Compatible / Anthropic / Gemini / Azure OpenAI / Ollama / Custom HTTP をサポート
- エージェント追跡：各エージェントの最新アクション、返却内容、受け渡し先、ローカル成果物パスを表示
- ローカル永続化：実行スナップショット、タイムライン、設定を SQLite に保存
- 成果物保存：タスク出力を `data/artifacts/<run-id>/` に自動保存

## 技術スタック

- バックエンド：Go 1.22+
- デスクトップ：Wails v2
- フロントエンド：React + TypeScript + Vite + ECharts
- ストレージ：SQLite（`modernc.org/sqlite`）

## ローカル開発

```bash
wails dev
```

## ビルド

```bash
wails build
```

生成物：

- `build/bin/maliang swarm.exe`

## ディレクトリ構成

- `internal/orchestrator`：計画、実行進行、ハンドオフ、追跡情報の整備
- `internal/storage`：SQLite スキーマと永続化
- `frontend/src`：ダッシュボード、オフィス画面、タスクボード、タイムライン、設定、モーダル
- `docs/swarm_enterprise_plan.md`：計画ドキュメント

## Release 自動化

GitHub Actions の Release ワークフローを設定済みです。

- ワークフロー：`.github/workflows/release.yml`
- トリガー：`v*` タグ（例 `v0.1.0`）の push、または手動実行
- 出力：Windows 実行ファイルをビルドし、GitHub Release に添付

## ライセンス

現在は OSS ライセンス未設定です。公開する場合は `LICENSE` の追加を推奨します。
