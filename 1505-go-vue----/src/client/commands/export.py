from typing import List

from client.api_client import APIClient


def cmd_export_batches(client: APIClient, args: List[str]) -> None:
    if len(args) != 2:
        print("用法: export batches <开始日期> <结束日期>")
        print("  日期格式: YYYY-MM-DD")
        print("示例: export batches 2024-01-01 2024-12-31")
        return
    
    start_date = args[0]
    end_date = args[1]
    
    try:
        table = client.get(f"/exports/batches?start_date={start_date}&end_date={end_date}")
        print(table)
    except Exception as e:
        print(f"导出失败: {e}")


def run_export_command(client: APIClient, cmd: str, args: List[str]) -> None:
    if cmd == "batches":
        cmd_export_batches(client, args)
    else:
        print(f"未知的导出命令: {cmd}")
        print("可用命令: batches")
