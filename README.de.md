# SensorFlow

Eine selbst gehostete Ereignisanalyse vom Tracking bis zur Visualisierung.

**Offizielle Sensors Data SDKs -> Ingestion-Service -> ClickHouse -> Apache Superset**

Sprachen: [English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Français](README.fr.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

> **Hinweis zu Drittanbieter-SDKs:** Dieses Repository vertreibt keine Sensors Data SDKs. SensorFlow ist unabhängig und weder verbunden noch unterstützt oder zertifiziert von Sensors Data. Beziehen Sie SDKs aus offiziellen Quellen und beachten Sie deren Lizenzen. Siehe [vollständiger Hinweis](THIRD_PARTY_NOTICES.md).

## Start

```bash
cd deploy/docker
docker compose up -d --build
cd ../..
go mod download
go run main.go
```

Der Ingestion-Service läuft unter `http://127.0.0.1:8081`, Superset unter `http://127.0.0.1:8088`.

## SDK-Anbindung

Dieses Repository veröffentlicht kein eigenes Client-SDK. Verwenden Sie die offiziellen Sensors Data SDKs für Android, iOS, Web, Mini-Programme und Server und setzen Sie `serverUrl` auf:

```text
https://your-domain.example/sensors/send/?token=YOUR_TOKEN
```

Ereignisse werden in `sensors.event` in ClickHouse gespeichert und können in Superset SQL Lab sowie im Ereignis-Dashboard analysiert werden. Verwenden Sie in Produktion HTTPS, starke Passwörter und einen eigenen `SUPERSET_SECRET_KEY`.
