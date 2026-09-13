# 使用 ClickHouse 和 Apache Superset 分析埋点数据

ClickHouse 是 SensorFlow 的事件分析存储引擎，Apache Superset 用于 SQL 查询、图表制作和数据看板展示。

在线体验：[查看 SensorFlow Superset 测试看板](https://superset.sensorflow.site/)。

示例查询：

```sql
SELECT event, count() AS event_count
FROM sensors.event
GROUP BY event
ORDER BY event_count DESC
LIMIT 20;
```

启动 Docker Compose 后访问 `http://127.0.0.1:8088`，在 Superset SQL Lab 中选择 ClickHouse 数据库，即可查询埋点数据并保存为图表或看板。
