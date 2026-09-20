# Docker 部署

部署分为两个阶段。第一阶段无需许可证，启动 Redis、ClickHouse 和 Apache Superset，导入明确标记的演示事件并创建演示看板；第二阶段把购买后的许可证安装到 `binaries/sensors-payload-decoder`，最后才启动 Go 接收服务。

在仓库根目录运行 `./install.sh`。安装器会检测已有 Redis 与 ClickHouse、生成缺失凭证、写入私有 `.env`、导入演示数据并启动 Superset，不要求许可证，也不会启动 ingestion。

```bash
cd ../..
./install.sh
```

查看演示看板后，到 `sensorflow.site` 下载许可证，再激活真实埋点接收：

```bash
./activate.sh
```

全新机器由安装器生成 Redis、ClickHouse 和 Superset 凭证；复用已有 Redis 或 ClickHouse 时，在安装过程中收集凭证，先传给 Superset，激活后再传给接收服务。

可覆盖变量包括 `REDIS_HOST`、`REDIS_TYPE`、`REDIS_PASSWORD`、`CLICKHOUSE_HOST`、`CLICKHOUSE_USER`、`CLICKHOUSE_PASSWORD`、`CLICKHOUSE_DB`、`SUPERSET_SECRET_KEY`、`SUPERSET_ADMIN_PASSWORD`、`CLICKHOUSE_SQLALCHEMY_URI` 以及公开端口变量。

所有公开端口默认只绑定 `127.0.0.1`。共享或生产部署前必须设置自己的凭证、配置 TLS 与备份，并确保许可证文件不进入 Git。

停止服务但保留 named volumes：

```bash
docker compose down
```
