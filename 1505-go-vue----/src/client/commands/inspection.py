from typing import List

from client.api_client import APIClient


def cmd_inspection_create(client: APIClient, args: List[str]) -> None:
    if len(args) < 3:
        print("用法: inspection create <批次ID> <结果:合格|不合格> [检测数值] [检验员] [备注]")
        print("  当结果为合格时，检测数值必填")
        return
    
    try:
        batch_id = int(args[0])
    except ValueError:
        print("错误: 批次ID必须是数字")
        return
    
    result = args[1]
    if result not in ["合格", "不合格"]:
        print('错误: 结果必须是 "合格" 或 "不合格"')
        return
    
    payload = {
        "batch_id": batch_id,
        "result": result,
    }
    
    if len(args) >= 3 and args[2]:
        try:
            payload["inspection_value"] = float(args[2])
        except ValueError:
            print("错误: 检测数值必须是数字")
            return
    
    if len(args) >= 4 and args[3]:
        payload["inspector"] = args[3]
    
    if len(args) >= 5 and args[4]:
        payload["notes"] = args[4]
    
    try:
        insp = client.post("/inspections", payload)
        print(f"质检记录创建成功，ID: {insp['id']}")
        if result == "不合格":
            print("  注意: 已自动生成待办通知品控主管处理")
    except Exception as e:
        print(f"创建失败: {e}")


def cmd_inspection_list(client: APIClient, args: List[str]) -> None:
    if len(args) != 1:
        print("用法: inspection list <批次ID>")
        return
    
    try:
        batch_id = int(args[0])
    except ValueError:
        print("错误: 批次ID必须是数字")
        return
    
    try:
        inspections = client.get(f"/inspections/batch/{batch_id}")
        if not inspections:
            print(f"批次 {batch_id} 没有质检记录")
            return
        
        print(f"批次 {batch_id} 共 {len(inspections)} 条质检记录:")
        for insp in inspections:
            val = f"{insp['inspection_value']}" if insp["inspection_value"] is not None else "无"
            inspector = insp["inspector"] or "未指定"
            print(f"  ID: {insp['id']}, 结果: {insp['result']}, 数值: {val}, 检验员: {inspector}")
            if insp["notes"]:
                print(f"    备注: {insp['notes']}")
    except Exception as e:
        print(f"查询失败: {e}")


def run_inspection_command(client: APIClient, cmd: str, args: List[str]) -> None:
    if cmd == "create":
        cmd_inspection_create(client, args)
    elif cmd == "list":
        cmd_inspection_list(client, args)
    else:
        print(f"未知的质检命令: {cmd}")
        print("可用命令: create, list")
