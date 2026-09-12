# SensorFlow

Uma plataforma auto-hospedada de análise de eventos, da coleta à visualização.

**SDKs oficiais do Sensors Data -> serviço de ingestão -> ClickHouse -> Apache Superset**

Idiomas: [English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Deutsch](README.de.md) | [Français](README.fr.md) | [Español](README.es.md)

> **Aviso sobre SDKs de terceiros:** Este repositório não distribui SDKs do Sensors Data. O SensorFlow é independente e não é afiliado, endossado ou certificado pelo Sensors Data. Obtenha os SDKs em fontes oficiais e cumpra suas licenças. Consulte o [aviso completo](THIRD_PARTY_NOTICES.md).

## Início

```bash
cd deploy/docker
docker compose up -d --build
cd ../..
go mod download
go run main.go
```

O serviço de ingestão funciona em `http://127.0.0.1:8081` e o Superset em `http://127.0.0.1:8088`.

## Integração dos SDKs

Este repositório não distribui um SDK cliente próprio. Use os SDKs oficiais do Sensors Data para Android, iOS, Web, miniprogramas e servidores, configurando `serverUrl` como:

```text
https://your-domain.example/sensors/send/?token=YOUR_TOKEN
```

Os eventos são armazenados em `sensors.event` no ClickHouse e podem ser consultados no SQL Lab e no painel do Superset. Em produção, use HTTPS, senhas fortes e um `SUPERSET_SECRET_KEY` exclusivo.
