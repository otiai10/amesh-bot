# Repository Guidelines

## プロジェクト構造とモジュール運用
SlackのEvent Subscriptionを受け取りアメッシュ画像を返すGo 1.24製Botです。`main.go`がCloud Run/GAE互換のエントリポイントで、Slack I/Oは`bot/`、HTTPルーティングは`controllers/`、メンション別ロジックは`commands/`、外部API/Storage連携は`service/`に置いています。`testdata/`はPNGやJSONのフィクスチャ、`vendor/`は固定依存、`app.dev.yaml`・`app.yaml`・`app-secrets.example.yaml`で環境を切り替えます。

## リクエストフローと主要パッケージ
1. Slackが`/slack/webhook`へイベント送信
2. `controllers.Controller.Webhook`が検証して`bot.Bot`に委譲
3. `bot.Bot.Handle`が`commands/`の実装を選択
4. コマンドがSlackへ返信（public/thread/forceThreadReplyで挙動制御）

## 開発・ビルド・テストコマンド
- `APP_CONFIG=app.dev.yaml SLACK_SIGNING_SECRET=xxxx PORT=8080 go run ./main.go` : ローカル実行。`ngrok http 8080`でSlackへ公開。
- `go test ./...` : controllers・commands・bot配下を一括テスト。
- `go test -coverprofile=coverage.out ./commands ./controllers` → `go tool cover -html=coverage.out` : 主要経路のカバレッジ確認。
- `GOOS=linux GOARCH=amd64 go build -o bin/amesh-bot ./...` : GAE/Cloud Run向けバイナリ確認。
- `go fmt ./... && go vet ./...` : 最小整形＋静的解析。

## 環境変数と設定
`GOOGLE_PROJECT_ID`、`GOOGLE_CUSTOMSEARCH_API_KEY`、`GOOGLE_CUSTOMSEARCH_ENGINE_ID`、`OPENAI_APIKEY`、`SLACK_SIGNING_SECRET`、`PORT`がローカルで必要です。シークレットの実値は未追跡の`app-secrets*.yaml`に置き、`codex --help`→`codex resume --last`→`codex mcp --list`の順で設定後の疎通を確認します。

## コーディングスタイルと命名規則
`go fmt`/`goimports`を必須とし、パッケージは小文字スネーク、ディレクトリはケバブケース、JSON/TOML/YAMLは2スペースインデント＋`snake_case`キーを徹底します。エラーハンドリングは`fmt.Errorf("context: %w", err)`でラップし、Slack向け文言は日本語150字以内、理由が必要な表現にはコメントで根拠URLを添えます。

## テスト指針
`*_test.go`はテーブル駆動で`[]struct{}`を用い、フィクスチャは`testdata/`から読み込みます。Slack Rate Limit、アメッシュAPIタイムアウト、Cloud Storage書き込み失敗など主要異常系を網羅し、新規コマンドでは正常/異常/Contextキャンセルの3ケースを追加してください。`go test ./commands -run TestAmesh`のような部分実行手順や`go test -race ./...`の結果をPRに添付します。

## コミットとプルリクエストガイドライン
コミットは日本語Conventional Commits（例: `feat: commands/forecast の通知強化`）で統一し、PRはスカッシュで1コミットにまとめます。テンプレートの概要/手動検証/言語チェック/リスク/関連リンクを必ず埋め、`go test ./...`や`GOOS=linux go build ...`など実行コマンドと結果ログを記載します。`config`や設定YAMLを変更した場合は対象プロファイル・更新キー・期待挙動を表で整理し、必要に応じてマスク済み`sessions/`抜粋や`codex resume --last`ログを添付します。

## ブランチ運用とデプロイ
`develop`ブランチはDEV環境へ自動デプロイ（`app.dev.yaml`＋`app-secrets.dev.yaml`）、`main`ブランチは本番へデプロイ（`app.yaml`＋`app-secrets.yaml`）。どちらもpush時にGitHub ActionsでGoテストと`gae-deploy`ワークフローが走ります。PRは`main`宛てに作成し、デプロイ前に`GOOS=linux`ビルドと`codex mcp --list`の健全性を確認してください。

## セキュリティと設定の注意
Slackトークン・GCP認証鍵・`app-secrets*.yaml`はgit管理下に置かず、`.gitignore`変更時は機密保護を再確認します。マシン固有の調整は環境変数や無視対象ファイルで扱い、権限拡張やサービスアカウントスコープ変更の際は`trust_level`を見直し、理由をIssue/PR本文に残して監査可能性を確保します。
