from datetime import datetime, timedelta, date
from typing import List, Optional, Dict, Tuple
from sqlalchemy.orm import Session
from sqlalchemy import and_, func
import numpy as np
from . import models
from .config import settings


class PipelineService:
    def __init__(self, db: Session):
        self.db = db

    def create_pipeline(self, data: dict) -> models.Pipeline:
        pipeline = models.Pipeline(**data)
        self.db.add(pipeline)
        self.db.commit()
        self.db.refresh(pipeline)
        return pipeline

    def get_pipeline(self, pipeline_id: int) -> Optional[models.Pipeline]:
        return self.db.query(models.Pipeline).filter(models.Pipeline.id == pipeline_id).first()

    def get_pipeline_by_code(self, code: str) -> Optional[models.Pipeline]:
        return self.db.query(models.Pipeline).filter(models.Pipeline.code == code).first()

    def list_pipelines(self, status: Optional[str] = None) -> List[models.Pipeline]:
        query = self.db.query(models.Pipeline)
        if status:
            query = query.filter(models.Pipeline.status == status)
        return query.all()

    def update_pipeline(self, pipeline_id: int, data: dict) -> Optional[models.Pipeline]:
        pipeline = self.get_pipeline(pipeline_id)
        if pipeline:
            for key, value in data.items():
                setattr(pipeline, key, value)
            self.db.commit()
            self.db.refresh(pipeline)
        return pipeline

    def delete_pipeline(self, pipeline_id: int) -> bool:
        pipeline = self.get_pipeline(pipeline_id)
        if pipeline:
            self.db.delete(pipeline)
            self.db.commit()
            return True
        return False


class SegmentService:
    def __init__(self, db: Session):
        self.db = db

    def create_segment(self, data: dict) -> models.Segment:
        risk_frequency_map = {
            models.RiskLevel.LOW: 14,
            models.RiskLevel.MEDIUM: 7,
            models.RiskLevel.HIGH: 3,
            models.RiskLevel.CRITICAL: 1
        }
        risk_level = data.get("risk_level", models.RiskLevel.MEDIUM)
        data["patrol_frequency_days"] = risk_frequency_map.get(risk_level, 7)
        segment = models.Segment(**data)
        self.db.add(segment)
        self.db.commit()
        self.db.refresh(segment)
        return segment

    def get_segment(self, segment_id: int) -> Optional[models.Segment]:
        return self.db.query(models.Segment).filter(models.Segment.id == segment_id).first()

    def list_segments(self, pipeline_id: Optional[int] = None) -> List[models.Segment]:
        query = self.db.query(models.Segment)
        if pipeline_id:
            query = query.filter(models.Segment.pipeline_id == pipeline_id)
        return query.all()

    def update_segment(self, segment_id: int, data: dict) -> Optional[models.Segment]:
        segment = self.get_segment(segment_id)
        if segment:
            if "risk_level" in data:
                risk_frequency_map = {
                    models.RiskLevel.LOW: 14,
                    models.RiskLevel.MEDIUM: 7,
                    models.RiskLevel.HIGH: 3,
                    models.RiskLevel.CRITICAL: 1
                }
                data["patrol_frequency_days"] = risk_frequency_map.get(data["risk_level"], 7)
            for key, value in data.items():
                setattr(segment, key, value)
            self.db.commit()
            self.db.refresh(segment)
        return segment

    def delete_segment(self, segment_id: int) -> bool:
        segment = self.get_segment(segment_id)
        if segment:
            self.db.delete(segment)
            self.db.commit()
            return True
        return False


class PressureService:
    def __init__(self, db: Session):
        self.db = db

    def create_pressure_point(self, data: dict) -> models.PressurePoint:
        point = models.PressurePoint(**data)
        self.db.add(point)
        self.db.commit()
        self.db.refresh(point)
        return point

    def get_pressure_point(self, point_id: int) -> Optional[models.PressurePoint]:
        return self.db.query(models.PressurePoint).filter(models.PressurePoint.id == point_id).first()

    def list_pressure_points(self, pipeline_id: int) -> List[models.PressurePoint]:
        return self.db.query(models.PressurePoint).filter(
            models.PressurePoint.pipeline_id == pipeline_id
        ).order_by(models.PressurePoint.position).all()

    def add_pressure_reading(self, pressure_point_id: int, pressure: float) -> models.PressureReading:
        reading = models.PressureReading(
            pressure_point_id=pressure_point_id,
            pressure=pressure
        )
        self.db.add(reading)
        self.db.commit()
        self.db.refresh(reading)
        return reading

    def get_adjacent_points(self, pipeline_id: int) -> List[Tuple[models.PressurePoint, models.PressurePoint]]:
        points = self.list_pressure_points(pipeline_id)
        pairs = []
        for i in range(len(points) - 1):
            pairs.append((points[i], points[i + 1]))
        return pairs

    def get_latest_reading(self, pressure_point_id: int) -> Optional[models.PressureReading]:
        return self.db.query(models.PressureReading).filter(
            models.PressureReading.pressure_point_id == pressure_point_id
        ).order_by(models.PressureReading.timestamp.desc()).first()

    def calculate_pressure_diff(self, start_point_id: int, end_point_id: int) -> Optional[float]:
        start_reading = self.get_latest_reading(start_point_id)
        end_reading = self.get_latest_reading(end_point_id)
        if start_reading and end_reading:
            return abs(start_reading.pressure - end_reading.pressure)
        return None

    def get_historical_diffs(self, pipeline_id: int, days: int = settings.PRESSURE_DIFF_STDDEV_DAYS) -> List[float]:
        cutoff = datetime.utcnow() - timedelta(days=days)
        diffs = self.db.query(models.PressureDiff).filter(
            and_(
                models.PressureDiff.pipeline_id == pipeline_id,
                models.PressureDiff.timestamp >= cutoff
            )
        ).all()
        return [d.pressure_diff for d in diffs]

    def calculate_stats(self, pipeline_id: int) -> Dict[str, float]:
        diffs = self.get_historical_diffs(pipeline_id)
        if len(diffs) == 0:
            return {"mean": 0.0, "stddev": 0.0, "count": 0}
        arr = np.array(diffs)
        return {
            "mean": float(np.mean(arr)),
            "stddev": float(np.std(arr)),
            "count": len(arr)
        }

    def check_leak_suspected(self, start_point_id: int, end_point_id: int, pipeline_id: int) -> Tuple[bool, Dict]:
        diff = self.calculate_pressure_diff(start_point_id, end_point_id)
        if diff is None:
            return False, {}

        stats = self.calculate_stats(pipeline_id)
        threshold = stats["mean"] + settings.LEAK_THRESHOLD_MULTIPLIER * stats["stddev"]

        if stats["count"] == 0:
            threshold = diff * 1.5

        is_suspected = diff > threshold

        return is_suspected, {
            "pressure_diff": diff,
            "mean": stats["mean"],
            "stddev": stats["stddev"],
            "threshold": threshold,
            "data_count": stats["count"]
        }

    def record_pressure_diff(self, pipeline_id: int, start_point_id: int, end_point_id: int, diff: float) -> models.PressureDiff:
        pd = models.PressureDiff(
            pipeline_id=pipeline_id,
            start_point_id=start_point_id,
            end_point_id=end_point_id,
            pressure_diff=diff
        )
        self.db.add(pd)
        self.db.commit()
        self.db.refresh(pd)
        return pd

    def aggregate_hourly_data(self):
        now = datetime.utcnow()
        start_of_hour = now.replace(minute=0, second=0, microsecond=0)
        hour_ago = start_of_hour - timedelta(hours=1)

        points = self.db.query(models.PressurePoint).all()
        for point in points:
            readings = self.db.query(models.PressureReading).filter(
                and_(
                    models.PressureReading.pressure_point_id == point.id,
                    models.PressureReading.timestamp >= hour_ago,
                    models.PressureReading.timestamp < start_of_hour,
                    models.PressureReading.is_aggregated == False
                )
            ).all()

            if readings:
                pressures = [r.pressure for r in readings]
                avg_pressure = sum(pressures) / len(pressures)

                self.db.query(models.PressureReading).filter(
                    and_(
                        models.PressureReading.pressure_point_id == point.id,
                        models.PressureReading.timestamp >= hour_ago,
                        models.PressureReading.timestamp < start_of_hour,
                        models.PressureReading.is_aggregated == False
                    )
                ).update({"is_aggregated": True})

                aggregated = models.PressureReading(
                    pressure_point_id=point.id,
                    pressure=avg_pressure,
                    timestamp=start_of_hour,
                    is_aggregated=True
                )
                self.db.add(aggregated)

        self.db.commit()


class AlarmService:
    def __init__(self, db: Session):
        self.db = db

    def create_alarm(self, data: dict) -> models.Alarm:
        alarm = models.Alarm(**data)
        self.db.add(alarm)
        self.db.commit()
        self.db.refresh(alarm)
        return alarm

    def get_alarm(self, alarm_id: int) -> Optional[models.Alarm]:
        return self.db.query(models.Alarm).filter(models.Alarm.id == alarm_id).first()

    def list_alarms(self, status: Optional[str] = None, severity: Optional[str] = None) -> List[models.Alarm]:
        query = self.db.query(models.Alarm)
        if status:
            query = query.filter(models.Alarm.status == status)
        if severity:
            query = query.filter(models.Alarm.severity == severity)
        return query.order_by(models.Alarm.created_at.desc()).all()

    def update_alarm_status(self, alarm_id: int, status: str, **kwargs) -> Optional[models.Alarm]:
        alarm = self.get_alarm(alarm_id)
        if not alarm:
            return None

        if status == models.AlarmStatus.FALSE_ALARM:
            if "false_alarm_reason" not in kwargs or not kwargs["false_alarm_reason"]:
                raise ValueError("False alarm requires a reason")
            alarm.false_alarm_reason = kwargs["false_alarm_reason"]

        alarm.status = status
        alarm.updated_at = datetime.utcnow()

        if status in [models.AlarmStatus.RESOLVED, models.AlarmStatus.FALSE_ALARM]:
            alarm.resolved_at = datetime.utcnow()

        if "handled_by" in kwargs:
            alarm.handled_by = kwargs["handled_by"]

        self.db.commit()
        self.db.refresh(alarm)
        return alarm

    def escalate_alarms(self):
        cutoff = datetime.utcnow() - timedelta(minutes=settings.ALARM_ESCALATION_MINUTES)

        alarms = self.db.query(models.Alarm).filter(
            and_(
                models.Alarm.status.in_([models.AlarmStatus.NEW]),
                models.Alarm.created_at < cutoff,
                models.Alarm.escalated == False
            )
        ).all()

        for alarm in alarms:
            alarm.status = models.AlarmStatus.URGENT
            alarm.severity = models.AlarmSeverity.URGENT
            alarm.escalated = True
            alarm.updated_at = datetime.utcnow()

        self.db.commit()
        return len(alarms)

    def send_reminders(self):
        cutoff = datetime.utcnow() - timedelta(minutes=settings.ALARM_REMINDER_MINUTES)

        alarms = self.db.query(models.Alarm).filter(
            and_(
                models.Alarm.status == models.AlarmStatus.IN_PROCESS,
                models.Alarm.updated_at < cutoff
            )
        ).all()

        for alarm in alarms:
            alarm.reminders += 1
            alarm.updated_at = datetime.utcnow()

        self.db.commit()
        return len(alarms)

    def create_leak_alarm(self, pipeline_id: int, pressure_diff_id: int, details: Dict) -> models.Alarm:
        return self.create_alarm({
            "alarm_type": "pressure_leak",
            "severity": models.AlarmSeverity.HIGH,
            "status": models.AlarmStatus.NEW,
            "pipeline_id": pipeline_id,
            "pressure_diff_id": pressure_diff_id,
            "title": "疑似泄漏报警",
            "description": f"压力差异常检测: 差值={details['pressure_diff']:.2f}, "
                          f"阈值={details['threshold']:.2f}, "
                          f"历史均值={details['mean']:.2f}, "
                          f"历史标准差={details['stddev']:.2f}, "
                          f"基于{details['data_count']}条历史数据"
        })


class PatrolService:
    def __init__(self, db: Session):
        self.db = db

    def create_daily_plan(self, plan_date: Optional[date] = None) -> List[models.PatrolPlan]:
        if plan_date is None:
            plan_date = date.today()

        pipelines = self.db.query(models.Pipeline).filter(
            models.Pipeline.status == models.PipelineStatus.ACTIVE
        ).all()

        plans = []
        for pipeline in pipelines:
            segments = self.db.query(models.Segment).filter(
                and_(
                    models.Segment.pipeline_id == pipeline.id,
                    models.Segment.status == models.SegmentStatus.OPERATIONAL
                )
            ).all()

            segments_to_patrol = []
            for segment in segments:
                last_patrol = self.db.query(models.PatrolRecord).filter(
                    models.PatrolRecord.segment_id == segment.id
                ).order_by(models.PatrolRecord.patrol_date.desc()).first()

                if not last_patrol:
                    segments_to_patrol.append(segment)
                else:
                    days_since_patrol = (plan_date - last_patrol.patrol_date).days
                    if days_since_patrol >= segment.patrol_frequency_days:
                        segments_to_patrol.append(segment)

            if segments_to_patrol:
                plan = models.PatrolPlan(
                    pipeline_id=pipeline.id,
                    plan_date=plan_date,
                    status="pending"
                )
                self.db.add(plan)
                self.db.flush()

                for idx, segment in enumerate(segments_to_patrol):
                    scheduled = datetime(plan_date.year, plan_date.month, plan_date.day, 9 + idx)
                    item = models.PatrolPlanItem(
                        plan_id=plan.id,
                        segment_id=segment.id,
                        scheduled_time=scheduled
                    )
                    self.db.add(item)

                plans.append(plan)

        self.db.commit()
        return plans

    def get_plan(self, plan_id: int) -> Optional[models.PatrolPlan]:
        return self.db.query(models.PatrolPlan).filter(models.PatrolPlan.id == plan_id).first()

    def list_plans(self, plan_date: Optional[date] = None) -> List[models.PatrolPlan]:
        query = self.db.query(models.PatrolPlan)
        if plan_date:
            query = query.filter(models.PatrolPlan.plan_date == plan_date)
        return query.order_by(models.PatrolPlan.plan_date.desc()).all()

    def add_patrol_record(self, data: dict) -> models.PatrolRecord:
        record = models.PatrolRecord(**data)
        self.db.add(record)
        self.db.commit()
        self.db.refresh(record)
        return record

    def list_records(self, segment_id: Optional[int] = None) -> List[models.PatrolRecord]:
        query = self.db.query(models.PatrolRecord)
        if segment_id:
            query = query.filter(models.PatrolRecord.segment_id == segment_id)
        return query.order_by(models.PatrolRecord.patrol_date.desc()).all()


class IntegrityService:
    def __init__(self, db: Session):
        self.db = db

    def add_integrity_record(self, data: dict) -> models.IntegrityRecord:
        record = models.IntegrityRecord(**data)
        self.db.add(record)
        self.db.commit()
        self.db.refresh(record)
        return record

    def list_records(self, pipeline_id: Optional[int] = None) -> List[models.IntegrityRecord]:
        query = self.db.query(models.IntegrityRecord)
        if pipeline_id:
            query = query.filter(models.IntegrityRecord.pipeline_id == pipeline_id)
        return query.order_by(models.IntegrityRecord.inspection_date.desc()).all()

    def check_wall_thickness(self, record: models.IntegrityRecord) -> Dict:
        pipeline = self.db.query(models.Pipeline).filter(
            models.Pipeline.id == record.pipeline_id
        ).first()

        if not pipeline:
            return {"needs_maintenance": False}

        threshold = pipeline.design_wall_thickness * settings.WALL_THICKNESS_THRESHOLD_RATIO
        needs_maintenance = record.wall_thickness < threshold

        return {
            "needs_maintenance": needs_maintenance,
            "design_thickness": pipeline.design_wall_thickness,
            "threshold": threshold,
            "current_thickness": record.wall_thickness,
            "ratio": record.wall_thickness / pipeline.design_wall_thickness
        }

    def create_maintenance_task(self, integrity_record: models.IntegrityRecord) -> Optional[models.MaintenanceTask]:
        check = self.check_wall_thickness(integrity_record)
        if not check["needs_maintenance"]:
            return None

        due_date = date.today() + timedelta(days=7)
        task = models.MaintenanceTask(
            task_type="wall_thickness",
            pipeline_id=integrity_record.pipeline_id,
            segment_id=integrity_record.segment_id,
            title="管道壁厚不足维修任务",
            description=f"当前壁厚 {check['current_thickness']:.2f}mm 低于设计壁厚 {check['design_thickness']:.2f}mm 的 80% (阈值 {check['threshold']:.2f}mm)，需要尽快维修。检测位置: {integrity_record.location or '未指定'}",
            priority="high",
            status="pending",
            due_date=due_date
        )
        self.db.add(task)
        self.db.commit()
        self.db.refresh(task)
        return task

    def list_maintenance_tasks(self, status: Optional[str] = None) -> List[models.MaintenanceTask]:
        query = self.db.query(models.MaintenanceTask)
        if status:
            query = query.filter(models.MaintenanceTask.status == status)
        return query.order_by(models.MaintenanceTask.created_at.desc()).all()

    def update_maintenance_task(self, task_id: int, data: dict) -> Optional[models.MaintenanceTask]:
        task = self.db.query(models.MaintenanceTask).filter(
            models.MaintenanceTask.id == task_id
        ).first()
        if task:
            for key, value in data.items():
                setattr(task, key, value)
            self.db.commit()
            self.db.refresh(task)
        return task


class ReportService:
    def __init__(self, db: Session):
        self.db = db

    def generate_daily_report(self, report_date: Optional[date] = None) -> models.DailyReport:
        if report_date is None:
            report_date = date.today()

        start_datetime = datetime(report_date.year, report_date.month, report_date.day)
        end_datetime = start_datetime + timedelta(days=1)

        total_pipelines = self.db.query(models.Pipeline).count()
        active_pipelines = self.db.query(models.Pipeline).filter(
            models.Pipeline.status == models.PipelineStatus.ACTIVE
        ).count()

        new_alarms = self.db.query(models.Alarm).filter(
            and_(
                models.Alarm.created_at >= start_datetime,
                models.Alarm.created_at < end_datetime
            )
        ).count()

        resolved_alarms = self.db.query(models.Alarm).filter(
            and_(
                models.Alarm.resolved_at >= start_datetime,
                models.Alarm.resolved_at < end_datetime
            )
        ).count()

        total_alarms = self.db.query(models.Alarm).count()

        pressure_readings = self.db.query(models.PressureReading).filter(
            and_(
                models.PressureReading.timestamp >= start_datetime,
                models.PressureReading.timestamp < end_datetime
            )
        ).count()

        patrols_completed = self.db.query(models.PatrolRecord).filter(
            models.PatrolRecord.patrol_date == report_date
        ).count()

        maintenance_tasks = self.db.query(models.MaintenanceTask).filter(
            and_(
                models.MaintenanceTask.created_at >= start_datetime,
                models.MaintenanceTask.created_at < end_datetime
            )
        ).count()

        content_lines = [
            f"油气管道监控系统日报 - {report_date}",
            "=" * 50,
            "",
            "一、管道概况",
            f"- 管道总数: {total_pipelines}",
            f"- 运行中管道: {active_pipelines}",
            "",
            "二、报警统计",
            f"- 历史总报警数: {total_alarms}",
            f"- 今日新增报警: {new_alarms}",
            f"- 今日处理报警: {resolved_alarms}",
            "",
            "三、监测数据",
            f"- 今日压力读数: {pressure_readings}",
            "",
            "四、巡检与维护",
            f"- 今日完成巡检: {patrols_completed}",
            f"- 今日生成维修任务: {maintenance_tasks}",
        ]

        report = self.db.query(models.DailyReport).filter(
            models.DailyReport.report_date == report_date
        ).first()

        if report:
            report.total_pipelines = total_pipelines
            report.active_pipelines = active_pipelines
            report.total_alarms = total_alarms
            report.new_alarms = new_alarms
            report.resolved_alarms = resolved_alarms
            report.pressure_readings_count = pressure_readings
            report.patrols_completed = patrols_completed
            report.maintenance_tasks = maintenance_tasks
            report.content = "\n".join(content_lines)
        else:
            report = models.DailyReport(
                report_date=report_date,
                total_pipelines=total_pipelines,
                active_pipelines=active_pipelines,
                total_alarms=total_alarms,
                new_alarms=new_alarms,
                resolved_alarms=resolved_alarms,
                pressure_readings_count=pressure_readings,
                patrols_completed=patrols_completed,
                maintenance_tasks=maintenance_tasks,
                content="\n".join(content_lines)
            )
            self.db.add(report)

        self.db.commit()
        self.db.refresh(report)
        return report

    def get_report(self, report_date: date) -> Optional[models.DailyReport]:
        return self.db.query(models.DailyReport).filter(
            models.DailyReport.report_date == report_date
        ).first()

    def list_reports(self) -> List[models.DailyReport]:
        return self.db.query(models.DailyReport).order_by(
            models.DailyReport.report_date.desc()
        ).all()
