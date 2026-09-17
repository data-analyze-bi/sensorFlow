<p align="center">
  <img src="assets/logo-sensorflow-zh-CN.svg" alt="SensorFlow — 自托管事件分析" width="720">
</p>

# SensorFlow

SensorFlow 是一个开源、自托管的用户行为分析与产品数据分析平台，用于采集、存储、查询和可视化用户事件。

它适合希望自主掌控数据、降低 SaaS 分析成本，并在自己的服务器上运行埋点分析系统的团队。

> **Superset 测试看板：** [https://superset.sensorflow.site/](https://superset.sensorflow.site/)

**神策各端 SDK -> 数据接收服务 -> ClickHouse -> Apache Superset**

![SensorFlow 运营概览](assets/operations-overview.svg)

核心优势：部署简单、服务器资源占用低、价格可预期，并兼容成熟的神策 SDK 采集生态与 Apache Superset 开放 BI 能力。

关键词：埋点、埋点分析、开源埋点平台、自托管埋点系统、用户行为埋点、用户行为分析、产品数据分析、事件分析平台、ClickHouse 埋点、神策 SDK 兼容、Apache Superset 数据看板。

如果你正在搜索“埋点平台”“埋点分析工具”或“开源用户行为分析系统”，SensorFlow 提供了一套基于 Go、ClickHouse 和 Apache Superset 的自托管方案。

语言：[English](README.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Deutsch](README.de.md) | [Français](README.fr.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

> **第三方 SDK 说明：** 本仓库不分发神策 SDK。SensorFlow 是独立项目，与神策数据不存在关联、背书或官方认证关系。请从神策官方渠道获取 SDK，并遵守对应许可及使用条款。详见[第三方 SDK 与商标声明](THIRD_PARTY_NOTICES.md#中文说明)。

## 组成

- 神策官方 SDK：采集 Android、iOS、Web、小程序及服务端事件。
- Go 接收服务：校验并处理上报事件。
- ClickHouse：保存事件和用户数据，执行分析查询。
- Apache Superset：提供 SQL 查询、图表与看板。
- Redis：提供接收链路所需的缓存能力。

## 使用场景

SensorFlow 适用于：

- Web、Android、iOS、小程序和服务端的用户行为埋点
- 产品数据分析、事件分析和用户行为分析
- 对数据隐私、数据驻留和基础设施自主权有要求的团队
- 基于 ClickHouse 的高性能事件数据查询
- 兼容神策 SDK 的自托管数据接收
- 使用 Apache Superset 构建内部数据分析看板
- 替代 Amplitude、Mixpanel 等托管式产品分析工具

## 为什么选择 SensorFlow

- **开源可控：** 可以查看代码、调整数据处理逻辑并自行部署。
- **完全自托管：** 用户事件和用户数据保存在自己的基础设施中。
- **部署简单：** 使用 Docker Compose 启动 ClickHouse、Redis 和 Superset。
- **查询灵活：** 直接使用 ClickHouse SQL 分析事件数据。
- **SDK 兼容：** 可将已有神策 SDK 应用接入 SensorFlow 数据接收服务。

## 中文文档

- [快速开始](docs/getting-started.zh-CN.md)
- [自托管部署指南](docs/self-hosted-deployment.zh-CN.md)
- [产品分析与用户行为埋点](docs/product-analytics.zh-CN.md)
- [ClickHouse 与 Superset 数据分析](docs/clickhouse-superset.zh-CN.md)
- [神策 SDK 接入](docs/sensors-sdk.zh-CN.md)
- [Amplitude 和 Mixpanel 开源替代方案](docs/alternatives.zh-CN.md)

## 埋点与用户行为分析指南

- [GitHub 搜索埋点：如何筛选开源数据埋点项目](https://sensorflow.site/resources/guides/github-search-event-tracking)
- [神策 SDK 数据迁移到自托管 ClickHouse](https://sensorflow.site/use-cases/sensors-sdk-to-clickhouse)
- [2026 开源埋点平台对比：SensorFlow、PostHog 与 Matomo](https://sensorflow.site/resources/comparisons/open-source-tracking-platforms)
- [数据埋点是什么：事件模型、采集方式与质量检查](https://sensorflow.site/resources/learn/what-is-event-tracking)
- [使用 ClickHouse 与 Superset 搭建用户行为分析](https://sensorflow.site/use-cases/clickhouse-superset-analytics)

## 启动数据基础设施

安装 Docker Engine 或 Docker Desktop 与 Docker Compose，然后运行：

```bash
cd deploy/docker
docker compose up -d --build
docker compose ps
```

Compose 仅启动基础设施，默认地址为 ClickHouse `8123/9000`、Redis `6379`、Superset `8088`。Go 接收服务需按下一节单独启动，默认端口为 `8081`。

启动 Compose 前必须通过环境变量设置数据库密码、Redis 密码、Superset 管理员密码和 `SUPERSET_SECRET_KEY`，具体见 Docker 部署文档。

## 启动接收服务

需要 Go 1.17 或更高版本。先从 SensorFlow 运营网站的控制台生成并下载许可证文件，然后在本仓库根目录执行：

```bash
mkdir -p binaries/licenses/manual
mv ~/Downloads/sensors-payload-decoder-* binaries/licenses/manual/decoder
mv ~/Downloads/*.verify.json binaries/licenses/manual/decoder.verify.json
ln -sfn manual binaries/licenses/current
chmod +x binaries/licenses/manual/decoder
```

确认 `conf/app.conf` 使用以下路径：

```ini
decoder_binary_path = ./binaries/licenses/current/decoder
```

接着检查 Redis 和 ClickHouse 配置并启动接收服务：

```bash
go mod download
go run main.go
```

更新许可证时，将运营后台下载的新许可证原子替换到 `binaries/licenses/current/decoder`，重新授予执行权限。不要将许可证文件提交到 Git。

## 接入应用

本仓库不发布自研客户端 SDK。请使用对应平台的神策官方 SDK，并将 `serverUrl` 设置为：

```text
https://your-domain.example/sensors/send/?token=YOUR_TOKEN
```

详细步骤见[神策 SDK 接入](docs/sensors-sdk.zh-CN.md)。

## 授权与商业服务

授权、托管运营和商业方案请访问 [sensorflow.site/plans](https://sensorflow.site/plans)，部署与账号文档请访问 [sensorflow.site/docs](https://sensorflow.site/docs)。

## 在 Superset 查询

打开 `http://127.0.0.1:8088`，在 SQL Lab 中选择 ClickHouse 数据库并执行：

```sql
SELECT event, count() AS event_count
FROM sensors.event
GROUP BY event
ORDER BY event_count DESC
LIMIT 20;
```

部署脚本还会创建埋点数据总览看板。常见问题见[中文 FAQ](FAQ.zh-CN.md)，基础设施说明见[Docker 部署](deploy/docker/README.zh-CN.md)。
