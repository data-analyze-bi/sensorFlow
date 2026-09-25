# Docker 一键部署

前置条件：Linux/macOS 上可用的 Docker Engine（或 Docker Desktop）和 Docker Compose。无需在主机安装 Go、Python、nc 或 openssl。

在仓库根目录运行：

```bash
./install.sh
```

如果已进入 `deploy/docker`，同样可以运行 `./install.sh`。不要把首次安装替换为 `docker compose up`：Compose 本身不会生成密钥，安装器会先补齐配置再启动服务。

安装器自动生成缺失的随机密钥和管理员密码，将配置保存到权限为 `600` 的 `deploy/docker/.env`，启动 Redis、ClickHouse 和 Superset，导入 `demo_*` 演示数据并创建图表。等待 Superset 健康检查通过后，输出可访问的看板地址，并在终端用**红色加粗**显示管理员账号和密码。设置 `NO_COLOR=1` 可关闭颜色。

重复执行会保留已有凭证、自定义配置和数据卷，可用于安装失败后的重试。已有 Redis 或 ClickHouse 地址保存在 `.env` 时会继续复用；首次安装检测到本机服务时可交互选择复用。

## 通过服务器 IP 查看图表

Superset 默认监听 `0.0.0.0:8088`。通过 SSH 安装时，安装器使用 SSH 会话中的服务器地址生成访问链接，例如：

```text
http://169.58.33.188:8088
http://169.58.33.188:8088/superset/dashboard/sensorflow-events-overview/
```

在浏览器登录后即可查看图表。允许服务器及云防火墙的入站 TCP 8088。Redis、ClickHouse 和尚未激活的 ingestion 不对公网开放；8081 是数据接收端口，不是图表页面。

没有 SSH 会话或服务器位于 NAT 后时，可在首次安装时指定地址：

```bash
SENSORFLOW_PUBLIC_HOST=你的服务器IP ./install.sh
```

已有安装可在 `deploy/docker/.env` 调整 `SUPERSET_PUBLIC_BASE_URL`、`SUPERSET_PORT`、`SUPERSET_BIND_ADDRESS`，然后在仓库根目录重新运行 `./install.sh`。仅需本机访问时设置 `SUPERSET_BIND_ADDRESS=127.0.0.1`。使用 IP 直连时为 HTTP；绑定域名后可启用 HTTPS。

## 绑定域名

建议为图表和埋点接口使用两个独立子域名。先将两个域名的 DNS A 记录都指向服务器公网 IP，仅在 IPv6 可用时添加 AAAA，并允许入站 TCP 80/443。

### 1. 给 Superset 绑定域名

修改 `deploy/docker/.env`：

```dotenv
SUPERSET_DOMAIN=analytics.example.com
SUPERSET_PUBLIC_BASE_URL=https://analytics.example.com/
SUPERSET_BIND_ADDRESS=127.0.0.1
```

浏览器访问 `https://analytics.example.com`，使用安装器输出的账号密码登录。这一域名只提供 Superset 页面。

### 2. 给 SensorFlow 绑定埋点域名

修改 `deploy/docker/.env`：

```dotenv
SENSORFLOW_DOMAIN=track.example.com
```

取得许可证后在仓库根目录运行 `./activate.sh`。SDK 上报地址为：

```text
https://track.example.com/sensors/send/?token=安装器生成的Token
```

这一域名只转发 `/sensors/*`，不会暴露 Superset。可以只配置其中一个域名；同时配置时必须使用两个不同的域名。

配置完成后，在 `deploy/docker` 目录执行：

```bash
docker compose up -d superset
docker compose --profile edge up -d caddy
docker compose logs --tail=50 caddy
```

Caddy 会自动申请和续期 HTTPS 证书。

## 激活真实埋点

第一阶段无需许可证，不启动真实数据接收服务。查看图表后，到 `sensorflow.site` 下载许可证，在仓库根目录运行：

```bash
./activate.sh
```

已有 `SENSORFLOW_DOMAIN` 时会复用埋点域名，SDK 接口为 `https://你的埋点域名/sensors/send/?token=你的Token`。

## 运维

在 `deploy/docker` 目录查看状态与日志：

```bash
docker compose ps
docker compose logs --tail=100 superset
```

停止服务但保留数据卷：

```bash
docker compose down
```

不要提交 `.env` 或许可证文件。备份数据库时同时保存 `.env`，尤其不要丢失或随意轮换 `SUPERSET_SECRET_KEY`。
