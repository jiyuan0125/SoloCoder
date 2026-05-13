import os
from datetime import datetime
from fastapi import FastAPI, Depends, HTTPException, Query
from fastapi.responses import JSONResponse
from sqlalchemy.orm import Session

from database import init_db, get_db
from quota_manager import QuotaManager
import schemas

app = FastAPI(title="Quota Limiter", description="FastAPI Quota Management Service")


@app.on_event("startup")
def startup_event():
    init_db()


@app.post("/quotas", response_model=schemas.QuotaSetResponse, status_code=201)
def set_quota(request: schemas.QuotaSetRequest, db: Session = Depends(get_db)):
    manager = QuotaManager(db)
    quota = manager.set_quota(request.user_id, request.resource_type, request.limit)
    return {
        "user_id": quota.user_id,
        "resource_type": quota.resource_type,
        "limit": quota.limit,
        "status": "set"
    }


@app.post("/quotas/reserve", response_model=schemas.QuotaReserveResponse, status_code=201)
def set_reservation(request: schemas.QuotaReserveRequest, db: Session = Depends(get_db)):
    manager = QuotaManager(db)
    reservation = manager.set_reservation(request.user_id, request.resource_type, request.reserved_amount)
    return {
        "user_id": reservation.user_id,
        "resource_type": reservation.resource_type,
        "reserved_amount": reservation.reserved_amount,
        "reserved_used": reservation.reserved_used,
        "status": "reserved"
    }


@app.get("/quotas/check", response_model=schemas.QuotaCheckResponse)
def check_quota(user_id: str = Query(...), resource_type: str = Query(...), db: Session = Depends(get_db)):
    manager = QuotaManager(db)
    result = manager.check_quota(user_id, resource_type)
    if result["remaining"] == 0:
        raise HTTPException(
            status_code=429,
            detail={
                "error": "Quota exceeded",
                "user_id": user_id,
                "resource_type": resource_type,
                "remaining": 0
            }
        )
    return result


@app.get("/quotas/usage", response_model=schemas.UserUsageResponse)
def get_user_usage(user_id: str = Query(...), db: Session = Depends(get_db)):
    manager = QuotaManager(db)
    return manager.get_user_usage(user_id)


@app.get("/quotas/usage/stats", response_model=schemas.ResourceStatsResponse)
def get_stats(
    start_time: datetime = Query(None),
    end_time: datetime = Query(None),
    db: Session = Depends(get_db)
):
    manager = QuotaManager(db)
    return manager.get_resource_stats(start_time, end_time)


@app.post("/quotas/consume", response_model=schemas.ConsumeResponse)
def consume(request: schemas.ConsumeRequest, db: Session = Depends(get_db)):
    manager = QuotaManager(db)
    result = manager.consume(request.user_id, request.resource_type, request.amount)
    
    if not result["success"]:
        raise HTTPException(
            status_code=429,
            detail={
                "error": result["message"],
                "user_id": request.user_id,
                "resource_type": request.resource_type,
                "remaining": result["remaining"]
            }
        )
    
    return {
        "user_id": request.user_id,
        "resource_type": request.resource_type,
        "amount": request.amount,
        "success": True,
        "remaining": result["remaining"]
    }


@app.post("/quotas/release", response_model=schemas.ReleaseResponse)
def release(request: schemas.ReleaseRequest, db: Session = Depends(get_db)):
    manager = QuotaManager(db)
    result = manager.release(request.user_id, request.resource_type, request.amount)
    
    return {
        "user_id": request.user_id,
        "resource_type": request.resource_type,
        "amount": request.amount,
        "success": result["success"],
        "released_from_reserved": result["released_from_reserved"],
        "released_from_shared": result["released_from_shared"]
    }


if __name__ == "__main__":
    import uvicorn
    port = int(os.getenv("PORT", 8000))
    uvicorn.run("main:app", host="0.0.0.0", port=port, reload=True)
