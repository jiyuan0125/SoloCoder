from datetime import datetime, timedelta, date
from sqlalchemy.orm import Session
from typing import List, Optional
from app.models import CrewMember, CrewGroup, CrewAssignment, DutyRecord, MonthlyWorkHours, Alert, AlertType, AlertStatus, CrewType
from app.schemas.models import CrewMemberCreate, CrewMemberUpdate, CrewGroupCreate, CrewGroupUpdate, DutyRecordCreate, DutyRecordUpdate


class CrewMemberService:
    @staticmethod
    def create_crew_member(db: Session, member: CrewMemberCreate):
        existing = db.query(CrewMember).filter(
            CrewMember.employee_id == member.employee_id
        ).first()
        if existing:
            raise ValueError(f"工号{member.employee_id}已存在")
        
        db_member = CrewMember(**member.model_dump())
        db.add(db_member)
        db.commit()
        db.refresh(db_member)
        return db_member

    @staticmethod
    def get_crew_member(db: Session, member_id: int):
        return db.query(CrewMember).filter(CrewMember.id == member_id).first()

    @staticmethod
    def get_crew_member_by_employee_id(db: Session, employee_id: str):
        return db.query(CrewMember).filter(CrewMember.employee_id == employee_id).first()

    @staticmethod
    def list_crew_members(db: Session, crew_type: Optional[CrewType] = None):
        query = db.query(CrewMember)
        if crew_type:
            query = query.filter(CrewMember.crew_type == crew_type)
        return query.all()

    @staticmethod
    def update_crew_member(db: Session, member_id: int, update: CrewMemberUpdate):
        member = CrewMemberService.get_crew_member(db, member_id)
        if not member:
            return None
        for key, value in update.model_dump(exclude_unset=True).items():
            setattr(member, key, value)
        db.commit()
        db.refresh(member)
        return member

    @staticmethod
    def delete_crew_member(db: Session, member_id: int):
        member = CrewMemberService.get_crew_member(db, member_id)
        if member:
            db.delete(member)
            db.commit()
            return True
        return False


class CrewGroupService:
    @staticmethod
    def create_crew_group(db: Session, group: CrewGroupCreate):
        existing = db.query(CrewGroup).filter(CrewGroup.group_code == group.group_code).first()
        if existing:
            raise ValueError(f"乘务组代码{group.group_code}已存在")
        
        db_group = CrewGroup(group_code=group.group_code, name=group.name)
        db.add(db_group)
        db.flush()
        
        for member_id in group.member_ids:
            member = CrewMemberService.get_crew_member(db, member_id)
            if not member:
                raise ValueError(f"乘务员ID {member_id} 不存在")
            
            assignment = CrewAssignment(
                crew_group_id=db_group.id,
                crew_member_id=member_id,
                is_lead=(member_id == group.lead_member_id)
            )
            db.add(assignment)
        
        db.commit()
        db.refresh(db_group)
        return db_group

    @staticmethod
    def get_crew_group(db: Session, group_id: int):
        return db.query(CrewGroup).filter(CrewGroup.id == group_id).first()

    @staticmethod
    def get_crew_group_by_code(db: Session, group_code: str):
        return db.query(CrewGroup).filter(CrewGroup.group_code == group_code).first()

    @staticmethod
    def list_crew_groups(db: Session):
        return db.query(CrewGroup).all()

    @staticmethod
    def get_group_members(db: Session, group_id: int):
        assignments = db.query(CrewAssignment).filter(
            CrewAssignment.crew_group_id == group_id
        ).all()
        member_ids = [a.crew_member_id for a in assignments]
        return db.query(CrewMember).filter(CrewMember.id.in_(member_ids)).all()


class DutyRulesService:
    MAX_CONTINUOUS_HOURS = 4
    MIN_REST_AFTER_CONTINUOUS = 30
    MAX_DAILY_HOURS = 8
    MAX_MONTHLY_HOURS = 160

    @staticmethod
    def get_crew_member_daily_work(db: Session, member_id: int, work_date: date) -> List[DutyRecord]:
        start_of_day = datetime.combine(work_date, datetime.min.time())
        end_of_day = datetime.combine(work_date, datetime.max.time())
        
        return db.query(DutyRecord).filter(
            DutyRecord.crew_member_id == member_id,
            DutyRecord.start_time >= start_of_day,
            DutyRecord.start_time <= end_of_day
        ).all()

    @staticmethod
    def calculate_daily_work_minutes(db: Session, member_id: int, work_date: date) -> int:
        records = DutyRulesService.get_crew_member_daily_work(db, member_id, work_date)
        return sum(r.work_duration_minutes for r in records if r.is_completed)

    @staticmethod
    def check_continuous_work_compliance(db: Session, member_id: int, 
                                         new_start_time: datetime, 
                                         expected_duration_minutes: int) -> tuple[bool, str]:
        work_date = new_start_time.date()
        daily_records = DutyRulesService.get_crew_member_daily_work(db, member_id, work_date)
        
        consecutive_minutes = expected_duration_minutes
        last_rest_time = None
        
        for record in daily_records:
            if not record.is_completed:
                continue
            
            if last_rest_time is None:
                consecutive_minutes += record.work_duration_minutes
            else:
                gap = (record.start_time - last_rest_time).total_seconds() / 60
                if gap >= DutyRulesService.MIN_REST_AFTER_CONTINUOUS:
                    consecutive_minutes = record.work_duration_minutes
                else:
                    consecutive_minutes += record.work_duration_minutes
            
            last_rest_time = record.end_time or record.start_time
        
        max_continuous_minutes = DutyRulesService.MAX_CONTINUOUS_HOURS * 60
        if consecutive_minutes > max_continuous_minutes:
            return False, (f"连续值乘时间{consecutive_minutes}分钟超过{DutyRulesService.MAX_CONTINUOUS_HOURS}小时，"
                          f"需休息{DutyRulesService.MIN_REST_AFTER_CONTINUOUS}分钟后再值乘")
        
        return True, "符合连续值乘规则"

    @staticmethod
    def check_daily_work_compliance(db: Session, member_id: int, work_date: date,
                                    additional_minutes: int = 0) -> tuple[bool, str]:
        current_minutes = DutyRulesService.calculate_daily_work_minutes(db, member_id, work_date)
        total_minutes = current_minutes + additional_minutes
        max_daily_minutes = DutyRulesService.MAX_DAILY_HOURS * 60
        
        if total_minutes > max_daily_minutes:
            return False, f"当日工作时间{total_minutes}分钟超过{DutyRulesService.MAX_DAILY_HOURS}小时限制"
        
        return True, "符合日工作时间规则"

    @staticmethod
    def get_monthly_work_hours(db: Session, member_id: int, year: int, month: int) -> MonthlyWorkHours:
        record = db.query(MonthlyWorkHours).filter(
            MonthlyWorkHours.crew_member_id == member_id,
            MonthlyWorkHours.year == year,
            MonthlyWorkHours.month == month
        ).first()
        
        if not record:
            record = MonthlyWorkHours(
                crew_member_id=member_id,
                year=year,
                month=month,
                total_minutes=0,
                max_limit_minutes=DutyRulesService.MAX_MONTHLY_HOURS * 60
            )
            db.add(record)
            db.commit()
            db.refresh(record)
        
        return record

    @staticmethod
    def check_monthly_work_compliance(db: Session, member_id: int, year: int, 
                                      month: int, additional_minutes: int = 0) -> tuple[bool, str]:
        monthly_record = DutyRulesService.get_monthly_work_hours(db, member_id, year, month)
        total_minutes = monthly_record.total_minutes + additional_minutes
        
        if total_minutes > monthly_record.max_limit_minutes:
            hours = total_minutes / 60
            max_hours = monthly_record.max_limit_minutes / 60
            return False, f"月度工作时间{hours:.1f}小时超过{max_hours}小时限制"
        
        return True, "符合月度工作时间规则"

    @staticmethod
    def check_crew_compliance_for_route(db: Session, crew_group_id: Optional[int], 
                                         route_duration_minutes: int,
                                         route_start_time: datetime) -> tuple[bool, str]:
        if not crew_group_id:
            return True, "未分配乘务组，跳过检查"
        
        members = CrewGroupService.get_group_members(db, crew_group_id)
        work_date = route_start_time.date()
        
        for member in members:
            if member.crew_type == CrewType.DRIVER:
                ok, msg = DutyRulesService.check_continuous_work_compliance(
                    db, member.id, route_start_time, route_duration_minutes
                )
                if not ok:
                    return False, f"司机{member.name}: {msg}"
                
                ok, msg = DutyRulesService.check_daily_work_compliance(
                    db, member.id, work_date, route_duration_minutes
                )
                if not ok:
                    return False, f"司机{member.name}: {msg}"
                
                ok, msg = DutyRulesService.check_monthly_work_compliance(
                    db, member.id, work_date.year, work_date.month, route_duration_minutes
                )
                if not ok:
                    return False, f"司机{member.name}: {msg}"
        
        return True, "乘务组符合值乘规则"
