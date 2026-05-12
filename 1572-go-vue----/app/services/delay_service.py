from datetime import datetime, timedelta
from sqlalchemy.orm import Session
from typing import List, Optional, Dict, Any
from app.models import Route, RouteStatus, CrewGroup, DelayAdjustment, DutyRecord, MonthlyWorkHours
from app.services.train_service import RouteService
from app.services.crew_service import CrewGroupService


class DelayService:
    DELAY_THRESHOLD_MINUTES = 30

    @staticmethod
    def check_and_handle_delay(db: Session, route_id: int, 
                                actual_departure: Optional[datetime] = None,
                                actual_arrival: Optional[datetime] = None) -> Optional[DelayAdjustment]:
        route = RouteService.get_route(db, route_id)
        if not route:
            raise ValueError("交路不存在")
        
        delay_minutes = 0
        
        if actual_departure and route.scheduled_departure:
            dep_delay = (actual_departure - route.scheduled_departure).total_seconds() / 60
            delay_minutes = max(delay_minutes, dep_delay)
        
        if actual_arrival and route.scheduled_arrival:
            arr_delay = (actual_arrival - route.scheduled_arrival).total_seconds() / 60
            delay_minutes = max(delay_minutes, arr_delay)
        
        route.delay_minutes = int(delay_minutes)
        
        if delay_minutes >= DelayService.DELAY_THRESHOLD_MINUTES:
            route.status = RouteStatus.DELAYED
            adjustment = DelayService.create_adjustment_plan(db, route, int(delay_minutes))
            db.commit()
            db.refresh(route)
            return adjustment
        else:
            db.commit()
            db.refresh(route)
        
        return None

    @staticmethod
    def has_following_routes(db: Session, crew_group_id: int, 
                              route_end_time: datetime, 
                              route_date: datetime) -> bool:
        end_of_day = datetime.combine(route_date.date(), datetime.max.time())
        
        following_routes = db.query(Route).filter(
            Route.crew_group_id == crew_group_id,
            Route.scheduled_departure > route_end_time,
            Route.scheduled_departure <= end_of_day,
            Route.status.in_([RouteStatus.SCHEDULED, RouteStatus.RUNNING])
        ).first()
        
        return following_routes is not None

    @staticmethod
    def create_adjustment_plan(db: Session, route: Route, delay_minutes: int) -> DelayAdjustment:
        route_end_time = route.scheduled_arrival + timedelta(minutes=delay_minutes)
        
        has_following = False
        if route.crew_group_id:
            has_following = DelayService.has_following_routes(
                db, route.crew_group_id, route_end_time, route.scheduled_departure
            )
        
        adjustment_type = "晚点信息通知"
        adjustment_plan = DelayService._generate_information_only_plan(route, delay_minutes)
        
        if has_following:
            adjustment_type = "乘务调整方案"
            adjustment_plan = DelayService._generate_crew_adjustment_plan(
                db, route, delay_minutes
            )
        
        adjustment = DelayAdjustment(
            route_id=route.id,
            delay_minutes=delay_minutes,
            adjustment_type=adjustment_type,
            adjustment_plan=adjustment_plan,
            notified_to="值班调度"
        )
        
        db.add(adjustment)
        db.commit()
        db.refresh(adjustment)
        
        return adjustment

    @staticmethod
    def _generate_information_only_plan(route: Route, delay_minutes: int) -> str:
        plan_lines = [
            "=" * 60,
            "乘务组晚点调整方案（仅通知）",
            "=" * 60,
            "",
            f"交路编号: {route.route_code}",
            f"交路区间: {route.departure_station} → {route.arrival_station}",
            "",
            "【晚点信息】",
            f"计划出发: {route.scheduled_departure.strftime('%Y-%m-%d %H:%M')}",
            f"计划到达: {route.scheduled_arrival.strftime('%Y-%m-%d %H:%M')}",
            f"晚点时长: {delay_minutes} 分钟",
            "",
            "【乘务安排】",
            "该乘务组当日无后续交路，无需调整人员排班。",
            "",
            "【备注】",
            "请值班调度密切关注晚点情况，及时通报相关部门。",
            "=" * 60,
        ]
        return "\n".join(plan_lines)

    @staticmethod
    def _generate_crew_adjustment_plan(db: Session, route: Route, delay_minutes: int) -> str:
        route_end_time = route.scheduled_arrival + timedelta(minutes=delay_minutes)
        route_date = route.scheduled_departure.date()
        end_of_day = datetime.combine(route_date, datetime.max.time())
        
        following_routes = db.query(Route).filter(
            Route.crew_group_id == route.crew_group_id,
            Route.scheduled_departure > route_end_time,
            Route.scheduled_departure <= end_of_day,
            Route.status.in_([RouteStatus.SCHEDULED, RouteStatus.RUNNING])
        ).order_by(Route.scheduled_departure).all()
        
        crew_group = CrewGroupService.get_crew_group(db, route.crew_group_id) if route.crew_group_id else None
        group_name = crew_group.name if crew_group else "未指定"
        
        plan_lines = [
            "=" * 60,
            "乘务组晚点调整方案",
            "=" * 60,
            "",
            f"交路编号: {route.route_code}",
            f"交路区间: {route.departure_station} → {route.arrival_station}",
            f"乘务组: {group_name}",
            "",
            "【晚点信息】",
            f"计划出发: {route.scheduled_departure.strftime('%Y-%m-%d %H:%M')}",
            f"计划到达: {route.scheduled_arrival.strftime('%Y-%m-%d %H:%M')}",
            f"预计新到达: {route_end_time.strftime('%Y-%m-%d %H:%M')}",
            f"晚点时长: {delay_minutes} 分钟",
            "",
            "【受影响的后续交路】",
        ]
        
        if following_routes:
            for i, fr in enumerate(following_routes, 1):
                gap = (fr.scheduled_departure - route_end_time).total_seconds() / 60
                plan_lines.append(f"  {i}. 交路 {fr.route_code}: {fr.departure_station}→{fr.arrival_station}")
                plan_lines.append(f"     计划出发: {fr.scheduled_departure.strftime('%H:%M')}")
                plan_lines.append(f"     预计衔接间隔: {int(gap)} 分钟")
                if gap < 30:
                    plan_lines.append(f"     ⚠️ 折返时间不足30分钟，建议调整")
        else:
            plan_lines.append("  无后续交路")
        
        plan_lines.extend([
            "",
            "【建议调整方案】",
            "方案一：待车调整",
            "  - 等待当前交路的动车组完成折返",
            "  - 后续交路全部顺延",
            "  - 适用：短时间晚点，影响范围有限",
            "",
            "方案二：交路调整",
            "  - 调配备用动车组接替后续交路",
            "  - 当前乘务组继续值乘备用车",
            "  - 适用：动车组紧张但有备用车",
            "",
            "方案三：人员轮换",
            "  - 安排备用乘务组接替后续交路",
            "  - 当前乘务组完成本交路后休息",
            "  - 适用：乘务组工时接近上限",
            "",
            "【调度操作建议】",
            "1. 立即通知车站和旅客晚点信息",
            "2. 检查备用动车组状态",
            "3. 核实乘务组剩余可用工时",
            "4. 选择最优方案并执行",
            "5. 更新系统中的交路状态",
            "",
            "=" * 60,
        ])
        
        return "\n".join(plan_lines)

    @staticmethod
    def recover_from_delay(db: Session, route_id: int) -> None:
        route = RouteService.get_route(db, route_id)
        if not route:
            raise ValueError("交路不存在")
        
        route.status = RouteStatus.RUNNING
        route.delay_minutes = 0
        
        if route.crew_group_id:
            crew_members = CrewGroupService.get_group_members(db, route.crew_group_id)
            for member in crew_members:
                DelayService.recalculate_monthly_hours(db, member.id)
        
        db.commit()
        db.refresh(route)

    @staticmethod
    def recalculate_monthly_hours(db: Session, member_id: int) -> None:
        now = datetime.utcnow()
        year = now.year
        month = now.month
        
        completed_records = db.query(DutyRecord).filter(
            DutyRecord.crew_member_id == member_id,
            DutyRecord.is_completed == True,
            DutyRecord.start_time >= datetime(year, month, 1)
        ).all()
        
        total_minutes = sum(r.work_duration_minutes for r in completed_records)
        
        from app.services.crew_service import DutyRulesService
        monthly_record = DutyRulesService.get_monthly_work_hours(db, member_id, year, month)
        monthly_record.total_minutes = total_minutes
        monthly_record.is_over_limit = total_minutes > monthly_record.max_limit_minutes
        
        db.commit()

    @staticmethod
    def get_delay_adjustments(db: Session, route_id: Optional[int] = None,
                               limit: int = 100) -> List[DelayAdjustment]:
        query = db.query(DelayAdjustment)
        if route_id:
            query = query.filter(DelayAdjustment.route_id == route_id)
        return query.order_by(DelayAdjustment.created_at.desc()).limit(limit).all()
