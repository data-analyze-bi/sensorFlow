# Docker 部署

此环境面向已购买授权的 SensorFlow 用户，会启动 Go 接收服务、Redis、ClickHouse 和 Apache Superset。

启动前请把交付的可执行 decoder 安装到 `../../binaries/licenses/current/decoder`，并把验证/许可证文件放在同一版本目录。许可证目录以只读方式挂载。

```bash
test -x ../../binaries/licenses/current/decoder
cp .env.example .env
# 在 .env 中设置自己的 SUPERSET_SECRET_KEY 和 SUPERSET_ADMIN_PASSWORD。
docker compose up -d --build
docker compose ps
docker compose logs ingestion
```

项目不定义客户密码。`REDIS_PASSWORD` 和 `CLICKHOUSE_PASSWORD` 为空时，内置本地数据库以无认证模式启动；设置非空值后会启用认证，并把相同值传给接收服务。启用 ClickHouse 密码时，还需设置包含相同 URL 编码凭证的 `CLICKHOUSE_SQLALCHEMY_URI` 供 Superset 使用。

可覆盖变量包括 `REDIS_HOST`、`REDIS_TYPE`、`REDIS_PASSWORD`、`CLICKHOUSE_HOST`、`CLICKHOUSE_USER`、`CLICKHOUSE_PASSWORD`、`CLICKHOUSE_DB`、`SUPERSET_SECRET_KEY`、`SUPERSET_ADMIN_PASSWORD`、`CLICKHOUSE_SQLALCHEMY_URI` 以及公开端口变量。

所有公开端口默认只绑定 `127.0.0.1`。共享或生产部署前必须设置自己的凭证、配置 TLS 与备份，并确保 decoder/license 文件不进入 Git。

停止服务但保留 named volumes：

```bash
docker compose down
```
