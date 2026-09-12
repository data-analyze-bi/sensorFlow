# 自托管用户行为分析部署

SensorFlow 由 Go 数据接收服务、ClickHouse 分析数据库、Redis 缓存和 Apache Superset 数据看板组成，适合在企业自己的服务器或云基础设施中部署。

## 生产环境检查清单

- 修改数据库、Redis 和 Superset 的默认密码。
- 设置高强度的 `SUPERSET_SECRET_KEY`。
- 通过 HTTPS 和反向代理暴露数据接收接口。
- 限制 ClickHouse、Redis 和 Superset 端口的访问范围。
- 定期备份 ClickHouse 数据并测试恢复流程。
- 不要提交密码、Token、客户数据或授权文件。

详细配置请参考 [Docker 部署文档](../deploy/docker/README.zh-CN.md)。
