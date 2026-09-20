<p align="center">
  <img src="assets/logo-sensorflow-zh-CN.svg" alt="SensorFlow — 开源自托管事件分析" width="720">
</p>

# SensorFlow

**面向已购买 SensorFlow 授权用户的开源、自托管事件分析平台：保留现有神策 SDK 埋点，将数据存入自己的 ClickHouse。**

```text
神策各端 SDK ──> Go 数据接收 ──> ClickHouse ──> Apache Superset
                          └─────> Redis
```

SensorFlow 是独立项目，与神策数据不存在关联、背书或官方认证关系。本仓库不分发神策官方 SDK。

[English](README.md) · [中文文档](docs/getting-started.zh-CN.md) · [官网](https://sensorflow.site/) · [Apache-2.0 许可证](LICENSE)

## 付费用户部署

前置条件：

- Docker Engine 或 Docker Desktop 与 Docker Compose v2
- 购买后获得的有效 SensorFlow payload decoder 和验证/许可证文件
- 业务应用中已配置的神策官方 SDK

克隆仓库，然后把交付文件安装到版本目录：

```bash
git clone https://github.com/data-analyze-bi/sensorFlow.git
cd sensorFlow
mkdir -p binaries/licenses/customer
mv ~/Downloads/sensors-payload-decoder-* binaries/licenses/customer/decoder
mv ~/Downloads/*.verify.json binaries/licenses/customer/decoder.verify.json
chmod +x binaries/licenses/customer/decoder
ln -sfn customer binaries/licenses/current
```

不要提交 decoder 或许可证文件。Compose 会把 `binaries/licenses` 以只读方式挂载到接收服务。启动前检查可执行文件：

```bash
test -x binaries/licenses/current/decoder
```

启动完整环境：

```bash
cd deploy/docker
cp .env.example .env
# 编辑 .env，设置自己的 SUPERSET_SECRET_KEY 和 SUPERSET_ADMIN_PASSWORD。
docker compose up -d --build
docker compose ps
```

Compose 会启动 Go 接收服务、Redis、ClickHouse 和 Apache Superset。默认端口只绑定 `127.0.0.1`：

- 接收服务：`127.0.0.1:8081`
- ClickHouse HTTP/native：`127.0.0.1:8123` / `127.0.0.1:9000`
- Redis：`127.0.0.1:6379`
- Superset：`127.0.0.1:8088`

## 接入神策 SDK

保留现有 SDK 埋点代码，只需把上报地址指向 SensorFlow：

```text
https://your-sensorflow.example/sensors/send/?token=YOUR_TOKEN
```

从真实客户端发送唯一命名的 `integration_test` 事件，然后验证真实 ClickHouse 数据：

```bash
cd deploy/docker
docker compose exec -T clickhouse clickhouse-client \
  ${CLICKHOUSE_PASSWORD:+--password "$CLICKHOUSE_PASSWORD"} \
  --query "SELECT time, event, distinct_id FROM sensors.event WHERE event = 'integration_test' ORDER BY time DESC LIMIT 10"
```

Superset 地址为 `http://127.0.0.1:8088`。启动前由用户设置 `SUPERSET_ADMIN_PASSWORD`，项目不定义用户密码。

## 生产要求

项目不定义 Redis、ClickHouse、MySQL 或客户账号密码。仅当密码变量为空时，Redis 与 ClickHouse 才以无密码模式启动，并且端口仍只绑定 `127.0.0.1`。生产部署前必须：

- 设置 `REDIS_PASSWORD`、`CLICKHOUSE_PASSWORD`、`SUPERSET_ADMIN_PASSWORD`、`SUPERSET_SECRET_KEY`，以及密码经过 URL 编码且保持一致的 `CLICKHOUSE_SQLALCHEMY_URI`。
- 通过反向代理提供 TLS，仅暴露必要的接收路径。
- 保持数据库端口不公开，配置备份，并监控 decoder 失败、接收延迟和 ClickHouse 磁盘水位。
- 根据发布策略固定并审核容器镜像版本。
- 切换生产流量前验证 SDK 版本、加密插件、用户关联、属性类型和事件时间。

继续阅读 [Docker 部署](deploy/docker/README.zh-CN.md)、[神策 SDK 接入](docs/sensors-sdk.zh-CN.md)、[中文 FAQ](FAQ.zh-CN.md)和[第三方声明](THIRD_PARTY_NOTICES.md#中文说明)。

## 开发验证

```bash
go test ./...
go build -o sensors main.go
docker compose -f deploy/docker/docker-compose.yml config
```

贡献说明见 [CONTRIBUTING.md](CONTRIBUTING.md)。不要提交客户数据、密码、decoder 二进制或许可证文件。
