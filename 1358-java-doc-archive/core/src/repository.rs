use std::collections::HashMap;
use std::sync::Mutex;

use crate::models::*;

pub trait Repository: Send + Sync {
    fn get_archive(&self, id: &str) -> Option<Archive>;
    fn get_all_archives(&self) -> Vec<Archive>;
    fn save_archive(&self, archive: Archive);
    fn update_archive(&self, archive: Archive) -> Result<(), ()>;

    fn get_user(&self, id: &str) -> Option<User>;
    fn get_all_users(&self) -> Vec<User>;
    fn save_user(&self, user: User);
    fn update_user(&self, user: User) -> Result<(), ()>;

    fn get_borrow(&self, id: &str) -> Option<Borrow>;
    fn get_all_borrows(&self) -> Vec<Borrow>;
    fn get_active_borrows_by_user(&self, user_id: &str) -> Vec<Borrow>;
    fn get_overdue_borrows(&self) -> Vec<Borrow>;
    fn save_borrow(&self, borrow: Borrow);
    fn update_borrow(&self, borrow: Borrow) -> Result<(), ()>;

    fn get_reservation(&self, id: &str) -> Option<Reservation>;
    fn get_all_reservations(&self) -> Vec<Reservation>;
    fn get_reservations_by_archive(&self, archive_id: &str) -> Vec<Reservation>;
    fn get_next_pending_reservation(&self, archive_id: &str) -> Option<Reservation>;
    fn save_reservation(&self, reservation: Reservation);
    fn update_reservation(&self, reservation: Reservation) -> Result<(), ()>;
    fn remove_reservation(&self, id: &str);

    fn get_approval_process(&self, id: &str) -> Option<ApprovalProcess>;
    fn save_approval_process(&self, process: ApprovalProcess);
    fn update_approval_process(&self, process: ApprovalProcess) -> Result<(), ()>;
}

pub struct InMemoryRepository {
    archives: Mutex<HashMap<String, Archive>>,
    users: Mutex<HashMap<String, User>>,
    borrows: Mutex<HashMap<String, Borrow>>,
    reservations: Mutex<HashMap<String, Reservation>>,
    approvals: Mutex<HashMap<String, ApprovalProcess>>,
}

impl InMemoryRepository {
    pub fn new() -> Self {
        Self {
            archives: Mutex::new(HashMap::new()),
            users: Mutex::new(HashMap::new()),
            borrows: Mutex::new(HashMap::new()),
            reservations: Mutex::new(HashMap::new()),
            approvals: Mutex::new(HashMap::new()),
        }
    }
}

impl Default for InMemoryRepository {
    fn default() -> Self {
        Self::new()
    }
}

impl Repository for InMemoryRepository {
    fn get_archive(&self, id: &str) -> Option<Archive> {
        self.archives.lock().ok().and_then(|map| map.get(id).cloned())
    }

    fn get_all_archives(&self) -> Vec<Archive> {
        self.archives
            .lock()
            .map(|map| map.values().cloned().collect())
            .unwrap_or_default()
    }

    fn save_archive(&self, archive: Archive) {
        if let Ok(mut map) = self.archives.lock() {
            map.insert(archive.id.clone(), archive);
        }
    }

    fn update_archive(&self, archive: Archive) -> Result<(), ()> {
        if let Ok(mut map) = self.archives.lock() {
            if map.contains_key(&archive.id) {
                map.insert(archive.id.clone(), archive);
                return Ok(());
            }
        }
        Err(())
    }

    fn get_user(&self, id: &str) -> Option<User> {
        self.users.lock().ok().and_then(|map| map.get(id).cloned())
    }

    fn get_all_users(&self) -> Vec<User> {
        self.users
            .lock()
            .map(|map| map.values().cloned().collect())
            .unwrap_or_default()
    }

    fn save_user(&self, user: User) {
        if let Ok(mut map) = self.users.lock() {
            map.insert(user.id.clone(), user);
        }
    }

    fn update_user(&self, user: User) -> Result<(), ()> {
        if let Ok(mut map) = self.users.lock() {
            if map.contains_key(&user.id) {
                map.insert(user.id.clone(), user);
                return Ok(());
            }
        }
        Err(())
    }

    fn get_borrow(&self, id: &str) -> Option<Borrow> {
        self.borrows.lock().ok().and_then(|map| map.get(id).cloned())
    }

    fn get_all_borrows(&self) -> Vec<Borrow> {
        self.borrows
            .lock()
            .map(|map| map.values().cloned().collect())
            .unwrap_or_default()
    }

    fn get_active_borrows_by_user(&self, user_id: &str) -> Vec<Borrow> {
        self.borrows
            .lock()
            .map(|map| {
                map.values()
                    .filter(|b| {
                        b.user_id == user_id
                            && matches!(b.status, BorrowStatus::Active | BorrowStatus::Overdue)
                    })
                    .cloned()
                    .collect()
            })
            .unwrap_or_default()
    }

    fn get_overdue_borrows(&self) -> Vec<Borrow> {
        let now = chrono::Utc::now();
        self.borrows
            .lock()
            .map(|map| {
                map.values()
                    .filter(|b| matches!(b.status, BorrowStatus::Active | BorrowStatus::Overdue) && b.due_date < now)
                    .cloned()
                    .collect()
            })
            .unwrap_or_default()
    }

    fn save_borrow(&self, borrow: Borrow) {
        if let Ok(mut map) = self.borrows.lock() {
            map.insert(borrow.id.clone(), borrow);
        }
    }

    fn update_borrow(&self, borrow: Borrow) -> Result<(), ()> {
        if let Ok(mut map) = self.borrows.lock() {
            if map.contains_key(&borrow.id) {
                map.insert(borrow.id.clone(), borrow);
                return Ok(());
            }
        }
        Err(())
    }

    fn get_reservation(&self, id: &str) -> Option<Reservation> {
        self.reservations
            .lock()
            .ok()
            .and_then(|map| map.get(id).cloned())
    }

    fn get_all_reservations(&self) -> Vec<Reservation> {
        self.reservations
            .lock()
            .map(|map| map.values().cloned().collect())
            .unwrap_or_default()
    }

    fn get_reservations_by_archive(&self, archive_id: &str) -> Vec<Reservation> {
        self.reservations
            .lock()
            .map(|map| {
                map.values()
                    .filter(|r| r.archive_id == archive_id)
                    .cloned()
                    .collect()
            })
            .unwrap_or_default()
    }

    fn get_next_pending_reservation(&self, archive_id: &str) -> Option<Reservation> {
        let mut reservations: Vec<Reservation> = self
            .reservations
            .lock()
            .map(|map| {
                map.values()
                    .filter(|r| {
                        r.archive_id == archive_id
                            && matches!(
                                r.status,
                                ReservationStatus::Pending | ReservationStatus::Notified
                            )
                    })
                    .cloned()
                    .collect()
            })
            .unwrap_or_default();

        reservations.sort_by_key(|r| r.queue_position);
        reservations.into_iter().next()
    }

    fn save_reservation(&self, reservation: Reservation) {
        if let Ok(mut map) = self.reservations.lock() {
            map.insert(reservation.id.clone(), reservation);
        }
    }

    fn update_reservation(&self, reservation: Reservation) -> Result<(), ()> {
        if let Ok(mut map) = self.reservations.lock() {
            if map.contains_key(&reservation.id) {
                map.insert(reservation.id.clone(), reservation);
                return Ok(());
            }
        }
        Err(())
    }

    fn remove_reservation(&self, id: &str) {
        if let Ok(mut map) = self.reservations.lock() {
            map.remove(id);
        }
    }

    fn get_approval_process(&self, id: &str) -> Option<ApprovalProcess> {
        self.approvals
            .lock()
            .ok()
            .and_then(|map| map.get(id).cloned())
    }

    fn save_approval_process(&self, process: ApprovalProcess) {
        if let Ok(mut map) = self.approvals.lock() {
            map.insert(process.id.clone(), process);
        }
    }

    fn update_approval_process(&self, process: ApprovalProcess) -> Result<(), ()> {
        if let Ok(mut map) = self.approvals.lock() {
            if map.contains_key(&process.id) {
                map.insert(process.id.clone(), process);
                return Ok(());
            }
        }
        Err(())
    }
}
