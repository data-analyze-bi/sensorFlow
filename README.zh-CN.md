<p align="center">
  <img src="assets/logo-sensorflow-zh-CN.svg" alt="SensorFlow — 自托管事件分析" width="720">
</p>

# SensorFlow

一套从采集到可视化的自托管埋点分析方案：

**神策各端 SDK -> 数据接收服务 -> ClickHouse -> Apache Superset**

![SensorFlow 运营概览](assets/operations-overview.svg)

核心优势：部署简单、服务器资源占用低、价格可预期，并兼容成熟的神策 SDK 采集生态与 Apache Superset 开放 BI 能力。

语言：[English](README.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Deutsch](README.de.md) | [Français](README.fr.md) | [Español](README.es.md) | [Português](README.pt-BR.md)

> **第三方 SDK 说明：** 本仓库不分发神策 SDK。SensorFlow 是独立项目，与神策数据不存在关联、背书或官方认证关系。请从神策官方渠道获取 SDK，并遵守对应许可及使用条款。详见[第三方 SDK 与商标声明](THIRD_PARTY_NOTICES.md#中文说明)。

## 组成

- 神策官方 SDK：采集 Android、iOS、Web、小程序及服务端事件。
- Go 接收服务：校验并处理上报事件。
- ClickHouse：保存事件和用户数据，执行分析查询。
- Apache Superset：提供 SQL 查询、图表与看板。
- Redis：提供接收链路所需的缓存能力。

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
