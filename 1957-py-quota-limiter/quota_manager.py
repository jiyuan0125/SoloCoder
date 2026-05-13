from typing import Optional, Tuple
from sqlalchemy.orm import Session
from models import QuotaLimit, QuotaReservation, QuotaUsage, QuotaLog
from datetime import datetime, timedelta


class QuotaManager:
    def __init__(self, db: Session):
        self.db = db
    
    def _log(self, user_id: str, resource_type: str, action: str, amount: int, result: str):
        log = QuotaLog(
            user_id=user_id,
            resource_type=resource_type,
            action=action,
            amount=amount,
            result=result
        )
        self.db.add(log)
        self.db.commit()
    
    def set_quota(self, user_id: str, resource_type: str, limit: int):
        quota = self.db.query(QuotaLimit).filter(
            QuotaLimit.user_id == user_id,
            QuotaLimit.resource_type == resource_type
        ).first()
        
        if quota:
            quota.limit = limit
        else:
            quota = QuotaLimit(
                user_id=user_id,
                resource_type=resource_type,
                limit=limit
            )
            self.db.add(quota)
        
        self.db.commit()
        self.db.refresh(quota)
        return quota
    
    def set_reservation(self, user_id: str, resource_type: str, reserved_amount: int):
        reservation = self.db.query(QuotaReservation).filter(
            QuotaReservation.user_id == user_id,
            QuotaReservation.resource_type == resource_type
        ).first()
        
        if reservation:
            reservation.reserved_amount = reserved_amount
            if reservation.reserved_used > reserved_amount:
                reservation.reserved_used = reserved_amount
        else:
            reservation = QuotaReservation(
                user_id=user_id,
                resource_type=resource_type,
                reserved_amount=reserved_amount,
                reserved_used=0
            )
            self.db.add(reservation)
        
        self.db.commit()
        self.db.refresh(reservation)
        return reservation
    
    def _get_or_create_usage(self, user_id: str, resource_type: str) -> QuotaUsage:
        usage = self.db.query(QuotaUsage).filter(
            QuotaUsage.user_id == user_id,
            QuotaUsage.resource_type == resource_type
        ).first()
        
        if not usage:
            usage = QuotaUsage(
                user_id=user_id,
                resource_type=resource_type,
                usage=0
            )
            self.db.add(usage)
            self.db.commit()
            self.db.refresh(usage)
        
        return usage
    
    def _get_limit(self, user_id: str, resource_type: str) -> int:
        quota = self.db.query(QuotaLimit).filter(
            QuotaLimit.user_id == user_id,
            QuotaLimit.resource_type == resource_type
        ).first()
        return quota.limit if quota else 0
    
    def _get_reservation(self, user_id: str, resource_type: str) -> Tuple[int, int]:
        reservation = self.db.query(QuotaReservation).filter(
            QuotaReservation.user_id == user_id,
            QuotaReservation.resource_type == resource_type
        ).first()
        
        if reservation:
            return reservation.reserved_amount, reservation.reserved_used
        return 0, 0
    
    def check_quota(self, user_id: str, resource_type: str) -> dict:
        limit = self._get_limit(user_id, resource_type)
        reserved_amount, reserved_used = self._get_reservation(user_id, resource_type)
        usage = self._get_or_create_usage(user_id, resource_type)
        
        total_limit = limit + reserved_amount
        total_used = usage.usage
        remaining = total_limit - total_used
        
        return {
            "user_id": user_id,
            "resource_type": resource_type,
            "total_limit": total_limit,
            "total_used": total_used,
            "remaining": max(0, remaining),
            "reserved_amount": reserved_amount,
            "reserved_used": reserved_used,
            "can_consume": remaining > 0
        }
    
    def consume(self, user_id: str, resource_type: str, amount: int = 1) -> dict:
        check = self.check_quota(user_id, resource_type)
        
        if check["remaining"] < amount:
            self._log(user_id, resource_type, "consume", amount, "rejected")
            return {
                "success": False,
                "remaining": check["remaining"],
                "message": "Quota exceeded"
            }
        
        usage = self._get_or_create_usage(user_id, resource_type)
        reserved_amount, reserved_used = self._get_reservation(user_id, resource_type)
        reservation = self.db.query(QuotaReservation).filter(
            QuotaReservation.user_id == user_id,
            QuotaReservation.resource_type == resource_type
        ).first()
        
        amount_to_consume = amount
        consumed_from_reserved = 0
        consumed_from_shared = 0
        
        if reservation and reserved_used < reserved_amount:
            available_reserved = reserved_amount - reserved_used
            consumed_from_reserved = min(amount_to_consume, available_reserved)
            reservation.reserved_used += consumed_from_reserved
            amount_to_consume -= consumed_from_reserved
        
        if amount_to_consume > 0:
            consumed_from_shared = amount_to_consume
        
        usage.usage += amount
        self.db.commit()
        self.db.refresh(usage)
        
        if reservation:
            self.db.refresh(reservation)
        
        check = self.check_quota(user_id, resource_type)
        self._log(user_id, resource_type, "consume", amount, "success")
        
        return {
            "success": True,
            "remaining": check["remaining"],
            "consumed_from_reserved": consumed_from_reserved,
            "consumed_from_shared": consumed_from_shared
        }
    
    def release(self, user_id: str, resource_type: str, amount: int = 1) -> dict:
        usage = self._get_or_create_usage(user_id, resource_type)
        
        if usage.usage <= 0:
            self._log(user_id, resource_type, "release", amount, "nothing_to_release")
            return {
                "success": True,
                "released_from_reserved": 0,
                "released_from_shared": 0
            }
        
        reserved_amount, reserved_used = self._get_reservation(user_id, resource_type)
        reservation = self.db.query(QuotaReservation).filter(
            QuotaReservation.user_id == user_id,
            QuotaReservation.resource_type == resource_type
        ).first()
        
        amount_to_release = min(amount, usage.usage)
        released_from_reserved = 0
        released_from_shared = 0
        
        if reservation and reserved_used > 0:
            released_from_reserved = min(amount_to_release, reserved_used)
            reservation.reserved_used -= released_from_reserved
            amount_to_release -= released_from_reserved
        
        if amount_to_release > 0:
            released_from_shared = amount_to_release
        
        usage.usage -= (released_from_reserved + released_from_shared)
        self.db.commit()
        self.db.refresh(usage)
        
        if reservation:
            self.db.refresh(reservation)
        
        self._log(user_id, resource_type, "release", released_from_reserved + released_from_shared, "success")
        
        return {
            "success": True,
            "released_from_reserved": released_from_reserved,
            "released_from_shared": released_from_shared
        }
    
    def get_user_usage(self, user_id: str) -> dict:
        usages = self.db.query(QuotaUsage).filter(
            QuotaUsage.user_id == user_id
        ).all()
        
        quota_limits = {
            q.resource_type: q.limit for q in self.db.query(QuotaLimit).filter(
                QuotaLimit.user_id == user_id
            ).all()
        }
        
        reservations = {
            r.resource_type: (r.reserved_amount, r.reserved_used) for r in self.db.query(QuotaReservation).filter(
                QuotaReservation.user_id == user_id
            ).all()
        }
        
        result = []
        for usage in usages:
            limit = quota_limits.get(usage.resource_type, 0)
            reserved_amount, reserved_used = reservations.get(usage.resource_type, (0, 0))
            total_limit = limit + reserved_amount
            remaining = total_limit - usage.usage
            
            result.append({
                "resource_type": usage.resource_type,
                "total_limit": total_limit,
                "total_used": usage.usage,
                "remaining": max(0, remaining),
                "reserved_amount": reserved_amount,
                "reserved_used": reserved_used
            })
        
        return {
            "user_id": user_id,
            "usage": result
        }
    
    def get_stats(self, start_time: datetime = None, end_time: datetime = None) -> dict:
        if end_time is None:
            end_time = datetime.now()
        if start_time is None:
            start_time = end_time - timedelta(days=1)
        
        logs = self.db.query(QuotaLog).filter(
            QuotaLog.timestamp >= start_time,
            QuotaLog.timestamp <= end_time
        ).all()
        
        total_created = sum(l.amount for l in logs if l.action == "consume" and l.result == "success")
        total_released = sum(l.amount for l in logs if l.action == "release" and l.result == "success")
        total_rejected = sum(l.amount for l in logs if l.action == "consume" and l.result == "rejected")
        
        return {
            "total_created": total_created,
            "total_released": total_released,
            "total_rejected": total_rejected,
            "start_time": start_time,
            "end_time": end_time
        }
    
    def get_resource_stats(self, start_time: datetime = None, end_time: datetime = None) -> dict:
        if end_time is None:
            end_time = datetime.now()
        if start_time is None:
            start_time = end_time - timedelta(days=1)
        
        logs = self.db.query(QuotaLog).filter(
            QuotaLog.timestamp >= start_time,
            QuotaLog.timestamp <= end_time
        ).all()
        
        resource_types = set(l.resource_type for l in logs)
        usages = self.db.query(QuotaUsage).all()
        
        usage_by_user = {}
        for u in usages:
            if u.resource_type not in usage_by_user:
                usage_by_user[u.resource_type] = {}
            usage_by_user[u.resource_type][u.user_id] = u.usage
        
        result = []
        for rt in resource_types:
            rt_logs = [l for l in logs if l.resource_type == rt]
            total_created = sum(l.amount for l in rt_logs if l.action == "consume" and l.result == "success")
            total_released = sum(l.amount for l in rt_logs if l.action == "release" and l.result == "success")
            total_rejected = sum(l.amount for l in rt_logs if l.action == "consume" and l.result == "rejected")
            
            users = usage_by_user.get(rt, {})
            sorted_users = sorted(users.items(), key=lambda x: x[1], reverse=True)
            top_users = [{"user_id": uid, "usage": count} for uid, count in sorted_users[:10]]
            
            result.append({
                "resource_type": rt,
                "total_created": total_created,
                "total_released": total_released,
                "total_rejected": total_rejected,
                "top_users": top_users
            })
        
        return {
            "resources": result,
            "start_time": start_time,
            "end_time": end_time
        }
