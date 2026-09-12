# SensorFlow

이벤트 수집부터 시각화까지 제공하는 셀프 호스팅 분석 시스템입니다.

**Sensors Data 공식 SDK -> 수집 서비스 -> ClickHouse -> Apache Superset**

언어: [English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [Deutsch](README.de.md) | [Français](README.fr.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

> **타사 SDK 고지:** 이 저장소는 Sensors Data SDK를 배포하지 않습니다. SensorFlow는 독립 프로젝트이며 Sensors Data와 제휴, 보증 또는 공식 인증 관계가 없습니다. 공식 경로에서 SDK를 받고 해당 라이선스를 준수하십시오. [전체 고지](THIRD_PARTY_NOTICES.md)

## 시작

```bash
cd deploy/docker
docker compose up -d --build
cd ../..
go mod download
go run main.go
```

수집 서비스는 `http://127.0.0.1:8081`, Superset은 `http://127.0.0.1:8088`에서 실행됩니다.

## SDK 연결

이 저장소는 자체 클라이언트 SDK를 배포하지 않습니다. Android, iOS, Web, 미니 프로그램 및 서버에서는 Sensors Data 공식 SDK를 사용하고 `serverUrl`을 다음과 같이 설정합니다.

```text
https://your-domain.example/sensors/send/?token=YOUR_TOKEN
```

이벤트는 ClickHouse의 `sensors.event`에 저장되며 Superset SQL Lab과 이벤트 대시보드에서 조회할 수 있습니다. 운영 환경에서는 HTTPS, 강력한 비밀번호와 별도의 `SUPERSET_SECRET_KEY`를 사용하십시오.
