from typing import List

from client.api_client import APIClient


def cmd_batch_list(client: APIClient, args: List[str]) -> None:
    batches = client.get("/batches")
    if not batches:
        print("没有批次")
        return
    
    print(f"共 {len(batches)} 个批次:")
    for b in batches:
        actual = f"{b['actual_quantity']:.2f}" if b["actual_quantity"] is not None else "未填写"
        status = "已通过质检" if b["has_passed_inspection"] else "未通过质检"
        print(f"  ID: {b['id']}, 配方ID: {b['recipe_id']}, 计划: {b['planned_quantity']:.2f}, 实际: {actual}, {status}")


def cmd_batch_create(client: APIClient, args: List[str]) -> None:
    if len(args) < 2:
        print("用法: batch create <配方ID> <计划数量> [实际数量]")
        return
    
    try:
        recipe_id = int(args[0])
        planned = float(args[1])
    except ValueError:
        print("错误: 配方ID和计划数量必须是数字")
        return
    
    payload = {
        "recipe_id": recipe_id,
        "planned_quantity": planned,
    }
    
    if len(args) >= 3:
        try:
            actual = float(args[2])
            payload["actual_quantity"] = actual
        except ValueError:
            print("错误: 实际数量必须是数字")
            return
    
    try:
        batch = client.post("/batches", payload)
        print(f"批次创建成功，ID: {batch['id']}")
    except Exception as e:
        print(f"创建失败: {e}")


def cmd_batch_get(client: APIClient, args: List[str]) -> None:
    if len(args) != 1:
        print("用法: batch get <批次ID>")
        return
    
    try:
        batch_id = int(args[0])
    except ValueError:
        print("错误: 批次ID必须是数字")
        return
    
    try:
        batch = client.get(f"/batches/{batch_id}")
        actual = f"{batch['actual_quantity']:.2f}" if batch["actual_quantity"] is not None else "未填写"
        status = "已通过质检" if batch["has_passed_inspection"] else "未通过质检"
        print(f"批次 ID: {batch['id']}")
        print(f"配方ID: {batch['recipe_id']}")
        print(f"计划数量: {batch['planned_quantity']:.2f}")
        print(f"实际数量: {actual}")
        print(f"状态: {status}")
        print(f"创建时间: {batch['created_at']}")
    except Exception as e:
        print(f"查询失败: {e}")


def cmd_batch_update_actual(client: APIClient, args: List[str]) -> None:
    if len(args) != 2:
        print("用法: batch update-actual <批次ID> <实际数量>")
        return
    
    try:
        batch_id = int(args[0])
        actual = float(args[1])
    except ValueError:
        print("错误: 批次ID和实际数量必须是数字")
        return
    
    try:
        batch = client.patch(f"/batches/{batch_id}/actual-quantity", {"actual_quantity": actual})
        print(f"实际数量更新成功，新值: {batch['actual_quantity']:.2f}")
    except Exception as e:
        print(f"更新失败: {e}")


def run_batch_command(client: APIClient, cmd: str, args: List[str]) -> None:
    if cmd == "list":
        cmd_batch_list(client, args)
    elif cmd == "create":
        cmd_batch_create(client, args)
    elif cmd == "get":
        cmd_batch_get(client, args)
    elif cmd == "update-actual":
        cmd_batch_update_actual(client, args)
    else:
        print(f"未知的批次命令: {cmd}")
        print("可用命令: list, create, get, update-actual")
