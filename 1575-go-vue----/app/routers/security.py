from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from typing import List, Dict, Any
from datetime import datetime
from app.database import get_db, Station, Zone, SecurityGate, Alert
from app.schemas import SecurityGateResponse, SecurityGateUpdate

router = APIRouter()

MIN_OPEN_GATES = 2

@router.post("/stations/{station_code}/security/{gate_number}/open", response_model=SecurityGateResponse)
def open_security_gate(station_code: str, gate_number: str, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    gate = db.query(SecurityGate).join(Zone).filter(
        Zone.station_id == station.id,
        SecurityGate.gate_number == gate_number
    ).first()
    
    if not gate:
        raise HTTPException(status_code=404, detail="安检通道不存在")
    
    if gate.is_faulty:
        raise HTTPException(status_code=400, detail="安检通道故障，无法开启")
    
    gate.is_open = True
    db.commit()
    db.refresh(gate)
    return gate

@router.post("/stations/{station_code}/security/{gate_number}/close", response_model=SecurityGateResponse)
def close_security_gate(station_code: str, gate_number: str, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    gate = db.query(SecurityGate).join(Zone).filter(
        Zone.station_id == station.id,
        SecurityGate.gate_number == gate_number
    ).first()
    
    if not gate:
        raise HTTPException(status_code=404, detail="安检通道不存在")
    
    available_gates = db.query(SecurityGate).join(Zone).filter(
        Zone.station_id == station.id,
        SecurityGate.zone_id == gate.zone_id,
        SecurityGate.is_open == True,
        SecurityGate.is_faulty == False
    ).count()
    
    if available_gates <= MIN_OPEN_GATES:
        raise HTTPException(status_code=400, detail=f"至少需要保留{MIN_OPEN_GATES}个开放的安检通道")
    
    gate.is_open = False
    db.commit()
    db.refresh(gate)
    return gate

@router.put("/stations/{station_code}/security/{gate_number}", response_model=SecurityGateResponse)
def update_security_gate(
    station_code: str, 
    gate_number: str, 
    gate_update: SecurityGateUpdate, 
    db: Session = Depends(get_db)
):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    gate = db.query(SecurityGate).join(Zone).filter(
        Zone.station_id == station.id,
        SecurityGate.gate_number == gate_number
    ).first()
    
    if not gate:
        raise HTTPException(status_code=404, detail="安检通道不存在")
    
    update_data = gate_update.dict(exclude_unset=True)
    for key, value in update_data.items():
        setattr(gate, key, value)
    
    db.commit()
    db.refresh(gate)
    return gate

@router.get("/stations/{station_code}/security/status", response_model=List[SecurityGateResponse])
def get_security_status(station_code: str, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    gates = db.query(SecurityGate).join(Zone).filter(
        Zone.station_id == station.id
    ).all()
    return gates

@router.post("/stations/{station_code}/security/{gate_number}/fault", response_model=SecurityGateResponse)
def mark_gate_faulty(station_code: str, gate_number: str, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    gate = db.query(SecurityGate).join(Zone).filter(
        Zone.station_id == station.id,
        SecurityGate.gate_number == gate_number
    ).first()
    
    if not gate:
        raise HTTPException(status_code=404, detail="安检通道不存在")
    
    gate.is_faulty = True
    gate.is_open = False
    db.commit()
    db.refresh(gate)
    
    alert = Alert(
        station_id=station.id,
        zone_id=gate.zone_id,
        alert_type="security_fault",
        severity="warning",
        message=f"安检通道 {gate_number} 发生故障，已自动关闭",
        is_active=True
    )
    db.add(alert)
    db.commit()
    
    return gate

@router.post("/stations/{station_code}/security/{gate_number}/repair", response_model=SecurityGateResponse)
def repair_gate(station_code: str, gate_number: str, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    gate = db.query(SecurityGate).join(Zone).filter(
        Zone.station_id == station.id,
        SecurityGate.gate_number == gate_number
    ).first()
    
    if not gate:
        raise HTTPException(status_code=404, detail="安检通道不存在")
    
    gate.is_faulty = False
    db.commit()
    db.refresh(gate)
    
    active_alerts = db.query(Alert).filter(
        Alert.station_id == station.id,
        Alert.zone_id == gate.zone_id,
        Alert.alert_type == "security_fault",
        Alert.is_active == True
    ).all()
    
    for alert in active_alerts:
        alert.is_active = False
        alert.resolved_at = datetime.utcnow()
    db.commit()
    
    return gate

@router.get("/stations/{station_code}/security/suggestions")
def get_gate_suggestions(station_code: str, db: Session = Depends(get_db)):
    station = db.query(Station).filter(Station.code == station_code).first()
    if not station:
        raise HTTPException(status_code=404, detail="车站不存在")
    
    zones = db.query(Zone).filter(Zone.station_id == station.id).all()
    suggestions = []
    
    for zone in zones:
        gates = db.query(SecurityGate).filter(
            SecurityGate.zone_id == zone.id,
            SecurityGate.is_faulty == False
        ).all()
        
        if not gates:
            continue
        
        open_gates = [g for g in gates if g.is_open]
        total_queue = sum(g.queue_length for g in gates)
        avg_queue_per_gate = total_queue / max(len(open_gates), 1) if open_gates else 0
        
        suggested_open_count = len(open_gates)
        if avg_queue_per_gate > 30:
            suggested_open_count = min(len(gates), len(open_gates) + 1)
        elif avg_queue_per_gate < 10 and len(open_gates) > MIN_OPEN_GATES:
            suggested_open_count = max(MIN_OPEN_GATES, len(open_gates) - 1)
        
        suggestions.append({
            "zone_id": zone.id,
            "zone_name": zone.name,
            "total_gates": len(gates),
            "open_gates": len(open_gates),
            "faulty_gates": sum(1 for g in gates if g.is_faulty),
            "total_queue": total_queue,
            "avg_queue_per_gate": round(avg_queue_per_gate, 2),
            "suggested_open_count": suggested_open_count,
            "action": "OPEN_MORE" if suggested_open_count > len(open_gates) else 
                      "CLOSE_SOME" if suggested_open_count < len(open_gates) else "MAINTAIN"
        })
    
    return {"station_code": station_code, "suggestions": suggestions}
