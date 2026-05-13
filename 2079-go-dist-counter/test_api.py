#!/usr/bin/env python3
import requests
import json
import sys

BASE_URL = "http://localhost:8080/api"
TIMEOUT = 5

def test_get(url):
    print(f"\n=== GET {url} ===")
    try:
        r = requests.get(f"{BASE_URL}{url}", timeout=TIMEOUT)
        print(f"Status: {r.status_code}")
        print(f"Response: {json.dumps(r.json(), indent=2, ensure_ascii=False)}")
        return r
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

def test_post(url, data):
    print(f"\n=== POST {url} ===")
    print(f"Data: {json.dumps(data)}")
    try:
        r = requests.post(f"{BASE_URL}{url}", json=data, timeout=TIMEOUT)
        print(f"Status: {r.status_code}")
        print(f"Response: {json.dumps(r.json(), indent=2, ensure_ascii=False)}")
        return r
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

def test_put(url, data):
    print(f"\n=== PUT {url} ===")
    print(f"Data: {json.dumps(data)}")
    try:
        r = requests.put(f"{BASE_URL}{url}", json=data, timeout=TIMEOUT)
        print(f"Status: {r.status_code}")
        print(f"Response: {json.dumps(r.json(), indent=2, ensure_ascii=False)}")
        return r
    except Exception as e:
        print(f"Error: {e}")
        sys.exit(1)

print("=" * 60)
print("开始测试分布式计数系统 API")
print("=" * 60)

# 1. 列出命名空间
test_get("/namespaces")

# 2. 递增计数器
test_post("/namespaces/myapp/counter1", {"delta": 5})

# 3. 递减计数器
test_post("/namespaces/myapp/counter1", {"delta": -3})

# 4. 设置边界
test_put("/namespaces/myapp/counter1/bounds", {"min": 0, "max": 10})

# 5. 测试超过上限 (应返回 409)
print("\n=== 测试超过上限 (应返回 409) ===")
test_post("/namespaces/myapp/counter1", {"delta": 10})

# 6. 获取单个计数器
test_get("/namespaces/myapp/counter1")

# 7. 获取命名空间下所有计数器
test_get("/namespaces/myapp")

# 8. 测试不存在的命名空间（应返回空列表）
test_get("/namespaces/nonexistent")

# 9. 设置值
test_put("/namespaces/myapp/counter2", {"value": 42})

# 10. 获取历史记录
test_get("/namespaces/myapp/counter1/history")

# 11. 批量操作
print("\n=== 测试批量操作 (应返回 207) ===")
test_post("/batch", {
    "items": [
        {"namespace": "myapp", "name": "counter1", "delta": 1},
        {"namespace": "myapp", "name": "counter2", "delta": -10},
        {"namespace": "another", "name": "newcounter", "delta": 5}
    ]
})

print("\n" + "=" * 60)
print("所有测试完成！")
print("=" * 60)
