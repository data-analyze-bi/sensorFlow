ALTER TABLE sensors.event ADD COLUMN IF NOT EXISTS channel LowCardinality(String) DEFAULT '';
ALTER TABLE sensors.event ADD COLUMN IF NOT EXISTS country LowCardinality(String) DEFAULT '';
ALTER TABLE sensors.event ADD COLUMN IF NOT EXISTS page_name String DEFAULT '';
ALTER TABLE sensors.event ADD COLUMN IF NOT EXISTS demo_hour String DEFAULT '';
ALTER TABLE sensors.event ADD COLUMN IF NOT EXISTS demo_day String DEFAULT '';
ALTER TABLE sensors.event ADD COLUMN IF NOT EXISTS revenue Float64 DEFAULT 0;

INSERT INTO sensors.event
    (time, event, distinct_id, os, network_type, app_id, app_version, app_name,
     device_id, channel, country, page_name, demo_hour, demo_day, revenue, is_first_day)
SELECT
    now64(3) - INTERVAL number * 3 HOUR,
    multiIf(number % 12 < 5, 'demo_app_open', number % 12 < 9, 'demo_product_view', number % 12 < 11, 'demo_add_to_cart', 'demo_purchase'),
    concat('demo-user-', toString(number % 48 + 1)),
    ['Android', 'iOS', 'Web'][number % 3 + 1],
    ['wifi', '5G', '4G'][number % 3 + 1],
    'sensorflow-demo',
    ['1.0.0', '1.1.0', '2.0.0'][number % 3 + 1],
    'SensorFlow Demo',
    concat('demo-device-', toString(number % 60 + 1)),
    ['Organic', 'Search Ads', 'Social', 'Partner'][number % 4 + 1],
    ['CN', 'US', 'SG', 'DE'][number % 4 + 1],
    ['Home', 'Pricing', 'Product', 'Checkout'][number % 4 + 1],
    formatDateTime(now() - INTERVAL number * 3 HOUR, '%m-%d %H:00'),
    formatDateTime(now() - INTERVAL number * 3 HOUR, '%Y-%m-%d'),
    if(number % 12 = 11, 49 + (number % 5) * 25, 0),
    number < 48
FROM numbers(240)
WHERE (SELECT count() FROM sensors.event WHERE app_id = 'sensorflow-demo') = 0;
