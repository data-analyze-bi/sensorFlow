# Frequently Asked Questions

## Which client SDK should I use?

Use the official Sensors Data SDK for Android, iOS, Web, mini programs, or your server language. This repository does not distribute a custom client SDK.

## Where should the SDK send events?

Set `serverUrl` to `https://your-domain.example/sensors/send/?token=YOUR_TOKEN`. Use HTTPS in production and keep the token out of public source code.

## Where is data stored?

Events are stored in `sensors.event`; user profile records are stored in `sensors.user` in ClickHouse.

## How do I confirm ingestion works?

Send an `integration_test` event, query it in ClickHouse, then run the same query in Superset SQL Lab.

## Why is a request rejected?

Confirm the license file at `binaries/sensors-payload-decoder`, request URL, and server logs. The license rejects processing when its embedded authorization check fails.

## Can I use this in production?

Run it behind HTTPS, replace every default credential, set a strong `SUPERSET_SECRET_KEY`, restrict database ports, back up volumes, and add monitoring before production rollout.

中文版本：[FAQ.zh-CN.md](FAQ.zh-CN.md)
