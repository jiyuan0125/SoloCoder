use chrono::{NaiveDate, Datelike};
use crate::error::{AppError, AppResult};

pub fn parse_date(date_str: &str) -> AppResult<NaiveDate> {
    NaiveDate::parse_from_str(date_str, "%Y-%m-%d")
        .map_err(|e| AppError::InvalidDate(format!("Invalid date format '{}': {}", date_str, e)))
}

pub fn add_days(date: NaiveDate, days: u32) -> NaiveDate {
    date + chrono::Duration::days(days as i64)
}

pub fn date_overlaps(start1: NaiveDate, end1: NaiveDate, start2: NaiveDate, end2: NaiveDate) -> bool {
    !(end1 < start2 || end2 < start1)
}

pub fn format_month(month: u32) -> String {
    let names = [
        "一月", "二月", "三月", "四月", "五月", "六月",
        "七月", "八月", "九月", "十月", "十一月", "十二月"
    ];
    if month >= 1 && month <= 12 {
        names[(month - 1) as usize].to_string()
    } else {
        format!("{}月", month)
    }
}

pub fn get_month(date: NaiveDate) -> u32 {
    date.month()
}
