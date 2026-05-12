from datetime import datetime
from sqlalchemy.orm import Session

from app.models import (
    DrainageDevice, Alarm, AlarmLevel, AlarmStatus, MaintenanceTask,
    DeviceStatus, ControlMode, MaintenanceTaskStatus
)


MAINTENANCE_CONTENT_MAP = {
    "default": """
    常规保养内容：
    1. 检查泵体密封性
    2. 清理进水口和出水口
    3. 检查电气连接
    4. 测试自动控制功能
    5. 记录运行参数
    """,
    "model_a": """
    Model A 专用保养内容：
    1. 更换机械密封件
    2. 检查叶轮磨损情况
    3. 润滑轴承
    4. 校准压力传感器
    5. 进行全功能测试
    """,
    "model_b": """
    Model B 专用保养内容：
    1. 清理过滤系统
    2. 检查浮球开关
    3. 测试过载保护
    4. 验证水位传感器校准
    5. 检查电缆绝缘
    """
}


class DrainageService:
    SHUTDOWN_DELAY_MINUTES = 3
    CONTINUOUS_RUNNING_ALARM_HOURS = 1

    @staticmethod
    def get_all_devices(db: Session):
        return db.query(DrainageDevice).all()

    @staticmethod
    def get_device(db: Session, device_id: int):
        return db.query(DrainageDevice).filter(DrainageDevice.id == device_id).first()

    @staticmethod
    def get_pending_maintenance_tasks(db: Session, device_id: int = None):
        query = db.query(MaintenanceTask).filter(
            MaintenanceTask.status == MaintenanceTaskStatus.PENDING.value
        )
        if device_id:
            query = query.filter(MaintenanceTask.drainage_device_id == device_id)
        return query.all()

    @staticmethod
    def create_device(db: Session, device_data: dict):
        device = DrainageDevice(**device_data)
        db.add(device)
        db.commit()
        db.refresh(device)
        return device

    @staticmethod
    def _check_device_available(device: DrainageDevice):
        if device.fault:
            raise ValueError(f"设备 {device.name} 处于故障状态，无法操作")
        return True

    @staticmethod
    def _get_maintenance_content(model: str) -> str:
        normalized_model = model.lower().replace("-", "_")
        return MAINTENANCE_CONTENT_MAP.get(normalized_model, MAINTENANCE_CONTENT_MAP["default"])

    @staticmethod
    def _check_and_create_maintenance_task(db: Session, device: DrainageDevice):
        runtime_since_last = device.cumulative_runtime - device.last_maintenance_runtime
        if runtime_since_last >= device.maintenance_interval_hours:
            existing_tasks = db.query(MaintenanceTask).filter(
                MaintenanceTask.drainage_device_id == device.id,
                MaintenanceTask.status == MaintenanceTaskStatus.PENDING.value
            ).first()
            
            if not existing_tasks:
                content = DrainageService._get_maintenance_content(device.model)
                task = MaintenanceTask(
                    drainage_device_id=device.id,
                    device_name=device.name,
                    device_model=device.model,
                    content=content
                )
                db.add(task)
                
                alarm = Alarm(
                    subsystem="drainage",
                    device_id=device.id,
                    device_name=device.name,
                    level=AlarmLevel.MEDIUM.value,
                    message=f"排水泵 {device.name} 累计运行时间已达到保养周期，请及时保养"
                )
                db.add(alarm)
                
                device.last_maintenance_runtime = device.cumulative_runtime
                return True
        return False

    @staticmethod
    def start_device(db: Session, device_id: int):
        device = DrainageService.get_device(db, device_id)
        if not device:
            return None, "设备不存在"
        
        try:
            DrainageService._check_device_available(device)
        except ValueError as e:
            return None, str(e)
        
        device.is_running = True
        device.status = DeviceStatus.RUNNING.value
        device.start_time = datetime.utcnow()
        device.shutdown_delay_active = False
        device.shutdown_delay_start = None
        db.commit()
        db.refresh(device)
        return device, "设备已启动"

    @staticmethod
    def stop_device(db: Session, device_id: int):
        device = DrainageService.get_device(db, device_id)
        if not device:
            return None, "设备不存在"
        
        try:
            DrainageService._check_device_available(device)
        except ValueError as e:
            return None, str(e)
        
        if device.is_running and device.start_time:
            runtime_hours = (datetime.utcnow() - device.start_time).total_seconds() / 3600.0
            device.cumulative_runtime += runtime_hours
        
        device.is_running = False
        device.status = DeviceStatus.STOPPED.value
        device.start_time = None
        device.shutdown_delay_active = False
        device.shutdown_delay_start = None
        
        DrainageService._check_and_create_maintenance_task(db, device)
        
        db.commit()
        db.refresh(device)
        return device, "设备已停止"

    @staticmethod
    def update_water_level(db: Session, device_id: int, level: float):
        device = DrainageService.get_device(db, device_id)
        if not device:
            return None
        
        device.water_level = level
        db.commit()
        db.refresh(device)
        return device

    @staticmethod
    def process_auto_control(db: Session, device_id: int):
        device = DrainageService.get_device(db, device_id)
        if not device:
            return
        
        if device.mode != ControlMode.AUTO.value or device.fault:
            return
        
        if device.is_running and device.start_time:
            runtime_hours = (datetime.utcnow() - device.start_time).total_seconds() / 3600.0
            if runtime_hours >= DrainageService.CONTINUOUS_RUNNING_ALARM_HOURS:
                existing_alarm = db.query(Alarm).filter(
                    Alarm.subsystem == "drainage",
                    Alarm.device_id == device.id,
                    Alarm.status == AlarmStatus.ACTIVE.value,
                    Alarm.message.like(f"%{device.name}%连续运行%")
                ).first()
                
                if not existing_alarm:
                    alarm = Alarm(
                        subsystem="drainage",
                        device_id=device.id,
                        device_name=device.name,
                        level=AlarmLevel.HIGH.value,
                        message=f"排水泵 {device.name} 连续运行已超过1小时，请检查"
                    )
                    db.add(alarm)
        
        if device.water_level >= device.warning_level and not device.is_running:
            device.is_running = True
            device.status = DeviceStatus.RUNNING.value
            device.start_time = datetime.utcnow()
            device.shutdown_delay_active = False
            device.shutdown_delay_start = None
            
            alarm = Alarm(
                subsystem="drainage",
                device_id=device.id,
                device_name=device.name,
                level=AlarmLevel.MEDIUM.value,
                message=f"排水泵 {device.name} 已自动启动：水位={device.water_level}m"
            )
            db.add(alarm)
        
        elif device.is_running and device.water_level <= device.safe_level:
            if not device.shutdown_delay_active:
                device.shutdown_delay_active = True
                device.shutdown_delay_start = datetime.utcnow()
            else:
                elapsed = (datetime.utcnow() - device.shutdown_delay_start).total_seconds() / 60
                if elapsed >= DrainageService.SHUTDOWN_DELAY_MINUTES:
                    if device.start_time:
                        runtime_hours = (datetime.utcnow() - device.start_time).total_seconds() / 3600.0
                        device.cumulative_runtime += runtime_hours
                    
                    device.is_running = False
                    device.status = DeviceStatus.STOPPED.value
                    device.start_time = None
                    device.shutdown_delay_active = False
                    device.shutdown_delay_start = None
                    
                    DrainageService._check_and_create_maintenance_task(db, device)
        
        elif device.is_running and device.water_level > device.safe_level and device.shutdown_delay_active:
            device.shutdown_delay_active = False
            device.shutdown_delay_start = None
        
        db.commit()

    @staticmethod
    def complete_maintenance_task(db: Session, task_id: int):
        task = db.query(MaintenanceTask).filter(MaintenanceTask.id == task_id).first()
        if not task:
            return None, "任务不存在"
        
        task.status = MaintenanceTaskStatus.COMPLETED.value
        task.completed_at = datetime.utcnow()
        
        alarms = db.query(Alarm).filter(
            Alarm.subsystem == "drainage",
            Alarm.device_id == task.drainage_device_id,
            Alarm.status == AlarmStatus.ACTIVE.value,
            Alarm.message.like(f"%{task.device_name}%保养%")
        ).all()
        
        for alarm in alarms:
            alarm.status = AlarmStatus.RESOLVED.value
            alarm.resolved_at = datetime.utcnow()
        
        db.commit()
        db.refresh(task)
        return task, "保养任务已完成"
