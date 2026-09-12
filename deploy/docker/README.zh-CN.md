# Docker 部署

此环境包含 MySQL、Redis、ClickHouse 和 Apache Superset。

此 Compose 仅启动基础设施；Go 数据接收服务需要在仓库根目录单独构建并运行。

首次启动接收服务前，请先在 SensorFlow 运营网站生成并下载许可证包，然后安装到当前版本目录：

```bash
mkdir -p binaries/licenses/manual
mv ~/Downloads/sensors-payload-decoder-* binaries/licenses/manual/decoder
mv ~/Downloads/*.verify.json binaries/licenses/manual/decoder.verify.json
ln -sfn manual binaries/licenses/current
chmod +x binaries/licenses/manual/decoder
```

确认 `conf/app.conf` 中 `decoder_binary_path = ./binaries/licenses/current/decoder`。更新许可证时，将新文件原子替换到该路径并重新授予执行权限。

```bash
export MYSQL_ROOT_PASSWORD='replace-with-strong-password'
export MYSQL_PASSWORD='replace-with-strong-password'
export REDIS_PASSWORD='replace-with-strong-password'
export CLICKHOUSE_PASSWORD='replace-with-strong-password'
export SUPERSET_SECRET_KEY='replace-with-long-random-secret'
export SUPERSET_ADMIN_PASSWORD='replace-with-strong-password'
export CLICKHOUSE_SQLALCHEMY_URI='clickhousedb://default:replace-with-url-encoded-password@clickhouse:8123/sensors'
docker compose up -d --build
docker compose ps
```

等待所有服务健康后，打开 `http://127.0.0.1:8088`，使用上面配置的管理员密码登录。

启动过程会初始化 Superset、创建管理员、配置 ClickHouse，并创建 `sensors.event` 数据集和埋点数据总览看板。

生产部署前必须通过环境变量覆盖 `SUPERSET_SECRET_KEY`、`SUPERSET_ADMIN_PASSWORD` 和数据库密码，同时限制数据库端口并配置备份。

停止服务但保留数据：

```bash
docker compose down
```
