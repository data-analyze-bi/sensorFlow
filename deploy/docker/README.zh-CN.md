# Docker 部署

此环境面向已购买授权的 SensorFlow 用户，会启动 Go 接收服务、Redis、ClickHouse 和 Apache Superset。

启动前请把交付的可执行 decoder 安装到 `../../binaries/licenses/current/decoder`，并把验证/许可证文件放在同一版本目录。许可证目录以只读方式挂载。

```bash
test -x ../../binaries/licenses/current/decoder
docker compose up -d --build
docker compose ps
docker compose logs ingestion
```

可覆盖的变量包括 `REDIS_PASSWORD`、`CLICKHOUSE_PASSWORD`、`SUPERSET_SECRET_KEY`、`SUPERSET_ADMIN_PASSWORD`、`CLICKHOUSE_SQLALCHEMY_URI`、`SENSORFLOW_PORT`、`CLICKHOUSE_HTTP_PORT`、`CLICKHOUSE_NATIVE_PORT`、`REDIS_PORT` 和 `SUPERSET_PORT`。

所有公开端口默认只绑定 `127.0.0.1`。共享或生产部署前必须替换全部默认凭证、配置 TLS 与备份，并确保 decoder/license 文件不进入 Git。

停止服务但保留 named volumes：

```bash
docker compose down
```
