import argparse
import ast
import json
import os
import sys

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from client.api_client import APIClient


def print_json(data):
    print(json.dumps(data, ensure_ascii=False, indent=2))


def cmd_recipes(args):
    client = APIClient()
    if args.action == "create":
        materials = ast.literal_eval(args.raw_materials)
        result = client.create_recipe(args.name, args.category, materials)
        print_json(result)
    elif args.action == "list":
        result = client.list_recipes()
        print_json(result)
    elif args.action == "get":
        result = client.get_recipe(args.id)
        print_json(result)


def cmd_materials(args):
    client = APIClient()
    if args.action == "upsert":
        result = client.upsert_material(args.name, args.current_stock, args.safety_stock)
        print_json(result)
    elif args.action == "list":
        result = client.list_materials(args.low_stock)
        print_json(result)
    elif args.action == "get":
        result = client.get_material(args.id)
        print_json(result)


def cmd_batches(args):
    client = APIClient()
    if args.action == "create":
        result = client.create_batch(args.recipe_id, args.plan_quantity, args.actual_quantity)
        print_json(result)
    elif args.action == "list":
        result = client.list_batches(args.start, args.end)
        print_json(result)
    elif args.action == "get":
        result = client.get_batch(args.id)
        print_json(result)


def cmd_qc(args):
    client = APIClient()
    measured_value = None
    if args.measured_value is not None:
        measured_value = float(args.measured_value)
    result = client.add_quality_check(
        args.batch_id,
        args.inspector,
        args.result,
        measured_value,
        args.notes or "",
    )
    print_json(result)


def cmd_todos(args):
    client = APIClient()
    if args.action == "list":
        result = client.list_todos()
        print_json(result)
    elif args.action == "resolve":
        result = client.resolve_todo(args.id, args.disposition, args.resolved_by)
        print_json(result)


def cmd_export(args):
    client = APIClient()
    result = client.export_batches(args.start, args.end)
    print(result)


def main():
    parser = argparse.ArgumentParser(description="食品加工厂生产管理客户端")
    subparsers = parser.add_subparsers(dest="command", required=True)

    recipes = subparsers.add_parser("recipes", help="配方管理")
    recipes_sub = recipes.add_subparsers(dest="action", required=True)
    create_recipe = recipes_sub.add_parser("create", help="创建配方")
    create_recipe.add_argument("--name", required=True)
    create_recipe.add_argument("--category", required=True)
    create_recipe.add_argument(
        "--raw-materials",
        required=True,
        help='原材料列表，例如 \'[{"name":"面粉","quantity":100,"is_key":true},{"name":"糖","quantity":50,"is_key":false}]\'',
    )
    recipes_sub.add_parser("list", help="列出配方")
    get_recipe = recipes_sub.add_parser("get", help="获取配方")
    get_recipe.add_argument("--id", required=True)

    materials = subparsers.add_parser("materials", help="原材料管理")
    materials_sub = materials.add_subparsers(dest="action", required=True)
    upsert_mat = materials_sub.add_parser("upsert", help="创建或更新原材料")
    upsert_mat.add_argument("--name", required=True)
    upsert_mat.add_argument("--current-stock", type=float, required=True)
    upsert_mat.add_argument("--safety-stock", type=float, required=True)
    list_mat = materials_sub.add_parser("list", help="列出原材料")
    list_mat.add_argument("--low-stock", action="store_true", help="仅显示库存不足")
    get_mat = materials_sub.add_parser("get", help="获取原材料")
    get_mat.add_argument("--id", required=True)

    batches = subparsers.add_parser("batches", help="批次管理")
    batches_sub = batches.add_subparsers(dest="action", required=True)
    create_batch = batches_sub.add_parser("create", help="创建批次")
    create_batch.add_argument("--recipe-id", required=True)
    create_batch.add_argument("--plan-quantity", type=int, required=True)
    create_batch.add_argument("--actual-quantity", type=int, required=True)
    list_batch = batches_sub.add_parser("list", help="列出批次")
    list_batch.add_argument("--start", help="开始日期 YYYY-MM-DD")
    list_batch.add_argument("--end", help="结束日期 YYYY-MM-DD")
    get_batch = batches_sub.add_parser("get", help="获取批次")
    get_batch.add_argument("--id", required=True)

    qc = subparsers.add_parser("qc", help="质检管理")
    qc.add_argument("--batch-id", required=True)
    qc.add_argument("--inspector", required=True)
    qc.add_argument("--result", required=True, choices=["pass", "fail"])
    qc.add_argument("--measured-value", help="检测数值（合格时必填）")
    qc.add_argument("--notes", default="", help="备注")

    todos = subparsers.add_parser("todos", help="待办管理")
    todos_sub = todos.add_subparsers(dest="action", required=True)
    todos_sub.add_parser("list", help="列出待办")
    resolve_todo = todos_sub.add_parser("resolve", help="处理待办")
    resolve_todo.add_argument("--id", required=True)
    resolve_todo.add_argument("--disposition", required=True, choices=["rework", "downgrade", "scrap"])
    resolve_todo.add_argument("--resolved-by", required=True)

    export_cmd = subparsers.add_parser("export", help="数据导出")
    export_cmd.add_argument("--start", help="开始日期 YYYY-MM-DD")
    export_cmd.add_argument("--end", help="结束日期 YYYY-MM-DD")

    args = parser.parse_args()

    try:
        if args.command == "recipes":
            cmd_recipes(args)
        elif args.command == "materials":
            cmd_materials(args)
        elif args.command == "batches":
            cmd_batches(args)
        elif args.command == "qc":
            cmd_qc(args)
        elif args.command == "todos":
            cmd_todos(args)
        elif args.command == "export":
            cmd_export(args)
    except RuntimeError as e:
        print(f"错误: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
