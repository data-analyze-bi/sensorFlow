# ClickHouse Analytics with Apache Superset

ClickHouse is SensorFlow's analytical storage engine. Apache Superset connects to ClickHouse for SQL exploration, charts, and dashboards.

Try the hosted demo: [View the SensorFlow Superset demo dashboard](https://superset.sensorflow.site/).

Example query:

```sql
SELECT event, count() AS event_count
FROM sensors.event
GROUP BY event
ORDER BY event_count DESC
LIMIT 20;
```

Run `./install.sh`, then open the server-IP URL printed by the installer. Use the configured ClickHouse connection in SQL Lab, then save useful queries as charts or dashboards.
