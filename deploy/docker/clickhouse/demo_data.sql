INSERT INTO sensors.event
    (time, event, distinct_id, os, network_type, app_id, app_version, app_name, device_id)
SELECT
    now64(3) - INTERVAL number HOUR,
    ['demo_app_open', 'demo_product_view', 'demo_add_to_cart', 'demo_purchase'][number % 4 + 1],
    concat('demo-user-', toString(number % 18 + 1)),
    ['Android', 'iOS', 'Web'][number % 3 + 1],
    ['wifi', '5G', '4G'][number % 3 + 1],
    'sensorflow-demo',
    ['1.0.0', '1.1.0', '2.0.0'][number % 3 + 1],
    'SensorFlow Demo',
    concat('demo-device-', toString(number % 24 + 1))
FROM numbers(96)
WHERE (SELECT count() FROM sensors.event WHERE app_id = 'sensorflow-demo') = 0;
