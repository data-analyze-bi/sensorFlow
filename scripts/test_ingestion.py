#!/usr/bin/env python3
import argparse
import base64
import json
import os
import secrets
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
import zlib
from pathlib import Path


ROOT_DIR = Path(__file__).resolve().parent.parent
ENV_FILE = ROOT_DIR / "deploy" / "docker" / ".env"


def read_env():
    values = {}
    if not ENV_FILE.exists():
        return values
    for line in ENV_FILE.read_text(encoding="utf-8").splitlines():
        if not line or line.lstrip().startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        values[key] = value.strip().strip("'\"")
    return values


def build_url(explicit_url, env):
    if explicit_url:
        return explicit_url
    token = env.get("SENSORFLOW_INGESTION_TOKEN", "")
    if not token:
        raise ValueError("未找到 Token，请传入完整接收 URL，或先运行 ./activate.sh。")
    domain = env.get("SENSORFLOW_DOMAIN", "")
    if domain:
        return f"https://{domain}/sensors/send/?token={token}"
    port = env.get("SENSORFLOW_PORT", "8081")
    return f"http://127.0.0.1:{port}/sensors/send/?token={token}"


def event(name, distinct_id, properties):
    now_ms = int(time.time() * 1000)
    return {
        "type": "track",
        "event": name,
        "distinct_id": distinct_id,
        "time": now_ms,
        "properties": {
            "$lib": "Python",
            "$lib_method": "code",
            "$lib_version": "sensorflow-test-1.0",
            "platform": "macos",
            "environment": "integration_test",
            **properties,
        },
    }


def send(url, events):
    raw = json.dumps(events, separators=(",", ":"), ensure_ascii=False).encode()
    encoded = base64.b64encode(zlib.compress(raw)).decode()
    body = urllib.parse.urlencode({"data_list": encoded, "gzip": "1"}).encode()
    request = urllib.request.Request(
        url,
        data=body,
        headers={"Content-Type": "application/x-www-form-urlencoded", "User-Agent": "SensorFlow-Ingestion-Test/1.0"},
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=20) as response:
        return response.status, response.read().decode("utf-8", errors="replace")


def main():
    parser = argparse.ArgumentParser(description="发送一组神策标准测试事件到 SensorFlow。")
    parser.add_argument("url", nargs="?", help="完整 serverUrl；省略时读取 deploy/docker/.env")
    parser.add_argument("--user", help="测试 distinct_id；省略时自动生成")
    args = parser.parse_args()

    try:
        url = build_url(args.url, read_env())
    except ValueError as error:
        print(error, file=sys.stderr)
        return 2

    distinct_id = args.user or f"sensorflow_test_{secrets.token_hex(4)}"
    order_id = f"order_{secrets.token_hex(4)}"
    events = [
        event("integration_test", distinct_id, {"test_source": "test-ingestion.sh"}),
        event("PageView", distinct_id, {"page_name": "SensorFlow Test", "channel": "wifi_test"}),
        event("UserSignup", distinct_id, {"signup_method": "email"}),
        event("UserLogin", distinct_id, {"login_method": "password"}),
        event("ProductView", distinct_id, {"product_id": "demo_product", "category": "analytics"}),
        event("AddToCart", distinct_id, {"product_id": "demo_product", "quantity": 1}),
        event("OrderCreated", distinct_id, {"order_id": order_id, "revenue": 99.0}),
        event("PaymentSuccess", distinct_id, {"order_id": order_id, "revenue": 99.0, "currency": "CNY"}),
    ]

    safe_url = urllib.parse.urlsplit(url)
    print(f"正在向 {safe_url.scheme}://{safe_url.netloc}{safe_url.path} 发送 {len(events)} 条测试事件…")
    try:
        status, response_body = send(url, events)
    except urllib.error.HTTPError as error:
        print(f"发送失败：HTTP {error.code} {error.read().decode(errors='replace')}", file=sys.stderr)
        return 1
    except OSError as error:
        print(f"连接失败：{error}", file=sys.stderr)
        return 1

    print(f"接收服务返回 HTTP {status}: {response_body}")
    print(f"测试用户：{distinct_id}")
    print("事件：" + ", ".join(item["event"] for item in events))
    print("请在服务机执行 ./verify-ingestion.sh " + distinct_id + " 确认 ClickHouse 入库。")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
