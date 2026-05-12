from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from typing import List
from datetime import datetime
from ..database import get_db
from ..models import Route, Flight, FlightType
from ..schemas import SimulationInput, SimulationReport, ConflictResponse, FlightCreate
from ..services.conflict_detector import detect_conflicts, get_time_difference, get_available_altitude

router = APIRouter(prefix="/simulation", tags=["simulation"])


@router.post("/run", response_model=SimulationReport)
def run_simulation(simulation: SimulationInput, db: Session = Depends(get_db)):
    if not simulation.flights:
        raise HTTPException(status_code=400, detail="No flights provided for simulation")
    
    route_id = simulation.flights[0].route_id
    route = db.query(Route).filter(Route.id == route_id).first()
    if not route:
        raise HTTPException(status_code=404, detail=f"Route {route_id} not found")
    
    for flight in simulation.flights:
        if flight.route_id != route_id:
            raise HTTPException(status_code=400, detail="All flights must be on the same route")
    
    temp_flights = []
    for i, flight_data in enumerate(simulation.flights):
        flight = Flight(
            id=i + 1,
            flight_number=flight_data.flight_number,
            route_id=flight_data.route_id,
            direction=flight_data.direction,
            altitude=flight_data.altitude,
            flight_type=flight_data.flight_type,
            estimated_entry_time=flight_data.estimated_entry_time
        )
        temp_flights.append(flight)
    
    conflicts = []
    min_interval = route.min_interval_minutes
    
    for i in range(len(temp_flights)):
        for j in range(i + 1, len(temp_flights)):
            f1 = temp_flights[i]
            f2 = temp_flights[j]
            
            if f1.direction == f2.direction:
                time_diff = get_time_difference(f1.estimated_entry_time, f2.estimated_entry_time)
                
                if time_diff < min_interval:
                    conflicts.append({
                        "id": len(conflicts) + 1,
                        "route_id": route_id,
                        "flight1_id": f1.id,
                        "flight2_id": f2.id,
                        "conflict_type": "horizontal",
                        "description": f"航班 {f1.flight_number} 和 {f2.flight_number} 预计进入航路时间差 {time_diff:.1f} 分钟，小于最小间隔 {min_interval} 分钟",
                        "detected_at": datetime.utcnow(),
                        "resolved": False,
                        "resolution_suggestion": f"建议将其中一个航班调整进入时间至少 {min_interval - time_diff:.1f} 分钟"
                    })
            
            if f1.altitude == f2.altitude:
                conflicts.append({
                    "id": len(conflicts) + 1,
                    "route_id": route_id,
                    "flight1_id": f1.id,
                    "flight2_id": f2.id,
                    "conflict_type": "vertical",
                    "description": f"航班 {f1.flight_number} 和 {f2.flight_number} 使用相同高度层 {f1.altitude}m",
                    "detected_at": datetime.utcnow(),
                    "resolved": False,
                    "resolution_suggestion": f"建议调整其中一个航班高度层，东单西双规则：向东飞行使用奇数高度层，向西使用偶数高度层"
                })
    
    suggestions = []
    if conflicts:
        suggestions.append(f"检测到 {len(conflicts)} 个潜在冲突")
        
        used_altitudes = {}
        for flight in temp_flights:
            key = flight.direction
            if key not in used_altitudes:
                used_altitudes[key] = []
            used_altitudes[key].append(flight.altitude)
        
        for conflict in conflicts:
            if conflict["conflict_type"] == "vertical":
                f1 = next(f for f in temp_flights if f.id == conflict["flight1_id"])
                new_alt = get_available_altitude(f1.direction, used_altitudes.get(f1.direction, []))
                suggestions.append(f"建议将 {f1.flight_number} 调整到 {new_alt}m 高度层")
    
    if not suggestions:
        suggestions.append("所有航班计划无冲突，可正常执行")
    
    conflict_responses = [ConflictResponse(**c) for c in conflicts]
    
    return SimulationReport(
        total_flights=len(temp_flights),
        conflicts=conflict_responses,
        suggestions=suggestions
    )
