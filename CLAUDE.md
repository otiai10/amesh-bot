# CLAUDE.md

Claude Code（claude.ai/code）が本リポジトリで作業するときの手引きです。指示・PRは必ず日本語でまとめてください。

## 概要
Slackで`@amesh`とメンションすると東京アメッシュの雨雲画像や関連情報を返すGo 1.24製Botです。Google App Engine StandardとCloud Runで動作し、`develop`→DEV、`main`→PRODに自動デプロイされます。

## プロジェクト構造
- `main.go` : ルーターとbotの初期化を行うエントリポイント
- `bot/` : Slackイベントを適切なコマンドへルーティングする実行エンジン
- `commands/` : `Amesh`, `Forecast`, `Image`, `Google`, `LGTM`, `Typhoon`, `AICompletion`など各メンション処理
- `controllers/` : `/slack/webhook`やOAuthなどのHTTPハンドラ
- `service/` : Slackクライアントや気象庁API、Cloud Storage/Firestore連携
- `testdata/` : PNG・JSONのフィクスチャ
- `sessions/`→進行中ログ、`archived_sessions/`→完了ログ、`log/`→実行トレース

## リクエストフロー
1. Slack Event Subscriptionが`/slack/webhook`にPOST
2. `controllers.Controller.Webhook`が署名検証とパース
3. `bot.Bot.Handle`が`commands.Command`（`Match`, `Execute`, `Help`）を選択
4. コマンドがpublic/thread/forceThreadReply方針に従ってSlackに返信

## 開発・ビルド・テストコマンド
```bash
APP_CONFIG=app.dev.yaml SLACK_SIGNING_SECRET=xxxx PORT=8080 go run ./main.go  # ローカル起動
ngrok http 8080                                                           # Slackへ公開
go test ./...                                                             # 全テスト
go test -coverprofile=coverage.out ./commands ./controllers               # カバレッジ
GOOS=linux GOARCH=amd64 go build -o bin/amesh-bot ./...                   # デプロイ確認
go fmt ./... && go vet ./...                                              # 整形+静的解析
codex --help && codex resume --last && codex mcp --list                   # config変更後の疎通
```
単体検証時は`go test -v -run TestFunctionName ./package`を使い、Race検知は`go test -race ./...`で実行します。

## 環境変数と設定
ローカルでは`GOOGLE_PROJECT_ID`、`GOOGLE_CUSTOMSEARCH_API_KEY`、`GOOGLE_CUSTOMSEARCH_ENGINE_ID`、`OPENAI_APIKEY`、`SLACK_SIGNING_SECRET`、`PORT`が必要です。実値は未追跡の`app-secrets*.yaml`に置き、`app.dev.yaml`（開発）と`app.yaml`（本番）を切り替えます。

## ブランチとデプロイ
- `develop`ブランチ: `app.dev.yaml`+`app-secrets.dev.yaml`でDEVへ自動デプロイ
- `main`ブランチ: `app.yaml`+`app-secrets.yaml`でPRODへデプロイ
PRは`main`宛て。push時にGoテストと`.github/workflows/gae-deploy.yml`が走ります。

## コーディングスタイルと命名
`go fmt`/`goimports`を必須とし、パッケージは小文字スネーク、ディレクトリはケバブケース、JSON/TOML/YAMLは2スペースインデントと`snake_case`キーを徹底します。エラーは`fmt.Errorf("context: %w", err)`でラップし、Slack向け文言は日本語150字以内、特異な表現はコメントで根拠（URL可）を添えてください。

## テスト指針
`*_test.go`はテーブル駆動で`[]struct{}`を用い、フィクスチャは`testdata/`から読み込みます。Slack Rate Limit、アメッシュAPIタイムアウト、Cloud Storage書き込み失敗など主要異常系を必ず追加し、新コマンドでは正常/異常/Contextキャンセルの3ケースを揃えます。カバレッジは`go test -coverprofile=coverage.out ./commands ./controllers`→`go tool cover -html=coverage.out`で確認し、必要に応じて`go test -race ./...`を実行して結果をPRに貼り付けます。

## コミットとPR運用
コミットは日本語Conventional Commits（例: `feat: commands/forecast の通知強化`）で統一し、PRはスカッシュ1件にまとめます。テンプレートの概要/手動検証/言語チェック/リスク/関連リンクを全て埋め、手動検証欄には実行した`go test ./...`や`GOOS=linux go build ...`等のコマンドと結果ログを列挙します。`config.toml`や設定YAMLを変更した場合は対象プロファイル・更新キー・期待挙動を表形式で記載し、必要に応じてマスク済み`sessions/`抜粋や`codex resume --last`の成功ログを添付してください。

## セキュリティと設定の注意
SlackトークンやGCP認証情報、`app-secrets*.yaml`はgitに含めず、`.gitignore`変更時は機密保護を再確認します。マシン固有の調整は環境変数または無視対象ファイルで扱い、権限拡張やサービスアカウントのスコープ変更前に`trust_level`を見直してIssue/PRで理由を共有します。
