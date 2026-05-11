from typing import List

from client.api_client import APIClient


VALID_STATUS = {"返工": "返工", "降级": "降级", "报废": "报废"}


def cmd_todo_list(client: APIClient, args: List[str]) -> None:
    try:
        todos = client.get("/todos")
        if not todos:
            print("没有待办事项")
            return
        
        print(f"共 {len(todos)} 个待办事项:")
        for t in todos:
            print(f"  ID: {t['id']}, 批次ID: {t['batch_id']}, 状态: {t['status']}")
            if t["notes"]:
                print(f"    备注: {t['notes']}")
    except Exception as e:
        print(f"查询失败: {e}")


def cmd_todo_process(client: APIClient, args: List[str]) -> None:
    if len(args) < 2:
        print("用法: todo process <待办ID> <状态:返工|降级|报废> [备注]")
        return
    
    try:
        todo_id = int(args[0])
    except ValueError:
        print("错误: 待办ID必须是数字")
        return
    
    status = args[1]
    if status not in VALID_STATUS:
        print('错误: 状态必须是 "返工"、"降级" 或 "报废"')
        return
    
    payload = {"status": status}
    if len(args) >= 3:
        payload["notes"] = args[2]
    
    try:
        todo = client.patch(f"/todos/{todo_id}", payload)
        print(f"待办处理成功，新状态: {todo['status']}")
    except Exception as e:
        print(f"处理失败: {e}")


def cmd_todo_get(client: APIClient, args: List[str]) -> None:
    if len(args) != 1:
        print("用法: todo get <待办ID>")
        return
    
    try:
        todo_id = int(args[0])
    except ValueError:
        print("错误: 待办ID必须是数字")
        return
    
    try:
        todo = client.get(f"/todos/{todo_id}")
        print(f"待办 ID: {todo['id']}")
        print(f"批次ID: {todo['batch_id']}")
        print(f"状态: {todo['status']}")
        if todo["assigned_to"]:
            print(f"负责人: {todo['assigned_to']}")
        if todo["notes"]:
            print(f"备注: {todo['notes']}")
        print(f"创建时间: {todo['created_at']}")
        print(f"更新时间: {todo['updated_at']}")
    except Exception as e:
        print(f"查询失败: {e}")


def run_todo_command(client: APIClient, cmd: str, args: List[str]) -> None:
    if cmd == "list":
        cmd_todo_list(client, args)
    elif cmd == "process":
        cmd_todo_process(client, args)
    elif cmd == "get":
        cmd_todo_get(client, args)
    else:
        print(f"未知的待办命令: {cmd}")
        print("可用命令: list, process, get")
