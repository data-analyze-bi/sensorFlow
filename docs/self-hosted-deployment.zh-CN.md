# 自托管用户行为分析部署

SensorFlow 由 Go 数据接收服务、ClickHouse 分析数据库、Redis 缓存和 Apache Superset 数据看板组成，适合在企业自己的服务器或云基础设施中部署。

## 安装流程

1. 安装 Docker Engine 和 Docker Compose v2。
2. 在仓库根目录运行 `./install.sh`。
3. 允许 TCP 8088 后，打开安装器输出的 `http://服务器IP:8088`，使用终端红色显示的 Superset 账号密码查看演示看板。
4. 从官网取得许可证后运行 `./activate.sh`，启动真实事件接收服务。
5. 将 SDK 上报地址配置为激活器输出的地址，发送一个 `integration_test` 事件并在 Superset/ClickHouse 中验证。
6. 需要域名和 HTTPS 时，再按 [Docker 部署文档](../deploy/docker/README.zh-CN.md#绑定域名)配置 Caddy。Caddy 是可选代理组件，不是 SensorFlow 的产品名称。

## 生产环境检查清单

- 修改数据库、Redis 和 Superset 的默认密码。
- 设置高强度的 `SUPERSET_SECRET_KEY`。
- 通过 HTTPS 和反向代理暴露数据接收接口。
- 限制 ClickHouse、Redis 和 Superset 端口的访问范围。
- 定期备份 ClickHouse 数据并测试恢复流程。
- 不要提交密码、Token、客户数据或授权文件。

详细配置请参考 [Docker 部署文档](../deploy/docker/README.zh-CN.md)。
