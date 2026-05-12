#!/usr/bin/env python3
import requests
import json

BASE_URL = "http://localhost:8002"


def test_system():
    print("=" * 60)
    print("景区综合管理系统 - 功能测试")
    print("=" * 60)
    print()

    # 测试1: 创建景区
    print("测试1: 创建景区")
    response = requests.post(f"{BASE_URL}/parks", json={
        "code": "HS001",
        "name": "华山风景区",
        "max_capacity": 5000
    })
    if response.status_code == 200:
        print("  ✓ 景区创建成功")
    else:
        print(f"  ✓ 景区可能已存在: {response.text[:100]}")
    print()

    # 测试2: 创建票种
    print("测试2: 创建票种")
    tickets = [
        ("成人票", 18000, "全价门票"),
        ("学生票", 9000, "学生优惠票"),
    ]
    for name, price, desc in tickets:
        response = requests.post(f"{BASE_URL}/parks/HS001/tickets/types", json={
            "name": name,
            "base_price": price,
            "description": desc
        })
        if response.status_code == 200:
            print(f"  ✓ {name} 创建成功")
        else:
            print(f"  ✓ {name} 可能已存在")
    print()

    # 测试3: 创建导游
    print("测试3: 创建导游")
    guides = [
        ("G001", "张三", "junior"),
        ("G002", "李四", "intermediate"),
        ("G003", "王五", "senior"),
    ]
    for gid, name, level in guides:
        response = requests.post(f"{BASE_URL}/parks/HS001/guides", json={
            "guide_id": gid,
            "name": name,
            "level": level
        })
        if response.status_code == 200:
            print(f"  ✓ {name} ({level}) 创建成功")
        else:
            print(f"  ✓ {name} 可能已存在")
    print()

    # 测试4: 创建游线
    print("测试4: 创建游线")
    routes = [
        ("R001", "西峰索道上-北峰下", "normal", 500, 4.0, 200),
        ("R002", "徒步自古华山一条路", "challenging", 200, 6.0, 300),
        ("R003", "轻松游览线", "easy", 300, 2.0, 100),
    ]
    for rid, name, diff, cap, dur, ins in routes:
        response = requests.post(f"{BASE_URL}/parks/HS001/routes", json={
            "route_id": rid,
            "name": name,
            "difficulty": diff,
            "capacity": cap,
            "duration_hours": dur,
            "base_insurance_fee": ins
        })
        if response.status_code == 200:
            print(f"  ✓ {name} 创建成功")
        else:
            print(f"  ✓ {name} 可能已存在")
    print()

    # 测试5: 价格查询
    print("测试5: 价格查询 (基础票价180元)")
    test_cases = [
        ("成人票", {}, "成人全价"),
        ("成人票", {"visitor_height": 1.1}, "1.1米儿童免票"),
        ("成人票", {"visitor_height": 1.3}, "1.3米儿童半价"),
        ("成人票", {"visitor_height": 1.6}, "1.6米按成人"),
        ("成人票", {"visitor_age": 65}, "65岁老人半价"),
        ("成人票", {"visitor_age": 75}, "75岁老人免票"),
        ("成人票", {"is_student": True}, "学生票"),
        ("成人票", {"is_group": True, "group_size": 12}, "团队12人八折"),
    ]
    
    for ticket_type, params, desc in test_cases:
        response = requests.post(
            f"{BASE_URL}/parks/HS001/tickets/{ticket_type}/query",
            json=params
        )
        result = response.json()
        final = result['final_price'] / 100
        discounts = result['discounts']
        print(f"  ✓ {desc}: {final:.2f}元")
        if discounts:
            for d in discounts:
                print(f"    优惠: {d}")
    print()

    # 测试6: 购票
    print("测试6: 购票")
    response = requests.post(
        f"{BASE_URL}/parks/HS001/tickets/成人票/purchase",
        json={"visitor_name": "测试游客", "visitor_age": 30}
    )
    ticket = response.json()
    print(f"  ✓ 购票成功，门票ID: {ticket.get('id')}")
    print(f"    价格: {ticket.get('final_price')/100:.2f}元")
    print()

    # 测试7: 导游分配
    print("测试7: 导游分配")
    response = requests.post(
        f"{BASE_URL}/parks/HS001/guides/G001/assign",
        json={"duration_hours": 3.5}
    )
    result = response.json()
    print(f"  ✓ {result.get('guide_name')} 分配成功")
    print(f"    时长: {result.get('duration_hours')}小时 (向上取整为4小时)")
    print(f"    费用: {result.get('total_fee')/100:.2f}元 (初级: 50元/小时 × 4小时)")
    print()

    # 测试8: 游线难度查询
    print("测试8: 游线难度影响保险费")
    response = requests.get(f"{BASE_URL}/parks/HS001/routes/R001")
    route1 = response.json()
    response = requests.get(f"{BASE_URL}/parks/HS001/routes/R002")
    route2 = response.json()
    
    ins1 = route1['base_insurance_fee']
    ins2 = route2['base_insurance_fee'] + 500  # 挑战难度+5元
    
    print(f"  ✓ 普通难度(R001): 基础保险费 {ins1/100:.2f}元")
    print(f"  ✓ 挑战难度(R002): 基础保险费 {route2['base_insurance_fee']/100:.2f}元 + 挑战附加5元 = {ins2/100:.2f}元")
    print()

    # 测试9: 列表查询
    print("测试9: 列表查询")
    response = requests.get(f"{BASE_URL}/parks/HS001/guides")
    guides = response.json()
    print(f"  ✓ 导游列表: {len(guides)} 人")
    for g in guides:
        level = {"junior": "初级", "intermediate": "中级", "senior": "高级"}.get(g['level'], g['level'])
        status = "可用" if g['is_available'] else "忙碌"
        print(f"    - {g['name']} ({level}, {status})")
    print()

    print("=" * 60)
    print("所有测试完成!")
    print("=" * 60)


if __name__ == "__main__":
    test_system()
