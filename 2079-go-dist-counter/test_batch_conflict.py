#!/usr/bin/env python3
import requests
import json

BASE_URL = "http://localhost:8080/api"

print("=" * 60)
print("测试包含边界冲突的批量操作 (应返回 207)")
print("=" * 60)

# 先设置一个有边界的计数器
r = requests.put(f"{BASE_URL}/namespaces/test/bounded_counter/bounds", 
                 json={"min": 0, "max": 5}, timeout=5)
print(f"\n设置边界 - Status: {r.status_code}")

# 先把值设到 5
r = requests.put(f"{BASE_URL}/namespaces/test/bounded_counter", 
                 json={"value": 5}, timeout=5)
print(f"设值到 5 - Status: {r.status_code}, Response: {r.json()}")

# 测试包含冲突的批量操作
print("\n=== 测试包含边界冲突的批量操作 ===")
payload = {
    "items": [
        {"namespace": "test", "name": "bounded_counter", "delta": 1},  # 会超过上限 5，失败
        {"namespace": "test", "name": "normal_counter", "delta": 3},    # 正常递增
        {"namespace": "test", "name": "another_normal", "delta": 10}    # 正常递增
    ]
}
print(f"Data: {json.dumps(payload, indent=2)}")
r = requests.post(f"{BASE_URL}/batch", json=payload, timeout=5)
print(f"Status: {r.status_code}")
print(f"Response: {json.dumps(r.json(), indent=2, ensure_ascii=False)}")

print("\n" + "=" * 60)
if r.status_code == 207:
    print("✓ 正确返回 207 状态码！")
else:
    print(f"✗ 期望 207，但得到 {r.status_code}")
print("=" * 60)
