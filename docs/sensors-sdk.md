# Sensors Data SDK Integration

SensorFlow uses the official Sensors Data SDKs and does not distribute a custom client SDK.

> ⚠️ **Important third-party SDK compatibility notice**
>
> SensorFlow supports the standard event upload flow used by official Sensors Data SDKs. SensorFlow is not affiliated with Sensors Data, and compatibility does not imply official certification or guarantees. SDK versions, encryption plugins, visual tracking, auto-track features, and future protocol changes may behave differently. Validate ingestion, identity mapping, property types, and stored results before production changes, and keep a rollback path.
>
> The SDKs are not included or redistributed here. Obtain them from official sources and comply with their separate licenses and terms. See [Third-Party SDK and Trademark Notice](../THIRD_PARTY_NOTICES.md).

## Ingestion URL

Configure the SDK `serverUrl` as:

```text
https://your-domain.example/sensors/send/?token=YOUR_TOKEN
```

Replace the domain and token, use HTTPS in production, and never commit a production token.

## Supported platforms

Use the official SDK for Android, iOS, JavaScript, WeChat Mini Program, or your server language. Follow that SDK's official installation and platform permission instructions.

For server applications, choose the official SDK for Java, Go, Python, Node.js, or PHP. Reuse an application-level client where appropriate, flush buffered events before shutdown, keep server and client user IDs consistent, and validate other official language SDKs before production use.

For every platform:

1. Install the official SDK.
2. Configure the ingestion URL.
3. Set a stable anonymous or signed-in user identifier.
4. Track business events with consistent names and properties.
5. Verify the event in ClickHouse and Superset.

## Official download links

- **Web / JavaScript:** [sensorsdata/sa-sdk-javascript](https://github.com/sensorsdata/sa-sdk-javascript)
- **iOS:** [sensorsdata/sa-sdk-ios](https://github.com/sensorsdata/sa-sdk-ios)
- **Android:** [sensorsdata/sa-sdk-android](https://github.com/sensorsdata/sa-sdk-android)
- **Mini Programs:** [sensorsdata/sa-sdk-miniprogram](https://github.com/sensorsdata/sa-sdk-miniprogram)
- **React Native:** [sensorsdata/react-native-sensors-analytics](https://github.com/sensorsdata/react-native-sensors-analytics)
- **Java:** [sensorsdata/sa-sdk-java](https://github.com/sensorsdata/sa-sdk-java)
- **Go:** [sensorsdata/sa-sdk-go](https://github.com/sensorsdata/sa-sdk-go)
- **Python:** [sensorsdata/sa-sdk-python](https://github.com/sensorsdata/sa-sdk-python)
- **Node.js:** [sensorsdata/sa-sdk-node](https://github.com/sensorsdata/sa-sdk-node)
- **PHP:** [sensorsdata/sa-sdk-php](https://github.com/sensorsdata/sa-sdk-php)

These are repositories under the official `sensorsdata` GitHub organization. Before selecting a version, review that version's `LICENSE` and official terms.

## Web / JavaScript

```javascript
import sensors from "sa-sdk-javascript";

sensors.init({
  server_url: "https://your-domain.example/sensors/send/?token=YOUR_TOKEN",
  is_track_single_page: true,
  use_client_time: true,
  send_type: "beacon",
});
sensors.quick("autoTrack");
sensors.login("USER_ID");
sensors.track("ButtonClick", { button_name: "Submit" });
```

## Android

```java
SAConfigOptions options = new SAConfigOptions(
    "https://your-domain.example/sensors/send/?token=YOUR_TOKEN"
);
options.setAutoTrackEventType(
    SensorsDataAPI.AutoTrackEventType.APP_START |
    SensorsDataAPI.AutoTrackEventType.APP_END |
    SensorsDataAPI.AutoTrackEventType.APP_CLICK
);
SensorsDataAPI.startWithConfigOptions(this, options);
SensorsDataAPI.sharedInstance().login("USER_ID");
SensorsDataAPI.sharedInstance().track("ButtonClick", properties);
```

## iOS

```swift
let options = SAConfigOptions(
    serverURL: "https://your-domain.example/sensors/send/?token=YOUR_TOKEN",
    launchOptions: launchOptions
)
options.enableAutoTrack = .default
SensorsAnalyticsSDK.start(configOptions: options)
SensorsAnalyticsSDK.sharedInstance()?.login("USER_ID")
SensorsAnalyticsSDK.sharedInstance()?.track("ButtonClick", withProperties: ["button_name": "Submit"])
```

## WeChat Mini Program

Add the SensorFlow HTTPS host to the request-domain allowlist before initialization.

```javascript
sensors.init({
  server_url: "https://your-domain.example/sensors/send/?token=YOUR_TOKEN",
  autotrack: {
    appLaunch: true, appShow: true, appHide: true,
    pageShow: true, pageShare: true,
  },
});
sensors.login("USER_ID");
sensors.track("ButtonClick", { button_name: "Submit" });
```

## React Native

```javascript
RNSensorsAnalyticsModule.init({
  serverUrl: "https://your-domain.example/sensors/send/?token=YOUR_TOKEN",
  autoTrackAppViewScreen: true,
});
RNSensorsAnalyticsModule.login("USER_ID");
RNSensorsAnalyticsModule.track("ButtonClick", { button_name: "Submit" });
```

## Server SDKs

Use the official Java, Go, Python, Node.js, or PHP SDK. Reuse one application-level instance, keep client and server login IDs consistent, and flush buffered events before shutdown.

```python
from sensorsanalytics import SensorsAnalytics

sa = SensorsAnalytics(server_url="https://your-domain.example/sensors/send/?token=YOUR_TOKEN")
sa.track("USER_ID", "ProductView", {"product_name": "Example", "price": 5999})
sa.flush()
sa.close()
```

SensorFlow is independent from Sensors Data Cloud. A Sensors Data Cloud project Scheme is not required for this endpoint. Configure privacy consent and auto-track scope for your jurisdiction, disable debug logging before release, and test optional plugins before production use.

## Validate collection

Use `integration_test` as the first test event with `platform` and `environment` properties. Confirm the request reaches `/sensors/send/`, inspect ingestion logs, then verify it with:

```sql
SELECT event, count()
FROM sensors.event
WHERE event = 'integration_test'
GROUP BY event;
```

Do not collect passwords, access tokens, identity numbers, or other sensitive data.

中文版本：[sensors-sdk.zh-CN.md](sensors-sdk.zh-CN.md)
