from typing import List, Optional
from datetime import date
from decimal import Decimal
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session

from core import (
    get_db, QuotaTransaction, Company,
    QuotaTransactionCreate, QuotaTransactionResponse,
    TransactionType, calculate_quota_balance, QuotaBalance
)

router = APIRouter()

@router.get("", response_model=List[QuotaTransactionResponse])
def list_transactions(
    company_id: Optional[int] = None,
    start_date: Optional[date] = None,
    end_date: Optional[date] = None,
    db: Session = Depends(get_db)
):
    query = db.query(QuotaTransaction)
    if company_id:
        query = query.filter(QuotaTransaction.company_id == company_id)
    if start_date:
        query = query.filter(QuotaTransaction.transaction_date >= start_date)
    if end_date:
        query = query.filter(QuotaTransaction.transaction_date <= end_date)
    return query.order_by(QuotaTransaction.transaction_date.desc()).all()

@router.post("", response_model=QuotaTransactionResponse, status_code=status.HTTP_201_CREATED)
def create_transaction(data: QuotaTransactionCreate, db: Session = Depends(get_db)):
    company = db.query(Company).filter(Company.id == data.company_id).first()
    if not company:
        raise HTTPException(status_code=404, detail="企业不存在")
    
    if data.transaction_type == TransactionType.SELL:
        balance = calculate_quota_balance(db, data.company_id)
        if balance["current_balance"] < data.quota_amount:
            raise HTTPException(
                status_code=400,
                detail=f"配额不足，当前可用配额: {balance['current_balance']}"
            )
    
    total_amount = data.quota_amount * data.price_per_unit
    
    transaction = QuotaTransaction(
        company_id=data.company_id,
        transaction_type=data.transaction_type,
        quota_amount=data.quota_amount,
        price_per_unit=data.price_per_unit,
        total_amount=total_amount,
        counterparty=data.counterparty,
        transaction_date=data.transaction_date,
        remarks=data.remarks
    )
    
    db.add(transaction)
    db.commit()
    db.refresh(transaction)
    return transaction

@router.get("/{transaction_id}", response_model=QuotaTransactionResponse)
def get_transaction(transaction_id: int, db: Session = Depends(get_db)):
    transaction = db.query(QuotaTransaction).filter(QuotaTransaction.id == transaction_id).first()
    if not transaction:
        raise HTTPException(status_code=404, detail="交易记录不存在")
    return transaction

@router.get("/balance/{company_id}", response_model=QuotaBalance)
def get_quota_balance(company_id: int, db: Session = Depends(get_db)):
    company = db.query(Company).filter(Company.id == company_id).first()
    if not company:
        raise HTTPException(status_code=404, detail="企业不存在")
    
    balance = calculate_quota_balance(db, company_id)
    if not balance:
        raise HTTPException(status_code=500, detail="计算配额余额失败")
    return QuotaBalance(**balance)
