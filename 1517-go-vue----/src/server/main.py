import os
from typing import List, Optional
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException, Query
from fastapi.responses import JSONResponse

from ..core import (
    AnimalRescueService,
    AnimalCreate, AnimalUpdate,
    AdopterCreate, AdopterUpdate,
    AdoptionApplicationCreate,
    FollowUpUpdate,
    DonationCreate,
    AppointmentCreate, AppointmentUpdate,
)


service: AnimalRescueService = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    global service
    service = AnimalRescueService()
    yield


app = FastAPI(
    title="动物救助站管理系统",
    description="动物救助站后端管理系统 - FastAPI",
    version="1.0.0",
    lifespan=lifespan
)


@app.get("/", tags=["根"])
async def root():
    return {"message": "动物救助站管理系统 API"}


@app.get("/health", tags=["根"])
async def health_check():
    return {"status": "healthy"}


@app.post("/animals", tags=["动物管理"], response_model_exclude_none=True)
async def create_animal(data: AnimalCreate):
    animal = service.create_animal(data)
    return animal.model_dump(mode='json')


@app.get("/animals", tags=["动物管理"])
async def list_animals(
    species: Optional[str] = Query(None, description="物种过滤"),
    health_status: Optional[str] = Query(None, description="健康状态过滤"),
    is_adoptable: Optional[bool] = Query(None, description="是否可领养")
):
    animals = service.list_animals(species, health_status, is_adoptable)
    return [a.model_dump(mode='json') for a in animals]


@app.get("/animals/{animal_id}", tags=["动物管理"])
async def get_animal(animal_id: str):
    animal = service.get_animal(animal_id)
    if not animal:
        raise HTTPException(status_code=404, detail="动物不存在")
    return animal.model_dump(mode='json')


@app.put("/animals/{animal_id}", tags=["动物管理"])
async def update_animal(animal_id: str, data: AnimalUpdate):
    animal = service.update_animal(animal_id, data)
    if not animal:
        raise HTTPException(status_code=404, detail="动物不存在")
    return animal.model_dump(mode='json')


@app.delete("/animals/{animal_id}", tags=["动物管理"])
async def delete_animal(animal_id: str):
    success = service.delete_animal(animal_id)
    if not success:
        raise HTTPException(status_code=404, detail="动物不存在")
    return {"message": "删除成功"}


@app.post("/adopters", tags=["领养人管理"])
async def create_adopter(data: AdopterCreate):
    adopter = service.create_adopter(data)
    return adopter.model_dump(mode='json')


@app.get("/adopters", tags=["领养人管理"])
async def list_adopters(
    has_bad_record: Optional[bool] = Query(None, description="是否有不良记录")
):
    adopters = service.list_adopters(has_bad_record)
    return [a.model_dump(mode='json') for a in adopters]


@app.get("/adopters/{adopter_id}", tags=["领养人管理"])
async def get_adopter(adopter_id: str):
    adopter = service.get_adopter(adopter_id)
    if not adopter:
        raise HTTPException(status_code=404, detail="领养人不存在")
    return adopter.model_dump(mode='json')


@app.put("/adopters/{adopter_id}", tags=["领养人管理"])
async def update_adopter(adopter_id: str, data: AdopterUpdate):
    adopter = service.update_adopter(adopter_id, data)
    if not adopter:
        raise HTTPException(status_code=404, detail="领养人不存在")
    return adopter.model_dump(mode='json')


@app.delete("/adopters/{adopter_id}", tags=["领养人管理"])
async def delete_adopter(adopter_id: str):
    success = service.delete_adopter(adopter_id)
    if not success:
        raise HTTPException(status_code=404, detail="领养人不存在")
    return {"message": "删除成功"}


@app.post("/adoptions", tags=["领养管理"])
async def create_adoption_application(data: AdoptionApplicationCreate):
    app, error = service.create_adoption_application(data)
    if error:
        raise HTTPException(status_code=400, detail=error)
    return app.model_dump(mode='json')


@app.get("/adoptions", tags=["领养管理"])
async def list_adoptions(
    animal_id: Optional[str] = Query(None),
    adopter_id: Optional[str] = Query(None),
    status: Optional[str] = Query(None)
):
    apps = service.list_adoption_applications(animal_id, adopter_id, status)
    return [a.model_dump(mode='json') for a in apps]


@app.get("/adoptions/{app_id}", tags=["领养管理"])
async def get_adoption(app_id: str):
    app = service.get_adoption_application(app_id)
    if not app:
        raise HTTPException(status_code=404, detail="申请不存在")
    return app.model_dump(mode='json')


@app.post("/adoptions/{app_id}/approve", tags=["领养管理"])
async def approve_application(app_id: str):
    app, error = service.approve_application(app_id)
    if error:
        raise HTTPException(status_code=400, detail=error)
    return app.model_dump(mode='json')


@app.post("/adoptions/{app_id}/approve-extra", tags=["领养管理"])
async def approve_with_extra_review(app_id: str):
    app, error = service.approve_with_extra_review(app_id)
    if error:
        raise HTTPException(status_code=400, detail=error)
    return app.model_dump(mode='json')


@app.post("/adoptions/{app_id}/reject", tags=["领养管理"])
async def reject_application(app_id: str):
    app, error = service.reject_application(app_id)
    if error:
        raise HTTPException(status_code=400, detail=error)
    return app.model_dump(mode='json')


@app.post("/adoptions/{app_id}/cancel", tags=["领养管理"])
async def cancel_application(app_id: str):
    app, error = service.cancel_application(app_id)
    if error:
        raise HTTPException(status_code=400, detail=error)
    return app.model_dump(mode='json')


@app.post("/adoptions/{app_id}/finalize", tags=["领养管理"])
async def finalize_adoption(app_id: str):
    from ..core import AdoptionApplicationUpdate, AdoptionStatus
    app = service.get_adoption_application(app_id)
    if not app:
        raise HTTPException(status_code=404, detail="申请不存在")
    if app.status != 'interview_passed':
        raise HTTPException(status_code=400, detail="只有面谈通过的申请可以完成领养")
    
    updated = service.update_adoption_application(
        app_id,
        AdoptionApplicationUpdate(status=AdoptionStatus.ADOPTED)
    )
    return updated.model_dump(mode='json')


@app.get("/follow-ups", tags=["回访管理"])
async def list_follow_ups(
    adoption_id: Optional[str] = Query(None),
    status: Optional[str] = Query(None),
    adopter_id: Optional[str] = Query(None)
):
    fus = service.list_follow_ups(adoption_id, status, adopter_id)
    return [f.model_dump(mode='json') for f in fus]


@app.get("/follow-ups/{fu_id}", tags=["回访管理"])
async def get_follow_up(fu_id: str):
    fu = service.get_follow_up(fu_id)
    if not fu:
        raise HTTPException(status_code=404, detail="回访不存在")
    return fu.model_dump(mode='json')


@app.put("/follow-ups/{fu_id}", tags=["回访管理"])
async def update_follow_up(fu_id: str, data: FollowUpUpdate):
    fu = service.update_follow_up(fu_id, data)
    if not fu:
        raise HTTPException(status_code=404, detail="回访不存在")
    return fu.model_dump(mode='json')


@app.post("/follow-ups/{fu_id}/complete", tags=["回访管理"])
async def complete_follow_up(fu_id: str, notes: Optional[str] = None):
    fu, error = service.complete_follow_up(fu_id, notes)
    if error:
        raise HTTPException(status_code=400, detail=error)
    return fu.model_dump(mode='json')


@app.post("/follow-ups/check-overdue", tags=["回访管理"])
async def check_overdue():
    overdue = service.check_overdue_follow_ups()
    return {"count": len(overdue), "items": [f.model_dump(mode='json') for f in overdue]}


@app.post("/donations", tags=["捐赠管理"])
async def create_donation(data: DonationCreate):
    don, error = service.create_donation(data)
    if error:
        raise HTTPException(status_code=400, detail=error)
    return don.model_dump(mode='json')


@app.get("/donations", tags=["捐赠管理"])
async def list_donations(
    donation_type: Optional[str] = Query(None)
):
    dons = service.list_donations(donation_type)
    return [d.model_dump(mode='json') for d in dons]


@app.get("/donations/stats", tags=["捐赠管理"])
async def donation_stats():
    stats = service.get_donation_stats()
    return stats.model_dump(mode='json')


@app.get("/donations/{don_id}", tags=["捐赠管理"])
async def get_donation(don_id: str):
    don = service.get_donation(don_id)
    if not don:
        raise HTTPException(status_code=404, detail="捐赠不存在")
    return don.model_dump(mode='json')


@app.post("/appointments", tags=["预约管理"])
async def create_appointment(data: AppointmentCreate):
    appt, error = service.create_appointment(data)
    if error:
        raise HTTPException(status_code=400, detail=error)
    return appt.model_dump(mode='json')


@app.get("/appointments", tags=["预约管理"])
async def list_appointments(
    adoption_id: Optional[str] = Query(None),
    status: Optional[str] = Query(None)
):
    appts = service.list_appointments(adoption_id, status)
    return [a.model_dump(mode='json') for a in appts]


@app.get("/appointments/{appt_id}", tags=["预约管理"])
async def get_appointment(appt_id: str):
    appt = service.get_appointment(appt_id)
    if not appt:
        raise HTTPException(status_code=404, detail="预约不存在")
    return appt.model_dump(mode='json')


@app.put("/appointments/{appt_id}", tags=["预约管理"])
async def update_appointment(appt_id: str, data: AppointmentUpdate):
    appt = service.update_appointment(appt_id, data)
    if not appt:
        raise HTTPException(status_code=404, detail="预约不存在")
    return appt.model_dump(mode='json')


@app.post("/appointments/{appt_id}/complete", tags=["预约管理"])
async def complete_appointment(appt_id: str, passed: bool, notes: Optional[str] = None):
    appt, error = service.complete_appointment(appt_id, passed, notes)
    if error:
        raise HTTPException(status_code=400, detail=error)
    return appt.model_dump(mode='json')


@app.post("/appointments/{appt_id}/cancel", tags=["预约管理"])
async def cancel_appointment(appt_id: str):
    appt, error = service.cancel_appointment(appt_id)
    if error:
        raise HTTPException(status_code=400, detail=error)
    return appt.model_dump(mode='json')


def run():
    import uvicorn
    port = int(os.environ.get("PORT", "8000"))
    uvicorn.run(app, host="0.0.0.0", port=port)


if __name__ == "__main__":
    run()
