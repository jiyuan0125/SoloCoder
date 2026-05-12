#!/usr/bin/env python3
import argparse
import sys
from datetime import date, time, datetime
from tabulate import tabulate

from app.database import SessionLocal, init_db
from app.models import (
    Booking, Venue, VenueType, User, Tournament, Match, Team,
    TrainingClass, ClassRegistration, TeamStatistics,
    BookingStatus, TournamentStatus, ClassStatus, ChargeType
)


def list_bookings(args):
    db = SessionLocal()
    try:
        query = db.query(Booking)
        
        if args.user:
            query = query.filter(Booking.user_id == args.user)
        if args.venue:
            query = query.filter(Booking.venue_id == args.venue)
        if args.status:
            query = query.filter(Booking.status == args.status)
        
        bookings = query.order_by(Booking.created_at.desc()).all()
        
        if not bookings:
            print("暂无预约记录")
            return
        
        headers = ["ID", "预约号", "用户ID", "场地ID", "日期", "时段", "金额", "状态"]
        rows = []
        for b in bookings:
            venue = db.query(Venue).filter(Venue.id == b.venue_id).first()
            venue_name = venue.name if venue else "未知"
            rows.append([
                b.id,
                b.booking_no,
                b.user_id,
                f"{b.venue_id}({venue_name})",
                b.booking_date,
                f"{b.start_time}-{b.end_time}",
                f"¥{b.final_amount:.2f}",
                b.status.value,
            ])
        
        print(tabulate(rows, headers=headers, tablefmt="grid"))
    finally:
        db.close()


def show_booking(args):
    db = SessionLocal()
    try:
        booking = db.query(Booking).filter(Booking.id == args.id).first()
        if not booking:
            print(f"预约 {args.id} 不存在")
            return
        
        print(f"\n{'='*50}")
        print(f"预约详情")
        print(f"{'='*50}")
        print(f"预约ID: {booking.id}")
        print(f"预约号: {booking.booking_no}")
        print(f"用户ID: {booking.user_id}")
        
        venue = db.query(Venue).filter(Venue.id == booking.venue_id).first()
        if venue:
            print(f"场地: {venue.name}")
        
        print(f"日期: {booking.booking_date}")
        print(f"时段: {booking.start_time} - {booking.end_time}")
        print(f"时长: {booking.hours} 小时")
        print(f"原始金额: ¥{booking.original_amount:.2f}")
        print(f"优惠金额: ¥{booking.discount_amount:.2f}")
        print(f"实付金额: ¥{booking.final_amount:.2f}")
        print(f"状态: {booking.status.value}")
        print(f"连续预约: {'是' if booking.is_continuous else '否'}")
        print(f"创建时间: {booking.created_at}")
        
        if booking.payment:
            print(f"\n支付信息:")
            print(f"  支付号: {booking.payment.payment_no}")
            print(f"  状态: {booking.payment.status.value}")
            print(f"  方式: {booking.payment.payment_method or '未知'}")
            if booking.payment.paid_at:
                print(f"  支付时间: {booking.payment.paid_at}")
        
        if booking.refund:
            print(f"\n退款信息:")
            print(f"  退款号: {booking.refund.refund_no}")
            print(f"  退款金额: ¥{booking.refund.refund_amount:.2f}")
            print(f"  退款比例: {booking.refund.refund_rate*100:.0f}%")
            print(f"  原因: {booking.refund.reason}")
        
        if booking.notes:
            print(f"备注: {booking.notes}")
        
        print(f"{'='*50}\n")
    finally:
        db.close()


def list_matches(args):
    db = SessionLocal()
    try:
        query = db.query(Match)
        
        if args.tournament:
            query = query.filter(Match.tournament_id == args.tournament)
        if args.team:
            from sqlalchemy import or_
            query = query.filter(
                or_(Match.home_team_id == args.team, Match.away_team_id == args.team)
            )
        if args.round:
            query = query.filter(Match.round_no == args.round)
        
        matches = query.order_by(Match.match_date, Match.start_time).all()
        
        if not matches:
            print("暂无比赛记录")
            return
        
        headers = ["ID", "赛事", "场次", "轮次", "主队", "客队", "日期", "时段", "比分", "状态"]
        rows = []
        for m in matches:
            tournament = db.query(Tournament).filter(Tournament.id == m.tournament_id).first()
            tour_name = tournament.name if tournament else "未知"
            
            home_team = db.query(Team).filter(Team.id == m.home_team_id).first()
            away_team = db.query(Team).filter(Team.id == m.away_team_id).first()
            
            home_name = home_team.name if home_team else "轮空"
            away_name = away_team.name if away_team else "轮空"
            
            if m.home_score is not None and m.away_score is not None:
                score = f"{m.home_score}:{m.away_score}"
            else:
                score = "-"
            
            rows.append([
                m.id,
                tour_name[:15],
                m.match_no,
                m.round_no,
                home_name,
                away_name,
                m.match_date,
                f"{m.start_time}-{m.end_time}",
                score,
                m.status.value,
            ])
        
        print(tabulate(rows, headers=headers, tablefmt="grid"))
    finally:
        db.close()


def list_tournaments(args):
    db = SessionLocal()
    try:
        tournaments = db.query(Tournament).order_by(Tournament.created_at.desc()).all()
        
        if not tournaments:
            print("暂无赛事")
            return
        
        headers = ["ID", "名称", "开始日期", "结束日期", "状态", "队伍数"]
        rows = []
        for t in tournaments:
            team_count = db.query(Team).filter(Team.tournament_id == t.id).count()
            rows.append([
                t.id,
                t.name,
                t.start_date,
                t.end_date,
                t.status.value,
                team_count,
            ])
        
        print(tabulate(rows, headers=headers, tablefmt="grid"))
    finally:
        db.close()


def list_classes(args):
    db = SessionLocal()
    try:
        query = db.query(TrainingClass)
        if args.status:
            query = query.filter(TrainingClass.status == args.status)
        
        classes = query.order_by(TrainingClass.created_at.desc()).all()
        
        if not classes:
            print("暂无培训班")
            return
        
        headers = ["ID", "班号", "名称", "教练", "开始", "结束", "时间", "人数", "状态"]
        rows = []
        for c in classes:
            reg_count = db.query(ClassRegistration).filter(
                ClassRegistration.class_id == c.id
            ).count()
            rows.append([
                c.id,
                c.class_no,
                c.name,
                c.instructor or "-",
                c.start_date,
                c.end_date,
                c.class_time,
                f"{reg_count}/{c.max_students}",
                c.status.value,
            ])
        
        print(tabulate(rows, headers=headers, tablefmt="grid"))
    finally:
        db.close()


def list_venues(args):
    db = SessionLocal()
    try:
        venues = db.query(Venue).order_by(Venue.id).all()
        
        if not venues:
            print("暂无场地")
            return
        
        headers = ["ID", "名称", "类型", "价格", "容量", "状态"]
        rows = []
        for v in venues:
            vtype = db.query(VenueType).filter(VenueType.id == v.venue_type_id).first()
            type_name = vtype.name if vtype else "未知"
            price = f"¥{vtype.price}/小时" if vtype and vtype.charge_type == ChargeType.PER_HOUR else f"¥{vtype.price}/场" if vtype else "-"
            rows.append([
                v.id,
                v.name,
                type_name,
                price,
                v.capacity,
                "启用" if v.is_active else "停用",
            ])
        
        print(tabulate(rows, headers=headers, tablefmt="grid"))
    finally:
        db.close()


def list_users(args):
    db = SessionLocal()
    try:
        users = db.query(User).order_by(User.id).all()
        
        if not users:
            print("暂无用户")
            return
        
        headers = ["ID", "用户名", "手机", "管理员", "注册时间"]
        rows = []
        for u in users:
            rows.append([
                u.id,
                u.username,
                u.phone,
                "是" if u.is_admin else "否",
                u.created_at,
            ])
        
        print(tabulate(rows, headers=headers, tablefmt="grid"))
    finally:
        db.close()


def tournament_stats(args):
    db = SessionLocal()
    try:
        stats = db.query(TeamStatistics).filter(
            TeamStatistics.tournament_id == args.tournament
        ).order_by(TeamStatistics.rank.asc().nullslast()).all()
        
        if not stats:
            print("暂无统计数据，请先生成统计")
            return
        
        tournament = db.query(Tournament).filter(
            Tournament.id == args.tournament
        ).first()
        
        if tournament:
            print(f"\n{tournament.name} - 积分榜")
            print("=" * 80)
        
        headers = ["排名", "队伍", "场次", "胜", "负", "平", "进球", "失球", "净胜", "积分"]
        rows = []
        for s in stats:
            team = db.query(Team).filter(Team.id == s.team_id).first()
            team_name = team.name if team else "未知"
            
            if team and (team.is_forfeited or team.is_withdrawn):
                status = "(弃权)" if team.is_forfeited else "(退赛)"
                team_name = f"{team_name} {status}"
            
            rows.append([
                s.rank or "-",
                team_name,
                s.matches_played if s.matches_played is not None else "-",
                s.wins if s.wins is not None else "-",
                s.losses if s.losses is not None else "-",
                s.draws if s.draws is not None else "-",
                s.goals_for if s.goals_for is not None else "-",
                s.goals_against if s.goals_against is not None else "-",
                s.goal_difference if s.goal_difference is not None else "-",
                s.points if s.points is not None else "-",
            ])
        
        print(tabulate(rows, headers=headers, tablefmt="grid"))
    finally:
        db.close()


def run_scheduler(args):
    from app.scheduler import run_scheduled_tasks_manual
    run_scheduled_tasks_manual()


def main():
    init_db()
    
    parser = argparse.ArgumentParser(
        prog="gym-cli",
        description="体育馆场地预约管理系统 - 命令行客户端"
    )
    
    subparsers = parser.add_subparsers(dest="command", help="可用命令")
    
    bookings_parser = subparsers.add_parser("bookings", help="查看预约列表")
    bookings_parser.add_argument("--user", type=int, help="按用户ID筛选")
    bookings_parser.add_argument("--venue", type=int, help="按场地ID筛选")
    bookings_parser.add_argument("--status", help="按状态筛选")
    bookings_parser.set_defaults(func=list_bookings)
    
    booking_parser = subparsers.add_parser("booking", help="查看预约详情")
    booking_parser.add_argument("id", type=int, help="预约ID")
    booking_parser.set_defaults(func=show_booking)
    
    matches_parser = subparsers.add_parser("matches", help="查看赛程")
    matches_parser.add_argument("--tournament", type=int, help="按赛事ID筛选")
    matches_parser.add_argument("--team", type=int, help="按队伍ID筛选")
    matches_parser.add_argument("--round", type=int, help="按轮次筛选")
    matches_parser.set_defaults(func=list_matches)
    
    tournaments_parser = subparsers.add_parser("tournaments", help="查看赛事列表")
    tournaments_parser.set_defaults(func=list_tournaments)
    
    stats_parser = subparsers.add_parser("stats", help="查看赛事统计")
    stats_parser.add_argument("tournament", type=int, help="赛事ID")
    stats_parser.set_defaults(func=tournament_stats)
    
    classes_parser = subparsers.add_parser("classes", help="查看培训班列表")
    classes_parser.add_argument("--status", help="按状态筛选")
    classes_parser.set_defaults(func=list_classes)
    
    venues_parser = subparsers.add_parser("venues", help="查看场地列表")
    venues_parser.set_defaults(func=list_venues)
    
    users_parser = subparsers.add_parser("users", help="查看用户列表")
    users_parser.set_defaults(func=list_users)
    
    scheduler_parser = subparsers.add_parser("scheduler", help="运行定时任务")
    scheduler_parser.set_defaults(func=run_scheduler)
    
    args = parser.parse_args()
    
    if not args.command:
        parser.print_help()
        return
    
    args.func(args)


if __name__ == "__main__":
    main()
