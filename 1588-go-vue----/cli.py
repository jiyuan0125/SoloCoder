#!/usr/bin/env python3
import argparse
import sys
from datetime import datetime
from sqlalchemy.orm import Session
from database import SessionLocal
import models
import services


def print_flight(flight):
    print(f"\n航班号: {flight.flight_number}")
    print(f"状态: {flight.status}")
    print(f"出发码头ID: {flight.departure_terminal_id}")
    print(f"到达码头ID: {flight.arrival_terminal_id}")
    print(f"计划出发: {flight.scheduled_departure}")
    print(f"计划到达: {flight.scheduled_arrival}")
    if flight.actual_departure:
        print(f"实际出发: {flight.actual_departure}")
    if flight.actual_arrival:
        print(f"实际到达: {flight.actual_arrival}")
    if flight.delay_minutes > 0:
        print(f"延误时间: {flight.delay_minutes} 分钟")
    print(f"预估客流: {flight.estimated_passengers}")
    print(f"实际载客: {flight.actual_passengers}")
    print(f"实际载车: {flight.actual_vehicles}")


def print_ship(ship):
    print(f"\n船舶名称: {ship.name}")
    print(f"状态: {ship.status}")
    print(f"载客量: {ship.passenger_capacity}")
    print(f"载车量: {ship.vehicle_capacity}")
    if ship.inspection_start_date:
        print(f"年检开始: {ship.inspection_start_date}")
    if ship.inspection_end_date:
        print(f"年检结束: {ship.inspection_end_date}")


def list_flights(args):
    db: Session = SessionLocal()
    try:
        flights = services.get_flights(
            db, 
            skip=args.skip, 
            limit=args.limit,
            status=args.status
        )
        if not flights:
            print("暂无航班数据")
            return
        for flight in flights:
            print_flight(flight)
    finally:
        db.close()


def search_flight(args):
    db: Session = SessionLocal()
    try:
        if args.number:
            flight = services.get_flight_by_number(db, args.number)
        else:
            flight = services.get_flight(db, args.id)
        
        if flight:
            print_flight(flight)
        else:
            print("未找到该航班")
    finally:
        db.close()


def list_ships(args):
    db: Session = SessionLocal()
    try:
        ships = services.get_ships(db, skip=args.skip, limit=args.limit, status=args.status)
        if not ships:
            print("暂无船舶数据")
            return
        for ship in ships:
            print_ship(ship)
    finally:
        db.close()


def search_ship(args):
    db: Session = SessionLocal()
    try:
        ship = services.get_ship(db, args.id)
        if ship:
            print_ship(ship)
        else:
            print("未找到该船舶")
    finally:
        db.close()


def list_terminals(args):
    db: Session = SessionLocal()
    try:
        terminals = services.get_terminals(db, skip=args.skip, limit=args.limit)
        if not terminals:
            print("暂无码头数据")
            return
        for t in terminals:
            print(f"\n码头ID: {t.id}")
            print(f"名称: {t.name}")
            print(f"位置: {t.location}")
            print(f"最大泊位数: {t.max_berths}")
    finally:
        db.close()


def list_berths(args):
    db: Session = SessionLocal()
    try:
        berths = services.get_berths(db, skip=args.skip, limit=args.limit, terminal_id=args.terminal_id)
        if not berths:
            print("暂无泊位数据")
            return
        for b in berths:
            print(f"\n泊位ID: {b.id}")
            print(f"所属码头ID: {b.terminal_id}")
            print(f"泊位号: {b.berth_number}")
            print(f"靠泊能力: {b.capacity}")
            print(f"状态: {b.status}")
    finally:
        db.close()


def check_departure(args):
    db: Session = SessionLocal()
    try:
        flight = services.get_flight(db, args.id)
        if not flight:
            print("未找到该航班")
            return
        can_go, message = services.can_depart(db, flight)
        print(f"开航检查结果: {message}")
    finally:
        db.close()


def main():
    parser = argparse.ArgumentParser(description='轮渡公司运营管理系统 - 命令行客户端')
    subparsers = parser.add_subparsers(dest='command', help='可用命令')
    
    flights_parser = subparsers.add_parser('flights', help='航班操作')
    flights_sub = flights_parser.add_subparsers(dest='action')
    
    flights_list = flights_sub.add_parser('list', help='列出航班')
    flights_list.add_argument('--skip', type=int, default=0, help='跳过数量')
    flights_list.add_argument('--limit', type=int, default=50, help='返回数量')
    flights_list.add_argument('--status', type=str, choices=['计划中', '已开航', '已到港', '取消', '延误'], help='按状态筛选')
    flights_list.set_defaults(func=list_flights)
    
    flights_search = flights_sub.add_parser('search', help='查询航班')
    flights_search_group = flights_search.add_mutually_exclusive_group(required=True)
    flights_search_group.add_argument('--id', type=int, help='航班ID')
    flights_search_group.add_argument('--number', type=str, help='航班号')
    flights_search.set_defaults(func=search_flight)
    
    flights_check = flights_sub.add_parser('check', help='检查开航条件')
    flights_check.add_argument('--id', type=int, required=True, help='航班ID')
    flights_check.set_defaults(func=check_departure)
    
    ships_parser = subparsers.add_parser('ships', help='船舶操作')
    ships_sub = ships_parser.add_subparsers(dest='action')
    
    ships_list = ships_sub.add_parser('list', help='列出船舶')
    ships_list.add_argument('--skip', type=int, default=0, help='跳过数量')
    ships_list.add_argument('--limit', type=int, default=50, help='返回数量')
    ships_list.add_argument('--status', type=str, choices=['可用', '年度检验中', '运营中'], help='按状态筛选')
    ships_list.set_defaults(func=list_ships)
    
    ships_search = ships_sub.add_parser('search', help='查询船舶')
    ships_search.add_argument('--id', type=int, required=True, help='船舶ID')
    ships_search.set_defaults(func=search_ship)
    
    terminals_parser = subparsers.add_parser('terminals', help='码头操作')
    terminals_sub = terminals_parser.add_subparsers(dest='action')
    
    terminals_list = terminals_sub.add_parser('list', help='列出码头')
    terminals_list.add_argument('--skip', type=int, default=0, help='跳过数量')
    terminals_list.add_argument('--limit', type=int, default=50, help='返回数量')
    terminals_list.set_defaults(func=list_terminals)
    
    berths_parser = subparsers.add_parser('berths', help='泊位操作')
    berths_sub = berths_parser.add_subparsers(dest='action')
    
    berths_list = berths_sub.add_parser('list', help='列出泊位')
    berths_list.add_argument('--skip', type=int, default=0, help='跳过数量')
    berths_list.add_argument('--limit', type=int, default=50, help='返回数量')
    berths_list.add_argument('--terminal-id', type=int, help='按码头筛选')
    berths_list.set_defaults(func=list_berths)
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        sys.exit(1)
    
    if hasattr(args, 'func'):
        args.func(args)
    else:
        parser.print_help()
        sys.exit(1)


if __name__ == '__main__':
    main()
