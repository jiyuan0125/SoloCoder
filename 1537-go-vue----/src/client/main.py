import argparse
import sys
from typing import List

from .api_client import APIClient
from .commands import Commands
from .config import DEFAULT_BASE_URL

def create_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="碳排放监测核算系统命令行客户端",
        formatter_class=argparse.RawDescriptionHelpFormatter
    )
    
    parser.add_argument(
        "--host",
        type=str,
        default=None,
        help="服务端主机地址（默认：127.0.0.1）"
    )
    parser.add_argument(
        "--port",
        type=int,
        default=None,
        help="服务端端口（默认：8000）"
    )
    parser.add_argument(
        "--base-url",
        type=str,
        default=DEFAULT_BASE_URL,
        help="完整服务端地址（默认：http://127.0.0.1:8000）"
    )
    
    subparsers = parser.add_subparsers(
        dest="command",
        title="可用命令",
        help="命令列表"
    )
    
    health = subparsers.add_parser("health", help="检查服务端健康状态")
    
    industries = subparsers.add_parser("industries", help="行业管理")
    industries_sub = industries.add_subparsers(dest="industry_cmd")
    
    industries_list = industries_sub.add_parser("list", help="列出所有行业")
    industries_create = industries_sub.add_parser("create", help="创建行业")
    industries_create.add_argument("--name", required=True, help="行业名称")
    industries_create.add_argument("--code", required=True, help="行业代码")
    industries_create.add_argument("--default-oxidation-rate", default="1.0000", help="默认氧化率（默认1.0000）")
    industries_create.add_argument("--min-oxidation-rate", default="0.9000", help="最小氧化率（默认0.9000）")
    industries_create.add_argument("--description", help="描述")
    
    industries_get = industries_sub.add_parser("get", help="获取行业详情")
    industries_get.add_argument("--id", type=int, required=True, help="行业ID")
    
    industries_update = industries_sub.add_parser("update", help="更新行业信息")
    industries_update.add_argument("--id", type=int, required=True, help="行业ID")
    industries_update.add_argument("--name", help="行业名称")
    industries_update.add_argument("--default-oxidation-rate", help="默认氧化率")
    industries_update.add_argument("--min-oxidation-rate", help="最小氧化率")
    industries_update.add_argument("--description", help="描述")
    
    industries_delete = industries_sub.add_parser("delete", help="删除行业")
    industries_delete.add_argument("--id", type=int, required=True, help="行业ID")
    
    companies = subparsers.add_parser("companies", help="企业管理")
    companies_sub = companies.add_subparsers(dest="company_cmd")
    
    companies_list = companies_sub.add_parser("list", help="列出所有企业")
    
    companies_create = companies_sub.add_parser("create", help="创建企业")
    companies_create.add_argument("--name", required=True, help="企业名称")
    companies_create.add_argument("--registration-no", required=True, help="企业注册号")
    companies_create.add_argument("--industry-id", type=int, required=True, help="行业ID")
    companies_create.add_argument("--annual-output", required=True, help="年产值")
    companies_create.add_argument("--address", help="地址")
    companies_create.add_argument("--contact-person", help="联系人")
    companies_create.add_argument("--contact-phone", help="联系电话")
    companies_create.add_argument("--initial-quota", default="0.0000", help="初始配额（默认0.0000）")
    
    companies_get = companies_sub.add_parser("get", help="获取企业详情")
    companies_get.add_argument("--id", type=int, required=True, help="企业ID")
    
    companies_update = companies_sub.add_parser("update", help="更新企业信息")
    companies_update.add_argument("--id", type=int, required=True, help="企业ID")
    companies_update.add_argument("--name", help="企业名称")
    companies_update.add_argument("--industry-id", type=int, help="行业ID")
    companies_update.add_argument("--annual-output", help="年产值")
    companies_update.add_argument("--address", help="地址")
    companies_update.add_argument("--contact-person", help="联系人")
    companies_update.add_argument("--contact-phone", help="联系电话")
    companies_update.add_argument("--initial-quota", help="初始配额")
    
    companies_delete = companies_sub.add_parser("delete", help="删除企业")
    companies_delete.add_argument("--id", type=int, required=True, help="企业ID")
    
    emission_sources = subparsers.add_parser("emission-sources", help="排放源管理")
    emission_sources_sub = emission_sources.add_subparsers(dest="source_cmd")
    
    sources_list = emission_sources_sub.add_parser("list", help="列出排放源")
    sources_list.add_argument("--company-id", type=int, help="按企业过滤")
    
    sources_create = emission_sources_sub.add_parser("create", help="创建排放源")
    sources_create.add_argument("--company-id", type=int, required=True, help="企业ID")
    sources_create.add_argument("--name", required=True, help="排放源名称")
    sources_create.add_argument("--code", required=True, help="排放源代码")
    sources_create.add_argument("--emission-type", required=True, help="排放类型")
    sources_create.add_argument("--emission-factor", required=True, help="排放因子")
    sources_create.add_argument("--oxidation-rate", required=True, help="氧化率")
    sources_create.add_argument("--unit", required=True, help="单位")
    sources_create.add_argument("--data-source", choices=["manual", "online"], default="manual", help="数据来源")
    sources_create.add_argument("--description", help="描述")
    
    sources_get = emission_sources_sub.add_parser("get", help="获取排放源详情")
    sources_get.add_argument("--id", type=int, required=True, help="排放源ID")
    
    sources_update = emission_sources_sub.add_parser("update", help="更新排放源信息")
    sources_update.add_argument("--id", type=int, required=True, help="排放源ID")
    sources_update.add_argument("--name", help="排放源名称")
    sources_update.add_argument("--emission-type", help="排放类型")
    sources_update.add_argument("--emission-factor", help="排放因子")
    sources_update.add_argument("--oxidation-rate", help="氧化率")
    sources_update.add_argument("--unit", help="单位")
    sources_update.add_argument("--data-source", choices=["manual", "online"], help="数据来源")
    sources_update.add_argument("--description", help="描述")
    
    sources_delete = emission_sources_sub.add_parser("delete", help="删除排放源")
    sources_delete.add_argument("--id", type=int, required=True, help="排放源ID")
    
    emission_data = subparsers.add_parser("emission-data", help="排放数据管理")
    emission_data_sub = emission_data.add_subparsers(dest="data_cmd")
    
    data_list = emission_data_sub.add_parser("list", help="列出排放数据")
    data_list.add_argument("--source-id", type=int, help="按排放源过滤")
    data_list.add_argument("--start-date", help="开始日期（YYYY-MM-DD）")
    data_list.add_argument("--end-date", help="结束日期（YYYY-MM-DD）")
    data_list.add_argument("--status", choices=["pending_review", "approved", "rejected", "auto_filled"], help="状态过滤")
    
    data_create = emission_data_sub.add_parser("create", help="创建排放数据")
    data_create.add_argument("--source-id", type=int, required=True, help="排放源ID")
    data_create.add_argument("--record-date", required=True, help="记录日期（YYYY-MM-DD）")
    data_create.add_argument("--record-hour", type=int, required=True, choices=range(24), help="记录小时（0-23）")
    data_create.add_argument("--activity-data", required=True, help="活动数据")
    data_create.add_argument("--data-source", choices=["manual", "online"], default="manual", help="数据来源")
    
    data_get = emission_data_sub.add_parser("get", help="获取排放数据详情")
    data_get.add_argument("--id", type=int, required=True, help="数据ID")
    
    data_update = emission_data_sub.add_parser("update", help="更新排放数据（审核）")
    data_update.add_argument("--id", type=int, required=True, help="数据ID")
    data_update.add_argument("--activity-data", help="活动数据")
    data_update.add_argument("--status", choices=["approved", "rejected", "pending_review"], help="状态（用于审核）")
    
    data_delete = emission_data_sub.add_parser("delete", help="删除排放数据")
    data_delete.add_argument("--id", type=int, required=True, help="数据ID")
    
    quota = subparsers.add_parser("quota", help="配额管理")
    quota_sub = quota.add_subparsers(dest="quota_cmd")
    
    quota_list = quota_sub.add_parser("list", help="列出配额交易记录")
    quota_list.add_argument("--company-id", type=int, help="按企业过滤")
    quota_list.add_argument("--start-date", help="开始日期（YYYY-MM-DD）")
    quota_list.add_argument("--end-date", help="结束日期（YYYY-MM-DD）")
    
    quota_create = quota_sub.add_parser("create", help="创建配额交易记录")
    quota_create.add_argument("--company-id", type=int, required=True, help="企业ID")
    quota_create.add_argument("--transaction-type", choices=["buy", "sell"], required=True, help="交易类型：buy/sell")
    quota_create.add_argument("--quota-amount", required=True, help="配额数量")
    quota_create.add_argument("--price-per-unit", required=True, help="单价")
    quota_create.add_argument("--transaction-date", required=True, help="交易日期（YYYY-MM-DD）")
    quota_create.add_argument("--counterparty", help="交易对手")
    quota_create.add_argument("--remarks", help="备注")
    
    quota_get = quota_sub.add_parser("get", help="获取交易记录详情")
    quota_get.add_argument("--id", type=int, required=True, help="交易记录ID")
    
    quota_balance = quota_sub.add_parser("balance", help="查询配额余额")
    quota_balance.add_argument("--company-id", type=int, required=True, help="企业ID")
    
    reports = subparsers.add_parser("reports", help="碳排放报告管理")
    reports_sub = reports.add_subparsers(dest="report_cmd")
    
    reports_list = reports_sub.add_parser("list", help="列出报告")
    reports_list.add_argument("--company-id", type=int, help="按企业过滤")
    reports_list.add_argument("--year", type=int, help="按年份过滤")
    reports_list.add_argument("--month", type=int, help="按月份过滤")
    reports_list.add_argument("--status", choices=["draft", "confirmed"], help="按状态过滤")
    
    reports_generate = reports_sub.add_parser("generate", help="生成月度报告")
    reports_generate.add_argument("--company-id", type=int, required=True, help="企业ID")
    reports_generate.add_argument("--year", type=int, required=True, help="年份")
    reports_generate.add_argument("--month", type=int, required=True, help="月份")
    
    reports_get = reports_sub.add_parser("get", help="获取报告详情")
    reports_get.add_argument("--id", type=int, required=True, help="报告ID")
    
    reports_summary = reports_sub.add_parser("summary", help="获取月度汇总")
    reports_summary.add_argument("--company-id", type=int, required=True, help="企业ID")
    reports_summary.add_argument("--year", type=int, required=True, help="年份")
    reports_summary.add_argument("--month", type=int, required=True, help="月份")
    
    reports_confirm = reports_sub.add_parser("confirm", help="确认报告（锁定数据）")
    reports_confirm.add_argument("--id", type=int, required=True, help="报告ID")
    
    reports_delete = reports_sub.add_parser("delete", help="删除报告")
    reports_delete.add_argument("--id", type=int, required=True, help="报告ID")
    
    analytics = subparsers.add_parser("analytics", help="数据分析")
    analytics_sub = analytics.add_subparsers(dest="analytics_cmd")
    
    analytics_ranking = analytics_sub.add_parser("ranking", help="行业碳排放强度排名")
    analytics_ranking.add_argument("--year", type=int, required=True, help="年份")
    
    return parser

def run_cli(args: List[str] = None) -> int:
    parser = create_parser()
    parsed_args = parser.parse_args(args)
    
    if not parsed_args.command:
        parser.print_help()
        return 0
    
    base_url = parsed_args.base_url
    if parsed_args.host or parsed_args.port:
        host = parsed_args.host or "127.0.0.1"
        port = parsed_args.port or 8000
        base_url = f"http://{host}:{port}"
    
    client = APIClient(base_url)
    commands = Commands(client)
    
    try:
        if parsed_args.command == "health":
            commands.health()
        
        elif parsed_args.command == "industries":
            if parsed_args.industry_cmd == "list":
                commands.list_industries()
            elif parsed_args.industry_cmd == "create":
                commands.create_industry(
                    parsed_args.name, parsed_args.code,
                    parsed_args.default_oxidation_rate, parsed_args.min_oxidation_rate,
                    parsed_args.description
                )
            elif parsed_args.industry_cmd == "get":
                commands.get_industry(parsed_args.id)
            elif parsed_args.industry_cmd == "update":
                commands.update_industry(
                    parsed_args.id, parsed_args.name,
                    parsed_args.default_oxidation_rate,
                    parsed_args.min_oxidation_rate,
                    parsed_args.description
                )
            elif parsed_args.industry_cmd == "delete":
                commands.delete_industry(parsed_args.id)
            else:
                industries.print_help()
        
        elif parsed_args.command == "companies":
            if parsed_args.company_cmd == "list":
                commands.list_companies()
            elif parsed_args.company_cmd == "create":
                commands.create_company(
                    parsed_args.name, parsed_args.registration_no,
                    parsed_args.industry_id, parsed_args.annual_output,
                    parsed_args.address, parsed_args.contact_person,
                    parsed_args.contact_phone, parsed_args.initial_quota
                )
            elif parsed_args.company_cmd == "get":
                commands.get_company(parsed_args.id)
            elif parsed_args.company_cmd == "update":
                commands.update_company(
                    parsed_args.id, parsed_args.name, parsed_args.industry_id,
                    parsed_args.annual_output, parsed_args.address,
                    parsed_args.contact_person, parsed_args.contact_phone,
                    parsed_args.initial_quota
                )
            elif parsed_args.company_cmd == "delete":
                commands.delete_company(parsed_args.id)
            else:
                companies.print_help()
        
        elif parsed_args.command == "emission-sources":
            if parsed_args.source_cmd == "list":
                commands.list_emission_sources(parsed_args.company_id)
            elif parsed_args.source_cmd == "create":
                commands.create_emission_source(
                    parsed_args.company_id, parsed_args.name, parsed_args.code,
                    parsed_args.emission_type, parsed_args.emission_factor,
                    parsed_args.oxidation_rate, parsed_args.unit,
                    parsed_args.data_source, parsed_args.description
                )
            elif parsed_args.source_cmd == "get":
                commands.get_emission_source(parsed_args.id)
            elif parsed_args.source_cmd == "update":
                commands.update_emission_source(
                    parsed_args.id, parsed_args.name, parsed_args.emission_type,
                    parsed_args.emission_factor, parsed_args.oxidation_rate,
                    parsed_args.unit, parsed_args.data_source, parsed_args.description
                )
            elif parsed_args.source_cmd == "delete":
                commands.delete_emission_source(parsed_args.id)
            else:
                emission_sources.print_help()
        
        elif parsed_args.command == "emission-data":
            if parsed_args.data_cmd == "list":
                commands.list_emission_data(
                    parsed_args.source_id, parsed_args.start_date,
                    parsed_args.end_date, parsed_args.status
                )
            elif parsed_args.data_cmd == "create":
                commands.create_emission_data(
                    parsed_args.source_id, parsed_args.record_date,
                    parsed_args.record_hour, parsed_args.activity_data,
                    parsed_args.data_source
                )
            elif parsed_args.data_cmd == "get":
                commands.get_emission_data(parsed_args.id)
            elif parsed_args.data_cmd == "update":
                commands.update_emission_data(
                    parsed_args.id, parsed_args.activity_data, parsed_args.status
                )
            elif parsed_args.data_cmd == "delete":
                commands.delete_emission_data(parsed_args.id)
            else:
                emission_data.print_help()
        
        elif parsed_args.command == "quota":
            if parsed_args.quota_cmd == "list":
                commands.list_transactions(
                    parsed_args.company_id, parsed_args.start_date, parsed_args.end_date
                )
            elif parsed_args.quota_cmd == "create":
                commands.create_transaction(
                    parsed_args.company_id, parsed_args.transaction_type,
                    parsed_args.quota_amount, parsed_args.price_per_unit,
                    parsed_args.transaction_date, parsed_args.counterparty,
                    parsed_args.remarks
                )
            elif parsed_args.quota_cmd == "get":
                commands.get_transaction(parsed_args.id)
            elif parsed_args.quota_cmd == "balance":
                commands.get_quota_balance(parsed_args.company_id)
            else:
                quota.print_help()
        
        elif parsed_args.command == "reports":
            if parsed_args.report_cmd == "list":
                commands.list_reports(
                    parsed_args.company_id, parsed_args.year,
                    parsed_args.month, parsed_args.status
                )
            elif parsed_args.report_cmd == "generate":
                commands.generate_report(
                    parsed_args.company_id, parsed_args.year, parsed_args.month
                )
            elif parsed_args.report_cmd == "get":
                commands.get_report(parsed_args.id)
            elif parsed_args.report_cmd == "summary":
                commands.get_monthly_summary(
                    parsed_args.company_id, parsed_args.year, parsed_args.month
                )
            elif parsed_args.report_cmd == "confirm":
                commands.confirm_report(parsed_args.id)
            elif parsed_args.report_cmd == "delete":
                commands.delete_report(parsed_args.id)
            else:
                reports.print_help()
        
        elif parsed_args.command == "analytics":
            if parsed_args.analytics_cmd == "ranking":
                commands.get_intensity_ranking(parsed_args.year)
            else:
                analytics.print_help()
        
        return 0
    
    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
        return 1

def main():
    sys.exit(run_cli())

if __name__ == "__main__":
    main()
