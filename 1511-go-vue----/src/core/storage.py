from typing import Dict, List, Optional
from datetime import date, datetime

from .models import (
    Plot, PlotCreate,
    HarvestPlan, HarvestPlanCreate,
    HarvestRecord, HarvestRecordCreate,
    ProcessingBatch, ProcessingBatchCreate, ProcessingBatchUpdate,
    QualityEvaluation, QualityEvaluationCreate,
    Alert, AlertType
)


class Storage:
    def __init__(self):
        self._plots: Dict[int, Plot] = {}
        self._harvest_plans: Dict[int, HarvestPlan] = {}
        self._harvest_records: Dict[int, HarvestRecord] = {}
        self._processing_batches: Dict[int, ProcessingBatch] = {}
        self._quality_evaluations: Dict[int, QualityEvaluation] = {}
        self._alerts: Dict[int, Alert] = {}
        
        self._plot_id_counter = 1
        self._harvest_plan_id_counter = 1
        self._harvest_record_id_counter = 1
        self._processing_batch_id_counter = 1
        self._quality_evaluation_id_counter = 1
        self._alert_id_counter = 1
    
    def create_plot(self, plot_create: PlotCreate) -> Plot:
        plot = Plot(
            id=self._plot_id_counter,
            **plot_create.model_dump()
        )
        self._plots[plot.id] = plot
        self._plot_id_counter += 1
        return plot
    
    def get_plot(self, plot_id: int) -> Optional[Plot]:
        return self._plots.get(plot_id)
    
    def list_plots(self) -> List[Plot]:
        return list(self._plots.values())
    
    def create_harvest_plan(self, plan_create: HarvestPlanCreate) -> HarvestPlan:
        plan = HarvestPlan(
            id=self._harvest_plan_id_counter,
            **plan_create.model_dump()
        )
        self._harvest_plans[plan.id] = plan
        self._harvest_plan_id_counter += 1
        return plan
    
    def get_harvest_plan(self, plan_id: int) -> Optional[HarvestPlan]:
        return self._harvest_plans.get(plan_id)
    
    def list_harvest_plans(self, plot_id: Optional[int] = None, 
                          plan_date: Optional[date] = None) -> List[HarvestPlan]:
        plans = list(self._harvest_plans.values())
        if plot_id:
            plans = [p for p in plans if p.plot_id == plot_id]
        if plan_date:
            plans = [p for p in plans if p.plan_date == plan_date]
        return plans
    
    def update_harvest_plan(self, plan_id: int, is_completed: bool) -> Optional[HarvestPlan]:
        if plan_id in self._harvest_plans:
            plan = self._harvest_plans[plan_id]
            plan.is_completed = is_completed
            return plan
        return None
    
    def create_harvest_record(self, record_create: HarvestRecordCreate) -> HarvestRecord:
        record = HarvestRecord(
            id=self._harvest_record_id_counter,
            **record_create.model_dump()
        )
        self._harvest_records[record.id] = record
        self._harvest_record_id_counter += 1
        
        if record.plan_id in self._harvest_plans:
            self._harvest_plans[record.plan_id].is_completed = True
        
        return record
    
    def get_harvest_record(self, record_id: int) -> Optional[HarvestRecord]:
        return self._harvest_records.get(record_id)
    
    def list_harvest_records(self, plan_id: Optional[int] = None) -> List[HarvestRecord]:
        records = list(self._harvest_records.values())
        if plan_id:
            records = [r for r in records if r.plan_id == plan_id]
        return records
    
    def has_processing_batch_for_harvest(self, harvest_record_id: int) -> bool:
        return any(
            batch.harvest_record_id == harvest_record_id
            for batch in self._processing_batches.values()
        )
    
    def create_processing_batch(self, batch_create: ProcessingBatchCreate) -> ProcessingBatch:
        batch = ProcessingBatch(
            id=self._processing_batch_id_counter,
            **batch_create.model_dump()
        )
        self._processing_batches[batch.id] = batch
        self._processing_batch_id_counter += 1
        return batch
    
    def get_processing_batch(self, batch_id: int) -> Optional[ProcessingBatch]:
        return self._processing_batches.get(batch_id)
    
    def list_processing_batches(self, harvest_record_id: Optional[int] = None) -> List[ProcessingBatch]:
        batches = list(self._processing_batches.values())
        if harvest_record_id:
            batches = [b for b in batches if b.harvest_record_id == harvest_record_id]
        return batches
    
    def update_processing_batch(self, batch_id: int, 
                               batch_update: ProcessingBatchUpdate) -> Optional[ProcessingBatch]:
        if batch_id in self._processing_batches:
            batch = self._processing_batches[batch_id]
            
            if batch_update.output_quantity > batch.input_quantity * 0.4:
                raise ValueError("成品量不能超过投叶量的40%")
            
            batch.output_quantity = batch_update.output_quantity
            batch.end_time = batch_update.end_time
            batch.notes = batch_update.notes
            batch.is_completed = True
            return batch
        return None
    
    def create_quality_evaluation(self, eval_create: QualityEvaluationCreate) -> QualityEvaluation:
        if not eval_create.sensory_description or not eval_create.sensory_description.strip():
            raise ValueError("感官描述不能为空")
        
        evaluation = QualityEvaluation(
            id=self._quality_evaluation_id_counter,
            **eval_create.model_dump()
        )
        self._quality_evaluations[evaluation.id] = evaluation
        self._quality_evaluation_id_counter += 1
        return evaluation
    
    def get_quality_evaluation(self, eval_id: int) -> Optional[QualityEvaluation]:
        return self._quality_evaluations.get(eval_id)
    
    def list_quality_evaluations(self, batch_id: Optional[int] = None) -> List[QualityEvaluation]:
        evaluations = list(self._quality_evaluations.values())
        if batch_id:
            evaluations = [e for e in evaluations if e.batch_id == batch_id]
        return evaluations
    
    def create_alert(self, alert_type: AlertType, message: str,
                    related_entity_id: Optional[int] = None,
                    related_entity_type: Optional[str] = None) -> Alert:
        alert = Alert(
            id=self._alert_id_counter,
            type=alert_type,
            message=message,
            related_entity_id=related_entity_id,
            related_entity_type=related_entity_type
        )
        self._alerts[alert.id] = alert
        self._alert_id_counter += 1
        return alert
    
    def list_alerts(self, alert_type: Optional[AlertType] = None, 
                   is_read: Optional[bool] = None) -> List[Alert]:
        alerts = list(self._alerts.values())
        if alert_type:
            alerts = [a for a in alerts if a.type == alert_type]
        if is_read is not None:
            alerts = [a for a in alerts if a.is_read == is_read]
        return alerts
    
    def mark_alert_read(self, alert_id: int) -> Optional[Alert]:
        if alert_id in self._alerts:
            self._alerts[alert_id].is_read = True
            return self._alerts[alert_id]
        return None
    
    def get_harvest_records_without_processing(self, since: datetime) -> List[HarvestRecord]:
        records = []
        for record in self._harvest_records.values():
            if record.harvest_time <= since:
                if not self.has_processing_batch_for_harvest(record.id):
                    records.append(record)
        return records
    
    def get_harvest_plans_for_date(self, target_date: date) -> List[HarvestPlan]:
        return [
            plan for plan in self._harvest_plans.values()
            if plan.plan_date == target_date and not plan.is_completed
        ]


storage = Storage()
