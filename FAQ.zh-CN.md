# 常见问题

## 客户端应该使用哪个 SDK？

Android、iOS、Web、小程序及服务端统一使用神策官方 SDK。本仓库暂不发布自研客户端 SDK。

## SDK 应把事件发送到哪里？

将 `serverUrl` 设置为 `https://your-domain.example/sensors/send/?token=YOUR_TOKEN`。生产环境使用 HTTPS，不要把 Token 写入公开源码。

## 数据保存在哪里？

事件保存在 ClickHouse 的 `sensors.event` 表，用户属性记录保存在 `sensors.user` 表。

## 如何确认数据链路正常？

先发送 `integration_test` 事件，再分别通过 ClickHouse 和 Superset SQL Lab 查询。两处都有数据即表示链路正常。

## 请求为什么被拒绝？

检查许可证文件、请求地址及服务日志。许可证内置授权校验失败时会拒绝解析。

## 可以直接用于生产吗？

上线前必须配置 HTTPS、替换所有默认密码、设置强 `SUPERSET_SECRET_KEY`、限制数据库端口、备份数据卷并配置监控。
