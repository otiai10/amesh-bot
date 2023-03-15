# Slackで @amesh って言うとアメッシュの画像出すbot

[![Go](https://github.com/otiai10/amesh-bot/actions/workflows/go.yml/badge.svg)](https://github.com/otiai10/amesh-bot/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/otiai10/amesh-bot)](https://goreportcard.com/report/github.com/otiai10/amesh-bot)
[![codecov](https://codecov.io/gh/otiai10/amesh-bot/branch/main/graph/badge.svg?token=WQ9SNJ5EO8)](https://codecov.io/gh/otiai10/amesh-bot)
[![Maintainability](https://api.codeclimate.com/v1/badges/2d4e967a9b401d12653e/maintainability)](https://codeclimate.com/github/otiai10/amesh-bot/maintainability)
[![GAE Deploy](https://github.com/otiai10/amesh-bot/actions/workflows/gae-deploy.yml/badge.svg)](https://github.com/otiai10/amesh-bot/actions/workflows/gae-deploy.yml)

<img width="60%" src="https://user-images.githubusercontent.com/931554/118962430-55c7de80-b9a0-11eb-9ce7-4845a72964bf.png" />

# これは何をするの

1. Slackの [Event Subscription](https://api.slack.com/events-api) からのRequestを受けるサーバです
2. `@amesh` と言われると、アメッシュ画像を取ってきて画像URLをSlackへ返します
3. あとは `@amesh help` を見てください

# どうやって導入するの

このボタンを押します。
<a href="https://slack.com/oauth/v2/authorize?client_id=2752107225.1411158530390&scope=app_mentions:read,chat:write,channels:read&user_scope="><img alt="Add to Slack" height="40" width="139" src="https://platform.slack-edge.com/img/add_to_slack.png" srcSet="https://platform.slack-edge.com/img/add_to_slack.png 1x, https://platform.slack-edge.com/img/add_to_slack@2x.png 2x" /></a> そしてこのボタンを押します。
<a href="https://github.com/sponsors/otiai10"><img width="145" alt="sponsor" src="https://user-images.githubusercontent.com/931554/119762444-b4cdac00-bee8-11eb-8eb9-88ba1b0211c4.png"></a>

# 問い合わせ

- https://github.com/otiai10/amesh-bot/issues
