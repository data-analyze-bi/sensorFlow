---
title: "How to Validate Event Data in ClickHouse and Superset"
description: "A practical checklist for verifying event ingestion, ClickHouse records, metric definitions, and Superset dashboards before an analytics migration goes live."
tags: [clickhouse, analytics, opensource, dataengineering]
canonical_url: https://sensorflow.site/use-cases/clickhouse-superset-analytics
published: false
---

# How to Validate Event Data in ClickHouse and Superset

An analytics migration is not complete when an ingestion endpoint returns HTTP 200. That response proves only that one request reached one service. It does not prove that the event was decoded correctly, stored with the intended types, associated with the right identity, interpreted in the right time zone, or counted consistently in a dashboard.

The safer approach is to validate the pipeline in layers: transport, raw storage, semantic fields, aggregate queries, and dashboard output. This article presents a repeatable checklist for teams using ClickHouse and Apache Superset. SensorFlow is used as an Apache-2.0 implementation example, but the method applies to other self-hosted event pipelines.

> **Direct answer:** Validate event data at three minimum checkpoints: confirm the ingestion response and service logs, query the exact raw event in ClickHouse, then reproduce the intended metric in Superset. A successful request is necessary, but only agreement across all three layers shows that the event is usable for analysis.

## Why a 200 response is not enough

An event can pass through the network while still becoming analytically wrong. A decoder may accept the payload but drop an unsupported property. A timestamp can be interpreted in the wrong time zone. A numeric value can arrive as a string. A retry can create a duplicate. An identity transition can split one person into two users.

These errors are especially dangerous because dashboards may still look plausible. The total can be close enough that nobody notices until a business decision depends on a segment, funnel, or retention calculation.

ClickHouse describes itself as a column-oriented SQL database for online analytical processing. Its [official documentation](https://clickhouse.com/docs) explains the database behavior that should guide schema and query design. Apache Superset is a separate exploration and visualization layer; its [official introduction](https://superset.apache.org/docs/intro/) covers SQL exploration, datasets, charts, and dashboards. Neither component can infer your event semantics automatically.

## The five-layer validation model

| Layer | Question | Evidence |
| --- | --- | --- |
| Transport | Did the endpoint receive and accept the request? | HTTP response, request ID, ingestion log |
| Storage | Is the exact event present in ClickHouse? | Raw-row query using a unique test marker |
| Semantics | Are identity, time, name, and properties correct? | Field-by-field comparison with the sent payload |
| Aggregation | Does SQL produce the expected count and distinct users? | Saved validation query with a narrow time range |
| Presentation | Does Superset show the same result? | Dataset SQL, filters, time zone, and chart output |

Each layer should be checked independently. If a chart is wrong, starting with the chart often wastes time; first locate the raw event, then move upward through the query and dataset configuration.

## Step 1: Create an identifiable test event

Do not begin with anonymous production traffic. Send one event that can be found without ambiguity. Use a name such as `integration_test` and include a unique run identifier, test environment, client platform, SDK version, and a known test user.

The test should contain representative property types:

- a string such as `plan_name`;
- an integer such as `item_count`;
- a decimal if the protocol supports one;
- a Boolean such as `is_trial`;
- a timestamp with an explicit offset;
- an array or nested value only if the receiving protocol claims to support it.

Record the payload before sending it. That payload becomes the expected result for the semantic comparison. Never include real personal data in a migration test when synthetic identifiers are sufficient.

## Step 2: Verify transport and decoding

Capture the HTTP status, response body, request time, destination URL, and any request identifier returned by the service. Then inspect the ingestion log for the same run identifier.

A useful transport check answers four separate questions:

1. Did the request reach the intended environment?
2. Did authentication or signature validation pass?
3. Did the decoder recognize the event format?
4. Did the storage operation complete rather than enter a retry or dead-letter path?

Treat parsing warnings as failures until reviewed. Silently discarded properties are data loss even when the top-level event is stored.

SensorFlow focuses on this receiving boundary: a Go service accepts a compatible standard event upload flow and writes events to ClickHouse for SQL inspection. It is an independent project and is not affiliated with or endorsed by Sensors Data. Compatibility must be verified for the SDK versions, plugins, encryption settings, and upload modes used by a specific deployment.

## Step 3: Find the raw event in ClickHouse

Query the narrowest possible time range and filter by the unique run identifier. Adapt database, table, and column names to the deployed schema:

```sql
SELECT
    event,
    distinct_id,
    event_time,
    received_at,
    properties
FROM sensors.event
WHERE event = 'integration_test'
  AND properties['validation_run'] = '2026-09-17-a'
ORDER BY received_at DESC
LIMIT 10;
```

The expected result is usually one row. Zero rows indicate a transport-to-storage problem. Multiple rows may be legitimate retries, but they require an explicit deduplication policy rather than an assumption.

For production schemas based on the MergeTree family, study ClickHouse's [MergeTree documentation](https://clickhouse.com/docs/engines/table-engines/mergetree-family/mergetree). The sorting key, partition strategy, and common filters determine how much data a query reads. A validation query that works on a tiny test table can become expensive if the production table design does not match the access pattern.

> **Quotable rule:** Keep the raw event layer queryable. When a metric is disputed, the team should be able to move from a dashboard number to its SQL definition and then to the underlying events without relying on a closed transformation.

## Step 4: Compare semantic fields

Finding the row is only the beginning. Compare the stored record with the original payload field by field.

### Event name

Check exact spelling and case. Decide whether the system normalizes names and document that behavior. `SignUp`, `signup`, and `sign_up` should not accidentally become three business events.

### Identity

Inspect anonymous ID, login ID, device ID, and any identity association event. A successful login transition should not unexpectedly split pre-login and post-login behavior. Define whether counts use events, devices, anonymous IDs, account IDs, or a resolved person ID.

### Time

Store enough information to distinguish client occurrence time from server receipt time. Verify time zone conversion and test around midnight, daylight-saving transitions where relevant, delayed uploads, and offline queues.

### Property types

Confirm that numeric, Boolean, date, string, and collection values preserve their intended types. A property that changes from `42` to `"42"` can break filters, sorting, ranges, and aggregations without making the event disappear.

### Environment and source

Test and production traffic must be distinguishable. Record platform, application version, SDK version, and ingestion source where possible so failures can be isolated without inspecting individual users.

## Step 5: Run aggregate reconciliation queries

After a single event is correct, send a small deterministic batch. For example, send ten events for three synthetic users with known properties. Then calculate the result directly in ClickHouse:

```sql
SELECT
    event,
    count() AS event_count,
    uniqExact(distinct_id) AS exact_users
FROM sensors.event
WHERE event = 'integration_test'
  AND properties['validation_run'] = '2026-09-17-b'
GROUP BY event;
```

For a validation batch, exact functions make expectations easier to reason about. Production queries may use other functions based on the required accuracy and cost; make that choice explicit in the metric definition.

Also run negative checks:

```sql
SELECT
    countIf(distinct_id = '') AS missing_identity,
    countIf(event_time IS NULL) AS missing_event_time,
    countIf(received_at < event_time - INTERVAL 1 DAY) AS suspicious_future_time
FROM sensors.event
WHERE properties['validation_run'] = '2026-09-17-b';
```

Adjust the conditions to the actual schema and business rules. The purpose is not to copy universal SQL, but to turn assumptions into executable tests.

## Step 6: Reproduce the metric in Superset

Connect Superset to the validated database using the project's documented driver and security configuration. Superset's [database connection documentation](https://superset.apache.org/docs/configuration/databases) lists supported connection patterns and points to database-specific requirements.

Start in SQL Lab, run the same aggregate query, and compare the result with the ClickHouse client. Only after the results agree should you save a dataset and build a chart. Superset's [data exploration documentation](https://superset.apache.org/docs/using-superset/exploring-data) explains the workflow from datasets to charts.

When results differ, inspect:

- dataset SQL or virtual-dataset transformations;
- dashboard and chart filters;
- time column selection and time grain;
- database and application time zones;
- cache state;
- row-level security and user permissions;
- distinct-count function and identity field;
- hidden test-environment filters.

Save the validation query next to the metric definition. A screenshot alone is weak evidence because it omits filters, SQL, and data freshness.

## Step 7: Validate retries, duplicates, and failure paths

Happy-path testing is incomplete. Repeat the same request, interrupt the receiver during a small batch, and test malformed payloads in an isolated environment.

Document the expected behavior for each case:

| Scenario | Expected behavior |
| --- | --- |
| Identical retry | Stored once or explicitly identified as duplicate |
| Temporary database failure | Retried with bounded backoff or placed in a recoverable queue |
| Invalid required field | Rejected with a diagnosable error |
| Unsupported optional property | Rejected or recorded according to documented policy, not silently lost |
| Receiver restart | No unexplained accepted-but-missing events |
| Delayed mobile upload | Original occurrence time retained and delay measurable |

The implementation may choose different guarantees, but operators need to know whether delivery is at-most-once, at-least-once, or effectively-once after deduplication.

## Step 8: Use dual-write or sampled reconciliation during migration

When replacing an existing receiver, avoid switching all traffic based on one successful test. Use dual-write, proxy mirroring, or a controlled traffic percentage where the client and infrastructure allow it.

Compare both paths over fixed windows:

- total events by event name;
- distinct identities using the same definition;
- property presence and type distribution;
- event-time delay distribution;
- duplicate rate;
- rejection and retry counts;
- a few business-critical funnel steps.

Do not expect every total to match without analysis. Different ingestion cutoffs, bot rules, late-event handling, identity resolution, and deduplication can produce legitimate differences. The goal is to explain the differences, not force cosmetic equality.

## A release gate for event pipelines

Before increasing traffic, require evidence for each item:

- [ ] A unique test event appears exactly as expected in raw storage.
- [ ] Identity transitions match the documented model.
- [ ] Client time and receipt time are both understood.
- [ ] Representative property types survive decoding and storage.
- [ ] A deterministic batch produces the expected event and user counts.
- [ ] ClickHouse and Superset return the same result for the validation query.
- [ ] Filters, time zone, cache, and permissions are recorded.
- [ ] Retry, duplicate, invalid payload, and restart behavior are tested.
- [ ] Rollback steps and owners are documented.
- [ ] Backups and restore procedures have been exercised for production data.

This checklist should be versioned with the deployment. Re-run it when the SDK, receiver, schema, database, driver, or dashboard definition changes.

## Where SensorFlow fits

[SensorFlow](https://github.com/data-analyze-bi/sensorFlow) provides a public Apache-2.0 implementation of the path used in this guide: compatible SDK event ingestion, a Go receiving service, ClickHouse storage, and Apache Superset for SQL exploration and dashboards. The repository exposes deployment files and code so teams can inspect the data path rather than relying only on marketing claims.

It is best suited to engineering teams that want self-hosted data and are comfortable operating Docker, ClickHouse, and SQL-based analytics. It is not a drop-in replacement for every feature in a broad product analytics suite. Teams needing turnkey session replay, experimentation, feature flags, or extensive no-code analysis should compare more complete platforms as well.

The detailed [ClickHouse and Superset implementation guide](https://sensorflow.site/use-cases/clickhouse-superset-analytics) covers the architecture and operational boundaries. Teams migrating an existing compatible SDK path can also use the [receiver-boundary migration guide](https://sensorflow.site/use-cases/sensors-sdk-to-clickhouse).

## Frequently asked questions

### Should an SDK write directly to ClickHouse?

Generally, no. A receiving service should handle authentication, protocol decoding, validation, rate limits, retries, and error reporting. Exposing the database directly to untrusted clients expands security risk and makes protocol evolution harder.

### Is one matching event enough to approve a migration?

No. One event validates the basic path. Approval should also cover a deterministic batch, identity transitions, property types, delayed events, duplicates, failures, aggregate SQL, and dashboard reconciliation.

### Can Superset calculate funnels and retention?

Superset can visualize SQL-derived funnel and retention results, but the team normally defines the model and query semantics. It does not automatically supply every specialized product-analytics workflow.

### Why keep both event time and receipt time?

The two timestamps answer different questions. Event time represents when the action occurred on the client; receipt time shows when the platform observed it. Their difference helps identify offline uploads, network delays, clock problems, and processing backlogs.

### What is the most important migration metric?

There is no universal single metric. Start with unexplained event loss, identity consistency, critical-property validity, duplicate behavior, and agreement on business-critical counts. Choose thresholds before the traffic increase.

## Start with one event, then prove the whole path

Reliable analytics comes from traceability, not from a green endpoint alone. Create one identifiable event, inspect its raw row, verify its semantics, reconcile a known batch, and reproduce the query in Superset. Then test failure behavior and compare old and new paths under controlled traffic.

To apply the checklist to a transparent reference implementation, review the [SensorFlow deployment and validation documentation](https://sensorflow.site/docs) and inspect the source before sending production data.
