#!/usr/bin/env python3
import requests
import json

BASE_URL = "http://127.0.0.1:3000"

def print_separator(title=""):
    print("\n" + "=" * 60)
    if title:
        print(f"  {title}")
        print("=" * 60)

def test_case(name, func):
    print_separator(f"测试: {name}")
    try:
        result = func()
        print(f"✓ 通过")
        return result
    except Exception as e:
        print(f"✗ 失败: {e}")
        return None

# 1. 创建车辆
print_separator("测试1: 创建车辆 (实际价值10万)")
vehicle = requests.post(f"{BASE_URL}/vehicles", json={
    "plate_number": "京A12345",
    "model": "丰田凯美瑞",
    "actual_value": "100000"
}).json()
print(f"车辆ID: {vehicle['id']}")
print(f"实际价值: {vehicle['actual_value']}")
vehicle_id = vehicle['id']

# 2. 创建理赔单
print_separator("测试2: 创建理赔单")
claim = requests.post(f"{BASE_URL}/claims", json={
    "vehicle_id": vehicle_id
}).json()
print(f"理赔单ID: {claim['id']}")
print(f"状态: {claim['status']}")
claim_id = claim['id']

# 3. 添加维修项目 (未超过80%)
print_separator("测试3: 添加维修项目 (5万, 未超过80%)")
claim = requests.post(f"{BASE_URL}/claims/{claim_id}/items", json={
    "name": "前保险杠更换",
    "item_type": "PartReplacement",
    "cost": "50000"
}).json()
print(f"维修项目数: {len(claim['repair_items'])}")
print(f"总维修费: 50000")
print(f"状态: {claim['status']}")
assert claim['status'] == 'Pending', "应该还是待处理状态"
assert len(claim['repair_items']) == 1, "应该有1个维修项目"

# 4. 继续添加 (仍未超过80%)
print_separator("测试4: 继续添加 (2万, 共7万, 仍未超过80%)")
claim = requests.post(f"{BASE_URL}/claims/{claim_id}/items", json={
    "name": "维修工时费",
    "item_type": "Labor",
    "cost": "20000"
}).json()
print(f"维修项目数: {len(claim['repair_items'])}")
print(f"状态: {claim['status']}")
assert claim['status'] == 'Pending', "应该还是待处理状态"
assert len(claim['repair_items']) == 2, "应该有2个维修项目"

# 5. 再次添加 (超过80%, 触发推定全损)
print_separator("测试5: 继续添加 (1.5万, 共8.5万, 超过80%=8万)")
claim = requests.post(f"{BASE_URL}/claims/{claim_id}/items", json={
    "name": "发动机大修",
    "item_type": "PartReplacement",
    "cost": "15000"
}).json()
print(f"维修项目数: {len(claim['repair_items'])}")
print(f"状态: {claim['status']}")
assert claim['status'] == 'TotalLoss', "应该是推定全损状态"
assert len(claim['repair_items']) == 0, "维修项目应该被清空"

# 6. 尝试给推定全损的理赔单添加项目
print_separator("测试6: 尝试给推定全损的理赔单添加项目 (应该失败)")
response = requests.post(f"{BASE_URL}/claims/{claim_id}/items", json={
    "name": "再添加一个",
    "item_type": "Labor",
    "cost": "1000"
})
print(f"状态码: {response.status_code}")
print(f"错误: {response.json().get('error')}")
assert response.status_code == 400, "应该返回400错误"

# 7. 结案推定全损的理赔单
print_separator("测试7: 结案推定全损 (残值2万)")
claim = requests.post(f"{BASE_URL}/claims/{claim_id}/settle", json={
    "salvage_value": "20000"
}).json()
print(f"状态: {claim['status']}")
print(f"残值: {claim.get('salvage_value')}")
print(f"赔付款: {claim.get('payout')}")
assert claim['status'] == 'Settled', "应该已结案"
assert claim['payout'] == '80000', "赔付款应该是10万 - 2万 = 8万"

# 8. 测试残值不能超过实际价值
print_separator("测试8: 测试残值不能超过实际价值")
claim2 = requests.post(f"{BASE_URL}/claims", json={"vehicle_id": vehicle_id}).json()
claim2_id = claim2['id']
# 添加一些项目触发推定全损
requests.post(f"{BASE_URL}/claims/{claim2_id}/items", json={
    "name": "全车维修",
    "item_type": "Labor",
    "cost": "90000"
})
response = requests.post(f"{BASE_URL}/claims/{claim2_id}/settle", json={
    "salvage_value": "150000"
})
print(f"状态码: {response.status_code}")
print(f"错误: {response.json().get('error')}")
assert response.status_code == 400, "应该返回400错误"

# 9. 测试有未结案理赔单不能修改实际价值
print_separator("测试9: 测试有未结案理赔单不能修改实际价值")
claim3 = requests.post(f"{BASE_URL}/claims", json={"vehicle_id": vehicle_id}).json()
response = requests.post(f"{BASE_URL}/vehicles/{vehicle_id}", json={
    "actual_value": "90000"
})
print(f"状态码: {response.status_code}")
print(f"错误: {response.json().get('error')}")
assert response.status_code == 400, "应该返回400错误"

# 10. 测试多个理赔单同时定损独立判断
print_separator("测试10: 创建第二辆车, 测试多个理赔单独立判断")
vehicle2 = requests.post(f"{BASE_URL}/vehicles", json={
    "plate_number": "京B67890",
    "model": "本田雅阁",
    "actual_value": "200000"
}).json()
vehicle2_id = vehicle2['id']

# 为两辆车各创建一个理赔单
claim_a = requests.post(f"{BASE_URL}/claims", json={"vehicle_id": vehicle_id}).json()
claim_b = requests.post(f"{BASE_URL}/claims", json={"vehicle_id": vehicle2_id}).json()

# 车辆1添加8.5万 (超过10万的80%)
claim_a = requests.post(f"{BASE_URL}/claims/{claim_a['id']}/items", json={
    "name": "高价维修",
    "item_type": "Labor",
    "cost": "85000"
}).json()

# 车辆2添加8.5万 (未超过20万的80%=16万)
claim_b = requests.post(f"{BASE_URL}/claims/{claim_b['id']}/items", json={
    "name": "维修项目",
    "item_type": "Labor",
    "cost": "85000"
}).json()

print(f"车辆1理赔单状态: {claim_a['status']} (推定全损预期)")
print(f"车辆1维修项目数: {len(claim_a['repair_items'])}")
print(f"车辆2理赔单状态: {claim_b['status']} (待处理预期)")
print(f"车辆2维修项目数: {len(claim_b['repair_items'])}")

assert claim_a['status'] == 'TotalLoss', "车辆1应该推定全损"
assert claim_b['status'] == 'Pending', "车辆2应该还是待处理"

# 11. 测试正常维修结案
print_separator("测试11: 测试正常维修结案 (非推定全损)")
# 结案车辆2的理赔单
claim_b = requests.post(f"{BASE_URL}/claims/{claim_b['id']}/settle", json={}).json()
print(f"状态: {claim_b['status']}")
print(f"赔付款: {claim_b.get('payout')}")
assert claim_b['status'] == 'Settled', "应该已结案"
assert claim_b['payout'] == '85000', "赔付款应该等于维修费"

print_separator("所有测试通过!")
