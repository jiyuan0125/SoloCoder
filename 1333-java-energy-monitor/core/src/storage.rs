use crate::models::{PointId, UsageRecord, WorkDayFlags};
use chrono::NaiveDate;
use std::collections::{HashMap, BTreeMap};
use std::sync::{Arc, RwLock};

#[derive(Debug, Default)]
pub struct InMemoryStorage {
    records: RwLock<HashMap<PointId, BTreeMap<NaiveDate, UsageRecord>>>,
    work_days: RwLock<WorkDayFlags>,
}

impl InMemoryStorage {
    pub fn new() -> Arc<Self> {
        Arc::new(Self {
            records: RwLock::new(HashMap::new()),
            work_days: RwLock::new(WorkDayFlags::new()),
        })
    }

    pub fn upsert_record(&self, record: UsageRecord) {
        let mut records = self.records.write().unwrap();
        let point_records = records.entry(record.point_id.clone()).or_insert_with(BTreeMap::new);
        point_records.insert(record.date, record);
    }

    pub fn get_record(&self, point_id: &PointId, date: &NaiveDate) -> Option<UsageRecord> {
        let records = self.records.read().unwrap();
        records
            .get(point_id)
            .and_then(|pr| pr.get(date))
            .cloned()
    }

    pub fn get_latest_usage_before(
        &self,
        point_id: &PointId,
        date: &NaiveDate,
    ) -> Option<(NaiveDate, UsageRecord)> {
        let records = self.records.read().unwrap();
        let point_records = records.get(point_id)?;

        point_records
            .range(..*date)
            .next_back()
            .map(|(d, r)| (*d, r.clone()))
    }

    pub fn get_all_records_for_point(&self, point_id: &PointId) -> Vec<UsageRecord> {
        let records = self.records.read().unwrap();
        records
            .get(point_id)
            .map(|pr| pr.values().cloned().collect())
            .unwrap_or_default()
    }

    pub fn get_all_points(&self) -> Vec<PointId> {
        let records = self.records.read().unwrap();
        records.keys().cloned().collect()
    }

    pub fn mark_shutdown(&self, date: NaiveDate) {
        let mut work_days = self.work_days.write().unwrap();
        work_days.mark_shutdown(date);
    }

    pub fn unmark_shutdown(&self, date: NaiveDate) {
        let mut work_days = self.work_days.write().unwrap();
        work_days.unmark_shutdown(date);
    }

    pub fn is_shutdown_day(&self, date: &NaiveDate) -> bool {
        let work_days = self.work_days.read().unwrap();
        work_days.is_shutdown_day(date)
    }

    pub fn should_skip_alert(&self, date: &NaiveDate) -> bool {
        let work_days = self.work_days.read().unwrap();
        work_days.should_skip_alert(date)
    }
}
