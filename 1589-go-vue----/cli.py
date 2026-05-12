#!/usr/bin/env python3
import os
import argparse
import json
import sys
from datetime import date, datetime
from dotenv import load_dotenv

load_dotenv()

try:
    import httpx
except ImportError:
    print("请先安装依赖: pip install httpx")
    sys.exit(1)

PORT = int(os.getenv("PORT", "8000"))
BASE_URL = f"http://localhost:{PORT}/api"


class RopewayCLI:
    def __init__(self, base_url: str = BASE_URL):
        self.base_url = base_url
        self.client = httpx.Client(timeout=30.0)

    def _request(self, method: str, endpoint: str, **kwargs):
        url = f"{self.base_url}{endpoint}"
        try:
            response = self.client.request(method, url, **kwargs)
            if response.status_code >= 400:
                print(f"错误 [{response.status_code}]: {response.text}")
                return None
            return response.json()
        except httpx.ConnectError:
            print(f"无法连接到服务器: {self.base_url}")
            print("请确保服务已启动: uvicorn app.main:app --port {PORT}")
            return None

    def status(self):
        data = self._request("GET", "/status")
        if data:
            status_map = {
                "normal": "正常运行",
                "decelerate": "减速运行",
                "pause": "暂停运行",
                "stop": "停止运行"
            }
            current = status_map.get(data["current_status"], data["current_status"])
            print("\n" + "=" * 50)
            print("索道运行状态")
            print("=" * 50)
            print(f"当前状态: {current}")
            if data["last_update"]:
                last_update = datetime.fromisoformat(data["last_update"].replace("Z", "+00:00"))
                print(f"最后更新: {last_update.strftime('%Y-%m-%d %H:%M:%S')}")
            if data["current_wind_speed"] is not None:
                print(f"当前风速: {data['current_wind_speed']} m/s")
            print(f"雷电预警: {'是' if data['has_lightning'] else '否'}")
            print(f"等待确认: {'是' if data['pending_confirmation'] else '否'}")
            print(f"滞后计数: {data['lag_counter']}/{data['lag_threshold']}")
            print("=" * 50 + "\n")
        return data

    def weather(self, minutes: int = 30):
        data = self._request("GET", f"/weather?minutes={minutes}")
        if data:
            print("\n" + "=" * 70)
            print(f"最近 {minutes} 分钟天气数据")
            print("=" * 70)
            print(f"{'时间':<20} {'风速(m/s)':<12} {'雷电':<8} {'温度(°C)':<10} {'湿度(%)':<8}")
            print("-" * 70)
            for item in data:
                ts = datetime.fromisoformat(item["timestamp"].replace("Z", "+00:00"))
                temp = item.get("temperature", "-")
                if temp is not None:
                    temp = f"{temp:.1f}"
                humidity = item.get("humidity", "-")
                if humidity is not None:
                    humidity = f"{humidity:.1f}"
                print(f"{ts.strftime('%Y-%m-%d %H:%M:%S'):<20} "
                      f"{item['wind_speed']:<12.1f} "
                      f"{'是' if item['has_lightning'] else '否':<8} "
                      f"{temp:<10} "
                      f"{humidity:<8}")
            print("=" * 70 + "\n")
        return data

    def history(self, limit: int = 20):
        data = self._request("GET", f"/status/history?limit={limit}")
        if data:
            status_map = {
                "normal": "正常",
                "decelerate": "减速",
                "pause": "暂停",
                "stop": "停止"
            }
            print("\n" + "=" * 80)
            print(f"最近 {limit} 条状态变更记录")
            print("=" * 80)
            print(f"{'时间':<20} {'从':<10} {'到':<10} {'原因':<40}")
            print("-" * 80)
            for item in data:
                ts = datetime.fromisoformat(item["timestamp"].replace("Z", "+00:00"))
                from_status = status_map.get(item.get("from_status") or "-", item.get("from_status") or "-")
                to_status = status_map.get(item["to_status"], item["to_status"])
                reason = item.get("reason", "")[:40]
                print(f"{ts.strftime('%Y-%m-%d %H:%M:%S'):<20} "
                      f"{from_status:<10} "
                      f"{to_status:<10} "
                      f"{reason:<40}")
            print("=" * 80 + "\n")
        return data

    def update_weather(self, wind_speed: float, lightning: bool = False):
        data = self._request("POST", "/weather", json={
            "wind_speed": wind_speed,
            "has_lightning": lightning
        })
        if data:
            print(f"天气数据已更新: 风速 {data['wind_speed']} m/s, 雷电: {'是' if data['has_lightning'] else '否'}")
            self.status()
        return data

    def recover(self, operator_id: int, notes: str = None):
        data = self._request("POST", "/status/recover", json={
            "operator_id": operator_id,
            "notes": notes
        })
        if data:
            if data["success"]:
                print(f"✓ 恢复成功: {data['reason']}")
            else:
                print(f"✗ 恢复失败: {data['reason']}")
            self.status()
        return data

    def inspections(self):
        data = self._request("GET", "/inspections?pending_only=true")
        if data:
            type_map = {
                "daily": "日常",
                "weekly": "周检",
                "monthly": "月检",
                "yearly": "年检"
            }
            status_map = {
                "pending": "待处理",
                "completed": "已完成",
                "urgent": "紧急"
            }
            print("\n" + "=" * 90)
            print("待处理检修任务")
            print("=" * 90)
            print(f"{'ID':<5} {'类型':<8} {'计划日期':<20} {'状态':<10} {'紧急':<6} {'问题':<40}")
            print("-" * 90)
            for item in data:
                scheduled = datetime.fromisoformat(item["scheduled_date"].replace("Z", "+00:00"))
                issues = (item.get("issues_found") or item.get("notes") or "")[:40]
                print(f"{item['id']:<5} "
                      f"{type_map.get(item['inspection_type'], item['inspection_type']):<8} "
                      f"{scheduled.strftime('%Y-%m-%d %H:%M'):<20} "
                      f"{status_map.get(item['status'], item['status']):<10} "
                      f"{'是' if item['is_urgent'] else '否':<6} "
                      f"{issues:<40}")
            print("=" * 90 + "\n")
        return data

    def tomorrow_check(self):
        data = self._request("GET", "/inspections/tomorrow-check")
        if data:
            print("\n" + "=" * 50)
            print("次日运营检查")
            print("=" * 50)
            print(f"允许运营: {'是' if data['can_operate'] else '否'}")
            print(f"未处理紧急问题数: {data['urgent_inspections_count']}")
            print(f"状态: {data['message']}")
            if not data['can_operate']:
                print("\n⚠️  警告：存在未处理的紧急问题，次日不允许运营")
            print("=" * 50 + "\n")
        return data

    def queue(self):
        data = self._request("GET", "/queue/status")
        if data:
            print("\n" + "=" * 50)
            print("排队状态")
            print("=" * 50)
            print(f"当前排队人数: {data['queue_count']}")
            print(f"吊厢总容量: {data['gondola_capacity']}")
            print(f"限流阈值: {data['threshold']} (容量×3)")
            print(f"是否限流: {'是' if data['is_limited'] else '否'}")
            if data['is_limited']:
                print("\n⚠️  警告：排队人数超过容量3倍，已触发限流")
            print("=" * 50 + "\n")
        return data

    def capacity(self):
        data = self._request("GET", "/capacity")
        if data:
            print("\n" + "=" * 50)
            print("吊厢容量配置")
            print("=" * 50)
            print(f"总容量: {data.get('total_capacity', 0)}")
            print(f"状态: {'激活' if data.get('is_active') else '未激活'}")
            if data.get('effective_date'):
                effective = datetime.fromisoformat(data['effective_date'].replace("Z", "+00:00"))
                print(f"生效日期: {effective.strftime('%Y-%m-%d %H:%M:%S')}")
            print("=" * 50 + "\n")
        return data

    def generate_report(self, report_date: str = None):
        date_param = f"?report_date={report_date}" if report_date else ""
        data = self._request("POST", f"/reports/generate{date_param}")
        if data:
            self._print_report(data)
        return data

    def get_report(self, report_date: str):
        data = self._request("GET", f"/reports/{report_date}")
        if data:
            self._print_report(data)
        return data

    def list_reports(self, start: str = None, end: str = None):
        params = []
        if start:
            params.append(f"start_date={start}")
        if end:
            params.append(f"end_date={end}")
        query = "?" + "&".join(params) if params else ""
        data = self._request("GET", f"/reports{query}")
        if data:
            if isinstance(data, list):
                print(f"\n共找到 {len(data)} 份报告")
                for item in data:
                    self._print_report(item)
        return data

    def _print_report(self, report: dict):
        report_date = datetime.fromisoformat(report["report_date"].replace("Z", "+00:00"))
        print("\n" + "=" * 70)
        print(f"每日运营报告 - {report_date.strftime('%Y-%m-%d')}")
        print("=" * 70)
        print(f"总运营时间: {report['total_operating_hours']:.2f} 小时")
        print(f"有效运营时间: {report['effective_operating_hours']:.2f} 小时")
        
        status = "⚠️  低于最低安全时长" if report["below_min_hours"] else "正常"
        print(f"时长状态: {status}")
        if report["below_min_hours"]:
            print("  ⚠️  警告：有效运营时间低于安全最低时长(4小时)，请标黄处理")
        
        print(f"\n乘客总数: {report['total_passengers']}")
        print(f"计费乘客: {report['total_passengers_counted']} (1.2米以下儿童不计费)")
        
        if report.get("weather_events"):
            print(f"\n天气事件: {report['weather_events']}")
        
        if report.get("urgent_inspections"):
            print(f"\n紧急检修: {report['urgent_inspections']}")
        
        print("=" * 70 + "\n")

    def config(self):
        data = self._request("GET", "/config")
        if data:
            print("\n" + "=" * 50)
            print("系统配置参数")
            print("=" * 50)
            print(f"减速风速阈值: {data['wind_speed_decelerate']} m/s")
            print(f"暂停风速阈值: {data['wind_speed_pause']} m/s")
            print(f"停止风速阈值: {data['wind_speed_stop']} m/s")
            print(f"滞后保护时间: {data['lag_protection_minutes']} 分钟")
            print(f"最低安全运营时长: {data['min_operating_hours']} 小时")
            print(f"排队限流倍数: {data['queue_capacity_multiplier']}x")
            print(f"儿童身高限制: {data['child_height_limit']} 米")
            print("=" * 50 + "\n")
        return data

    def close(self):
        self.client.close()


def main():
    parser = argparse.ArgumentParser(
        description="索道运行管理系统命令行客户端",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  python cli.py status                    # 查看当前状态
  python cli.py weather                   # 查看最近30分钟天气
  python cli.py weather --minutes 60      # 查看最近1小时天气
  python cli.py history                   # 查看状态变更历史
  python cli.py update-weather 9.5        # 更新风速为9.5 m/s
  python cli.py update-weather 15 --lightning  # 更新风速并标记雷电
  python cli.py recover --operator 1      # 操作员1确认恢复运行
  python cli.py inspections               # 查看待处理检修
  python cli.py tomorrow-check            # 检查次日是否允许运营
  python cli.py queue                     # 查看排队状态
  python cli.py capacity                  # 查看吊厢容量
  python cli.py generate-report           # 生成今日报告
  python cli.py report 2024-01-15         # 查看指定日期报告
  python cli.py config                    # 查看系统配置
        """
    )
    
    parser.add_argument("--port", type=int, default=PORT, help=f"服务端口 (默认: {PORT})")
    
    subparsers = parser.add_subparsers(dest="command", help="可用命令")
    
    subparsers.add_parser("status", help="查看当前运行状态")
    
    weather_parser = subparsers.add_parser("weather", help="查看天气数据")
    weather_parser.add_argument("--minutes", type=int, default=30, help="查看最近N分钟数据")
    
    history_parser = subparsers.add_parser("history", help="查看状态变更历史")
    history_parser.add_argument("--limit", type=int, default=20, help="显示条数")
    
    update_parser = subparsers.add_parser("update-weather", help="更新天气数据")
    update_parser.add_argument("wind_speed", type=float, help="风速 (m/s)")
    update_parser.add_argument("--lightning", action="store_true", help="是否有雷电")
    
    recover_parser = subparsers.add_parser("recover", help="确认恢复运行")
    recover_parser.add_argument("--operator", type=int, required=True, help="操作员ID")
    recover_parser.add_argument("--notes", type=str, help="备注信息")
    
    subparsers.add_parser("inspections", help="查看待处理检修")
    subparsers.add_parser("tomorrow-check", help="检查次日运营许可")
    subparsers.add_parser("queue", help="查看排队状态")
    subparsers.add_parser("capacity", help="查看吊厢容量")
    
    report_parser = subparsers.add_parser("generate-report", help="生成每日报告")
    report_parser.add_argument("--date", type=str, help="报告日期 (YYYY-MM-DD)")
    
    get_report_parser = subparsers.add_parser("report", help="查看指定日期报告")
    get_report_parser.add_argument("date", type=str, help="报告日期 (YYYY-MM-DD)")
    
    list_reports_parser = subparsers.add_parser("reports", help="列出报告")
    list_reports_parser.add_argument("--start", type=str, help="开始日期 (YYYY-MM-DD)")
    list_reports_parser.add_argument("--end", type=str, help="结束日期 (YYYY-MM-DD)")
    
    subparsers.add_parser("config", help="查看系统配置")
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        return
    
    cli = RopewayCLI(f"http://localhost:{args.port}/api")
    
    try:
        if args.command == "status":
            cli.status()
        elif args.command == "weather":
            cli.weather(args.minutes)
        elif args.command == "history":
            cli.history(args.limit)
        elif args.command == "update-weather":
            cli.update_weather(args.wind_speed, args.lightning)
        elif args.command == "recover":
            cli.recover(args.operator, args.notes)
        elif args.command == "inspections":
            cli.inspections()
        elif args.command == "tomorrow-check":
            cli.tomorrow_check()
        elif args.command == "queue":
            cli.queue()
        elif args.command == "capacity":
            cli.capacity()
        elif args.command == "generate-report":
            cli.generate_report(args.date)
        elif args.command == "report":
            cli.get_report(args.date)
        elif args.command == "reports":
            cli.list_reports(args.start, args.end)
        elif args.command == "config":
            cli.config()
    finally:
        cli.close()


if __name__ == "__main__":
    main()
