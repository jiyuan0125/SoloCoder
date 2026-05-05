from datetime import datetime

from shared.models import (
    ReturnOrder as SharedReturnOrder,
    ReturnStatus,
)


class ReturnOrder(SharedReturnOrder):
    def is_expired(self) -> bool:
        now = datetime.now()
        if self.status == ReturnStatus.APPLIED:
            return now > self.expiry_date
        if self.status == ReturnStatus.INSPECTED and self.stockin_deadline:
            return now > self.stockin_deadline
        return False
