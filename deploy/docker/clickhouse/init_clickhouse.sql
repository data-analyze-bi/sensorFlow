-- ClickHouse 数据库初始化脚本
-- 字段名与服务端写入前的 RemoveDollarSign/FlattenJSON 结果保持一致。

CREATE DATABASE IF NOT EXISTS sensors;

USE sensors;

CREATE TABLE IF NOT EXISTS sensors.event (
    `time` DateTime64(3) DEFAULT now64(3),
    `ds` Date DEFAULT toDate(`time`),
    `event` LowCardinality(String) DEFAULT '',
    `distinct_id` String DEFAULT '',
    `anonymous_id` String DEFAULT '',

    `lib` LowCardinality(String) DEFAULT '',
    `lib_method` LowCardinality(String) DEFAULT '',
    `lib_version` String DEFAULT '',
    `_flush_time` UInt64 DEFAULT 0,

    `os` LowCardinality(String) DEFAULT '',
    `os_version` String DEFAULT '',
    `model` String DEFAULT '',
    `brand` String DEFAULT '',
    `manufacturer` String DEFAULT '',
    `screen_width` UInt16 DEFAULT 0,
    `screen_height` UInt16 DEFAULT 0,
    `wifi` Bool DEFAULT false,
    `network_type` LowCardinality(String) DEFAULT '',

    `app_id` String DEFAULT '',
    `app_version` String DEFAULT '',
    `app_name` String DEFAULT '',
    `is_first_day` Bool DEFAULT false,
    `timezone_offset` Int16 DEFAULT 0,
    `device_id` String DEFAULT '',
    `user_id` String DEFAULT '',
    `channel` LowCardinality(String) DEFAULT '',
    `country` LowCardinality(String) DEFAULT '',
    `page_name` String DEFAULT '',
    `demo_hour` String DEFAULT '',
    `demo_day` String DEFAULT '',
    `revenue` Float64 DEFAULT 0,

    INDEX idx_event (`event`) TYPE bloom_filter GRANULARITY 1,
    INDEX idx_distinct_id (`distinct_id`) TYPE bloom_filter GRANULARITY 1,
    INDEX idx_time (`time`) TYPE minmax GRANULARITY 1
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(`ds`)
ORDER BY (`distinct_id`, `event`, `time`)
TTL toDateTime(`time`) + INTERVAL 1 YEAR
SETTINGS index_granularity = 8192;

CREATE TABLE IF NOT EXISTS sensors.user (
    `distinct_id` String DEFAULT '',
    `ds` Date DEFAULT today(),
    `user_id` String DEFAULT '',
    `first_visit_time` DateTime64(3) DEFAULT now64(3),
    `last_visit_time` DateTime64(3) DEFAULT now64(3),

    `name` String DEFAULT '',
    `email` String DEFAULT '',
    `phone` String DEFAULT '',
    `avatar` String DEFAULT '',

    INDEX idx_distinct_id (`distinct_id`) TYPE bloom_filter GRANULARITY 1,
    INDEX idx_user_id (`user_id`) TYPE bloom_filter GRANULARITY 1
) ENGINE = ReplacingMergeTree(last_visit_time)
ORDER BY (`distinct_id`)
SETTINGS index_granularity = 8192;

SHOW TABLES FROM sensors;
