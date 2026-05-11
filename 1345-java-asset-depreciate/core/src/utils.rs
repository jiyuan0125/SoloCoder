use chrono::Local;

pub fn today() -> chrono::NaiveDate {
    Local::now().date_naive()
}
