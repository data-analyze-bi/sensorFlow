# Product Analytics and Event Tracking

SensorFlow provides self-hosted product analytics for web, mobile, mini-program, and server applications. Existing Sensors Data SDK clients can send events to the SensorFlow ingestion endpoint without a custom client SDK.

## Event flow

1. An application records an event with its platform SDK.
2. The SDK sends the event to the SensorFlow ingestion endpoint.
3. The Go service validates and processes the payload.
4. ClickHouse stores event and user data for analytical queries.
5. Apache Superset visualizes product metrics and user behavior.

Typical analyses include event volume, unique devices, retention inputs, conversion funnels, and top events. Adapt the SQL and dashboards to your product's event schema.
