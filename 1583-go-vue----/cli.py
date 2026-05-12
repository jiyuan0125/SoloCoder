#!/usr/bin/env python3
import sys
import requests
import json
from datetime import datetime

BASE_URL = "http://localhost:8000/api"


def print_menu():
    print("\n" + "=" * 50)
    print("     火车站客服管理系统 - 命令行客户端")
    print("=" * 50)
    print("1. 问询记录管理")
    print("2. 失物登记")
    print("3. 拾物登记")
    print("4. 投诉管理")
    print("5. 查询指标统计")
    print("6. 执行手动匹配")
    print("0. 退出")
    print("=" * 50)


def inquiry_menu():
    while True:
        print("\n--- 问询记录管理 ---")
        print("1. 新建问询")
        print("2. 查询所有问询")
        print("3. 按编号查询")
        print("0. 返回")
        
        choice = input("请选择: ").strip()
        if choice == "0":
            break
        elif choice == "1":
            name = input("问询人姓名(可选): ").strip()
            phone = input("问询人电话(可选): ").strip()
            content = input("问询内容: ").strip()
            if not content:
                print("内容不能为空")
                continue
            
            data = {"content": content}
            if name:
                data["inquirer_name"] = name
            if phone:
                data["inquirer_phone"] = phone
            
            try:
                r = requests.post(f"{BASE_URL}/inquiries/", json=data)
                if r.status_code == 200:
                    print(f"创建成功！编号: {r.json()['record_number']}")
                else:
                    print(f"失败: {r.text}")
            except Exception as e:
                print(f"连接失败: {e}")
        
        elif choice == "2":
            try:
                r = requests.get(f"{BASE_URL}/inquiries/")
                for item in r.json():
                    print(f"{item['record_number']} - {item['content'][:40]}...")
            except Exception as e:
                print(f"失败: {e}")
        
        elif choice == "3":
            num = input("请输入编号(WX开头): ").strip()
            try:
                r = requests.get(f"{BASE_URL}/inquiries/by-number/{num}")
                if r.status_code == 200:
                    item = r.json()
                    print(f"编号: {item['record_number']}")
                    print(f"内容: {item['content']}")
                    if item['response']:
                        print(f"回复: {item['response']}")
                else:
                    print("未找到")
            except Exception as e:
                print(f"失败: {e}")


def lost_item_menu():
    print("\n--- 失物登记 ---")
    item_type = input("物品类型: ").strip()
    description = input("物品描述: ").strip()
    location = input("遗失地点: ").strip()
    value = float(input("物品价值(元): ") or "0")
    name = input("失主姓名: ").strip()
    phone = input("失主电话: ").strip()
    
    data = {
        "item_type": item_type,
        "description": description,
        "location": location,
        "value": value,
        "lost_time": datetime.now().isoformat()
    }
    if name:
        data["owner_name"] = name
    if phone:
        data["owner_phone"] = phone
    
    try:
        r = requests.post(f"{BASE_URL}/lost-found/lost/", json=data)
        if r.status_code == 200:
            print(f"登记成功！ID: {r.json()['id']}")
        else:
            print(f"失败: {r.text}")
    except Exception as e:
        print(f"连接失败: {e}")


def found_item_menu():
    print("\n--- 拾物登记 ---")
    item_type = input("物品类型: ").strip()
    description = input("物品描述: ").strip()
    location = input("拾取地点: ").strip()
    value = float(input("物品价值(元): ") or "0")
    name = input("拾主姓名: ").strip()
    phone = input("拾主电话: ").strip()
    
    data = {
        "item_type": item_type,
        "description": description,
        "location": location,
        "value": value,
        "found_time": datetime.now().isoformat()
    }
    if name:
        data["finder_name"] = name
    if phone:
        data["finder_phone"] = phone
    
    try:
        r = requests.post(f"{BASE_URL}/lost-found/found/", json=data)
        if r.status_code == 200:
            print(f"登记成功！ID: {r.json()['id']}")
        else:
            print(f"失败: {r.text}")
    except Exception as e:
        print(f"连接失败: {e}")


def complaint_menu():
    while True:
        print("\n--- 投诉管理 ---")
        print("1. 新建投诉")
        print("2. 开始处理投诉")
        print("3. 发送回复")
        print("4. 旅客确认关闭")
        print("5. 查询所有投诉")
        print("0. 返回")
        
        choice = input("请选择: ").strip()
        if choice == "0":
            break
        
        elif choice == "1":
            content = input("投诉内容: ").strip()
            if not content:
                print("内容不能为空")
                continue
            sev = input("严重程度(1-一般, 2-严重): ").strip()
            severity = "严重投诉" if sev == "2" else "一般投诉"
            
            data = {"content": content, "severity": severity}
            try:
                r = requests.post(f"{BASE_URL}/complaints/", json=data)
                if r.status_code == 200:
                    print(f"登记成功！ID: {r.json()['id']}")
                else:
                    print(f"失败: {r.text}")
            except Exception as e:
                print(f"连接失败: {e}")
        
        elif choice == "2":
            cid = input("投诉ID: ").strip()
            handler = input("处理人姓名: ").strip()
            try:
                r = requests.post(
                    f"{BASE_URL}/complaints/{cid}/start-processing",
                    params={"handler": handler}
                )
                print(f"状态: {r.status_code} - {r.text[:200]}")
            except Exception as e:
                print(f"失败: {e}")
        
        elif choice == "3":
            cid = input("投诉ID: ").strip()
            resp = input("回复内容: ").strip()
            try:
                r = requests.post(
                    f"{BASE_URL}/complaints/{cid}/send-response",
                    params={"response_content": resp}
                )
                print(f"状态: {r.status_code} - {r.text[:200]}")
            except Exception as e:
                print(f"失败: {e}")
        
        elif choice == "4":
            cid = input("投诉ID: ").strip()
            try:
                r = requests.post(f"{BASE_URL}/complaints/{cid}/confirm-close")
                print(f"状态: {r.status_code} - {r.text[:200]}")
            except Exception as e:
                print(f"失败: {e}")
        
        elif choice == "5":
            try:
                r = requests.get(f"{BASE_URL}/complaints/")
                for c in r.json():
                    print(f"ID:{c['id']} [{c['status']}] {c['severity']} - {c['content'][:30]}...")
            except Exception as e:
                print(f"失败: {e}")


def show_statistics():
    try:
        r = requests.get(f"{BASE_URL}/statistics/dashboard")
        s = r.json()
        print("\n" + "=" * 40)
        print("      今日统计")
        print("=" * 40)
        print(f"问询量:       {s['today_inquiry_count']}")
        print(f"失物登记:     {s['today_lost_items_count']}")
        print(f"拾物登记:     {s['today_found_items_count']}")
        print(f"今日匹配数:   {s['today_matches_count']}")
        print("=" * 40)
        print("      本月统计")
        print("=" * 40)
        print(f"投诉总量:     {s['monthly_complaint_count']}")
        print(f"平均处理时长: {s['avg_processing_hours']:.1f} 小时")
        print(f"逾期投诉:     {s['overdue_count']}")
        print("=" * 40)
    except Exception as e:
        print(f"获取统计失败: {e}")


def main():
    while True:
        print_menu()
        choice = input("请选择: ").strip()
        
        if choice == "0":
            print("再见！")
            sys.exit(0)
        elif choice == "1":
            inquiry_menu()
        elif choice == "2":
            lost_item_menu()
        elif choice == "3":
            found_item_menu()
        elif choice == "4":
            complaint_menu()
        elif choice == "5":
            show_statistics()
        elif choice == "6":
            print("手动触发匹配需要直接调用API...")
            print("(系统每天凌晨自动执行匹配)")


if __name__ == "__main__":
    main()
