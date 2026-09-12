# SensorFlow 快速开始

SensorFlow 是开源、自托管的产品数据分析和用户行为分析平台。本指南介绍如何启动 ClickHouse、Redis、Apache Superset 数据基础设施，以及 Go 事件接收服务。

## 环境要求

- Docker Engine 或 Docker Desktop，以及 Docker Compose
- Go 1.17 或更高版本
- 按主 README 配置可用的授权解码器

## 启动数据基础设施

```bash
cd deploy/docker
docker compose up -d --build
docker compose ps
```

启动前请配置数据库密码、Redis 密码、Superset 管理员密码和 `SUPERSET_SECRET_KEY`。

## 启动事件接收服务

在仓库根目录执行：

```bash
go mod download
go run main.go
```

服务默认监听 `8081` 端口。接下来可参考[神策 SDK 接入](sensors-sdk.zh-CN.md)发送事件。
