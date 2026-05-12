import requests
import json
import time

BASE_URL = "http://localhost:8000"

def test_rate_limit():
    print("=== 测试限流功能 ===")
    
    print("\n1. 测试 /api/fast/test 端点（窗口10秒，最大3次）")
    for i in range(5):
        response = requests.get(f"{BASE_URL}/api/fast/test")
        print(f"  请求 {i+1}: 状态码 {response.status_code}")
        if response.status_code == 429:
            data = response.json()
            print(f"    限流信息: {json.dumps(data, indent=2, ensure_ascii=False)}")
            print(f"    Retry-After 头: {response.headers.get('Retry-After')}")
            break
    
    print("\n2. 测试 /api/slow/test 端点（窗口15秒，最大2次）")
    for i in range(4):
        response = requests.get(f"{BASE_URL}/api/slow/test")
        print(f"  请求 {i+1}: 状态码 {response.status_code}")
        if response.status_code == 429:
            data = response.json()
            print(f"    限流信息: {json.dumps(data, indent=2, ensure_ascii=False)}")
            print(f"    Retry-After 头: {response.headers.get('Retry-After')}")
            break
    
    print("\n3. 查看当前使用量")
    response = requests.get(f"{BASE_URL}/rate-limits/usage")
    if response.status_code == 200:
        print(f"  使用量信息: {json.dumps(response.json(), indent=2, ensure_ascii=False)}")
    
    print("\n4. 测试配置删除功能")
    response = requests.delete(f"{BASE_URL}/rate-limits/%2Fapi%2Ffast")
    print(f"  删除 /api/fast 配置: 状态码 {response.status_code}")
    
    response = requests.get(f"{BASE_URL}/rate-limits")
    if response.status_code == 200:
        print(f"  剩余配置: {json.dumps(response.json(), indent=2, ensure_ascii=False)}")

if __name__ == "__main__":
    test_rate_limit()
