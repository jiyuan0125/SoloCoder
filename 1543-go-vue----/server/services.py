from sqlalchemy.orm import Session
from sqlalchemy import and_
from datetime import datetime, timedelta
from typing import Optional, List, Dict
from server.models import (
    Watershed, Reservoir, MonitoringStation, HydrologicalData,
    Alert, DispatchRecommendation, ReservoirOperation, FloodPhase, FloodEvent,
    AlertLevel, DispatchStatus, FloodPhaseType
)
from server.database import SessionLocal

RAIN_WARNING_THRESHOLD = 50.0
HEAVY_RAIN_WARNING_THRESHOLD = 100.0


def calculate_6h_rainfall(db: Session, station_id: int, before_time: Optional[datetime] = None) -> float:
    if before_time is None:
        before_time = datetime.utcnow()
    start_time = before_time - timedelta(hours=6)
    records = db.query(HydrologicalData).filter(
        and_(
            HydrologicalData.station_id == station_id,
            HydrologicalData.recorded_at >= start_time,
            HydrologicalData.recorded_at <= before_time
        )
    ).all()
    total = sum(r.value for r in records)
    return round(total, 2)


def get_or_create_active_flood_event(db: Session, watershed_id: int, alert: Alert) -> Optional[FloodEvent]:
    active_event = db.query(FloodEvent).filter(
        and_(
            FloodEvent.watershed_id == watershed_id,
            FloodEvent.is_active == True
        )
    ).first()
    if active_event:
        return active_event
    watershed = db.query(Watershed).filter(Watershed.id == watershed_id).first()
    watershed_name = watershed.name if watershed else "Unknown"
    event = FloodEvent(
        watershed_id=watershed_id,
        title=f"{watershed_name}洪水事件 - {datetime.utcnow().strftime('%Y-%m-%d %H:%M')}",
        description=f"由{alert.alert_message}触发",
        trigger_alert_id=alert.id,
        is_active=True
    )
    db.add(event)
    db.flush()
    phase = FloodPhase(
        flood_event_id=event.id,
        phase_type=FloodPhaseType.WARNING.value,
        description="预警触发，洪水事件开始"
    )
    db.add(phase)
    db.flush()
    return event


def check_and_create_alerts(db: Session, hydrological_data: HydrologicalData) -> List[Alert]:
    station = db.query(MonitoringStation).filter(
        MonitoringStation.id == hydrological_data.station_id
    ).first()
    if not station:
        return []
    alerts = []
    if station.station_type == "rain":
        total_6h = hydrological_data.rainfall_6h or 0
        if total_6h >= HEAVY_RAIN_WARNING_THRESHOLD:
            alert = create_alert(db, station, AlertLevel.HEAVY_RAIN_WARNING.value,
                               f"大暴雨预警：{station.name}站6小时降雨量达{total_6h}mm",
                               total_6h, HEAVY_RAIN_WARNING_THRESHOLD)
            alerts.append(alert)
        elif total_6h >= RAIN_WARNING_THRESHOLD:
            alert = create_alert(db, station, AlertLevel.RAIN_WARNING.value,
                               f"暴雨预警：{station.name}站6小时降雨量达{total_6h}mm",
                               total_6h, RAIN_WARNING_THRESHOLD)
            alerts.append(alert)
    elif station.station_type == "water":
        threshold = station.warning_threshold
        if threshold and hydrological_data.value >= threshold:
            alert = create_alert(db, station, AlertLevel.FLOOD_WARNING.value,
                               f"洪水预警：{station.name}站水位{hydrological_data.value}m超过警戒水位{threshold}m",
                               hydrological_data.value, threshold)
            alerts.append(alert)
    return alerts


def create_alert(db: Session, station: MonitoringStation, alert_level: str,
                 message: str, trigger_value: float, threshold: Optional[float]) -> Alert:
    existing = db.query(Alert).filter(
        and_(
            Alert.station_id == station.id,
            Alert.alert_level == alert_level,
            Alert.is_active == True
        )
    ).first()
    if existing:
        return existing
    alert = Alert(
        station_id=station.id,
        alert_level=alert_level,
        alert_message=message,
        trigger_value=trigger_value,
        threshold=threshold,
        is_active=True
    )
    db.add(alert)
    db.flush()
    event = get_or_create_active_flood_event(db, station.watershed_id, alert)
    if event:
        alert.flood_event_id = event.id
        associate_active_data_to_event(db, event)
    db.flush()
    return alert


def associate_active_data_to_event(db: Session, event: FloodEvent):
    six_hours_ago = datetime.utcnow() - timedelta(hours=6)
    stations = db.query(MonitoringStation).filter(
        MonitoringStation.watershed_id == event.watershed_id
    ).all()
    station_ids = [s.id for s in stations]
    records = db.query(HydrologicalData).filter(
        and_(
            HydrologicalData.station_id.in_(station_ids),
            HydrologicalData.recorded_at >= six_hours_ago,
            HydrologicalData.flood_event_id == None
        )
    ).all()
    for record in records:
        record.flood_event_id = event.id
    db.flush()


def get_reservoir_current_capacity(reservoir: Reservoir) -> Optional[float]:
    if not reservoir.capacity_curve:
        return None
    curve = reservoir.capacity_curve
    level = reservoir.current_level
    levels = sorted([float(k) for k in curve.keys()])
    if not levels:
        return None
    if level <= levels[0]:
        return curve[str(levels[0])]
    if level >= levels[-1]:
        return curve[str(levels[-1])]
    for i in range(len(levels) - 1):
        l1, l2 = levels[i], levels[i + 1]
        if l1 <= level <= l2:
            v1 = curve[str(l1)]
            v2 = curve[str(l2)]
            ratio = (level - l1) / (l2 - l1)
            return round(v1 + (v2 - v1) * ratio, 2)
    return curve[str(levels[-1])]


def calculate_discharge(db: Session, reservoir: Reservoir) -> Optional[Dict]:
    if reservoir.flood_limit_level is None or reservoir.current_level <= reservoir.flood_limit_level:
        return None
    current_capacity = get_reservoir_current_capacity(reservoir)
    downstream_safe = reservoir.downstream_safe_discharge
    capacity_ratio = 0.0
    if reservoir.capacity and current_capacity:
        capacity_ratio = current_capacity / reservoir.capacity
    if downstream_safe is None:
        downstream_safe = 100.0
    excess_ratio = 0.0
    if reservoir.warning_level and reservoir.flood_limit_level:
        excess = reservoir.current_level - reservoir.flood_limit_level
        range_val = max(reservoir.warning_level - reservoir.flood_limit_level, 1.0)
        excess_ratio = min(excess / range_val, 1.5)
    base_discharge = reservoir.current_discharge
    recommended = base_discharge + downstream_safe * excess_ratio * max(0.5, capacity_ratio)
    recommended = min(recommended, downstream_safe * 1.5)
    recommended = round(max(recommended, 0), 2)
    rationale_parts = []
    if current_capacity:
        rationale_parts.append(f"当前库容约{current_capacity}立方米")
    if downstream_safe:
        rationale_parts.append(f"下游河道安全泄量{downstream_safe}立方米/秒")
    rationale_parts.append(f"当前水位{reservoir.current_level}m，超过汛限水位{reservoir.flood_limit_level}m")
    rationale = "；".join(rationale_parts) + f"。建议泄洪流量调整为{recommended}立方米/秒。"
    return {
        "recommended_discharge": recommended,
        "rationale": rationale,
        "current_capacity": current_capacity,
        "downstream_safe": downstream_safe
    }


def generate_dispatch_recommendations(db: Session, flood_event: FloodEvent) -> List[DispatchRecommendation]:
    reservoir_ids = []
    if flood_event.watershed_id:
        reservoirs = db.query(Reservoir).filter(
            Reservoir.watershed_id == flood_event.watershed_id
        ).all()
        reservoir_ids = [r.id for r in reservoirs]
    else:
        active_reservoirs = db.query(Reservoir).filter(
            Reservoir.is_discharging == True
        ).all()
        reservoir_ids = [r.id for r in active_reservoirs]
    recommendations = []
    for rid in reservoir_ids:
        reservoir = db.query(Reservoir).filter(Reservoir.id == rid).first()
        if not reservoir:
            continue
        calc = calculate_discharge(db, reservoir)
        if calc is None:
            continue
        existing = db.query(DispatchRecommendation).filter(
            and_(
                DispatchRecommendation.flood_event_id == flood_event.id,
                DispatchRecommendation.reservoir_id == rid,
                DispatchRecommendation.status == DispatchStatus.PENDING.value
            )
        ).first()
        if existing:
            continue
        rec = DispatchRecommendation(
            flood_event_id=flood_event.id,
            reservoir_id=rid,
            recommended_discharge=calc["recommended_discharge"],
            rationale=calc["rationale"],
            current_capacity=calc["current_capacity"],
            downstream_safe=calc["downstream_safe"],
            status=DispatchStatus.PENDING.value
        )
        db.add(rec)
        recommendations.append(rec)
    db.flush()
    return recommendations


def execute_dispatch(db: Session, dispatch: DispatchRecommendation, operator: str) -> ReservoirOperation:
    reservoir = db.query(Reservoir).filter(Reservoir.id == dispatch.reservoir_id).first()
    if not reservoir:
        raise ValueError("Reservoir not found")
    operation = ReservoirOperation(
        reservoir_id=reservoir.id,
        flood_event_id=dispatch.flood_event_id,
        dispatch_id=dispatch.id,
        previous_discharge=reservoir.current_discharge,
        new_discharge=dispatch.recommended_discharge,
        previous_level=reservoir.current_level,
        reason=dispatch.rationale,
        operated_by=operator
    )
    db.add(operation)
    reservoir.current_discharge = dispatch.recommended_discharge
    if dispatch.recommended_discharge > 0:
        reservoir.is_discharging = True
    dispatch.status = DispatchStatus.EXECUTED.value
    dispatch.executed_at = datetime.utcnow()
    db.flush()
    return operation


def evaluate_flood_situation(db: Session) -> List[Dict]:
    active_events = db.query(FloodEvent).filter(FloodEvent.is_active == True).all()
    results = []
    for event in active_events:
        recs = generate_dispatch_recommendations(db, event)
        results.append({
            "event_id": event.id,
            "title": event.title,
            "new_recommendations": len(recs)
        })
    db.commit()
    return results


def resolve_alerts_for_event(db: Session, event: FloodEvent):
    now = datetime.utcnow()
    for alert in event.alerts:
        if alert.is_active:
            alert.is_active = False
            alert.resolved_at = now


def end_flood_event(db: Session, event_id: int) -> FloodEvent:
    event = db.query(FloodEvent).filter(FloodEvent.id == event_id).first()
    if not event:
        raise ValueError("Flood event not found")
    now = datetime.utcnow()
    event.is_active = False
    event.ended_at = now
    resolve_alerts_for_event(db, event)
    current_phase = db.query(FloodPhase).filter(
        and_(
            FloodPhase.flood_event_id == event.id,
            FloodPhase.ended_at == None
        )
    ).first()
    if current_phase:
        current_phase.ended_at = now
    end_phase = FloodPhase(
        flood_event_id=event.id,
        phase_type=FloodPhaseType.ENDED.value,
        description="洪水事件结束",
        ended_at=now
    )
    db.add(end_phase)
    db.commit()
    return event
