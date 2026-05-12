from datetime import datetime, timedelta, date
from sqlalchemy.orm import Session
from typing import List, Optional, Dict, Any
from app.models import Train, Route, RouteStatus, RouteOptimization, DutyRecord, MonthlyWorkHours, Alert, AlertType, AlertStatus, CrewType
from app.services.train_service import RouteService, TrainService
from app.services.crew_service import DutyRulesService


class OptimizationService:
    MIN_UTILIZATION_HOURS_PER_DAY = 8
    MAX_TURNAROUND_MINUTES = 120

    @staticmethod
    def calculate_train_utilization(db: Session, train_id: int, 
                                    start_date: date, end_date: date) -> Dict[str, Any]:
        train = TrainService.get_train(db, train_id)
        if not train:
            raise ValueError("列车不存在")
        
        start_dt = datetime.combine(start_date, datetime.min.time())
        end_dt = datetime.combine(end_date, datetime.max.time())
        
        routes = db.query(Route).filter(
            Route.train_id == train_id,
            Route.scheduled_departure >= start_dt,
            Route.scheduled_departure <= end_dt,
            Route.status != RouteStatus.CANCELLED
        ).all()
        
        total_service_minutes = 0
        total_turnaround_minutes = 0
        turnaround_count = 0
        
        routes_sorted = sorted(routes, key=lambda r: r.scheduled_departure)
        
        for i, route in enumerate(routes_sorted):
            start_time = route.actual_departure or route.scheduled_departure
            end_time = route.actual_arrival or route.scheduled_arrival
            service_minutes = (end_time - start_time).total_seconds() / 60
            total_service_minutes += service_minutes
            
            if i > 0:
                prev_route = routes_sorted[i - 1]
                prev_end = prev_route.actual_arrival or prev_route.scheduled_arrival
                current_start = route.scheduled_departure
                turnaround = (current_start - prev_end).total_seconds() / 60
                if turnaround > 0:
                    total_turnaround_minutes += turnaround
                    turnaround_count += 1
        
        days_diff = (end_date - start_date).days + 1
        avg_daily_hours = (total_service_minutes / 60) / days_diff if days_diff > 0 else 0
        avg_turnaround = total_turnaround_minutes / turnaround_count if turnaround_count > 0 else 0
        
        return {
            "train_id": train_id,
            "train_number": train.train_number,
            "total_service_minutes": total_service_minutes,
            "total_service_hours": total_service_minutes / 60,
            "avg_daily_hours": avg_daily_hours,
            "avg_turnaround_minutes": avg_turnaround,
            "turnaround_count": turnaround_count,
            "days_analyzed": days_diff,
            "route_count": len(routes)
        }

    @staticmethod
    def analyze_all_trains(db: Session, days: int = 7) -> List[Dict[str, Any]]:
        end_date = date.today()
        start_date = end_date - timedelta(days=days - 1)
        
        trains = TrainService.list_trains(db)
        results = []
        
        for train in trains:
            try:
                utilization = OptimizationService.calculate_train_utilization(
                    db, train.id, start_date, end_date
                )
                results.append(utilization)
            except Exception:
                continue
        
        return results

    @staticmethod
    def generate_optimization_suggestions(db: Session, days: int = 7) -> List[RouteOptimization]:
        end_date = date.today()
        start_date = end_date - timedelta(days=days - 1)
        
        trains = TrainService.list_trains(db)
        suggestions = []
        
        for train in trains:
            try:
                utilization = OptimizationService.calculate_train_utilization(
                    db, train.id, start_date, end_date
                )
                
                if utilization["avg_daily_hours"] < OptimizationService.MIN_UTILIZATION_HOURS_PER_DAY:
                    suggestion = OptimizationService._create_low_utilization_suggestion(
                        db, train, utilization
                    )
                    suggestions.append(suggestion)
                
                if utilization["avg_turnaround_minutes"] > OptimizationService.MAX_TURNAROUND_MINUTES:
                    suggestion = OptimizationService._create_long_turnaround_suggestion(
                        db, train, utilization
                    )
                    suggestions.append(suggestion)
                    
            except Exception:
                continue
        
        return suggestions

    @staticmethod
    def _create_low_utilization_suggestion(db: Session, train: Train, 
                                            utilization: Dict[str, Any]) -> RouteOptimization:
        description = (
            f"动车组{train.train_number}日均运用{utilization['avg_daily_hours']:.1f}小时，"
            f"低于{OptimizationService.MIN_UTILIZATION_HOURS_PER_DAY}小时标准"
        )
        suggestion = (
            f"【优化建议】\n"
            f"动车组：{train.train_number}（{train.train_type}）\n"
            f"问题：日均运用{utilization['avg_daily_hours']:.1f}小时，利用率偏低\n"
            f"分析期：{utilization['days_analyzed']}天，共执行{utilization['route_count']}趟交路\n"
            f"\n"
            f"建议措施：\n"
            f"1. 重新规划交路图，增加该动车组的日间运营交路\n"
            f"2. 考虑与其他动车组进行交路互补，平衡整体利用率\n"
            f"3. 如该车型运量不足，可考虑调配至客流高峰线路\n"
            f"4. 若为备用车，可安排高峰时段加车任务"
        )
        
        opt = RouteOptimization(
            train_id=train.id,
            optimization_type="低利用率",
            description=description,
            suggestion=suggestion,
            avg_utilization_hours=utilization["avg_daily_hours"],
            avg_turnaround_minutes=None
        )
        db.add(opt)
        db.commit()
        db.refresh(opt)
        return opt

    @staticmethod
    def _create_long_turnaround_suggestion(db: Session, train: Train, 
                                            utilization: Dict[str, Any]) -> RouteOptimization:
        description = (
            f"动车组{train.train_number}平均折返时间{utilization['avg_turnaround_minutes']:.0f}分钟，"
            f"超过{OptimizationService.MAX_TURNAROUND_MINUTES}分钟标准"
        )
        suggestion = (
            f"【优化建议】\n"
            f"动车组：{train.train_number}（{train.train_type}）\n"
            f"问题：平均折返时间{utilization['avg_turnaround_minutes']:.0f}分钟，过长\n"
            f"涉及折返：{utilization['turnaround_count']}次\n"
            f"\n"
            f"建议措施：\n"
            f"1. 优化交路接续，减少折返等待时间\n"
            f"2. 检查车站接发能力，协调调度安排\n"
            f"3. 考虑在折返站安排快速整备作业\n"
            f"4. 适当增加该方向的交路密度，填补时间空隙"
        )
        
        opt = RouteOptimization(
            train_id=train.id,
            optimization_type="折返时间过长",
            description=description,
            suggestion=suggestion,
            avg_utilization_hours=utilization["avg_daily_hours"],
            avg_turnaround_minutes=utilization["avg_turnaround_minutes"]
        )
        db.add(opt)
        db.commit()
        db.refresh(opt)
        return opt

    @staticmethod
    def get_optimizations(db: Session, train_id: Optional[int] = None,
                          is_implemented: Optional[bool] = None) -> List[RouteOptimization]:
        query = db.query(RouteOptimization)
        if train_id:
            query = query.filter(RouteOptimization.train_id == train_id)
        if is_implemented is not None:
            query = query.filter(RouteOptimization.is_implemented == is_implemented)
        return query.order_by(RouteOptimization.created_at.desc()).all()

    @staticmethod
    def mark_implemented(db: Session, optimization_id: int) -> RouteOptimization:
        opt = db.query(RouteOptimization).filter(RouteOptimization.id == optimization_id).first()
        if not opt:
            raise ValueError("优化建议不存在")
        opt.is_implemented = True
        db.commit()
        db.refresh(opt)
        return opt


class WorkHoursService:
    @staticmethod
    def complete_duty_record(db: Session, record_id: int, 
                              end_time: Optional[datetime] = None) -> DutyRecord:
        record = db.query(DutyRecord).filter(DutyRecord.id == record_id).first()
        if not record:
            raise ValueError("值乘记录不存在")
        if record.is_completed:
            raise ValueError("值乘记录已完成")
        
        actual_end = end_time or datetime.utcnow()
        record.end_time = actual_end
        record.work_duration_minutes = int(
            (actual_end - record.start_time).total_seconds() / 60
        )
        record.is_completed = True
        
        db.commit()
        db.refresh(record)
        
        WorkHoursService.accumulate_to_monthly(db, record.crew_member_id, record)
        
        return record

    @staticmethod
    def accumulate_to_monthly(db: Session, crew_member_id: int, 
                               record: DutyRecord) -> MonthlyWorkHours:
        work_date = record.start_time.date()
        year = work_date.year
        month = work_date.month
        
        monthly_record = DutyRulesService.get_monthly_work_hours(db, crew_member_id, year, month)
        monthly_record.total_minutes += record.work_duration_minutes
        
        if monthly_record.total_minutes > monthly_record.max_limit_minutes:
            monthly_record.is_over_limit = True
            WorkHoursService.create_overtime_alert(db, crew_member_id, monthly_record)
        else:
            monthly_record.is_over_limit = False
        
        db.commit()
        db.refresh(monthly_record)
        
        return monthly_record

    @staticmethod
    def create_overtime_alert(db: Session, crew_member_id: int,
                               monthly_record: MonthlyWorkHours) -> Alert:
        from app.services.crew_service import CrewMemberService
        
        member = CrewMemberService.get_crew_member(db, crew_member_id)
        member_name = member.name if member else f"员工{crew_member_id}"
        
        total_hours = monthly_record.total_minutes / 60
        max_hours = monthly_record.max_limit_minutes / 60
        over_hours = total_hours - max_hours
        
        alert = Alert(
            alert_type=AlertType.OVERTIME,
            title=f"司机月度工时超限预警",
            message=(
                f"司机 {member_name} 本月累计工作 {total_hours:.1f} 小时，"
                f"已超过 {max_hours} 小时上限，超限 {over_hours:.1f} 小时。\n"
                f"请调度主任立即调整该司机的排班计划，安排必要休息。"
            ),
            target_role="调度主任",
            status=AlertStatus.PENDING,
            related_crew_member_id=crew_member_id
        )
        
        db.add(alert)
        db.commit()
        db.refresh(alert)
        
        return alert

    @staticmethod
    def get_monthly_report(db: Session, year: int, month: int, 
                           crew_type: Optional[CrewType] = None) -> List[Dict[str, Any]]:
        from app.services.crew_service import CrewMemberService
        
        members = CrewMemberService.list_crew_members(db, crew_type)
        report = []
        
        for member in members:
            monthly = DutyRulesService.get_monthly_work_hours(db, member.id, year, month)
            report.append({
                "crew_member_id": member.id,
                "employee_id": member.employee_id,
                "name": member.name,
                "crew_type": member.crew_type.value,
                "total_minutes": monthly.total_minutes,
                "total_hours": monthly.total_minutes / 60,
                "max_limit_hours": monthly.max_limit_minutes / 60,
                "remaining_hours": (monthly.max_limit_minutes - monthly.total_minutes) / 60,
                "is_over_limit": monthly.is_over_limit,
                "last_recalculated": monthly.last_recalculated
            })
        
        return sorted(report, key=lambda x: x["total_minutes"], reverse=True)

    @staticmethod
    def get_alerts(db: Session, status: Optional[AlertStatus] = None,
                   alert_type: Optional[AlertType] = None,
                   limit: int = 100) -> List[Alert]:
        query = db.query(Alert)
        if status:
            query = query.filter(Alert.status == status)
        if alert_type:
            query = query.filter(Alert.alert_type == alert_type)
        return query.order_by(Alert.created_at.desc()).limit(limit).all()

    @staticmethod
    def acknowledge_alert(db: Session, alert_id: int) -> Alert:
        alert = db.query(Alert).filter(Alert.id == alert_id).first()
        if not alert:
            raise ValueError("预警不存在")
        alert.status = AlertStatus.ACKNOWLEDGED
        alert.acknowledged_at = datetime.utcnow()
        db.commit()
        db.refresh(alert)
        return alert

    @staticmethod
    def resolve_alert(db: Session, alert_id: int) -> Alert:
        alert = db.query(Alert).filter(Alert.id == alert_id).first()
        if not alert:
            raise ValueError("预警不存在")
        alert.status = AlertStatus.RESOLVED
        db.commit()
        db.refresh(alert)
        return alert
