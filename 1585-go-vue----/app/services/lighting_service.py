from sqlalchemy.orm import Session

from app.models import LightingDevice, LightingGroup, DeviceStatus, ControlMode


class LightingService:
    MAX_BRIGHTNESS_STEP = 10
    MIN_BRIGHTNESS = 0
    MAX_BRIGHTNESS = 100

    @staticmethod
    def get_all_devices(db: Session):
        return db.query(LightingDevice).all()

    @staticmethod
    def get_device(db: Session, device_id: int):
        return db.query(LightingDevice).filter(LightingDevice.id == device_id).first()

    @staticmethod
    def create_device(db: Session, device_data: dict):
        device = LightingDevice(**device_data)
        db.add(device)
        db.commit()
        db.refresh(device)
        return device

    @staticmethod
    def _check_device_available(device: LightingDevice):
        if device.fault:
            raise ValueError(f"设备 {device.name} 处于故障状态，无法操作")
        return True

    @staticmethod
    def _calculate_step_brightness(current: int, target: int) -> int:
        diff = target - current
        if abs(diff) <= LightingService.MAX_BRIGHTNESS_STEP:
            return target
        step = LightingService.MAX_BRIGHTNESS_STEP if diff > 0 else -LightingService.MAX_BRIGHTNESS_STEP
        return current + step

    @staticmethod
    def _clamp_brightness(value: int) -> int:
        return max(LightingService.MIN_BRIGHTNESS, min(LightingService.MAX_BRIGHTNESS, value))

    @staticmethod
    def set_brightness(db: Session, device_id: int, brightness: int):
        device = LightingService.get_device(db, device_id)
        if not device:
            return None, "设备不存在"
        
        try:
            LightingService._check_device_available(device)
        except ValueError as e:
            return None, str(e)
        
        target = LightingService._clamp_brightness(brightness)
        device.target_brightness = target
        new_brightness = LightingService._calculate_step_brightness(device.brightness, target)
        device.brightness = new_brightness
        
        if new_brightness > 0:
            device.status = DeviceStatus.RUNNING.value
        else:
            device.status = DeviceStatus.STOPPED.value
        
        db.commit()
        db.refresh(device)
        return device, f"亮度已调整至 {new_brightness}%"

    @staticmethod
    def set_group(db: Session, device_id: int, group: LightingGroup):
        device = LightingService.get_device(db, device_id)
        if not device:
            return None, "设备不存在"
        
        device.group = group.value
        db.commit()
        db.refresh(device)
        return device, f"已切换到{group.value}组"

    @staticmethod
    def update_auto_brightness(db: Session, entrance_factor: float = 1.3, middle_factor: float = 1.0):
        devices = LightingService.get_all_devices(db)
        for device in devices:
            if device.mode != ControlMode.AUTO.value or device.fault:
                continue
            
            base_brightness = 50
            if device.group == LightingGroup.ENTRANCE.value:
                target = int(base_brightness * entrance_factor)
            elif device.group == LightingGroup.EXIT.value:
                target = int(base_brightness * entrance_factor)
            else:
                target = int(base_brightness * middle_factor)
            
            target = LightingService._clamp_brightness(target)
            device.target_brightness = target
            device.brightness = LightingService._calculate_step_brightness(device.brightness, target)
            
            if device.brightness > 0:
                device.status = DeviceStatus.RUNNING.value
            else:
                device.status = DeviceStatus.STOPPED.value
        
        db.commit()
