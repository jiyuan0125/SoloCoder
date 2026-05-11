from datetime import datetime, timedelta
from typing import List, Optional, Tuple
from sqlalchemy.orm import Session
from sqlalchemy import and_, func

from .config import settings
from .models import (
    Station, StationType, RawData, DataQualityStatus, QualityRecord,
    AggregatedData, AggregationGranularity, AggregationMissingStatus,
    Calibration
)
from .schemas import (
    StationCreate, StationUpdate, RawDataCreate,
    CalibrationCreate, CalibrationUpdate, QualityRecordReview
)


class StationService:
    @staticmethod
    def create(db: Session, station: StationCreate) -> Station:
        db_station = Station(
            name=station.name,
            station_type=StationType(station.station_type.value),
            device_model=station.device_model
        )
        db.add(db_station)
        db.commit()
        db.refresh(db_station)
        return db_station

    @staticmethod
    def get_by_id(db: Session, station_id: int) -> Optional[Station]:
        return db.query(Station).filter(Station.id == station_id).first()

    @staticmethod
    def get_all(db: Session, skip: int = 0, limit: int = 100) -> Tuple[List[Station], int]:
        query = db.query(Station)
        total = query.count()
        stations = query.order_by(Station.id).offset(skip).limit(limit).all()
        return stations, total

    @staticmethod
    def update(db: Session, station_id: int, data: StationUpdate) -> Optional[Station]:
        station = StationService.get_by_id(db, station_id)
        if not station:
            return None

        update_data = data.model_dump(exclude_unset=True)
        if 'station_type' in update_data:
            update_data['station_type'] = StationType(update_data['station_type'].value)

        for key, value in update_data.items():
            setattr(station, key, value)

        db.commit()
        db.refresh(station)
        return station

    @staticmethod
    def delete(db: Session, station_id: int) -> bool:
        station = StationService.get_by_id(db, station_id)
        if not station:
            return False
        db.delete(station)
        db.commit()
        return True


class QualityService:
    @staticmethod
    def check_single_value(data_type: StationType, value: float) -> Tuple[bool, Optional[str]]:
        if data_type == StationType.RAINFALL:
            if value < settings.RAINFALL_MIN:
                return False, f"降雨值为负值: {value}mm"
            if value > settings.RAINFALL_MAX:
                return False, f"降雨值超过上限: {value}mm (最大{settings.RAINFALL_MAX}mm)"
        return True, None

    @staticmethod
    def check_water_level_change(db: Session, station_id: int, new_data: RawData) -> Tuple[bool, Optional[str]]:
        one_hour_ago = new_data.timestamp - timedelta(hours=1)

        previous = db.query(RawData).filter(
            and_(
                RawData.station_id == station_id,
                RawData.data_type == StationType.WATER_LEVEL,
                RawData.timestamp < new_data.timestamp,
                RawData.timestamp >= one_hour_ago,
                RawData.quality_status != DataQualityStatus.INVALID,
                RawData.id != new_data.id
            )
        ).order_by(RawData.timestamp.desc()).first()

        if previous:
            change = abs(new_data.value - previous.value)
            if change > settings.WATER_LEVEL_CHANGE_LIMIT:
                return False, f"水位1小时变化超过{settings.WATER_LEVEL_CHANGE_LIMIT}米: {change}米"

        return True, None

    @staticmethod
    def create_quality_record(
        db: Session,
        station_id: int,
        raw_data_id: int,
        issue_type: str,
        description: str
    ) -> QualityRecord:
        record = QualityRecord(
            station_id=station_id,
            raw_data_id=raw_data_id,
            issue_type=issue_type,
            description=description
        )
        db.add(record)
        db.commit()
        db.refresh(record)
        return record

    @staticmethod
    def validate_raw_data(
        db: Session,
        station_id: int,
        raw_data: RawData,
        check_water_level: bool = True
    ) -> List[QualityRecord]:
        records = []

        is_valid, reason = QualityService.check_single_value(
            raw_data.data_type, raw_data.value
        )
        if not is_valid:
            raw_data.quality_status = DataQualityStatus.SUSPICIOUS
            record = QualityService.create_quality_record(
                db, station_id, raw_data.id, "VALUE_OUT_OF_RANGE", reason
            )
            records.append(record)

        if check_water_level and raw_data.data_type == StationType.WATER_LEVEL:
            is_valid, reason = QualityService.check_water_level_change(db, station_id, raw_data)
            if not is_valid:
                if raw_data.quality_status == DataQualityStatus.VALID:
                    raw_data.quality_status = DataQualityStatus.SUSPICIOUS
                record = QualityService.create_quality_record(
                    db, station_id, raw_data.id, "EXCESSIVE_CHANGE", reason
                )
                records.append(record)

        if records:
            db.commit()

        return records

    @staticmethod
    def get_quality_records(
        db: Session,
        station_id: int,
        skip: int = 0,
        limit: int = 100
    ) -> Tuple[List[QualityRecord], int]:
        query = db.query(QualityRecord).filter(QualityRecord.station_id == station_id)
        total = query.count()
        records = query.order_by(QualityRecord.detected_at.desc()).offset(skip).limit(limit).all()
        return records, total

    @staticmethod
    def review_quality_record(
        db: Session,
        station_id: int,
        record_id: int,
        review_data: QualityRecordReview
    ) -> Optional[QualityRecord]:
        record = db.query(QualityRecord).filter(
            and_(
                QualityRecord.id == record_id,
                QualityRecord.station_id == station_id
            )
        ).first()

        if not record:
            return None

        record.reviewed = review_data.reviewed
        record.reviewer = review_data.reviewer
        record.review_notes = review_data.review_notes
        record.reviewed_at = datetime.utcnow()

        if record.raw_data_id:
            raw_data = db.query(RawData).filter(RawData.id == record.raw_data_id).first()
            if raw_data:
                if review_data.reviewed:
                    raw_data.quality_status = DataQualityStatus.INVALID
                else:
                    existing_suspicious = db.query(QualityRecord).filter(
                        and_(
                            QualityRecord.raw_data_id == raw_data.id,
                            QualityRecord.reviewed == False
                        )
                    ).first()
                    if not existing_suspicious:
                        raw_data.quality_status = DataQualityStatus.VALID

                db.commit()
                AggregationService.mark_for_recalculation(db, raw_data.station_id, raw_data.timestamp)

        db.commit()
        db.refresh(record)
        return record


class RawDataService:
    @staticmethod
    def create(
        db: Session,
        station_id: int,
        data: RawDataCreate,
        run_quality_check: bool = True
    ) -> RawData:
        raw_data = RawData(
            station_id=station_id,
            data_type=StationType(data.data_type.value),
            timestamp=data.timestamp,
            value=data.value,
            quality_status=DataQualityStatus.VALID
        )
        db.add(raw_data)
        db.commit()
        db.refresh(raw_data)

        if run_quality_check:
            QualityService.validate_raw_data(db, station_id, raw_data)

        AggregationService.mark_for_recalculation(db, station_id, raw_data.timestamp)

        return raw_data

    @staticmethod
    def create_bulk(
        db: Session,
        station_id: int,
        data_list: List[RawDataCreate],
        run_quality_check: bool = True
    ) -> List[RawData]:
        created = []
        for data in data_list:
            raw_data = RawData(
                station_id=station_id,
                data_type=StationType(data.data_type.value),
                timestamp=data.timestamp,
                value=data.value,
                quality_status=DataQualityStatus.VALID
            )
            db.add(raw_data)
            db.flush()
            created.append(raw_data)

        db.commit()

        if run_quality_check:
            for raw_data in created:
                QualityService.validate_raw_data(db, station_id, raw_data, check_water_level=False)

            for i, raw_data in enumerate(created):
                if raw_data.data_type == StationType.WATER_LEVEL:
                    for j in range(max(0, i - 6), i):
                        prev = created[j]
                        if prev.data_type == StationType.WATER_LEVEL:
                            time_diff = (raw_data.timestamp - prev.timestamp).total_seconds() / 3600
                            if 0 < time_diff <= 1:
                                change = abs(raw_data.value - prev.value)
                                if change > settings.WATER_LEVEL_CHANGE_LIMIT:
                                    raw_data.quality_status = DataQualityStatus.SUSPICIOUS
                                    QualityService.create_quality_record(
                                        db, station_id, raw_data.id, "EXCESSIVE_CHANGE",
                                        f"水位1小时变化超过{settings.WATER_LEVEL_CHANGE_LIMIT}米: {change}米"
                                    )
                                    break

        for raw_data in created:
            AggregationService.mark_for_recalculation(db, station_id, raw_data.timestamp)

        return created

    @staticmethod
    def get_by_station(
        db: Session,
        station_id: int,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
        skip: int = 0,
        limit: int = 100
    ) -> Tuple[List[RawData], int]:
        query = db.query(RawData).filter(RawData.station_id == station_id)

        if start_time:
            query = query.filter(RawData.timestamp >= start_time)
        if end_time:
            query = query.filter(RawData.timestamp <= end_time)

        total = query.count()
        data = query.order_by(RawData.timestamp.desc()).offset(skip).limit(limit).all()
        return data, total


class AggregationService:
    @staticmethod
    def get_hour_boundary(timestamp: datetime) -> Tuple[datetime, datetime]:
        period_start = timestamp.replace(minute=0, second=0, microsecond=0)
        period_end = period_start + timedelta(hours=1)
        return period_start, period_end

    @staticmethod
    def get_day_boundary(timestamp: datetime) -> Tuple[datetime, datetime]:
        period_start = timestamp.replace(hour=0, minute=0, second=0, microsecond=0)
        period_end = period_start + timedelta(days=1)
        return period_start, period_end

    @staticmethod
    def get_expected_samples_per_hour(station_type: StationType) -> int:
        if station_type in (StationType.WATER_LEVEL, StationType.RAINFALL):
            return 6
        elif station_type == StationType.FLOW:
            return 1
        return 0

    @staticmethod
    def aggregate_hourly_for_type(
        db: Session,
        station: Station,
        period_start: datetime,
        period_end: datetime,
        data_type: StationType
    ) -> Optional[AggregatedData]:
        samples = db.query(RawData).filter(
            and_(
                RawData.station_id == station.id,
                RawData.data_type == data_type,
                RawData.timestamp >= period_start,
                RawData.timestamp < period_end,
                RawData.quality_status != DataQualityStatus.INVALID
            )
        ).all()

        if not samples:
            return None

        expected_samples = AggregationService.get_expected_samples_per_hour(data_type)
        sample_count = len(samples)
        missing_ratio = 1.0 - (sample_count / expected_samples) if expected_samples > 0 else 0
        is_missing = missing_ratio > settings.AGGREGATION_MISSING_RATIO

        values = [s.value for s in samples]

        agg = AggregatedData(
            station_id=station.id,
            data_type=data_type,
            granularity=AggregationGranularity.HOURLY,
            period_start=period_start,
            period_end=period_end,
            average_value=sum(values) / len(values) if values else None,
            max_value=max(values) if values else None,
            min_value=min(values) if values else None,
            missing_status=AggregationMissingStatus.MISSING if is_missing else AggregationMissingStatus.COMPLETE,
            sample_count=sample_count,
            device_model=station.device_model
        )

        return agg

    @staticmethod
    def aggregate_daily_flow(
        db: Session,
        station: Station,
        period_start: datetime,
        period_end: datetime
    ) -> Optional[AggregatedData]:
        hourly_aggs = db.query(AggregatedData).filter(
            and_(
                AggregatedData.station_id == station.id,
                AggregatedData.data_type == StationType.FLOW,
                AggregatedData.granularity == AggregationGranularity.HOURLY,
                AggregatedData.period_start >= period_start,
                AggregatedData.period_start < period_end,
                AggregatedData.missing_status == AggregationMissingStatus.COMPLETE
            )
        ).all()

        if not hourly_aggs:
            return None

        expected_hours = 24
        valid_hours = len(hourly_aggs)
        missing_ratio = 1.0 - (valid_hours / expected_hours)
        is_missing = missing_ratio > settings.AGGREGATION_MISSING_RATIO

        values = [agg.average_value for agg in hourly_aggs if agg.average_value is not None]

        agg = AggregatedData(
            station_id=station.id,
            data_type=StationType.FLOW,
            granularity=AggregationGranularity.DAILY,
            period_start=period_start,
            period_end=period_end,
            average_value=sum(values) / len(values) if values else None,
            max_value=max(values) if values else None,
            min_value=None,
            missing_status=AggregationMissingStatus.MISSING if is_missing else AggregationMissingStatus.COMPLETE,
            sample_count=len(values),
            device_model=station.device_model
        )

        return agg

    @staticmethod
    def delete_existing_aggregations(
        db: Session,
        station_id: int,
        data_type: StationType,
        granularity: AggregationGranularity,
        period_start: datetime
    ):
        existing = db.query(AggregatedData).filter(
            and_(
                AggregatedData.station_id == station_id,
                AggregatedData.data_type == data_type,
                AggregatedData.granularity == granularity,
                AggregatedData.period_start == period_start
            )
        ).first()

        if existing:
            db.delete(existing)
            db.commit()

    @staticmethod
    def mark_for_recalculation(db: Session, station_id: int, timestamp: datetime):
        station = StationService.get_by_id(db, station_id)
        if not station:
            return

        hour_start, hour_end = AggregationService.get_hour_boundary(timestamp)

        AggregationService.delete_existing_aggregations(
            db, station_id, station.station_type, AggregationGranularity.HOURLY, hour_start
        )

        if station.station_type == StationType.FLOW:
            day_start, day_end = AggregationService.get_day_boundary(timestamp)
            AggregationService.delete_existing_aggregations(
                db, station_id, StationType.FLOW, AggregationGranularity.DAILY, day_start
            )

    @staticmethod
    def aggregate_for_period(
        db: Session,
        station: Station,
        start_time: datetime,
        end_time: datetime
    ) -> List[AggregatedData]:
        results = []

        if station.station_type in (StationType.WATER_LEVEL, StationType.RAINFALL):
            current = start_time.replace(minute=0, second=0, microsecond=0)
            while current < end_time:
                period_end = current + timedelta(hours=1)
                agg = AggregationService.aggregate_hourly_for_type(
                    db, station, current, period_end, station.station_type
                )
                if agg:
                    AggregationService.delete_existing_aggregations(
                        db, station.id, station.station_type, AggregationGranularity.HOURLY, current
                    )
                    db.add(agg)
                    results.append(agg)
                current = period_end

        elif station.station_type == StationType.FLOW:
            current = start_time.replace(hour=0, minute=0, second=0, microsecond=0)
            while current < end_time:
                period_end = current + timedelta(days=1)

                hour_ptr = current
                while hour_ptr < min(period_end, end_time):
                    hour_end = hour_ptr + timedelta(hours=1)
                    agg = AggregationService.aggregate_hourly_for_type(
                        db, station, hour_ptr, hour_end, StationType.FLOW
                    )
                    if agg:
                        AggregationService.delete_existing_aggregations(
                            db, station.id, StationType.FLOW, AggregationGranularity.HOURLY, hour_ptr
                        )
                        db.add(agg)
                    hour_ptr = hour_end

                daily_agg = AggregationService.aggregate_daily_flow(db, station, current, period_end)
                if daily_agg:
                    AggregationService.delete_existing_aggregations(
                        db, station.id, StationType.FLOW, AggregationGranularity.DAILY, current
                    )
                    db.add(daily_agg)
                    results.append(daily_agg)

                current = period_end

        db.commit()
        return results

    @staticmethod
    def get_aggregated_data(
        db: Session,
        station_id: int,
        granularity: Optional[AggregationGranularity] = None,
        start_time: Optional[datetime] = None,
        end_time: Optional[datetime] = None,
        skip: int = 0,
        limit: int = 100
    ) -> Tuple[List[AggregatedData], int]:
        query = db.query(AggregatedData).filter(AggregatedData.station_id == station_id)

        if granularity:
            query = query.filter(AggregatedData.granularity == granularity)
        if start_time:
            query = query.filter(AggregatedData.period_start >= start_time)
        if end_time:
            query = query.filter(AggregatedData.period_end <= end_time)

        total = query.count()
        data = query.order_by(AggregatedData.period_start.desc()).offset(skip).limit(limit).all()
        return data, total


class CalibrationService:
    @staticmethod
    def create(db: Session, station_id: int, data: CalibrationCreate) -> Calibration:
        calibration = Calibration(
            station_id=station_id,
            last_calibration_date=data.last_calibration_date,
            next_calibration_date=data.next_calibration_date,
            is_done=data.is_done
        )
        db.add(calibration)
        db.commit()
        db.refresh(calibration)
        return calibration

    @staticmethod
    def get_by_station(db: Session, station_id: int, skip: int = 0, limit: int = 100) -> Tuple[List[Calibration], int]:
        query = db.query(Calibration).filter(Calibration.station_id == station_id)
        total = query.count()
        calibrations = query.order_by(Calibration.next_calibration_date.desc()).offset(skip).limit(limit).all()
        return calibrations, total

    @staticmethod
    def update(
        db: Session,
        station_id: int,
        calibration_id: int,
        data: CalibrationUpdate
    ) -> Optional[Calibration]:
        calibration = db.query(Calibration).filter(
            and_(
                Calibration.id == calibration_id,
                Calibration.station_id == station_id
            )
        ).first()

        if not calibration:
            return None

        update_data = data.model_dump(exclude_unset=True)
        for key, value in update_data.items():
            setattr(calibration, key, value)

        db.commit()
        db.refresh(calibration)
        return calibration

    @staticmethod
    def get_pending_todos(db: Session) -> List[dict]:
        now = datetime.utcnow()
        warning_date = now + timedelta(days=settings.CALIBRATION_WARNING_DAYS)

        stations = db.query(Station).all()
        results = []

        for station in stations:
            latest = db.query(Calibration).filter(
                Calibration.station_id == station.id
            ).order_by(Calibration.next_calibration_date.desc()).first()

            if latest and not latest.is_done and latest.next_calibration_date <= warning_date:
                days_remaining = (latest.next_calibration_date - now).days
                results.append({
                    'station_id': station.id,
                    'station_name': station.name,
                    'next_calibration_date': latest.next_calibration_date,
                    'days_remaining': days_remaining
                })

        return results
