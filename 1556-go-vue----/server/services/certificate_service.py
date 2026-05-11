from datetime import date
from typing import List, Optional
from sqlalchemy.orm import Session
from server.models import Certificate
from server.schemas import CertificateCreate, CertificateUpdate
from server.config import settings


class CertificateService:
    @staticmethod
    def calculate_certificate_status(expiry_date: date) -> tuple[str, str, int]:
        today = date.today()
        days_to_expiry = (expiry_date - today).days
        
        if days_to_expiry < 0:
            return "过期", "红色", days_to_expiry
        elif days_to_expiry <= settings.CERTIFICATE_WARNING_DAYS_ORANGE:
            return "即将过期", "橙色", days_to_expiry
        elif days_to_expiry <= settings.CERTIFICATE_WARNING_DAYS_YELLOW:
            return "即将过期", "黄色", days_to_expiry
        else:
            return "有效", "绿色", days_to_expiry
    
    @staticmethod
    def get_certificates_by_crew(db: Session, crew_id: int) -> List[Certificate]:
        return db.query(Certificate).filter(Certificate.crew_id == crew_id).all()
    
    @staticmethod
    def create_certificate(db: Session, certificate_data: CertificateCreate) -> Certificate:
        validity_days = (certificate_data.expiry_date - certificate_data.issue_date).days
        status, _, _ = CertificateService.calculate_certificate_status(certificate_data.expiry_date)
        
        certificate = Certificate(
            crew_id=certificate_data.crew_id,
            certificate_type=certificate_data.certificate_type,
            certificate_number=certificate_data.certificate_number,
            issue_date=certificate_data.issue_date,
            expiry_date=certificate_data.expiry_date,
            validity_days=validity_days,
            status=status,
        )
        db.add(certificate)
        db.commit()
        db.refresh(certificate)
        return certificate
    
    @staticmethod
    def update_certificate(db: Session, certificate_id: int, certificate_data: CertificateUpdate) -> Optional[Certificate]:
        certificate = db.query(Certificate).filter(Certificate.id == certificate_id).first()
        if not certificate:
            return None
        
        update_data = certificate_data.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(certificate, key, value)
        
        if "expiry_date" in update_data or "issue_date" in update_data:
            certificate.validity_days = (certificate.expiry_date - certificate.issue_date).days
            certificate.status, _, _ = CertificateService.calculate_certificate_status(certificate.expiry_date)
        
        db.commit()
        db.refresh(certificate)
        return certificate
    
    @staticmethod
    def has_expired_certificate(db: Session, crew_id: int) -> bool:
        today = date.today()
        expired_certificates = db.query(Certificate).filter(
            Certificate.crew_id == crew_id,
            Certificate.expiry_date < today
        ).first()
        return expired_certificates is not None
    
    @staticmethod
    def get_all_warning_certificates(db: Session, min_days: int = 90) -> List[Certificate]:
        today = date.today()
        warning_date = today.replace(day=today.day + min_days)
        return db.query(Certificate).filter(
            Certificate.expiry_date <= warning_date
        ).all()
