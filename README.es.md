# SensorFlow

Una plataforma autoalojada de analítica de eventos, desde la captura hasta la visualización.

**SDK oficiales de Sensors Data -> servicio de ingesta -> ClickHouse -> Apache Superset**

Idiomas: [English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Deutsch](README.de.md) | [Français](README.fr.md) | [Português](README.pt-BR.md)

> **Aviso sobre SDK de terceros:** Este repositorio no distribuye SDK de Sensors Data. SensorFlow es independiente y no está afiliado, respaldado ni certificado por Sensors Data. Obtén los SDK de fuentes oficiales y cumple sus licencias. Consulta el [aviso completo](THIRD_PARTY_NOTICES.md).

## Inicio

```bash
cd deploy/docker
docker compose up -d --build
cd ../..
go mod download
go run main.go
```

El servicio de ingesta funciona en `http://127.0.0.1:8081` y Superset en `http://127.0.0.1:8088`.

## Integración de SDK

Este repositorio no publica un SDK cliente propio. Usa los SDK oficiales de Sensors Data para Android, iOS, Web, miniprogramas y servidores, y configura `serverUrl` como:

```text
https://your-domain.example/sensors/send/?token=YOUR_TOKEN
```

Los eventos se almacenan en `sensors.event` de ClickHouse y se consultan mediante SQL Lab y el panel de Superset. En producción usa HTTPS, contraseñas seguras y un `SUPERSET_SECRET_KEY` propio.
