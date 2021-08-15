# Bot

どのようなコマンドを登録されているかに関わらず、基本的なbotの動きとして、`Handle`を提供する。
直接、WebhookのControllerから呼ばれれる対象。
もちろん、Controllersも、Bot自身も、どのようなコマンドが登録されているかは関知しない。
くわしくは、[main.go](https://github.com/otiai10/amesh-bot/blob/main/main.go)において`Commands`を登録している部分を参照のこと。
