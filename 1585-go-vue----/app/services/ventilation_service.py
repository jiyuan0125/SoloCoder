from datetime import datetime
from sqlalchemy.orm import Session

from app.models import VentilationDevice, Alarm, AlarmLevel, AlarmStatus, DeviceStatus, ControlMode


class VentilationService:
    SHUTDOWN_DELAY_MINUTES = 5

    @staticmethod
    def get_all_devices(db: Session):
        return db.query(VentilationDevice).all()

    @staticmethod
    def get_device(db: Session, device_id: int):
        return db.query(VentilationDevice).filter(VentilationDevice.id == device_id).first()

    @staticmethod
    def create_device(db: Session, device_data: dict):
        device = VentilationDevice(**device_data)
        db.add(device)
        db.commit()
        db.refresh(device)
        return device

    @staticmethod
    def _check_device_available(device: VentilationDevice):
        if device.fault:
            raise ValueError(f"设备 {device.name} 处于故障状态，无法操作")
        return True

    @staticmethod
    def start_device(db: Session, device_id: int):
        device = VentilationService.get_device(db, device_id)
        if not device:
            return None, "设备不存在"
        
        try:
            VentilationService._check_device_available(device)
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
        device = VentilationService.get_device(db, device_id)
        if not device:
            return None, "设备不存在"
        
        try:
            VentilationService._check_device_available(device)
        except ValueError as e:
            return None, str(e)
        
        device.is_running = False
        device.status = DeviceStatus.STOPPED.value
        device.start_time = None
        device.shutdown_delay_active = False
        device.shutdown_delay_start = None
        db.commit()
        db.refresh(device)
        return device, "设备已停止"

    @staticmethod
    def set_mode(db: Session, device_id: int, mode: ControlMode):
        device = VentilationService.get_device(db, device_id)
        if not device:
            return None, "设备不存在"
        
        device.mode = mode.value
        db.commit()
        db.refresh(device)
        return device, f"已切换到{mode.value}模式"

    @staticmethod
    def update_sensor_data(db: Session, device_id: int, co_level: float = None, visibility: float = None):
        device = VentilationService.get_device(db, device_id)
        if not device:
            return None
        
        if co_level is not None:
            device.co_level = co_level
        if visibility is not None:
            device.visibility = visibility
        
        db.commit()
        db.refresh(device)
        return device

    @staticmethod
    def process_auto_control(db: Session, device_id: int):
        device = VentilationService.get_device(db, device_id)
        if not device:
            return
        
        if device.mode != ControlMode.AUTO.value or device.fault:
            return
        
        co_exceeds = device.co_level > device.co_threshold
        visibility_below = device.visibility < device.visibility_threshold
        needs_running = co_exceeds or visibility_below
        all_safe = device.co_level <= device.co_threshold and device.visibility >= device.visibility_threshold
        
        if needs_running and not device.is_running:
            device.is_running = True
            device.status = DeviceStatus.RUNNING.value
            device.start_time = datetime.utcnow()
            device.shutdown_delay_active = False
            device.shutdown_delay_start = None
            
            alarm = Alarm(
                subsystem="ventilation",
                device_id=device.id,
                device_name=device.name,
                level=AlarmLevel.MEDIUM.value,
                message=f"通风设备 {device.name} 已自动启动：CO={device.co_level}ppm，能见度={device.visibility}m"
            )
            db.add(alarm)
        
        elif device.is_running and all_safe:
            if not device.shutdown_delay_active:
                device.shutdown_delay_active = True
                device.shutdown_delay_start = datetime.utcnow()
            else:
                elapsed = (datetime.utcnow() - device.shutdown_delay_start).total_seconds() / 60
                if elapsed >= VentilationService.SHUTDOWN_DELAY_MINUTES:
                    device.is_running = False
                    device.status = DeviceStatus.STOPPED.value
                    device.start_time = None
                    device.shutdown_delay_active = False
                    device.shutdown_delay_start = None
        elif device.is_running and needs_running and device.shutdown_delay_active:
            device.shutdown_delay_active = False
            device.shutdown_delay_start = None
        
        db.commit()
