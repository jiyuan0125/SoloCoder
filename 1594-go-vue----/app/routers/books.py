from typing import List, Optional
from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from sqlalchemy import or_

from app.database import get_db
from app.models import Book, BookCopy, CopyStatus
from app.schemas import (
    BookCreate,
    BookUpdate,
    BookCopyCreate,
    BookCopyUpdate,
    Book as BookSchema,
    BookCopy as BookCopySchema,
)


router = APIRouter(prefix="/books", tags=["books"])


@router.post("", response_model=BookSchema)
def create_book(book_data: BookCreate, db: Session = Depends(get_db)):
    existing = db.query(Book).filter(Book.isbn == book_data.isbn).first()
    if existing:
        raise HTTPException(status_code=400, detail="ISBN已存在")
    
    book = Book(**book_data.model_dump())
    db.add(book)
    db.commit()
    db.refresh(book)
    return book


@router.get("", response_model=List[BookSchema])
def list_books(
    skip: int = 0,
    limit: int = 100,
    search: Optional[str] = None,
    db: Session = Depends(get_db),
):
    query = db.query(Book)
    if search:
        query = query.filter(
            or_(
                Book.title.contains(search),
                Book.author.contains(search),
                Book.isbn.contains(search),
            )
        )
    return query.offset(skip).limit(limit).all()


@router.get("/{book_id}", response_model=BookSchema)
def get_book(book_id: int, db: Session = Depends(get_db)):
    book = db.query(Book).filter(Book.id == book_id).first()
    if not book:
        raise HTTPException(status_code=404, detail="图书不存在")
    return book


@router.put("/{book_id}", response_model=BookSchema)
def update_book(book_id: int, update_data: BookUpdate, db: Session = Depends(get_db)):
    book = db.query(Book).filter(Book.id == book_id).first()
    if not book:
        raise HTTPException(status_code=404, detail="图书不存在")
    
    for key, value in update_data.model_dump(exclude_unset=True).items():
        setattr(book, key, value)
    
    db.commit()
    db.refresh(book)
    return book


@router.delete("/{book_id}")
def delete_book(book_id: int, db: Session = Depends(get_db)):
    book = db.query(Book).filter(Book.id == book_id).first()
    if not book:
        raise HTTPException(status_code=404, detail="图书不存在")
    
    for copy in book.copies:
        if copy.status not in [CopyStatus.AVAILABLE.value, CopyStatus.MAINTENANCE.value]:
            raise HTTPException(status_code=400, detail="存在已借出或已预约的副本，无法删除")
    
    db.delete(book)
    db.commit()
    return {"message": "删除成功"}


@router.post("/copies", response_model=BookCopySchema)
def create_copy(copy_data: BookCopyCreate, db: Session = Depends(get_db)):
    book = db.query(Book).filter(Book.id == copy_data.book_id).first()
    if not book:
        raise HTTPException(status_code=404, detail="图书不存在")
    
    existing = db.query(BookCopy).filter(BookCopy.copy_number == copy_data.copy_number).first()
    if existing:
        raise HTTPException(status_code=400, detail="副本编号已存在")
    
    copy = BookCopy(**copy_data.model_dump())
    db.add(copy)
    db.commit()
    db.refresh(copy)
    return copy


@router.get("/copies/", response_model=List[BookCopySchema])
def list_copies(
    book_id: Optional[int] = None,
    status: Optional[CopyStatus] = None,
    skip: int = 0,
    limit: int = 100,
    db: Session = Depends(get_db),
):
    query = db.query(BookCopy)
    if book_id:
        query = query.filter(BookCopy.book_id == book_id)
    if status:
        query = query.filter(BookCopy.status == status.value)
    return query.offset(skip).limit(limit).all()


@router.get("/copies/{copy_id}", response_model=BookCopySchema)
def get_copy(copy_id: int, db: Session = Depends(get_db)):
    copy = db.query(BookCopy).filter(BookCopy.id == copy_id).first()
    if not copy:
        raise HTTPException(status_code=404, detail="副本不存在")
    return copy


@router.put("/copies/{copy_id}", response_model=BookCopySchema)
def update_copy(copy_id: int, update_data: BookCopyUpdate, db: Session = Depends(get_db)):
    copy = db.query(BookCopy).filter(BookCopy.id == copy_id).first()
    if not copy:
        raise HTTPException(status_code=404, detail="副本不存在")
    
    for key, value in update_data.model_dump(exclude_unset=True).items():
        if key == "status":
            value = value.value
        setattr(copy, key, value)
    
    db.commit()
    db.refresh(copy)
    return copy


@router.delete("/copies/{copy_id}")
def delete_copy(copy_id: int, db: Session = Depends(get_db)):
    copy = db.query(BookCopy).filter(BookCopy.id == copy_id).first()
    if not copy:
        raise HTTPException(status_code=404, detail="副本不存在")
    
    if copy.status not in [CopyStatus.AVAILABLE.value, CopyStatus.MAINTENANCE.value]:
        raise HTTPException(status_code=400, detail="该副本已借出或已预约，无法删除")
    
    db.delete(copy)
    db.commit()
    return {"message": "删除成功"}
