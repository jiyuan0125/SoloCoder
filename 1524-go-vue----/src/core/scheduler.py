from typing import List, Optional
from .services import BatchingService, FurnaceService, ServiceException
from .models import BatchingOrder, Furnace


class Scheduler:
    @staticmethod
    def schedule() -> List[BatchingOrder]:
        assigned_orders = []
        
        idle_furnaces = FurnaceService.get_idle_furnaces()
        pending_orders = BatchingService.get_pending_orders()
        
        for furnace in idle_furnaces:
            if not pending_orders:
                break
            
            for i, order in enumerate(pending_orders):
                if BatchingService.check_capacity(order, furnace):
                    try:
                        assigned_order = BatchingService.assign_order_to_furnace(
                            order.id, furnace.id
                        )
                        assigned_orders.append(assigned_order)
                        pending_orders.pop(i)
                        break
                    except ServiceException:
                        continue
        
        return assigned_orders
    
    @staticmethod
    def get_pending_orders() -> List[BatchingOrder]:
        return BatchingService.get_pending_orders()
    
    @staticmethod
    def get_idle_furnaces() -> List[Furnace]:
        return FurnaceService.get_idle_furnaces()


scheduler = Scheduler()
