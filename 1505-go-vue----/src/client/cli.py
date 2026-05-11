import sys
from typing import List

from client.api_client import APIClient
from client.commands.recipe import run_recipe_command
from client.commands.batch import run_batch_command
from client.commands.inspection import run_inspection_command
from client.commands.todo import run_todo_command
from client.commands.material import run_material_command
from client.commands.export import run_export_command


def print_help() -> None:
    print("食品加工厂生产管理系统 - 命令行客户端")
    print()
    print("用法: python -m client <资源> <命令> [参数...]")
    print()
    print("资源:")
    print("  recipe     配方管理")
    print("  batch      批次管理")
    print("  inspection 质检管理")
    print("  todo       待办事项")
    print("  material   原材料库存")
    print("  export     数据导出")
    print()
    print("配方管理命令:")
    print("  recipe list                    列出所有配方")
    print("  recipe get <ID>                查看配方详情")
    print("  recipe create <名称> <类别> <材料:用量:是否关键>...  创建配方")
    print()
    print("批次管理命令:")
    print("  batch list                     列出所有批次")
    print("  batch get <ID>                 查看批次详情")
    print("  batch create <配方ID> <计划数量> [实际数量]  创建批次")
    print("  batch update-actual <ID> <实际数量>  更新实际数量")
    print()
    print("质检管理命令:")
    print("  inspection list <批次ID>       查看批次质检记录")
    print("  inspection create <批次ID> <合格|不合格> [数值] [检验员] [备注]")
    print()
    print("待办事项命令:")
    print("  todo list                      列出所有待办")
    print("  todo get <ID>                  查看待办详情")
    print("  todo process <ID> <返工|降级|报废> [备注]")
    print()
    print("原材料命令:")
    print("  material list                  列出所有原材料库存")
    print("  material create <名称> <当前库存> <安全库存>")
    print("  material alert                 查看库存不足提醒")
    print("  material update-stock <ID> <新库存>")
    print()
    print("数据导出命令:")
    print("  export batches <开始日期> <结束日期>  导出批次汇总 (YYYY-MM-DD)")
    print()
    print("环境变量:")
    print("  SERVER_HOST    服务端主机 (默认: 127.0.0.1)")
    print("  SERVER_PORT    服务端端口 (默认: 8000)")


def main(argv: List[str]) -> int:
    if len(argv) < 2:
        print_help()
        return 0
    
    resource = argv[1]
    cmd = argv[2] if len(argv) > 2 else None
    args = argv[3:] if len(argv) > 3 else []
    
    client = APIClient()
    
    try:
        if resource == "recipe":
            if cmd is None:
                print("缺少命令")
                print("可用命令: list, create, get")
                return 1
            run_recipe_command(client, cmd, args)
        elif resource == "batch":
            if cmd is None:
                print("缺少命令")
                print("可用命令: list, create, get, update-actual")
                return 1
            run_batch_command(client, cmd, args)
        elif resource == "inspection":
            if cmd is None:
                print("缺少命令")
                print("可用命令: create, list")
                return 1
            run_inspection_command(client, cmd, args)
        elif resource == "todo":
            if cmd is None:
                print("缺少命令")
                print("可用命令: list, process, get")
                return 1
            run_todo_command(client, cmd, args)
        elif resource == "material":
            if cmd is None:
                print("缺少命令")
                print("可用命令: list, create, alert, update-stock")
                return 1
            run_material_command(client, cmd, args)
        elif resource == "export":
            if cmd is None:
                print("缺少命令")
                print("可用命令: batches")
                return 1
            run_export_command(client, cmd, args)
        elif resource in ["help", "-h", "--help"]:
            print_help()
            return 0
        else:
            print(f"未知的资源: {resource}")
            print_help()
            return 1
    except Exception as e:
        print(f"错误: {e}")
        return 1
    
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
