from typing import List, Optional
from datetime import date, datetime
from fastapi import APIRouter, Depends, HTTPException, status, Query
from sqlalchemy.orm import Session
from app.database import get_db
from app.models import Borrowing, Collection, CollectionStatus, InventoryLog
from app.schemas import BorrowingCreate, ReturnBorrowing, BorrowingResponse

router = APIRouter()


@router.get("/", response_model=List[BorrowingResponse])
def list_borrowings(
    status: Optional[str] = Query(None, description="按状态筛选"),
    borrower: Optional[str] = Query(None, description="按借方筛选"),
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db)
):
    query = db.query(Borrowing)
    if status:
        query = query.filter(Borrowing.status == status)
    if borrower:
        query = query.filter(Borrowing.borrower.contains(borrower))

    borrowings = query.order_by(Borrowing.created_at.desc()).offset(skip).limit(limit).all()

    results = []
    for b in borrowings:
        collection = db.query(Collection).filter(Collection.id == b.collection_id).first()
        results.append(BorrowingResponse(
            id=b.id,
            collection_id=b.collection_id,
            collection_name=collection.name if collection else None,
            borrower=b.borrower,
            contact=b.contact,
            borrow_date=b.borrow_date,
            expected_return_date=b.expected_return_date,
            actual_return_date=b.actual_return_date,
            purpose=b.purpose,
            status=b.status,
            created_at=b.created_at,
            updated_at=b.updated_at
        ))
    return results


@router.post("/", response_model=BorrowingResponse, status_code=status.HTTP_201_CREATED)
def create_borrowing(borrowing: BorrowingCreate, db: Session = Depends(get_db)):
    today = date.today()
    if borrowing.borrow_date < today:
        raise HTTPException(status_code=400, detail="出借日期不能是过去的日期")

    if borrowing.expected_return_date < borrowing.borrow_date:
        raise HTTPException(status_code=400, detail="预计归还日期不能早于出借日期")

    collection = db.query(Collection).filter(Collection.id == borrowing.collection_id).first()
    if not collection:
        raise HTTPException(status_code=404, detail="藏品不存在")

    if collection.status == CollectionStatus.BORROWED.value:
        raise HTTPException(status_code=400, detail="该藏品已在外借中")

    if collection.status == CollectionStatus.ON_EXHIBIT.value:
        raise HTTPException(status_code=400, detail="该藏品正在展出中，无法外借")

    if collection.status != CollectionStatus.IN_STOCK.value:
        raise HTTPException(status_code=400, detail=f"该藏品状态为 '{collection.status}'，不在库，无法外借")

    collection.status = CollectionStatus.BORROWED.value

    db_borrowing = Borrowing(
        collection_id=borrowing.collection_id,
        borrower=borrowing.borrower,
        contact=borrowing.contact,
        borrow_date=borrowing.borrow_date,
        expected_return_date=borrowing.expected_return_date,
        purpose=borrowing.purpose
    )
    db.add(db_borrowing)

    log = InventoryLog(
        collection_id=collection.id,
        operation_type="外借出库",
        operator=borrowing.operator,
        notes=f"藏品外借: {borrowing.borrower}"
    )
    db.add(log)

    db.commit()
    db.refresh(db_borrowing)

    return BorrowingResponse(
        id=db_borrowing.id,
        collection_id=db_borrowing.collection_id,
        collection_name=collection.name,
        borrower=db_borrowing.borrower,
        contact=db_borrowing.contact,
        borrow_date=db_borrowing.borrow_date,
        expected_return_date=db_borrowing.expected_return_date,
        actual_return_date=db_borrowing.actual_return_date,
        purpose=db_borrowing.purpose,
        status=db_borrowing.status,
        created_at=db_borrowing.created_at,
        updated_at=db_borrowing.updated_at
    )


@router.get("/{borrowing_id}", response_model=BorrowingResponse)
def get_borrowing(borrowing_id: int, db: Session = Depends(get_db)):
    borrowing = db.query(Borrowing).filter(Borrowing.id == borrowing_id).first()
    if not borrowing:
        raise HTTPException(status_code=404, detail="外借记录不存在")

    collection = db.query(Collection).filter(Collection.id == borrowing.collection_id).first()

    return BorrowingResponse(
        id=borrowing.id,
        collection_id=borrowing.collection_id,
        collection_name=collection.name if collection else None,
        borrower=borrowing.borrower,
        contact=borrowing.contact,
        borrow_date=borrowing.borrow_date,
        expected_return_date=borrowing.expected_return_date,
        actual_return_date=borrowing.actual_return_date,
        purpose=borrowing.purpose,
        status=borrowing.status,
        created_at=borrowing.created_at,
        updated_at=borrowing.updated_at
    )


@router.post("/{borrowing_id}/return")
def return_borrowing(borrowing_id: int, data: ReturnBorrowing, db: Session = Depends(get_db)):
    borrowing = db.query(Borrowing).filter(Borrowing.id == borrowing_id).first()
    if not borrowing:
        raise HTTPException(status_code=404, detail="外借记录不存在")

    if borrowing.status != "外借中":
        raise HTTPException(status_code=400, detail=f"该外借记录状态为 '{borrowing.status}'，无法归还")

    collection = db.query(Collection).filter(Collection.id == borrowing.collection_id).first()
    if not collection:
        raise HTTPException(status_code=404, detail="关联藏品不存在")

    actual_return = data.actual_return_date or date.today()
    borrowing.actual_return_date = actual_return
    borrowing.status = "已归还"
    collection.status = CollectionStatus.IN_STOCK.value

    log = InventoryLog(
        collection_id=collection.id,
        operation_type="外借归还入库",
        operator=data.operator,
        notes=f"藏品归还入库，外借方: {borrowing.borrower}"
    )
    db.add(log)

    db.commit()
    db.refresh(borrowing)

    return {
        "message": "藏品已归还",
        "borrowing_id": borrowing_id,
        "actual_return_date": actual_return
    }


@router.get("/overdue/today")
def get_overdue_borrowings(db: Session = Depends(get_db)):
    today = date.today()
    overdue = db.query(Borrowing).filter(
        Borrowing.status == "外借中",
        Borrowing.expected_return_date < today
    ).all()

    results = []
    for b in overdue:
        collection = db.query(Collection).filter(Collection.id == b.collection_id).first()
        overdue_days = (today - b.expected_return_date).days
        results.append({
            "id": b.id,
            "collection_id": b.collection_id,
            "collection_name": collection.name if collection else None,
            "borrower": b.borrower,
            "expected_return_date": b.expected_return_date,
            "overdue_days": overdue_days
        })

    return {"today": today, "overdue_count": len(results), "overdue_items": results}
