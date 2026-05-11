from typing import Tuple
from src.core.models.schemas import (
    Reactor, SensorReading, AlarmLevel, AlarmRecord
)


WARNING_THRESHOLD_RATIO = 0.9
CRITICAL_THRESHOLD_RATIO = 1.0


class AlarmService:
    def determine_alarm_level(
        self, reactor: Reactor, temperature: float, pressure: float
    ) -> Tuple[AlarmLevel, bool]:
        temp_warning = reactor.design_temperature_max * WARNING_THRESHOLD_RATIO
        temp_critical = reactor.design_temperature_max * CRITICAL_THRESHOLD_RATIO
        press_warning = reactor.design_pressure_max * WARNING_THRESHOLD_RATIO
        press_critical = reactor.design_pressure_max * CRITICAL_THRESHOLD_RATIO

        temp_alarm = AlarmLevel.NORMAL
        if temperature >= temp_critical:
            temp_alarm = AlarmLevel.CRITICAL
        elif temperature >= temp_warning:
            temp_alarm = AlarmLevel.WARNING

        press_alarm = AlarmLevel.NORMAL
        if pressure >= press_critical:
            press_alarm = AlarmLevel.CRITICAL
        elif pressure >= press_warning:
            press_alarm = AlarmLevel.WARNING

        temp_is_warning_or_higher = temp_alarm in (AlarmLevel.WARNING, AlarmLevel.CRITICAL)
        press_is_warning_or_higher = press_alarm in (AlarmLevel.WARNING, AlarmLevel.CRITICAL)
        
        is_high_risk = temp_is_warning_or_higher and press_is_warning_or_higher

        if is_high_risk:
            if temp_alarm == AlarmLevel.CRITICAL or press_alarm == AlarmLevel.CRITICAL:
                overall_level = AlarmLevel.CRITICAL
            else:
                overall_level = AlarmLevel.HIGH_RISK
        else:
            if temp_alarm == AlarmLevel.CRITICAL or press_alarm == AlarmLevel.CRITICAL:
                overall_level = AlarmLevel.CRITICAL
            elif temp_alarm == AlarmLevel.WARNING or press_alarm == AlarmLevel.WARNING:
                overall_level = AlarmLevel.WARNING
            else:
                overall_level = AlarmLevel.NORMAL

        return overall_level, is_high_risk

    def create_alarm_record(
        self, reading: SensorReading, reactor: Reactor
    ) -> AlarmRecord:
        alarm_messages = []
        
        temp_warning = reactor.design_temperature_max * WARNING_THRESHOLD_RATIO
        temp_critical = reactor.design_temperature_max * CRITICAL_THRESHOLD_RATIO
        press_warning = reactor.design_pressure_max * WARNING_THRESHOLD_RATIO
        press_critical = reactor.design_pressure_max * CRITICAL_THRESHOLD_RATIO

        if reading.temperature >= temp_critical:
            alarm_messages.append(
                f"温度紧急报警: {reading.temperature}°C (上限: {reactor.design_temperature_max}°C)"
            )
        elif reading.temperature >= temp_warning:
            alarm_messages.append(
                f"温度预警: {reading.temperature}°C (预警值: {temp_warning}°C)"
            )

        if reading.pressure >= press_critical:
            alarm_messages.append(
                f"压力紧急报警: {reading.pressure}MPa (上限: {reactor.design_pressure_max}MPa)"
            )
        elif reading.pressure >= press_warning:
            alarm_messages.append(
                f"压力预警: {reading.pressure}MPa (预警值: {press_warning}MPa)"
            )

        if reading.is_high_risk:
            alarm_messages.append("【高危】温度和压力同时超标！")

        return AlarmRecord(
            reactor_id=reading.reactor_id,
            alarm_level=reading.alarm_level,
            is_high_risk=reading.is_high_risk,
            message="; ".join(alarm_messages),
            sensor_reading_id=reading.id,
            timestamp=reading.timestamp
        )


_alarm_service_instance = AlarmService()


def get_alarm_service() -> AlarmService:
    return _alarm_service_instance
