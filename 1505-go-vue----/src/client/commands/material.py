from typing import List

from client.api_client import APIClient


def cmd_material_list(client: APIClient, args: List[str]) -> None:
    try:
        materials = client.get("/materials")
        if not materials:
            print("没有原材料库存")
            return
        
        print(f"共 {len(materials)} 种原材料:")
        for m in materials:
            alert = ""
            if m["current_stock"] < m["safety_stock"]:
                alert = " [库存不足!]"
            print(f"  ID: {m['id']}, 名称: {m['name']}, 当前库存: {m['current_stock']:.2f}, 安全库存: {m['safety_stock']:.2f}{alert}")
    except Exception as e:
        print(f"查询失败: {e}")


def cmd_material_create(client: APIClient, args: List[str]) -> None:
    if len(args) != 3:
        print("用法: material create <名称> <当前库存> <安全库存>")
        return
    
    name = args[0]
    try:
        current = float(args[1])
        safety = float(args[2])
    except ValueError:
        print("错误: 库存数量必须是数字")
        return
    
    payload = {
        "name": name,
        "current_stock": current,
        "safety_stock": safety,
    }
    
    try:
        material = client.post("/materials", payload)
        print(f"原材料创建成功，ID: {material['id']}")
    except Exception as e:
        print(f"创建失败: {e}")


def cmd_material_alert(client: APIClient, args: List[str]) -> None:
    try:
        materials = client.get("/materials/alert/low-stock")
        if not materials:
            print("所有原材料库存充足")
            return
        
        print(f"库存不足的原材料 ({len(materials)} 种):")
        for m in materials:
            print(f"  ID: {m['id']}, 名称: {m['name']}, 当前: {m['current_stock']:.2f}, 安全线: {m['safety_stock']:.2f}")
            print(f"    建议补货数量: {m['safety_stock'] - m['current_stock']:.2f}")
    except Exception as e:
        print(f"查询失败: {e}")


def cmd_material_update_stock(client: APIClient, args: List[str]) -> None:
    if len(args) != 2:
        print("用法: material update-stock <材料ID> <新库存数量>")
        return
    
    try:
        material_id = int(args[0])
        stock = float(args[1])
    except ValueError:
        print("错误: 材料ID和库存数量必须是数字")
        return
    
    try:
        material = client.patch(f"/materials/{material_id}/stock", {"current_stock": stock})
        print(f"库存更新成功，新值: {material['current_stock']:.2f}")
        if material["current_stock"] < material["safety_stock"]:
            print(f"  警告: 库存低于安全线 ({material['safety_stock']:.2f})")
    except Exception as e:
        print(f"更新失败: {e}")


def run_material_command(client: APIClient, cmd: str, args: List[str]) -> None:
    if cmd == "list":
        cmd_material_list(client, args)
    elif cmd == "create":
        cmd_material_create(client, args)
    elif cmd == "alert":
        cmd_material_alert(client, args)
    elif cmd == "update-stock":
        cmd_material_update_stock(client, args)
    else:
        print(f"未知的原材料命令: {cmd}")
        print("可用命令: list, create, alert, update-stock")
