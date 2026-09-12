# ClickHouse Analytics with Apache Superset

ClickHouse is SensorFlow's analytical storage engine. Apache Superset connects to ClickHouse for SQL exploration, charts, and dashboards.

Example query:

```sql
SELECT event, count() AS event_count
FROM sensors.event
GROUP BY event
ORDER BY event_count DESC
LIMIT 20;
```

Open Superset at `http://127.0.0.1:8088` after starting Docker Compose. Use the configured ClickHouse connection in SQL Lab, then save useful queries as charts or dashboards.
