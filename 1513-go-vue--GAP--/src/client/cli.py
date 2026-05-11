import sys
import argparse

from .api_client import APIClient

OPERATION_TYPES = {
    "fertilization": "施肥",
    "pesticide": "用药",
    "irrigation": "灌溉",
}

HARVEST_STATUSES = {
    "planned": "计划中",
    "in_progress": "进行中",
    "completed": "已完成",
    "cancelled": "已取消",
}

TODO_TYPES = {
    "safety_interval": "安全间隔期",
    "gap_inspection": "GAP年检",
}

TODO_STATUSES = {
    "pending": "待处理",
    "completed": "已完成",
    "overdue": "已逾期",
}


class GAPCLI:
    def __init__(self):
        self.client = APIClient()

    def run(self):
        parser = self._create_parser()
        args = parser.parse_args()

        if hasattr(args, 'func'):
            try:
                args.func(args)
            except RuntimeError as e:
                print(f"错误: {e}")
                sys.exit(1)
            except Exception as e:
                print(f"意外错误: {e}")
                sys.exit(1)
        else:
            parser.print_help()

    def _create_parser(self) -> argparse.ArgumentParser:
        parser = argparse.ArgumentParser(
            prog="gap-cli",
            description="中药材GAP生产管理系统命令行客户端"
        )
        subparsers = parser.add_subparsers(dest="command", help="可用命令")

        self._add_health_cmd(subparsers)
        self._add_plot_cmds(subparsers)
        self._add_operation_cmds(subparsers)
        self._add_harvest_cmds(subparsers)
        self._add_processing_cmds(subparsers)
        self._add_todo_cmds(subparsers)
        self._add_export_cmds(subparsers)

        return parser

    def _add_health_cmd(self, subparsers):
        health_parser = subparsers.add_parser("health", help="检查服务端健康状态")
        health_parser.set_defaults(func=self._cmd_health)

    def _add_plot_cmds(self, subparsers):
        plot_create = subparsers.add_parser("plot-create", help="创建地块")
        plot_create.add_argument("--name", required=True, help="地块名称")
        plot_create.add_argument("--area", type=float, required=True, help="地块面积(亩)")
        plot_create.add_argument("--variety", required=True, help="种植品种")
        plot_create.add_argument("--planting-date", required=True, help="种植日期(YYYY-MM-DD)")
        plot_create.add_argument("--expected-harvest", required=True, help="预计采收日期(YYYY-MM-DD)")
        plot_create.add_argument("--notes", help="备注")
        plot_create.set_defaults(func=self._cmd_plot_create)

        plot_list = subparsers.add_parser("plot-list", help="列出所有地块")
        plot_list.set_defaults(func=self._cmd_plot_list)

        plot_get = subparsers.add_parser("plot-get", help="获取地块详情")
        plot_get.add_argument("plot_id", help="地块ID")
        plot_get.set_defaults(func=self._cmd_plot_get)

        plot_delete = subparsers.add_parser("plot-delete", help="删除地块")
        plot_delete.add_argument("plot_id", help="地块ID")
        plot_delete.set_defaults(func=self._cmd_plot_delete)

    def _add_operation_cmds(self, subparsers):
        op_create = subparsers.add_parser("op-create", help="记录农事操作")
        op_create.add_argument("--plot-id", required=True, help="地块ID")
        op_create.add_argument("--type", required=True, choices=OPERATION_TYPES.keys(), help="操作类型")
        op_create.add_argument("--date", required=True, help="操作日期(YYYY-MM-DD)")
        op_create.add_argument("--details", required=True, help="操作详情")
        op_create.add_argument("--quantity", help="用量")
        op_create.add_argument("--safety-interval", type=int, help="安全间隔期(天)")
        op_create.set_defaults(func=self._cmd_op_create)

        op_list = subparsers.add_parser("op-list", help="列出地块农事操作")
        op_list.add_argument("plot_id", help="地块ID")
        op_list.set_defaults(func=self._cmd_op_list)

    def _add_harvest_cmds(self, subparsers):
        hv_create = subparsers.add_parser("hv-create", help="创建采收任务")
        hv_create.add_argument("--plot-id", required=True, help="地块ID")
        hv_create.add_argument("--planned-date", required=True, help="计划采收日期(YYYY-MM-DD)")
        hv_create.add_argument("--quantity", type=float, help="采收数量")
        hv_create.add_argument("--unit", default="kg", help="单位(默认kg)")
        hv_create.add_argument("--notes", help="备注")
        hv_create.set_defaults(func=self._cmd_hv_create)

        hv_list = subparsers.add_parser("hv-list", help="列出采收任务")
        hv_list.add_argument("--plot-id", help="按地块筛选")
        hv_list.set_defaults(func=self._cmd_hv_list)

        hv_update = subparsers.add_parser("hv-update", help="更新采收任务")
        hv_update.add_argument("harvest_id", help="采收ID")
        hv_update.add_argument("--status", choices=HARVEST_STATUSES.keys(), help="状态")
        hv_update.add_argument("--actual-date", help="实际采收日期(YYYY-MM-DD)")
        hv_update.add_argument("--quantity", type=float, help="采收数量")
        hv_update.add_argument("--quality", help="质量状态")
        hv_update.add_argument("--notes", help="备注")
        hv_update.set_defaults(func=self._cmd_hv_update)

    def _add_processing_cmds(self, subparsers):
        pr_create = subparsers.add_parser("pr-create", help="创建加工记录")
        pr_create.add_argument("--harvest-id", required=True, help="采收ID")
        pr_create.add_argument("--batch", required=True, help="加工批号")
        pr_create.add_argument("--input-qty", type=float, required=True, help="投料量")
        pr_create.add_argument("--output-qty", type=float, help="成品量")
        pr_create.add_argument("--unit", default="kg", help="单位(默认kg)")
        pr_create.add_argument("--date", required=True, help="加工日期(YYYY-MM-DD)")
        pr_create.add_argument("--details", required=True, help="加工详情")
        pr_create.set_defaults(func=self._cmd_pr_create)

        pr_list = subparsers.add_parser("pr-list", help="列出加工记录")
        pr_list.add_argument("--harvest-id", help="按采收筛选")
        pr_list.set_defaults(func=self._cmd_pr_list)

        pr_update = subparsers.add_parser("pr-update", help="更新加工记录")
        pr_update.add_argument("processing_id", help="加工ID")
        pr_update.add_argument("--output-qty", type=float, help="成品量")
        pr_update.add_argument("--details", help="加工详情")
        pr_update.set_defaults(func=self._cmd_pr_update)

    def _add_todo_cmds(self, subparsers):
        todo_list = subparsers.add_parser("todo-list", help="列出待办事项")
        todo_list.set_defaults(func=self._cmd_todo_list)

        todo_done = subparsers.add_parser("todo-done", help="标记待办完成")
        todo_done.add_argument("todo_id", help="待办ID")
        todo_done.set_defaults(func=self._cmd_todo_done)

        todo_refresh = subparsers.add_parser("todo-refresh", help="刷新待办事项")
        todo_refresh.set_defaults(func=self._cmd_todo_refresh)

    def _add_export_cmds(self, subparsers):
        exp_batch = subparsers.add_parser("export-batch", help="导出单批次记录")
        exp_batch.add_argument("harvest_id", help="采收ID")
        exp_batch.add_argument("--output", "-o", help="输出文件路径")
        exp_batch.set_defaults(func=self._cmd_export_batch)

        exp_all = subparsers.add_parser("export-all", help="导出所有批次记录")
        exp_all.add_argument("--output", "-o", help="输出文件路径")
        exp_all.set_defaults(func=self._cmd_export_all)

    def _cmd_health(self, args):
        result = self.client.health_check()
        print(f"状态: {result['status']}")
        print(f"服务: {result['service']}")

    def _cmd_plot_create(self, args):
        data = {
            "name": args.name,
            "area": args.area,
            "variety": args.variety,
            "planting_date": args.planting_date,
            "expected_harvest_date": args.expected_harvest,
        }
        if args.notes:
            data["notes"] = args.notes
        result = self.client.create_plot(data)
        self._print_plot(result)

    def _cmd_plot_list(self, args):
        plots = self.client.list_plots()
        if not plots:
            print("暂无地块记录")
            return
        print(f"{'ID':<15} {'名称':<20} {'品种':<15} {'面积(亩)':<10} {'种植日期':<12} {'预计采收':<12}")
        print("-" * 80)
        for p in plots:
            print(f"{p['id']:<15} {p['name']:<20} {p['variety']:<15} {p['area']:<10} {p['planting_date']:<12} {p['expected_harvest_date']:<12}")

    def _cmd_plot_get(self, args):
        plot = self.client.get_plot(args.plot_id)
        self._print_plot(plot)

    def _cmd_plot_delete(self, args):
        result = self.client.delete_plot(args.plot_id)
        print(result["message"])

    def _cmd_op_create(self, args):
        data = {
            "plot_id": args.plot_id,
            "operation_type": args.type,
            "operation_date": args.date,
            "details": args.details,
        }
        if args.quantity:
            data["quantity"] = args.quantity
        if args.safety_interval:
            data["safety_interval_days"] = args.safety_interval
        result = self.client.create_operation(data)
        self._print_operation(result)

    def _cmd_op_list(self, args):
        ops = self.client.list_operations(args.plot_id)
        if not ops:
            print("暂无农事操作记录")
            return
        print(f"{'ID':<15} {'类型':<8} {'日期':<12} {'详情':<30} {'用量':<15} {'间隔期':<8}")
        print("-" * 90)
        for op in ops:
            op_type = OPERATION_TYPES.get(op['operation_type'], op['operation_type'])
            qty = op.get('quantity', '-') or '-'
            interval = op.get('safety_interval_days', '-') or '-'
            if interval != '-':
                interval = f"{interval}天"
            print(f"{op['id']:<15} {op_type:<8} {op['operation_date']:<12} {op['details'][:28]:<30} {qty:<15} {interval:<8}")

    def _cmd_hv_create(self, args):
        data = {
            "plot_id": args.plot_id,
            "planned_date": args.planned_date,
        }
        if args.quantity:
            data["quantity"] = args.quantity
            data["unit"] = args.unit
        if args.notes:
            data["notes"] = args.notes
        result = self.client.create_harvest(data)
        self._print_harvest(result)

    def _cmd_hv_list(self, args):
        harvests = self.client.list_harvests(args.plot_id)
        if not harvests:
            print("暂无采收任务")
            return
        print(f"{'ID':<15} {'地块ID':<15} {'状态':<10} {'计划日期':<12} {'实际日期':<12} {'数量':<10} {'单位':<6}")
        print("-" * 80)
        for h in harvests:
            status = HARVEST_STATUSES.get(h['status'], h['status'])
            actual = h.get('actual_date', '-') or '-'
            qty = h.get('quantity', '-') or '-'
            unit = h.get('unit', '-') or '-'
            print(f"{h['id']:<15} {h['plot_id']:<15} {status:<10} {h['planned_date']:<12} {actual:<12} {qty:<10} {unit:<6}")

    def _cmd_hv_update(self, args):
        data = {}
        if args.status:
            data["status"] = args.status
        if args.actual_date:
            data["actual_date"] = args.actual_date
        if args.quantity is not None:
            data["quantity"] = args.quantity
        if args.quality:
            data["quality_status"] = args.quality
        if args.notes:
            data["notes"] = args.notes
        result = self.client.update_harvest(args.harvest_id, data)
        self._print_harvest(result)

    def _cmd_pr_create(self, args):
        data = {
            "harvest_id": args.harvest_id,
            "batch_number": args.batch,
            "input_quantity": args.input_qty,
            "processing_date": args.date,
            "details": args.details,
            "unit": args.unit,
        }
        if args.output_qty is not None:
            data["output_quantity"] = args.output_qty
        result = self.client.create_processing(data)
        self._print_processing(result)

    def _cmd_pr_list(self, args):
        processings = self.client.list_processings(args.harvest_id)
        if not processings:
            print("暂无加工记录")
            return
        print(f"{'ID':<15} {'采收ID':<15} {'批号':<20} {'日期':<12} {'投料':<10} {'成品':<10} {'单位':<6}")
        print("-" * 90)
        for p in processings:
            out = p.get('output_quantity', '-') or '-'
            print(f"{p['id']:<15} {p['harvest_id']:<15} {p['batch_number']:<20} {p['processing_date']:<12} {p['input_quantity']:<10} {out:<10} {p.get('unit', 'kg'):<6}")

    def _cmd_pr_update(self, args):
        data = {}
        if args.output_qty is not None:
            data["output_quantity"] = args.output_qty
        if args.details:
            data["details"] = args.details
        result = self.client.update_processing(args.processing_id, data)
        self._print_processing(result)

    def _cmd_todo_list(self, args):
        todos = self.client.list_todos()
        if not todos:
            print("暂无待办事项")
            return
        print(f"{'ID':<15} {'类型':<12} {'状态':<10} {'截止日期':<12} {'标题':<40}")
        print("-" * 90)
        for t in todos:
            t_type = TODO_TYPES.get(t['todo_type'], t['todo_type'])
            t_status = TODO_STATUSES.get(t['status'], t['status'])
            print(f"{t['id']:<15} {t_type:<12} {t_status:<10} {t['due_date']:<12} {t['title'][:38]:<40}")

    def _cmd_todo_done(self, args):
        result = self.client.update_todo(args.todo_id, {"status": "completed"})
        print(f"已标记待办 '{result['title']}' 为完成状态")

    def _cmd_todo_refresh(self, args):
        result = self.client.refresh_todos()
        print(result["message"])

    def _cmd_export_batch(self, args):
        content = self.client.export_batch(args.harvest_id)
        if args.output:
            with open(args.output, "w", encoding="utf-8") as f:
                f.write(content)
            print(f"已导出到: {args.output}")
        else:
            print(content)

    def _cmd_export_all(self, args):
        content = self.client.export_all()
        if args.output:
            with open(args.output, "w", encoding="utf-8") as f:
                f.write(content)
            print(f"已导出到: {args.output}")
        else:
            print(content)

    @staticmethod
    def _print_plot(plot: dict):
        print("=" * 50)
        print(f"地块ID: {plot['id']}")
        print(f"地块名称: {plot['name']}")
        print(f"种植品种: {plot['variety']}")
        print(f"地块面积: {plot['area']} 亩")
        print(f"种植日期: {plot['planting_date']}")
        print(f"预计采收: {plot['expected_harvest_date']}")
        if plot.get('notes'):
            print(f"备注: {plot['notes']}")
        print("=" * 50)

    @staticmethod
    def _print_operation(op: dict):
        op_type = OPERATION_TYPES.get(op['operation_type'], op['operation_type'])
        print("=" * 50)
        print(f"操作ID: {op['id']}")
        print(f"地块ID: {op['plot_id']}")
        print(f"操作类型: {op_type}")
        print(f"操作日期: {op['operation_date']}")
        print(f"操作详情: {op['details']}")
        if op.get('quantity'):
            print(f"用量: {op['quantity']}")
        if op.get('safety_interval_days'):
            print(f"安全间隔期: {op['safety_interval_days']}天")
        print("=" * 50)

    @staticmethod
    def _print_harvest(h: dict):
        status = HARVEST_STATUSES.get(h['status'], h['status'])
        print("=" * 50)
        print(f"采收ID: {h['id']}")
        print(f"地块ID: {h['plot_id']}")
        print(f"状态: {status}")
        print(f"计划日期: {h['planned_date']}")
        if h.get('actual_date'):
            print(f"实际日期: {h['actual_date']}")
        if h.get('quantity'):
            print(f"数量: {h['quantity']} {h.get('unit', 'kg')}")
        if h.get('quality_status'):
            print(f"质量状态: {h['quality_status']}")
        if h.get('notes'):
            print(f"备注: {h['notes']}")
        print("=" * 50)

    @staticmethod
    def _print_processing(p: dict):
        print("=" * 50)
        print(f"加工ID: {p['id']}")
        print(f"采收ID: {p['harvest_id']}")
        print(f"批号: {p['batch_number']}")
        print(f"加工日期: {p['processing_date']}")
        print(f"投料量: {p['input_quantity']} {p.get('unit', 'kg')}")
        if p.get('output_quantity') is not None:
            print(f"成品量: {p['output_quantity']} {p.get('unit', 'kg')}")
            yield_rate = (p['output_quantity'] / p['input_quantity'] * 100) if p['input_quantity'] > 0 else 0
            print(f"出成率: {yield_rate:.1f}%")
        print(f"详情: {p['details']}")
        print("=" * 50)


def main():
    cli = GAPCLI()
    cli.run()


if __name__ == "__main__":
    main()
