# Docker 部署

此环境面向已购买授权的 SensorFlow 用户，会启动 Go 接收服务、Redis、ClickHouse 和 Apache Superset。

启动前请把交付的可执行 decoder 安装到 `../../binaries/licenses/current/decoder`，并把验证/许可证文件放在同一版本目录。许可证目录以只读方式挂载。

在仓库根目录运行 `./install.sh`。安装器会检测已有 Redis 与 ClickHouse，仅在可能复用已有服务时询问连接信息，自动生成缺失凭证、写入私有 `.env` 并启动所需 Compose 服务。

```bash
cd ../..
./install.sh
```

全新机器由安装器生成 Redis、ClickHouse 和 Superset 凭证；复用已有 Redis 或 ClickHouse 时，在安装过程中收集凭证，并统一传递给接收服务和 Superset。

可覆盖变量包括 `REDIS_HOST`、`REDIS_TYPE`、`REDIS_PASSWORD`、`CLICKHOUSE_HOST`、`CLICKHOUSE_USER`、`CLICKHOUSE_PASSWORD`、`CLICKHOUSE_DB`、`SUPERSET_SECRET_KEY`、`SUPERSET_ADMIN_PASSWORD`、`CLICKHOUSE_SQLALCHEMY_URI` 以及公开端口变量。

所有公开端口默认只绑定 `127.0.0.1`。共享或生产部署前必须设置自己的凭证、配置 TLS 与备份，并确保 decoder/license 文件不进入 Git。

停止服务但保留 named volumes：

```bash
docker compose down
```
