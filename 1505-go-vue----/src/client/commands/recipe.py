import json
from typing import List

from client.api_client import APIClient


def cmd_recipe_list(client: APIClient, args: List[str]) -> None:
    recipes = client.get("/recipes")
    if not recipes:
        print("没有配方")
        return
    
    print(f"共 {len(recipes)} 个配方:")
    for r in recipes:
        print(f"  ID: {r['id']}, 名称: {r['name']}, 类别: {r['product_category']}")
        print(f"    原材料 ({len(r['ingredients'])} 种):")
        for ing in r["ingredients"]:
            key_mark = " [关键]" if ing["is_key_material"] else ""
            print(f"      - {ing['material_name']}: {ing['quantity']}{key_mark}")


def cmd_recipe_create(client: APIClient, args: List[str]) -> None:
    if len(args) < 4:
        print("用法: recipe create <名称> <产品类别> <材料1:用量1:是否关键> <材料2:用量2:是否关键> [...]")
        print("  是否关键: true 或 false")
        print("示例: recipe create 面包 烘焙 面粉:500:true 酵母:10:true 水:300:false")
        return
    
    name = args[0]
    category = args[1]
    ingredient_args = args[2:]
    
    if len(ingredient_args) < 2:
        print("错误: 至少需要两种原材料")
        return
    
    ingredients = []
    for arg in ingredient_args:
        parts = arg.split(":")
        if len(parts) != 3:
            print(f"错误: 原材料格式无效: {arg}")
            return
        try:
            material_name = parts[0]
            quantity = float(parts[1])
            is_key = parts[2].lower() == "true"
            ingredients.append({
                "material_name": material_name,
                "quantity": quantity,
                "is_key_material": is_key,
            })
        except ValueError:
            print(f"错误: 原材料用量必须是数字: {arg}")
            return
    
    payload = {
        "name": name,
        "product_category": category,
        "ingredients": ingredients,
    }
    
    try:
        recipe = client.post("/recipes", payload)
        print(f"配方创建成功，ID: {recipe['id']}")
    except Exception as e:
        print(f"创建失败: {e}")


def cmd_recipe_get(client: APIClient, args: List[str]) -> None:
    if len(args) != 1:
        print("用法: recipe get <配方ID>")
        return
    
    try:
        recipe_id = int(args[0])
    except ValueError:
        print("错误: 配方ID必须是数字")
        return
    
    try:
        recipe = client.get(f"/recipes/{recipe_id}")
        print(f"配方 ID: {recipe['id']}")
        print(f"名称: {recipe['name']}")
        print(f"产品类别: {recipe['product_category']}")
        print(f"创建时间: {recipe['created_at']}")
        print(f"原材料 ({len(recipe['ingredients'])} 种):")
        for ing in recipe["ingredients"]:
            key_mark = " [关键]" if ing["is_key_material"] else ""
            print(f"  - {ing['material_name']}: {ing['quantity']}{key_mark}")
    except Exception as e:
        print(f"查询失败: {e}")


def run_recipe_command(client: APIClient, cmd: str, args: List[str]) -> None:
    if cmd == "list":
        cmd_recipe_list(client, args)
    elif cmd == "create":
        cmd_recipe_create(client, args)
    elif cmd == "get":
        cmd_recipe_get(client, args)
    else:
        print(f"未知的配方命令: {cmd}")
        print("可用命令: list, create, get")
