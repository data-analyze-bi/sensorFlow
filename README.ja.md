# SensorFlow

イベント収集から可視化までを提供するセルフホスト型分析基盤です。

**Sensors Data 公式SDK -> 受信サービス -> ClickHouse -> Apache Superset**

言語: [English](README.md) | [简体中文](README.zh-CN.md) | [한국어](README.ko.md) | [Deutsch](README.de.md) | [Français](README.fr.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

> **第三者 SDK に関する注意:** 本リポジトリは Sensors Data SDK を配布しません。SensorFlow は独立したプロジェクトであり、Sensors Data との提携、承認、認定関係はありません。公式提供元から SDK を取得し、そのライセンスに従ってください。[詳細](THIRD_PARTY_NOTICES.md)

## 起動

```bash
cd deploy/docker
docker compose up -d --build
cd ../..
go mod download
go run main.go
```

受信サービスは `http://127.0.0.1:8081`、Superset は `http://127.0.0.1:8088` で利用できます。

## SDK連携

独自クライアントSDKは配布していません。Android、iOS、Web、ミニプログラム、サーバーでは Sensors Data の公式SDKを利用し、`serverUrl` を次のように設定します。

```text
https://your-domain.example/sensors/send/?token=YOUR_TOKEN
```

イベントは ClickHouse の `sensors.event` に保存され、Superset の SQL Lab とイベント概要ダッシュボードから確認できます。本番環境では HTTPS、強力なパスワード、独自の `SUPERSET_SECRET_KEY` を必ず設定してください。
