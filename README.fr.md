# SensorFlow

Une plateforme auto-hébergée d'analyse des événements, de la collecte à la visualisation.

**SDK officiels Sensors Data -> service d'ingestion -> ClickHouse -> Apache Superset**

Langues : [English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Deutsch](README.de.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

> **Avis SDK tiers :** Ce dépôt ne distribue aucun SDK Sensors Data. SensorFlow est un projet indépendant, sans affiliation, approbation ni certification de Sensors Data. Obtenez les SDK auprès des sources officielles et respectez leurs licences. Voir [l'avis complet](THIRD_PARTY_NOTICES.md).

## Démarrage

```bash
cd deploy/docker
docker compose up -d --build
cd ../..
go mod download
go run main.go
```

Le service d'ingestion écoute sur `http://127.0.0.1:8081` et Superset sur `http://127.0.0.1:8088`.

## Connexion des SDK

Ce dépôt ne distribue pas de SDK client personnalisé. Utilisez les SDK officiels Sensors Data pour Android, iOS, Web, mini-programmes et serveurs, puis configurez `serverUrl` :

```text
https://your-domain.example/sensors/send/?token=YOUR_TOKEN
```

Les événements sont stockés dans `sensors.event` sur ClickHouse et consultables dans SQL Lab et le tableau de bord Superset. En production, activez HTTPS et remplacez tous les mots de passe ainsi que `SUPERSET_SECRET_KEY`.
