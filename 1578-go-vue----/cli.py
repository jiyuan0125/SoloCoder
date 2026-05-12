#!/usr/bin/env python3
import argparse
import sys
from datetime import datetime

from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

from app import models, services
from app.database import Base

DB_URL = "sqlite:///./waybill.db"


def get_session():
    engine = create_engine(DB_URL, connect_args={"check_same_thread": False})
    Base.metadata.create_all(bind=engine)
    Session = sessionmaker(autocommit=False, autoflush=False, bind=engine)
    return Session()


def print_rate_list(rates):
    if not rates:
        print("暂无费率数据")
        return
    
    print(f"\n{'ID':<4} {'价区对':<12} {'费率(分/吨)':<15} {'生效时间':<20} {'失效时间'}")
    print("-" * 80)
    for r in rates:
        eff_to = r.effective_to.strftime("%Y-%m-%d %H:%M") if r.effective_to else "未失效"
        print(f"{r.id:<4} {r.from_zone}->{r.to_zone:<9} {r.rate_per_ton_fen:<15} {r.effective_from.strftime('%Y-%m-%d %H:%M'):<20} {eff_to}")


def print_waybill_list(waybills):
    if not waybills:
        print("暂无货票数据")
        return
    
    print(f"\n{'货票编号':<20} {'发站':<10} {'到站':<10} {'重量(吨)':<10} {'危险品':<6} {'状态':<10} {'总运费(元)':<12} {'创建时间'}")
    print("-" * 100)
    for wb in waybills:
        print(f"{wb.waybill_no:<20} {wb.from_station_name:<10} {wb.to_station_name:<10} {float(wb.weight_ton):<10} {'是' if wb.is_hazardous else '否':<6} {wb.status:<10} {wb.total_charge_fen/100:<12.2f} {wb.created_at.strftime('%Y-%m-%d %H:%M')}")


def cmd_list_rates(args):
    db = get_session()
    try:
        if args.from_zone and args.to_zone:
            rate = services.get_current_rate(db, args.from_zone, args.to_zone)
            if rate:
                print_rate_list([rate])
            else:
                print(f"价区 {args.from_zone} -> {args.to_zone} 无当前生效费率")
        else:
            rates = services.get_all_rates(db)
            print_rate_list(rates)
    finally:
        db.close()


def cmd_list_waybills(args):
    db = get_session()
    try:
        start_date = None
        end_date = None
        if args.start_date:
            start_date = datetime.strptime(args.start_date, "%Y-%m-%d")
        if args.end_date:
            end_date = datetime.strptime(args.end_date, "%Y-%m-%d")
        
        waybills = services.list_waybills(
            db,
            start_date=start_date,
            end_date=end_date,
            status=args.status,
            customer_code=args.customer,
        )
        print_waybill_list(waybills)
    finally:
        db.close()


def cmd_waybill_detail(args):
    db = get_session()
    try:
        if args.waybill_no:
            wb = services.get_waybill_by_no(db, args.waybill_no)
        else:
            wb = services.get_waybill(db, args.id)
        
        if not wb:
            print("货票不存在")
            return
        
        print("\n" + "=" * 60)
        print("货票详情")
        print("=" * 60)
        print(f"货票编号: {wb.waybill_no}")
        print(f"发站: {wb.from_station_name}")
        print(f"到站: {wb.to_station_name}")
        print(f"重量: {float(wb.weight_ton)} 吨")
        print(f"危险品: {'是' if wb.is_hazardous else '否'}")
        print(f"客户编码: {wb.customer_code or '无'}")
        print("-" * 60)
        print(f"基础费率: {wb.basic_rate_fen} 分/吨")
        print(f"基础运费: {wb.basic_charge_fen / 100:.2f} 元")
        print(f"危险品上浮: {wb.hazardous_surcharge_fen / 100:.2f} 元")
        print(f"折扣: {wb.discount_fen / 100:.2f} 元")
        print(f"总运费: {wb.total_charge_fen / 100:.2f} 元")
        print("-" * 60)
        print(f"状态: {wb.status}")
        print(f"创建时间: {wb.created_at.strftime('%Y-%m-%d %H:%M:%S')}")
        if wb.settled_at:
            print(f"结算时间: {wb.settled_at.strftime('%Y-%m-%d %H:%M:%S')}")
        print("=" * 60)
    finally:
        db.close()


def cmd_monthly_report(args):
    db = get_session()
    try:
        report = services.get_monthly_report(db, args.year, args.month)
        
        print("\n" + "=" * 60)
        print(f"月度对账报表 - {report.year_month}")
        print("=" * 60)
        print(f"总票数: {report.total_count}")
        print(f"总重量: {float(report.total_weight_ton)} 吨")
        print(f"总运费: {report.total_amount_fen / 100:.2f} 元")
        print(f"已结算: {report.settled_amount_fen / 100:.2f} 元")
        print(f"未结算: {report.unsettled_amount_fen / 100:.2f} 元")
        print("=" * 60)
        
        customer_stats = services.get_customer_monthly_stats(db, args.year, args.month)
        if customer_stats:
            print("\n客户月度统计:")
            print(f"{'客户编码':<15} {'票数':<8} {'总金额(元)':<15} {'已结算(元)':<15}")
            print("-" * 55)
            for stat in customer_stats:
                print(f"{stat.customer_code:<15} {stat.total_count:<8} {stat.total_amount_fen/100:<15.2f} {stat.settled_amount_fen/100:<15.2f}")
    finally:
        db.close()


def main():
    parser = argparse.ArgumentParser(description="货票管理系统命令行客户端")
    subparsers = parser.add_subparsers(dest="command", help="可用命令")
    
    rates_parser = subparsers.add_parser("rates", help="查询费率")
    rates_parser.add_argument("--from", dest="from_zone", type=int, help="出发价区 (1-8)")
    rates_parser.add_argument("--to", dest="to_zone", type=int, help="到达价区 (1-8)")
    rates_parser.set_defaults(func=cmd_list_rates)
    
    list_parser = subparsers.add_parser("list", help="查询货票列表")
    list_parser.add_argument("--start-date", help="开始日期 (YYYY-MM-DD)")
    list_parser.add_argument("--end-date", help="结束日期 (YYYY-MM-DD)")
    list_parser.add_argument("--status", help="状态 (issued/settled)")
    list_parser.add_argument("--customer", help="客户编码")
    list_parser.set_defaults(func=cmd_list_waybills)
    
    detail_parser = subparsers.add_parser("detail", help="查询货票详情")
    detail_group = detail_parser.add_mutually_exclusive_group(required=True)
    detail_group.add_argument("--id", type=int, help="货票ID")
    detail_group.add_argument("--no", dest="waybill_no", help="货票编号")
    detail_parser.set_defaults(func=cmd_waybill_detail)
    
    report_parser = subparsers.add_parser("report", help="月度对账报表")
    report_parser.add_argument("--year", type=int, required=True, help="年份")
    report_parser.add_argument("--month", type=int, required=True, help="月份 (1-12)")
    report_parser.set_defaults(func=cmd_monthly_report)
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(0)
    
    args.func(args)


if __name__ == "__main__":
    main()
