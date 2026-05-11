use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use chrono::{Utc, Duration, NaiveDate, Local};
use uuid::Uuid;

use crate::models::*;
use crate::error::*;

const MAX_FAILED_ATTEMPTS: u32 = 5;
const LOCK_DURATION_MINUTES: i64 = 30;
const PROXY_FEEDBACK_HOURS: i64 = 2;
const STRANDED_DAYS: i64 = 3;
const RETURN_DAYS: i64 = 7;

#[derive(Clone)]
pub struct ExpressService {
    inner: Arc<RwLock<ExpressStore>>,
}

struct ExpressStore {
    packages: HashMap<Uuid, Package>,
    packages_by_code: HashMap<String, Uuid>,
    packages_by_phone: HashMap<String, Vec<Uuid>>,
    daily_counters: HashMap<NaiveDate, DailyCodeCounter>,
    proxy_records: Vec<ProxyRecord>,
    notifications: Vec<Notification>,
}

impl ExpressService {
    pub fn new() -> Self {
        Self {
            inner: Arc::new(RwLock::new(ExpressStore {
                packages: HashMap::new(),
                packages_by_code: HashMap::new(),
                packages_by_phone: HashMap::new(),
                daily_counters: HashMap::new(),
                proxy_records: Vec::new(),
                notifications: Vec::new(),
            })),
        }
    }

    pub async fn inbound_package(&self, req: InboundRequest) -> Result<InboundResult, ExpressError> {
        self.validate_inbound(&req)?;

        let today = Local::now().date_naive();
        let pickup_code = self.generate_pickup_code(today).await;

        let package = Package {
            id: Uuid::new_v4(),
            tracking_number: req.tracking_number,
            recipient_phone: req.recipient_phone.clone(),
            recipient_name: req.recipient_name,
            pickup_code: pickup_code.clone(),
            courier_name: req.courier_name,
            courier_phone: req.courier_phone,
            express_company: req.express_company,
            status: PackageStatus::InStock,
            created_at: Utc::now(),
            picked_up_at: None,
            is_proxy: false,
            proxy_phone: None,
            proxy_name: None,
            id_last_four: None,
            failed_attempts: 0,
            locked_until: None,
            stranded_reminded: false,
            return_notified: false,
        };

        let phone = package.recipient_phone.clone();
        let code = package.pickup_code.clone();
        let id = package.id;

        {
            let mut store = self.inner.write().await;
            store.packages.insert(id, package.clone());
            store.packages_by_code.insert(code, id);
            store.packages_by_phone.entry(phone).or_default().push(id);
        }

        let notification = self.create_arrival_notification(&package).await;

        if let Some(notif) = notification.clone() {
            let mut store = self.inner.write().await;
            store.notifications.push(notif);
        }

        Ok(InboundResult {
            package,
            notification,
        })
    }

    async fn generate_pickup_code(&self, date: NaiveDate) -> String {
        let mut store = self.inner.write().await;
        
        let counter = store.daily_counters.entry(date).or_insert_with(|| DailyCodeCounter {
            date,
            counter: 0,
            letter_prefix: ' ',
        });

        counter.counter += 1;
        let mut code = counter.counter;
        let mut prefix = ' ';

        if code > 999 {
            let letter_offset = ((code - 1) / 999) as u8;
            prefix = (b'A' + letter_offset) as char;
            code = ((code - 1) % 999) + 1;
        }

        if prefix == ' ' {
            format!("{:06}", code)
        } else {
            format!("{}{:05}", prefix, code)
        }
    }

    async fn create_arrival_notification(&self, package: &Package) -> Option<Notification> {
        let today_start = Local::now().date_naive().and_hms_opt(0, 0, 0)?.and_utc();
        let now = Utc::now();
        
        let store = self.inner.read().await;
        
        let same_day_codes: Vec<String> = store.packages.values()
            .filter(|p| {
                p.recipient_phone == package.recipient_phone
                    && p.created_at >= today_start
                    && p.created_at <= now
                    && p.status == PackageStatus::InStock
                    && p.pickup_code != package.pickup_code
            })
            .map(|p| p.pickup_code.clone())
            .collect();

        let mut all_codes = same_day_codes;
        all_codes.push(package.pickup_code.clone());

        if all_codes.len() > 1 {
            let message = format!(
                "【快递驿站】您有{}个包裹已到达，取件码：{}",
                all_codes.len(),
                all_codes.join("、")
            );
            Some(Notification {
                id: Uuid::new_v4(),
                package_id: package.id,
                recipient_phone: package.recipient_phone.clone(),
                pickup_codes: all_codes,
                notification_type: NotificationType::Arrival,
                sent_at: now,
                message,
            })
        } else {
            let message = format!(
                "【快递驿站】您的包裹已到达，取件码：{}",
                package.pickup_code
            );
            Some(Notification {
                id: Uuid::new_v4(),
                package_id: package.id,
                recipient_phone: package.recipient_phone.clone(),
                pickup_codes: vec![package.pickup_code.clone()],
                notification_type: NotificationType::Arrival,
                sent_at: now,
                message,
            })
        }
    }

    pub async fn pickup_package(&self, req: PickupRequest) -> Result<PickupResult, ExpressError> {
        let mut store = self.inner.write().await;

        let package_id = *store.packages_by_code
            .get(&req.pickup_code)
            .ok_or(ExpressError::PackageNotFound)?;

        let package = store.packages.get_mut(&package_id)
            .ok_or(ExpressError::PackageNotFound)?;

        if package.status == PackageStatus::PickedUp {
            return Err(ExpressError::PackageAlreadyPickedUp);
        }

        let now = Utc::now();

        if let Some(locked_until) = package.locked_until {
            if now < locked_until {
                return Err(ExpressError::PackageLocked);
            } else {
                package.locked_until = None;
                package.failed_attempts = 0;
            }
        }

        package.failed_attempts += 1;

        if package.failed_attempts >= MAX_FAILED_ATTEMPTS {
            package.locked_until = Some(now + Duration::minutes(LOCK_DURATION_MINUTES));
            return Err(ExpressError::PackageLocked);
        }

        package.status = PackageStatus::PickedUp;
        package.picked_up_at = Some(now);
        package.failed_attempts = 0;

        Ok(PickupResult {
            package: package.clone(),
            success: true,
            message: "取件成功".to_string(),
            locked: false,
            locked_until: None,
        })
    }

    pub async fn manual_verify_pickup(&self, req: ManualVerifyRequest) -> Result<PickupResult, ExpressError> {
        let mut store = self.inner.write().await;

        let package = store.packages.get_mut(&req.package_id)
            .ok_or(ExpressError::PackageNotFound)?;

        if package.status == PackageStatus::PickedUp {
            return Err(ExpressError::PackageAlreadyPickedUp);
        }

        let phone_matched = package.recipient_phone == req.phone
            || package.proxy_phone.as_ref().map_or(false, |p| p == &req.phone);

        let id_matched = package.id_last_four.as_ref()
            .map_or(false, |id| id == &req.id_last_four);

        if !phone_matched || !id_matched {
            return Err(ExpressError::VerificationFailed);
        }

        let now = Utc::now();
        package.status = PackageStatus::PickedUp;
        package.picked_up_at = Some(now);
        package.failed_attempts = 0;
        package.locked_until = None;

        Ok(PickupResult {
            package: package.clone(),
            success: true,
            message: "人工核验取件成功".to_string(),
            locked: false,
            locked_until: None,
        })
    }

    pub async fn register_proxy(&self, req: ProxyRequest) -> Result<(ProxyRecord, Notification), ExpressError> {
        let mut store = self.inner.write().await;

        let package = store.packages.get_mut(&req.package_id)
            .ok_or(ExpressError::PackageNotFound)?;

        if package.is_proxy {
            return Err(ExpressError::AlreadyProxy);
        }

        if package.status == PackageStatus::PickedUp {
            return Err(ExpressError::PackageAlreadyPickedUp);
        }

        if !self.is_valid_phone(&req.proxy_phone) {
            return Err(ExpressError::InvalidPhone);
        }

        let now = Utc::now();
        let proxy_phone = req.proxy_phone.clone();
        let proxy_name = req.proxy_name.clone();
        let recipient_phone = package.recipient_phone.clone();
        let pickup_code = package.pickup_code.clone();

        let proxy_record = ProxyRecord {
            id: Uuid::new_v4(),
            package_id: req.package_id,
            proxy_phone: proxy_phone.clone(),
            proxy_name: proxy_name.clone(),
            recipient_phone: recipient_phone.clone(),
            notified_at: now,
            feedback: None,
            feedback_at: None,
        };

        package.is_proxy = true;
        package.proxy_phone = Some(proxy_phone.clone());
        package.proxy_name = proxy_name;

        let notification = Notification {
            id: Uuid::new_v4(),
            package_id: req.package_id,
            recipient_phone: recipient_phone.clone(),
            pickup_codes: vec![pickup_code.clone()],
            notification_type: NotificationType::ProxyRequested,
            sent_at: now,
            message: format!(
                "【快递驿站】您的包裹（取件码：{}）有人申请代收，代收人手机号：{}。请在{}小时内反馈未委托，否则视为同意。",
                pickup_code, proxy_phone, PROXY_FEEDBACK_HOURS
            ),
        };

        store.proxy_records.push(proxy_record.clone());
        store.notifications.push(notification.clone());

        Ok((proxy_record, notification))
    }

    pub async fn feedback_proxy(&self, req: ProxyFeedbackRequest) -> Result<(ProxyRecord, Option<Notification>), ExpressError> {
        let now = Utc::now();
        let (proxy_record_index, mut proxy_record) = {
            let store = self.inner.read().await;
            let (idx, record) = store.proxy_records.iter().enumerate()
                .find(|(_, r)| r.package_id == req.package_id && r.feedback.is_none())
                .ok_or(ExpressError::PackageNotFound)?;
            (idx, record.clone())
        };

        let mut store = self.inner.write().await;

        let package = store.packages.get_mut(&req.package_id)
            .ok_or(ExpressError::PackageNotFound)?;

        proxy_record.feedback = Some(req.feedback.clone());
        proxy_record.feedback_at = Some(now);

        let pickup_code = package.pickup_code.clone();
        let recipient_phone = package.recipient_phone.clone();

        let notification = match req.feedback {
            ProxyFeedback::Rejected => {
                package.status = PackageStatus::Abnormal;
                Some(Notification {
                    id: Uuid::new_v4(),
                    package_id: req.package_id,
                    recipient_phone: recipient_phone.clone(),
                    pickup_codes: vec![pickup_code.clone()],
                    notification_type: NotificationType::ProxyRejected,
                    sent_at: now,
                    message: format!(
                        "【快递驿站】您的包裹（取件码：{}）代收已被拒绝，包裹已标记异常。",
                        pickup_code
                    ),
                })
            }
            ProxyFeedback::Confirmed => {
                Some(Notification {
                    id: Uuid::new_v4(),
                    package_id: req.package_id,
                    recipient_phone: recipient_phone.clone(),
                    pickup_codes: vec![pickup_code.clone()],
                    notification_type: NotificationType::ProxyConfirmed,
                    sent_at: now,
                    message: format!(
                        "【快递驿站】您已确认代收，包裹（取件码：{}）可由代收人领取。",
                        pickup_code
                    ),
                })
            }
        };

        store.proxy_records[proxy_record_index] = proxy_record.clone();
        if let Some(notif) = notification.clone() {
            store.notifications.push(notif);
        }

        Ok((proxy_record, notification))
    }

    pub async fn check_expired_proxies(&self) -> Vec<Package> {
        let now = Utc::now();
        let threshold = now - Duration::hours(PROXY_FEEDBACK_HOURS);
        
        let to_mark: Vec<Uuid> = {
            let store = self.inner.read().await;
            store.proxy_records.iter()
                .filter(|r| r.feedback.is_none() && r.notified_at < threshold)
                .map(|r| r.package_id)
                .collect()
        };

        let mut store = self.inner.write().await;
        let mut abnormal_packages = Vec::new();

        for package_id in to_mark {
            if let Some(package) = store.packages.get_mut(&package_id) {
                if package.status == PackageStatus::InStock {
                    package.status = PackageStatus::Abnormal;
                    abnormal_packages.push(package.clone());
                }
            }
        }

        abnormal_packages
    }

    pub async fn check_stranded_packages(&self) -> (Vec<Package>, Vec<Package>) {
        let now = Utc::now();
        let stranded_threshold = now - Duration::days(STRANDED_DAYS);
        let return_threshold = now - Duration::days(RETURN_DAYS);

        let package_ids: Vec<Uuid> = {
            let store = self.inner.read().await;
            store.packages.keys().cloned().collect()
        };

        let mut store = self.inner.write().await;
        let mut stranded = Vec::new();
        let mut to_return = Vec::new();

        for id in package_ids {
            let package = match store.packages.get_mut(&id) {
                Some(p) if p.status == PackageStatus::InStock => p,
                _ => continue,
            };

            if package.created_at < return_threshold && !package.return_notified {
                package.status = PackageStatus::Returned;
                package.return_notified = true;
                
                let notif = Notification {
                    id: Uuid::new_v4(),
                    package_id: id,
                    recipient_phone: package.recipient_phone.clone(),
                    pickup_codes: vec![package.pickup_code.clone()],
                    notification_type: NotificationType::Return,
                    sent_at: now,
                    message: format!(
                        "【快递驿站】您的包裹（取件码：{}）已超过{}天未取，已通知快递公司退回。",
                        package.pickup_code, RETURN_DAYS
                    ),
                };
                
                to_return.push(package.clone());
                store.notifications.push(notif);
            } else if package.created_at < stranded_threshold && !package.stranded_reminded {
                package.status = PackageStatus::Stranded;
                package.stranded_reminded = true;
                
                let notif = Notification {
                    id: Uuid::new_v4(),
                    package_id: id,
                    recipient_phone: package.recipient_phone.clone(),
                    pickup_codes: vec![package.pickup_code.clone()],
                    notification_type: NotificationType::Stranded,
                    sent_at: now,
                    message: format!(
                        "【快递驿站】您的包裹（取件码：{}）已超过{}天未取，请尽快领取。",
                        package.pickup_code, STRANDED_DAYS
                    ),
                };
                
                stranded.push(package.clone());
                store.notifications.push(notif);
            }
        }

        (stranded, to_return)
    }

    pub async fn get_package(&self, id: Uuid) -> Option<Package> {
        let store = self.inner.read().await;
        store.packages.get(&id).cloned()
    }

    pub async fn get_package_by_code(&self, code: &str) -> Option<Package> {
        let store = self.inner.read().await;
        store.packages_by_code.get(code)
            .and_then(|id| store.packages.get(id))
            .cloned()
    }

    pub async fn get_packages_by_phone(&self, phone: &str) -> Vec<Package> {
        let store = self.inner.read().await;
        store.packages_by_phone.get(phone)
            .map(|ids| ids.iter()
                .filter_map(|id| store.packages.get(id))
                .cloned()
                .collect())
            .unwrap_or_default()
    }

    pub async fn get_all_packages(&self) -> Vec<Package> {
        let store = self.inner.read().await;
        store.packages.values().cloned().collect()
    }

    pub async fn get_proxy_records(&self) -> Vec<ProxyRecord> {
        let store = self.inner.read().await;
        store.proxy_records.clone()
    }

    pub async fn get_notifications(&self) -> Vec<Notification> {
        let store = self.inner.read().await;
        store.notifications.clone()
    }

    fn validate_inbound(&self, req: &InboundRequest) -> Result<(), ExpressError> {
        if req.tracking_number.is_empty() {
            return Err(ExpressError::InvalidTrackingNumber);
        }
        if !self.is_valid_phone(&req.recipient_phone) {
            return Err(ExpressError::InvalidPhone);
        }
        Ok(())
    }

    fn is_valid_phone(&self, phone: &str) -> bool {
        phone.len() == 11 && phone.chars().all(|c| c.is_ascii_digit())
    }
}

impl Default for ExpressService {
    fn default() -> Self {
        Self::new()
    }
}
