# SensorFlow 快速开始

SensorFlow 是开源、自托管的产品数据分析和用户行为分析平台。本指南介绍从安装 Docker 到查看 Superset 演示看板，再到激活真实事件接收的完整流程。

## 环境要求

- Docker Engine 或 Docker Desktop，以及 Docker Compose
- 可从浏览器访问的服务器公网 IP（查看看板需要开放 TCP 8088）

## 一键安装并查看演示看板

```bash
git clone https://github.com/data-analyze-bi/sensorFlow.git
cd sensorFlow
./install.sh
```

安装器自动生成缺失凭证并保存在 `deploy/docker/.env`，完成后用红色显示 Superset 管理员账号和密码。打开安装器输出的 `http://服务器IP:8088` 查看演示看板。Caddy 是可选的域名/HTTPS 代理，第一阶段不需要启动。

## 下一步：激活真实事件接收

查看演示确认部署正常后，从官网取得许可证，在仓库根目录执行：

```bash
./activate.sh
```

激活后再按照[神策 SDK 接入](sensors-sdk.zh-CN.md)配置上报地址并发送 `integration_test` 事件验证。

## 可选：绑定域名

将域名 DNS 指向服务器后，参考 [Docker 部署文档中的域名和 HTTPS](../deploy/docker/README.zh-CN.md#绑定域名)。Caddy 只负责反向代理和证书，产品名称仍是 SensorFlow。
