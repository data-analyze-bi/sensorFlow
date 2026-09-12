# 神策各端 SDK 接入

SensorFlow 使用神策官方客户端与服务端 SDK，当前不发布自研客户端 SDK。

> ⚠️ **关于第三方 SDK 兼容性的重要说明**
>
> SensorFlow 兼容神策官方 SDK 的标准事件上报流程，但 SensorFlow 与神策数据不是关联公司，该兼容性也不代表神策官方认证或承诺。不同 SDK 版本、加密插件、可视化埋点、全埋点和后续协议变更可能存在差异。升级 SDK 或启用扩展功能前，请先在测试环境验证事件接收、用户关联、属性类型和入库结果；生产切换前建议保留回滚方案。
>
> 本仓库不包含或再分发神策 SDK。请从神策官方渠道获取，并遵守其独立许可及使用条款。详见[第三方 SDK 与商标声明](../THIRD_PARTY_NOTICES.md#中文说明)。

## 接收地址

初始化 SDK 时，将 `serverUrl` 设置为：

```text
https://your-domain.example/sensors/send/?token=YOUR_TOKEN
```

替换实际域名和 Token；生产环境必须使用 HTTPS，且不得把生产 Token 提交到公开源码。

## 支持平台

Android、iOS、JavaScript、微信小程序及服务端语言均使用神策对应的官方 SDK。依赖安装、自动采集和平台权限以该 SDK 官方说明为准。

各平台遵循相同流程：安装 SDK、配置接收地址、设置稳定用户标识、上报统一命名的业务事件、最后在 ClickHouse 和 Superset 验证。

## 各端下载与安装入口

- **Web / JavaScript**：[sensorsdata/sa-sdk-javascript](https://github.com/sensorsdata/sa-sdk-javascript)
- **iOS**：[sensorsdata/sa-sdk-ios](https://github.com/sensorsdata/sa-sdk-ios)
- **Android**：[sensorsdata/sa-sdk-android](https://github.com/sensorsdata/sa-sdk-android)
- **微信及其他小程序**：[sensorsdata/sa-sdk-miniprogram](https://github.com/sensorsdata/sa-sdk-miniprogram)
- **React Native**：[sensorsdata/react-native-sensors-analytics](https://github.com/sensorsdata/react-native-sensors-analytics)
- **Java**：[sensorsdata/sa-sdk-java](https://github.com/sensorsdata/sa-sdk-java)
- **Go**：[sensorsdata/sa-sdk-go](https://github.com/sensorsdata/sa-sdk-go)
- **Python**：[sensorsdata/sa-sdk-python](https://github.com/sensorsdata/sa-sdk-python)
- **Node.js**：[sensorsdata/sa-sdk-node](https://github.com/sensorsdata/sa-sdk-node)
- **PHP**：[sensorsdata/sa-sdk-php](https://github.com/sensorsdata/sa-sdk-php)

以上均为 `sensorsdata` GitHub 官方组织仓库。下载和依赖安装以神策官方说明为准；选择版本前请查看该版本仓库中的 `LICENSE` 与使用条款。SensorFlow 只要求 SDK 最终将标准事件发送到本文的接收地址。

## Web

安装神策 JavaScript SDK 后，将 `server_url` 指向 SensorFlow：

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
sensors.track("ButtonClick", { button_name: "Submit", page_name: "Registration" });
```

可使用 npm、yarn 或神策官方支持的浏览器引入方式安装。生产环境请关闭调试日志，并检查 Beacon、CORS 与域名策略。

## Android

按神策 Android SDK 官方文档添加依赖，在 `Application` 初始化时使用 SensorFlow 接收地址：

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

SensorFlow 与神策云服务无关，无需配置神策云项目 Scheme。请按隐私合规要求选择自动采集范围和用户授权时机，发布前关闭日志并验证混淆、网络权限和 HTTPS。

## iOS

按神策 iOS SDK 官方文档安装依赖，并在应用启动时配置接收地址：

```swift
import SensorsAnalyticsSDK

let options = SAConfigOptions(
    serverURL: "https://your-domain.example/sensors/send/?token=YOUR_TOKEN",
    launchOptions: launchOptions
)
options.enableLog = true
options.enableAutoTrack = .default
SensorsAnalyticsSDK.start(configOptions: options)
SensorsAnalyticsSDK.sharedInstance()?.login("USER_ID")
SensorsAnalyticsSDK.sharedInstance()?.track(
    "ButtonClick",
    withProperties: ["button_name": "Submit"]
)
```

SensorFlow 与神策云服务无关，无需配置神策云项目 Scheme。先验证自动采集、登录 ID、自定义属性和生命周期事件，正式发布前关闭调试日志。

## 微信小程序

从神策官方渠道获取 SDK，将 SensorFlow HTTPS 域名加入微信公众平台的 `request` 合法域名：

```javascript
import sensors from "./utils/sensorsdata.min.js";

sensors.init({
  server_url: "https://your-domain.example/sensors/send/?token=YOUR_TOKEN",
  autotrack: {
    appLaunch: true,
    appShow: true,
    appHide: true,
    pageShow: true,
    pageShare: true,
  },
  show_name: true,
});
sensors.login("USER_ID");
sensors.track("ButtonClick", { button_name: "Submit" });
```

按业务和隐私要求选择自动采集项，并在真机及正式域名环境验证请求。

## React Native

按神策官方文档安装 npm 包并完成 iOS、Android 两端原生初始化：

```javascript
import RNSensorsAnalyticsModule from "react-native-sensors-analytics";

RNSensorsAnalyticsModule.init({
  serverUrl: "https://your-domain.example/sensors/send/?token=YOUR_TOKEN",
  autoTrackAppViewScreen: true,
});
RNSensorsAnalyticsModule.login("USER_ID");
RNSensorsAnalyticsModule.track("ButtonClick", { button_name: "Submit" });
```

确保 JavaScript 包与两端原生 SDK 版本兼容，升级 React Native 或原生 SDK 前完成回归测试。

## 服务端语言

服务端请选择神策官方对应语言 SDK，将接收地址配置为 SensorFlow，并在登录、注册、支付等业务成功后上报事件。用户 ID 应与客户端登录 ID 保持一致，避免同一用户被拆分。

- **Java**：使用神策 Java SDK，在服务初始化时创建 Consumer，并复用同一个 `SensorsAnalytics` 实例。
- **Go**：使用神策 Go SDK，配置网络 Consumer 的接收地址；进程退出前确保缓冲事件已经发送。
- **Python**：使用神策 Python SDK，根据 Web 服务生命周期创建和关闭 Consumer。
- **Node.js**：使用神策 Node.js SDK，在应用级单例中配置接收地址，避免每次请求重复初始化。
- **PHP**：使用神策 PHP SDK；长驻进程和传统 PHP-FPM 应分别按生命周期管理 Consumer。
- **其他语言**：若神策提供对应官方 SDK，先确认其支持自定义接收地址并在测试环境验证；不要自行拼装未验证的请求格式直接用于生产。

Python 示例：

```python
from sensorsanalytics import SensorsAnalytics

sa = SensorsAnalytics(
    server_url="https://your-domain.example/sensors/send/?token=YOUR_TOKEN"
)
sa.track(
    distinct_id="USER_ID",
    event_name="ProductView",
    properties={"product_name": "Example", "price": 5999},
)
sa.flush()
sa.close()
```

## 接入检查

1. 先在测试环境发送一条 `integration_test` 事件。
2. 确认接收服务返回成功，且日志中没有字段校验错误。
3. 在 ClickHouse 查询事件与用户属性。
4. 在 Superset 创建数据集和图表，验证日活、注册、登录及转化指标。
5. 验证完成后再切换生产域名和 Token。

首次可发送 `integration_test` 事件，并附带 `platform`、`environment` 属性。查询：

```sql
SELECT event, count()
FROM sensors.event
WHERE event = 'integration_test'
GROUP BY event;
```

不要采集密码、访问 Token、身份证号等敏感信息。
