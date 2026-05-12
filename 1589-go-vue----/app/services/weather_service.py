from datetime import datetime, timedelta
from typing import Optional, Tuple
from sqlalchemy.orm import Session
from app.models import RopewayStatus, WeatherData, StatusRecord
from app.config import (
    WIND_SPEED_DECELERATE,
    WIND_SPEED_PAUSE,
    WIND_SPEED_STOP,
    LAG_PROTECTION_MINUTES
)


_lag_counters = {}
_pending_confirmation = False


class WeatherMonitor:
    def __init__(self, db: Session):
        self.db = db

    @property
    def current_status(self) -> RopewayStatus:
        latest = self.db.query(StatusRecord).order_by(
            StatusRecord.timestamp.desc()
        ).first()
        if latest:
            return latest.to_status
        return RopewayStatus.NORMAL

    @property
    def pending_confirmation(self) -> bool:
        global _pending_confirmation
        return _pending_confirmation

    @pending_confirmation.setter
    def pending_confirmation(self, value: bool):
        global _pending_confirmation
        _pending_confirmation = value

    def _get_status_order(self) -> list:
        return [
            RopewayStatus.NORMAL,
            RopewayStatus.DECELERATE,
            RopewayStatus.PAUSE,
            RopewayStatus.STOP
        ]

    def _get_status_index(self, status: RopewayStatus) -> int:
        return self._get_status_order().index(status)

    def get_threshold_for_status(self, status: RopewayStatus) -> float:
        if status == RopewayStatus.DECELERATE:
            return WIND_SPEED_DECELERATE
        if status == RopewayStatus.PAUSE:
            return WIND_SPEED_PAUSE
        if status == RopewayStatus.STOP:
            return WIND_SPEED_STOP
        return 0.0

    def get_next_downgrade_status(self, current: RopewayStatus, wind_speed: float) -> Optional[RopewayStatus]:
        status_order = self._get_status_order()
        current_idx = self._get_status_index(current)
        
        if current_idx >= len(status_order) - 1:
            return None
        
        next_status = status_order[current_idx + 1]
        threshold = self.get_threshold_for_status(next_status)
        
        if wind_speed >= threshold:
            return next_status
        
        return None

    def _get_lag_counter(self) -> int:
        return _lag_counters.get('global', 0)
    
    def _set_lag_counter(self, value: int):
        _lag_counters['global'] = value

    def _is_wind_safe_for_transition(self, wind_speed: float, next_status: RopewayStatus) -> bool:
        threshold = self.get_threshold_for_status(next_status)
        return wind_speed < threshold

    def check_lag_protection(self, wind_speed: float, next_status: RopewayStatus) -> bool:
        current = self.current_status
        if current == next_status:
            return True
        
        if self._is_wind_safe_for_transition(wind_speed, next_status):
            self._set_lag_counter(0)
            return False
        
        new_count = self._get_lag_counter() + 1
        self._set_lag_counter(new_count)
        return new_count >= LAG_PROTECTION_MINUTES

    def update_weather(self, wind_speed: float, has_lightning: bool, 
                       temperature: Optional[float] = None,
                       humidity: Optional[float] = None) -> Tuple[WeatherData, Optional[StatusRecord]]:
        current = self.current_status
        
        weather_data = WeatherData(
            wind_speed=wind_speed,
            has_lightning=has_lightning,
            temperature=temperature,
            humidity=humidity
        )
        self.db.add(weather_data)
        self.db.flush()
        
        status_record = None
        
        if has_lightning:
            if current != RopewayStatus.STOP:
                status_record = self._record_transition(
                    current, RopewayStatus.STOP,
                    "检测到雷电，立即停止",
                    weather_data.id
                )
                self._set_lag_counter(0)
        else:
            if wind_speed < WIND_SPEED_DECELERATE:
                if current != RopewayStatus.NORMAL:
                    self.pending_confirmation = True
            else:
                next_status = self.get_next_downgrade_status(current, wind_speed)
                
                if next_status and self.can_transition(current, next_status):
                    if self.check_lag_protection(wind_speed, next_status):
                        status_record = self._record_transition(
                            current, next_status,
                            self._get_reason_for_transition(current, next_status, wind_speed),
                            weather_data.id
                        )
                        self._set_lag_counter(0)
        
        self.db.commit()
        self.db.refresh(weather_data)
        if status_record:
            self.db.refresh(status_record)
        
        return weather_data, status_record

    def can_transition(self, from_status: RopewayStatus, to_status: RopewayStatus) -> bool:
        status_order = self._get_status_order()
        from_idx = self._get_status_index(from_status)
        to_idx = self._get_status_index(to_status)
        
        if to_idx > from_idx:
            return to_idx == from_idx + 1
        
        if to_idx < from_idx:
            return to_idx == 0
        
        return False

    def _get_reason_for_transition(self, from_status: RopewayStatus, to_status: RopewayStatus, wind_speed: float) -> str:
        threshold = self.get_threshold_for_status(to_status)
        
        if to_status == RopewayStatus.DECELERATE:
            return f"风速 {wind_speed} m/s 超过 {threshold} m/s，减速运行"
        if to_status == RopewayStatus.PAUSE:
            return f"风速 {wind_speed} m/s 超过 {threshold} m/s，暂停运行"
        if to_status == RopewayStatus.STOP:
            return f"风速 {wind_speed} m/s 超过 {threshold} m/s，停止运行"
        return "状态变更"

    def _record_transition(self, from_status: RopewayStatus, to_status: RopewayStatus,
                           reason: str, weather_data_id: Optional[int] = None,
                           operator_id: Optional[int] = None) -> StatusRecord:
        record = StatusRecord(
            from_status=from_status,
            to_status=to_status,
            reason=reason,
            weather_data_id=weather_data_id,
            operator_id=operator_id
        )
        self.db.add(record)
        self.db.flush()
        return record

    def confirm_recovery(self, operator_id: int, notes: Optional[str] = None) -> Optional[StatusRecord]:
        current = self.current_status
        if current == RopewayStatus.NORMAL:
            return None
        
        latest_weather = self.db.query(WeatherData).order_by(
            WeatherData.timestamp.desc()
        ).first()
        
        if latest_weather and latest_weather.wind_speed >= WIND_SPEED_DECELERATE:
            return None
        
        if latest_weather and latest_weather.has_lightning:
            return None
        
        reason = "操作员确认天气恢复"
        if notes:
            reason += f"：{notes}"
        
        record = self._record_transition(
            current, RopewayStatus.NORMAL, reason,
            operator_id=operator_id
        )
        self.db.commit()
        self.db.refresh(record)
        self.pending_confirmation = False
        self._set_lag_counter(0)
        return record

    def get_recent_weather(self, minutes: int = 30) -> list:
        cutoff = datetime.utcnow() - timedelta(minutes=minutes)
        return self.db.query(WeatherData).filter(
            WeatherData.timestamp >= cutoff
        ).order_by(WeatherData.timestamp.desc()).all()

    def get_status_history(self, limit: int = 20) -> list:
        return self.db.query(StatusRecord).order_by(
            StatusRecord.timestamp.desc()
        ).limit(limit).all()
