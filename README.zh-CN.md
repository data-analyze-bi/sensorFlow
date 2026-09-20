<p align="center">
  <img src="assets/logo-sensorflow-zh-CN.svg" alt="SensorFlow — 开源自托管事件分析" width="720">
</p>

# SensorFlow

**开源、自托管事件分析平台：先部署并查看 Superset 演示图表，确认价值后再用许可证激活真实神策 SDK 埋点接收。**

```text
神策各端 SDK ──> Go 数据接收 ──> ClickHouse ──> Apache Superset
                          └─────> Redis
```

SensorFlow 是独立项目，与神策数据不存在关联、背书或官方认证关系。本仓库不分发神策官方 SDK。

[English](README.md) · [中文文档](docs/getting-started.zh-CN.md) · [官网](https://sensorflow.site/) · [Apache-2.0 许可证](LICENSE)

## 第一阶段：部署并查看演示

前置条件：

- Docker Engine 或 Docker Desktop 与 Docker Compose v2

克隆仓库并运行一键安装：

```bash
git clone https://github.com/data-analyze-bi/sensorFlow.git
cd sensorFlow
./install.sh
```

第一阶段不需要许可证。安装器会检测 Redis 和 ClickHouse、启动缺失服务、导入明确标记为 `demo_*` 的演示事件，并创建 Superset 数据集与看板，最后显示登录信息。演示看板包含事件数、用户数、DAU、新用户、购买人数、Demo GMV、付费转化率、小时/日趋势、核心漏斗阶段，以及渠道、页面、国家地区、操作系统、网络和版本分布。此时不会启动真实 SDK 埋点接收服务。

## 第二阶段：激活真实埋点接收

确认演示效果后，到 [sensorflow.site](https://sensorflow.site/) 下载 SensorFlow 许可证，保留在 `~/Downloads` 并执行：

```bash
./activate.sh
```

激活程序会把许可证自动安装为 `binaries/sensors-payload-decoder`，配套验证文件安装为 `binaries/sensors-payload-decoder.verify.json`，最后才启动 ingestion。已有 Redis、ClickHouse、Superset、演示数据和私有配置均不会被覆盖。

Compose 会启动 Go 接收服务、Redis、ClickHouse 和 Apache Superset。默认端口只绑定 `127.0.0.1`：

- 接收服务：`127.0.0.1:8081`
- ClickHouse HTTP/native：`127.0.0.1:8123` / `127.0.0.1:9000`
- Redis：`127.0.0.1:6379`
- Superset：`127.0.0.1:8088`

## 激活后接入神策 SDK

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

Superset 地址为 `http://127.0.0.1:8088`，安装完成后会显示自动生成的管理员初始密码。

## 生产要求

项目不定义 Redis、ClickHouse、MySQL 或客户账号密码。仅当密码变量为空时，Redis 与 ClickHouse 才以无密码模式启动，并且端口仍只绑定 `127.0.0.1`。生产部署前必须：

- 设置 `REDIS_PASSWORD`、`CLICKHOUSE_PASSWORD`、`SUPERSET_ADMIN_PASSWORD`、`SUPERSET_SECRET_KEY`，以及密码经过 URL 编码且保持一致的 `CLICKHOUSE_SQLALCHEMY_URI`。
- 通过反向代理提供 TLS，仅暴露必要的接收路径。
- 保持数据库端口不公开，配置备份，并监控许可证处理失败、接收延迟和 ClickHouse 磁盘水位。
- 根据发布策略固定并审核容器镜像版本。
- 切换生产流量前验证 SDK 版本、加密插件、用户关联、属性类型和事件时间。

继续阅读 [Docker 部署](deploy/docker/README.zh-CN.md)、[神策 SDK 接入](docs/sensors-sdk.zh-CN.md)、[中文 FAQ](FAQ.zh-CN.md)和[第三方声明](THIRD_PARTY_NOTICES.md#中文说明)。

## 开发验证

```bash
go test ./...
go build -o sensors main.go
docker compose -f deploy/docker/docker-compose.yml config
```

贡献说明见 [CONTRIBUTING.md](CONTRIBUTING.md)。不要提交客户数据、密码或许可证文件。
