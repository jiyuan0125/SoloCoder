import requests
import json

BASE = "http://localhost:8090/api"

def test():
    print("=" * 50)
    print("测试1：创建项目 - 预算精确匹配")
    print("=" * 50)
    r = requests.post(f"{BASE}/projects", json={
        "name": "API测试项目",
        "principal": "测试人员",
        "start_date": "2026-01-01",
        "end_date": "2026-12-31",
        "total_amount": "500000.50",
        "budget": {
            "equipment": "200000.25",
            "material": "100000.00",
            "travel": "80000.00",
            "labor": "70000.00",
            "expert": "30000.00",
            "other": "20000.25"
        }
    })
    print(f"状态码: {r.status_code}")
    print(f"响应: {r.json()}")
    assert r.status_code == 201, "应该创建成功"
    project_id = r.json()["id"]
    print(f"项目ID: {project_id}")
    print()

    print("=" * 50)
    print("测试2：创建项目 - 预算差一分钱")
    print("=" * 50)
    r = requests.post(f"{BASE}/projects", json={
        "name": "应该失败的项目",
        "principal": "测试",
        "start_date": "2026-01-01",
        "end_date": "2026-12-31",
        "total_amount": "10000.00",
        "budget": {
            "equipment": "3000.00",
            "material": "2000.00",
            "travel": "1500.00",
            "labor": "1500.00",
            "expert": "1000.00",
            "other": "1000.01"
        }
    })
    print(f"状态码: {r.status_code}")
    print(f"响应: {r.json()}")
    assert r.status_code == 400, "应该返回400"
    print()

    print("=" * 50)
    print("测试3：创建报销单 - 正常金额")
    print("=" * 50)
    r = requests.post(f"{BASE}/reimbursements", json={
        "project_id": project_id,
        "category": "equipment",
        "amount": "50000.00",
        "reason": "购买服务器",
        "date": "2026-05-12"
    })
    print(f"状态码: {r.status_code}")
    print(f"响应: {r.json()}")
    assert r.status_code == 201, "应该创建成功"
    reimb_id = r.json()["id"]
    print(f"报销单ID: {reimb_id}")
    print()

    print("=" * 50)
    print("测试4：创建报销单 - 金额为0")
    print("=" * 50)
    r = requests.post(f"{BASE}/reimbursements", json={
        "project_id": project_id,
        "category": "material",
        "amount": "0.00",
        "reason": "零金额测试",
        "date": "2026-05-12"
    })
    print(f"状态码: {r.status_code}")
    print(f"响应: {r.json()}")
    assert r.status_code == 400, "应该返回400"
    print()

    print("=" * 50)
    print("测试5：创建报销单 - 金额为负数")
    print("=" * 50)
    r = requests.post(f"{BASE}/reimbursements", json={
        "project_id": project_id,
        "category": "material",
        "amount": "-100.00",
        "reason": "负数测试",
        "date": "2026-05-12"
    })
    print(f"状态码: {r.status_code}")
    print(f"响应: {r.json()}")
    assert r.status_code == 400, "应该返回400"
    print()

    print("=" * 50)
    print("测试6：查看待办审批列表")
    print("=" * 50)
    r = requests.get(f"{BASE}/reimbursements/pending")
    print(f"状态码: {r.status_code}")
    data = r.json()
    print(f"待办数量: {len(data)}")
    print(f"列表: {json.dumps(data, ensure_ascii=False, indent=2)}")
    assert len(data) >= 1, "应该至少有一个待办"
    print()

    print("=" * 50)
    print("测试7：审批报销单")
    print("=" * 50)
    r = requests.post(f"{BASE}/reimbursements/{reimb_id}/approve")
    print(f"状态码: {r.status_code}")
    print(f"响应: {r.json()}")
    assert r.status_code == 200, "应该审批成功"
    print()

    print("=" * 50)
    print("测试8：已批准的报销单不能再修改")
    print("=" * 50)
    r = requests.post(f"{BASE}/reimbursements/{reimb_id}/approve")
    print(f"状态码: {r.status_code}")
    print(f"响应: {r.json()}")
    assert r.status_code == 400, "应该返回400"
    print()

    print("=" * 50)
    print("测试9：项目结题")
    print("=" * 50)
    r = requests.post(f"{BASE}/projects/{project_id}/close")
    print(f"状态码: {r.status_code}")
    print(f"响应: {r.json()}")
    assert r.status_code == 200, "应该结题成功"
    print()

    print("=" * 50)
    print("测试10：已结题项目不能生成报销单")
    print("=" * 50)
    r = requests.post(f"{BASE}/reimbursements", json={
        "project_id": project_id,
        "category": "material",
        "amount": "1000.00",
        "reason": "已结题项目测试",
        "date": "2026-05-12"
    })
    print(f"状态码: {r.status_code}")
    print(f"响应: {r.json()}")
    assert r.status_code == 400, "应该返回400"
    print()

    print("=" * 50)
    print("测试11：生成决算报告")
    print("=" * 50)
    r = requests.get(f"{BASE}/projects/{project_id}/final-report")
    print(f"状态码: {r.status_code}")
    data = r.json()
    print(f"响应: {json.dumps(data, ensure_ascii=False, indent=2)}")
    assert r.status_code == 200, "应该生成成功"
    print()

    print("=" * 50)
    print("测试12：未结题项目不能生成决算报告")
    print("=" * 50)
    r = requests.get(f"{BASE}/projects/P202605122/final-report")
    print(f"状态码: {r.status_code}")
    print(f"响应: {r.json()}")
    assert r.status_code == 400, "应该返回400"
    print()

    print("=" * 50)
    print("所有测试通过！")
    print("=" * 50)

if __name__ == "__main__":
    test()
