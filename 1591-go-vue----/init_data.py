#!/usr/bin/env python3
import requests
import json

BASE_URL = "http://localhost:8000"


def create_park(code, name, max_capacity):
    response = requests.post(f"{BASE_URL}/parks", json={
        "code": code,
        "name": name,
        "max_capacity": max_capacity
    })
    if response.status_code == 200:
        print(f"✓ 景区创建成功: {name}")
        return response.json()
    else:
        print(f"✗ 景区创建失败: {response.json().get('detail')}")
        return None


def create_ticket_type(park_code, name, base_price, description=""):
    response = requests.post(f"{BASE_URL}/parks/{park_code}/tickets/types", json={
        "name": name,
        "base_price": base_price,
        "description": description
    })
    if response.status_code == 200:
        print(f"  ✓ 票种创建成功: {name} ({base_price/100:.2f}元)")
        return response.json()
    else:
        print(f"  ✗ 票种创建失败: {response.json().get('detail')}")
        return None


def create_guide(park_code, guide_id, name, level):
    response = requests.post(f"{BASE_URL}/parks/{park_code}/guides", json={
        "guide_id": guide_id,
        "name": name,
        "level": level
    })
    if response.status_code == 200:
        level_name = {"junior": "初级", "intermediate": "中级", "senior": "高级"}
        print(f"  ✓ 导游创建成功: {name} ({level_name.get(level, level)})")
        return response.json()
    else:
        print(f"  ✗ 导游创建失败: {response.json().get('detail')}")
        return None


def create_route(park_code, route_id, name, difficulty, capacity, duration_hours, base_insurance_fee=0):
    response = requests.post(f"{BASE_URL}/parks/{park_code}/routes", json={
        "route_id": route_id,
        "name": name,
        "difficulty": difficulty,
        "capacity": capacity,
        "duration_hours": duration_hours,
        "base_insurance_fee": base_insurance_fee
    })
    if response.status_code == 200:
        diff_name = {"easy": "简单", "normal": "普通", "challenging": "挑战"}
        print(f"  ✓ 游线创建成功: {name} ({diff_name.get(difficulty, difficulty)})")
        return response.json()
    else:
        print(f"  ✗ 游线创建失败: {response.json().get('detail')}")
        return None


def main():
    print("=" * 60)
    print("初始化景区综合管理系统数据")
    print("=" * 60)
    print()

    print("创建景区...")
    huashan = create_park("HS001", "华山风景区", 5000)
    print()

    if huashan:
        print("创建票种...")
        create_ticket_type("HS001", "成人票", 18000, "全价门票")
        create_ticket_type("HS001", "儿童票", 9000, "1.2-1.5米儿童")
        create_ticket_type("HS001", "学生票", 9000, "全日制学生，凭学生证")
        print()

        print("创建导游...")
        create_guide("HS001", "G001", "张三", "junior")
        create_guide("HS001", "G002", "李四", "intermediate")
        create_guide("HS001", "G003", "王五", "senior")
        print()

        print("创建游线...")
        create_route("HS001", "R001", "西峰索道上-北峰下", "normal", 500, 4.0, 200)
        create_route("HS001", "R002", "徒步自古华山一条路", "challenging", 200, 6.0, 300)
        create_route("HS001", "R003", "轻松游览线", "easy", 300, 2.0, 100)
        print()

    print("=" * 60)
    print("初始化完成!")
    print("=" * 60)
    print()
    print("可用命令示例:")
    print("  python cli.py list-parks")
    print("  python cli.py list-tickets HS001")
    print("  python cli.py list-guides HS001")
    print("  python cli.py list-routes HS001")
    print("  python cli.py query-price HS001 成人票 --age 65")
    print("  python cli.py purchase HS001 成人票 --visitor-name 测试游客")
    print()


if __name__ == "__main__":
    main()
