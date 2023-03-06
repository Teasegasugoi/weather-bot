# weather-bot
天気情報をSlackに垂れ流すbot, Yahooの[気象情報API](https://developer.yahoo.co.jp/webapi/map/openlocalplatform/v1/weather.html)を使用しています

以下で動作確認済み
```
go version go1.19.3 darwin/arm64
```
## 準備
1. Slack側でトークンを発行
2. 気象情報APIを利用するためにClient ID を発行
3. `.env.example` を複製し, `.env`を作成
4. `.env`に必要情報を記入

トークン・Client ID発行方法は各自で調べてくださいm(_ _)m

## 使い方
```
$ go run main.go
```
実行している間,1時間に1回指定したチャンネルに気象情報を通知します