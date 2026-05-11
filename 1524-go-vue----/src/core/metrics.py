from datetime import datetime, date
from typing import List
from .models import Metrics, BatchingOrder, BatchingOrderStatus
from .storage import storage


class MetricsService:
    @staticmethod
    def get_daily_metrics(target_date: date = None) -> Metrics:
        if target_date is None:
            target_date = date.today()
        
        all_orders = storage.get_all_batching_orders()
        
        completed_orders = [
            o for o in all_orders
            if o.status == BatchingOrderStatus.COMPLETED
            and o.completed_at
            and o.completed_at.date() == target_date
        ]
        
        total_output = sum(
            o.output_weight for o in completed_orders
            if o.output_weight is not None
        )
        
        total_energy = sum(
            o.energy_consumed for o in completed_orders
            if o.energy_consumed is not None
        )
        
        active_furnaces = set()
        for order in completed_orders:
            if order.furnace_id:
                active_furnaces.add(order.furnace_id)
        
        average_output = 0.0
        if active_furnaces:
            average_output = total_output / len(active_furnaces)
        
        temperature_exceed_count = 0
        for furnace_id, readings in storage.temperature_readings.items():
            for reading in readings:
                if (reading.is_alert 
                    and reading.timestamp.date() == target_date):
                    temperature_exceed_count += 1
        
        return Metrics(
            date=target_date.isoformat(),
            total_output=round(total_output, 2),
            average_output_per_furnace=round(average_output, 2),
            total_energy_consumed=round(total_energy, 2),
            temperature_exceed_count=temperature_exceed_count,
            completed_orders=len(completed_orders)
        )


metrics_service = MetricsService()
