from datetime import datetime, date, timedelta
from typing import Optional, List, Dict, Any
from sqlalchemy.orm import Session
from sqlalchemy import func, and_

from .models import (
    Officer, Channel, Schedule, CheckIn, Contraband, TrafficLog,
    ShiftType, RoleType, ChannelType, ChannelStatus,
    DisposalType, ContrabandType
)
from .schemas import (
    ChannelStatusInfo, ChannelCapacityAlert, TrafficSummary
)


SHIFT_TIMES = {
    ShiftType.MORNING: (6, 14),
    ShiftType.MIDDAY: (14, 22),
    ShiftType.NIGHT: (22, 30),
}


def get_current_shift(dt: datetime) -> ShiftType:
    hour = dt.hour
    for shift, (start, end) in SHIFT_TIMES.items():
        adjusted_end = end % 24
        if start <= hour < adjusted_end or (start > adjusted_end and (hour >= start or hour < adjusted_end)):
            return shift
    return ShiftType.MORNING


class ScheduleService:
    @staticmethod
    def check_consecutive_days(db: Session, officer_id: int, schedule_date: date) -> int:
        six_days_before = schedule_date - timedelta(days=6)
        schedules = db.query(Schedule).filter(
            Schedule.officer_id == officer_id,
            Schedule.schedule_date >= six_days_before,
            Schedule.schedule_date <= schedule_date
        ).all()
        
        date_set = {s.schedule_date for s in schedules}
        date_set.add(schedule_date)
        
        max_consecutive = 0
        current_consecutive = 1
        sorted_dates = sorted(date_set)
        
        for i in range(1, len(sorted_dates)):
            if (sorted_dates[i] - sorted_dates[i-1]).days == 1:
                current_consecutive += 1
                max_consecutive = max(max_consecutive, current_consecutive)
            else:
                current_consecutive = 1
        
        max_consecutive = max(max_consecutive, current_consecutive)
        return max_consecutive

    @staticmethod
    def check_night_to_morning(db: Session, officer_id: int, schedule_date: date, new_shift: ShiftType) -> bool:
        day_before = schedule_date - timedelta(days=1)
        prev_schedule = db.query(Schedule).filter(
            Schedule.officer_id == officer_id,
            Schedule.schedule_date == day_before,
            Schedule.shift == ShiftType.NIGHT
        ).first()
        
        if prev_schedule and new_shift == ShiftType.MORNING:
            return True
        return False

    @staticmethod
    def validate_schedule(db: Session, officer_id: int, channel_id: int, schedule_date: date, shift: ShiftType) -> List[str]:
        errors = []
        
        consecutive = ScheduleService.check_consecutive_days(db, officer_id, schedule_date)
        if consecutive > 6:
            errors.append(f"连续工作天数不能超过6天，当前已有{consecutive}天")
        
        if ScheduleService.check_night_to_morning(db, officer_id, schedule_date, shift):
            errors.append("夜班后不能接早班，需要至少休息12小时")
        
        officer = db.query(Officer).filter(Officer.id == officer_id).first()
        channel = db.query(Channel).filter(Channel.id == channel_id).first()
        
        if officer and channel:
            existing = db.query(Schedule).filter(
                Schedule.channel_id == channel_id,
                Schedule.schedule_date == schedule_date,
                Schedule.shift == shift,
                Schedule.officer.has(role=officer.role)
            ).first()
            
            if existing and existing.officer_id != officer_id:
                if officer.role == RoleType.HANDCHECK and channel.channel_type == ChannelType.VIP:
                    handcheck_count = db.query(Schedule).filter(
                        Schedule.channel_id == channel_id,
                        Schedule.schedule_date == schedule_date,
                        Schedule.shift == shift,
                        Schedule.officer.has(role=RoleType.HANDCHECK)
                    ).count()
                    if handcheck_count >= 2:
                        errors.append("该班次已有两名手检员")
                elif officer.role != RoleType.HANDCHECK:
                    errors.append(f"该班次已有{officer.role.value}")
        
        return errors


class ChannelService:
    @staticmethod
    def check_channel_staffing(db: Session, channel_id: int, schedule_date: date, shift: ShiftType) -> ChannelStatusInfo:
        channel = db.query(Channel).filter(Channel.id == channel_id).first()
        if not channel:
            return ChannelStatusInfo(
                channel_id=channel_id,
                channel_name="Unknown",
                status=ChannelStatus.CLOSED,
                scanner_present=False,
                handcheck_count=0,
                verifier_present=False,
                is_valid=False,
                missing_roles=["通道不存在"]
            )
        
        schedules = db.query(Schedule).filter(
            Schedule.channel_id == channel_id,
            Schedule.schedule_date == schedule_date,
            Schedule.shift == shift
        ).all()
        
        roles = {}
        for sched in schedules:
            officer = sched.officer
            role = officer.role
            if role not in roles:
                roles[role] = 0
            roles[role] += 1
        
        scanner_present = roles.get(RoleType.SCANNER, 0) >= 1
        handcheck_count = roles.get(RoleType.HANDCHECK, 0)
        verifier_present = roles.get(RoleType.VERIFIER, 0) >= 1
        
        missing_roles = []
        is_valid = True
        
        if not scanner_present:
            missing_roles.append("开机员")
            is_valid = False
        if not verifier_present:
            missing_roles.append("验证员")
            is_valid = False
        
        min_handcheck = 2 if channel.channel_type == ChannelType.VIP else 1
        if handcheck_count < min_handcheck:
            missing_roles.append(f"手检员（需要{min_handcheck}人，当前{handcheck_count}人）")
            is_valid = False
        
        return ChannelStatusInfo(
            channel_id=channel_id,
            channel_name=channel.name,
            status=ChannelStatus.OPEN if is_valid else ChannelStatus.CLOSED,
            scanner_present=scanner_present,
            handcheck_count=handcheck_count,
            verifier_present=verifier_present,
            is_valid=is_valid,
            missing_roles=missing_roles
        )

    @staticmethod
    def update_channel_status(db: Session, channel_id: int) -> Channel:
        channel = db.query(Channel).filter(Channel.id == channel_id).first()
        if not channel:
            return None
        
        now = datetime.now()
        current_shift = get_current_shift(now)
        status_info = ChannelService.check_channel_staffing(db, channel_id, now.date(), current_shift)
        
        if status_info.is_valid and channel.status != ChannelStatus.OPEN:
            channel.status = ChannelStatus.OPEN
            db.commit()
            db.refresh(channel)
        elif not status_info.is_valid and channel.status != ChannelStatus.CLOSED:
            channel.status = ChannelStatus.CLOSED
            db.commit()
            db.refresh(channel)
        
        return channel

    @staticmethod
    def check_capacity_alert(db: Session, channel_id: int) -> Optional[ChannelCapacityAlert]:
        channel = db.query(Channel).filter(Channel.id == channel_id).first()
        if not channel:
            return None
        
        now = datetime.now()
        latest_traffic = db.query(TrafficLog).filter(
            TrafficLog.channel_id == channel_id,
            TrafficLog.log_date == now.date()
        ).order_by(TrafficLog.hour.desc()).first()
        
        if not latest_traffic:
            return None
        
        max_allowed = channel.capacity_per_hour * 3
        should_add = latest_traffic.queue_length > max_allowed
        
        if should_add:
            open_channels = db.query(Channel).filter(
                Channel.status == ChannelStatus.OPEN
            ).count()
            if open_channels < 2:
                should_add = False
                reason = "排队过长但开放通道不足2个，暂不建议增加"
            else:
                reason = f"排队人数{latest_traffic.queue_length}超过容量3倍({max_allowed})，建议增加通道"
        else:
            reason = "当前排队在容量范围内"
        
        return ChannelCapacityAlert(
            channel_id=channel_id,
            channel_name=channel.name,
            current_queue=latest_traffic.queue_length,
            capacity_per_hour=channel.capacity_per_hour,
            max_allowed=max_allowed,
            should_add_channel=should_add,
            reason=reason
        )


class ContrabandService:
    @staticmethod
    def validate_contraband(item_type: ContrabandType, disposal_type: DisposalType, police_badge: Optional[str]) -> List[str]:
        errors = []
        
        dangerous_types = [ContrabandType.EXPLOSIVE, ContrabandType.SIMULATED_WEAPON]
        
        if item_type in dangerous_types:
            if disposal_type != DisposalType.HANDOVER:
                errors.append("爆炸物和仿真武器必须移交公安，不能自弃")
            if not police_badge or not police_badge.strip():
                errors.append("移交公安时必须记录接收民警警号")
        else:
            if disposal_type == DisposalType.HANDOVER and not police_badge:
                errors.append("移交公安时必须记录接收民警警号")
        
        return errors


class CheckInService:
    @staticmethod
    def find_today_schedule(db: Session, officer_id: int, check_in_time: datetime) -> Optional[Schedule]:
        today = check_in_time.date()
        shift = get_current_shift(check_in_time)
        
        schedule = db.query(Schedule).filter(
            Schedule.officer_id == officer_id,
            Schedule.schedule_date == today,
            Schedule.shift == shift
        ).first()
        
        return schedule

    @staticmethod
    def create_check_in(db: Session, officer_id: int, check_in_time: Optional[datetime] = None) -> CheckIn:
        if check_in_time is None:
            check_in_time = datetime.now()
        
        schedule = CheckInService.find_today_schedule(db, officer_id, check_in_time)
        
        check_in = CheckIn(
            officer_id=officer_id,
            schedule_id=schedule.id if schedule else None,
            check_in_time=check_in_time,
            is_temporary=(schedule is None)
        )
        
        db.add(check_in)
        db.commit()
        db.refresh(check_in)
        
        return check_in


class TrafficService:
    @staticmethod
    def get_channel_traffic_summary(db: Session, channel_id: int, start_date: date, end_date: date) -> TrafficSummary:
        logs = db.query(TrafficLog).filter(
            TrafficLog.channel_id == channel_id,
            TrafficLog.log_date >= start_date,
            TrafficLog.log_date <= end_date
        ).all()
        
        total_passengers = sum(log.passenger_count for log in logs)
        max_queue = max(log.queue_length for log in logs) if logs else 0
        avg_queue = sum(log.queue_length for log in logs) / len(logs) if logs else 0.0
        
        return TrafficSummary(
            channel_id=channel_id,
            total_passengers=total_passengers,
            max_queue_length=max_queue,
            avg_queue_length=round(avg_queue, 2),
            date_range=f"{start_date} ~ {end_date}"
        )
