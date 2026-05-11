use std::collections::HashMap;
use std::sync::Mutex;

use chrono::Utc;
use rust_decimal::prelude::*;
use rust_decimal::Decimal;
use uuid::Uuid;

use crate::models::*;
use crate::errors::BookingError;

pub struct BookingService {
    menu_sets: Mutex<HashMap<Uuid, MenuSet>>,
    menu_items: Mutex<HashMap<Uuid, MenuItem>>,
    bookings: Mutex<HashMap<Uuid, Booking>>,
}

impl BookingService {
    pub fn new() -> Self {
        let service = BookingService {
            menu_sets: Mutex::new(HashMap::new()),
            menu_items: Mutex::new(HashMap::new()),
            bookings: Mutex::new(HashMap::new()),
        };
        service.initialize_default_data();
        service
    }

    fn initialize_default_data(&self) {
        let items = vec![
            MenuItem::new("清蒸鲈鱼", "热菜", Decimal::new(288, 0)),
            MenuItem::new("红烧海参", "热菜", Decimal::new(388, 0)),
            MenuItem::new("北京烤鸭", "热菜", Decimal::new(268, 0)),
            MenuItem::new("宫保鸡丁", "热菜", Decimal::new(88, 0)),
            MenuItem::new("水煮鱼", "热菜", Decimal::new(168, 0)),
            MenuItem::new("糖醋里脊", "热菜", Decimal::new(68, 0)),
            MenuItem::new("白灼虾", "热菜", Decimal::new(198, 0)),
            MenuItem::new("红烧肉", "热菜", Decimal::new(98, 0)),
            MenuItem::new("凉拌黄瓜", "凉菜", Decimal::new(28, 0)),
            MenuItem::new("口水鸡", "凉菜", Decimal::new(68, 0)),
            MenuItem::new("夫妻肺片", "凉菜", Decimal::new(58, 0)),
            MenuItem::new("皮蛋豆腐", "凉菜", Decimal::new(22, 0)),
            MenuItem::new("蛋花汤", "汤类", Decimal::new(28, 0)),
            MenuItem::new("紫菜蛋汤", "汤类", Decimal::new(18, 0)),
            MenuItem::new("酸辣汤", "汤类", Decimal::new(32, 0)),
            MenuItem::new("水果拼盘", "主食", Decimal::new(88, 0)),
            MenuItem::new("扬州炒饭", "主食", Decimal::new(48, 0)),
            MenuItem::new("长寿面", "主食", Decimal::new(38, 0)),
        ];

        let mut menu_items = self.menu_items.lock().unwrap();
        for item in items {
            menu_items.insert(item.id, item);
        }

        let wedding_items = menu_items.values()
            .filter(|i| i.category == "热菜")
            .take(4)
            .cloned()
            .chain(menu_items.values().filter(|i| i.category == "凉菜").take(2).cloned())
            .chain(menu_items.values().filter(|i| i.category == "汤类").take(1).cloned())
            .chain(menu_items.values().filter(|i| i.category == "主食").take(1).cloned())
            .collect::<Vec<_>>();

        let wedding_set = MenuSet::new("豪华婚宴套系", BanquetType::Wedding, wedding_items, Decimal::new(5000, 0));
        
        let birthday_items = menu_items.values()
            .filter(|i| i.category == "热菜")
            .take(3)
            .cloned()
            .chain(menu_items.values().filter(|i| i.category == "凉菜").take(2).cloned())
            .chain(menu_items.values().filter(|i| i.category == "汤类").take(1).cloned())
            .chain(menu_items.values().filter(|i| i.name == "长寿面").cloned())
            .collect::<Vec<_>>();
        
        let birthday_set = MenuSet::new("福寿双全寿宴", BanquetType::Birthday, birthday_items, Decimal::new(3000, 0));
        
        let graduation_items = menu_items.values()
            .filter(|i| i.category == "热菜")
            .take(3)
            .cloned()
            .chain(menu_items.values().filter(|i| i.category == "凉菜").take(1).cloned())
            .chain(menu_items.values().filter(|i| i.category == "汤类").take(1).cloned())
            .chain(menu_items.values().filter(|i| i.category == "主食").take(1).cloned())
            .collect::<Vec<_>>();
        
        let graduation_set = MenuSet::new("金榜题名升学宴", BanquetType::Graduation, graduation_items, Decimal::new(2500, 0));
        
        let business_items = menu_items.values()
            .filter(|i| i.category == "热菜")
            .take(4)
            .cloned()
            .chain(menu_items.values().filter(|i| i.category == "凉菜").take(2).cloned())
            .chain(menu_items.values().filter(|i| i.category == "汤类").take(1).cloned())
            .chain(menu_items.values().filter(|i| i.category == "主食").take(1).cloned())
            .collect::<Vec<_>>();
        
        let business_set = MenuSet::new("商务尊享套系", BanquetType::Business, business_items, Decimal::new(4000, 0));

        let mut menu_sets = self.menu_sets.lock().unwrap();
        menu_sets.insert(wedding_set.id, wedding_set);
        menu_sets.insert(birthday_set.id, birthday_set);
        menu_sets.insert(graduation_set.id, graduation_set);
        menu_sets.insert(business_set.id, business_set);
    }

    pub fn list_menu_sets(&self) -> Vec<MenuSet> {
        let menu_sets = self.menu_sets.lock().unwrap();
        menu_sets.values().cloned().collect()
    }

    pub fn get_menu_set(&self, id: Uuid) -> Option<MenuSet> {
        let menu_sets = self.menu_sets.lock().unwrap();
        menu_sets.get(&id).cloned()
    }

    pub fn list_menu_items(&self) -> Vec<MenuItem> {
        let menu_items = self.menu_items.lock().unwrap();
        menu_items.values().cloned().collect()
    }

    pub fn create_booking(&self, req: CreateBookingRequest) -> Result<Booking, BookingError> {
        if req.event_date < Utc::now() {
            return Err(BookingError::InvalidEventDate);
        }
        if req.booked_tables == 0 {
            return Err(BookingError::InvalidTableCount);
        }
        if req.deposit < Decimal::ZERO {
            return Err(BookingError::InvalidDeposit);
        }

        let menu_sets = self.menu_sets.lock().unwrap();
        if !menu_sets.contains_key(&req.menu_set_id) {
            return Err(BookingError::MenuSetNotFound);
        }

        let booking = Booking {
            id: Uuid::new_v4(),
            customer_name: req.customer_name,
            customer_phone: req.customer_phone,
            banquet_type: req.banquet_type,
            menu_set_id: req.menu_set_id,
            booked_tables: req.booked_tables,
            event_date: req.event_date,
            deposit: req.deposit,
            created_at: Utc::now(),
            status: BookingStatus::Confirmed,
            actual_tables: None,
            cancelled_at: None,
            refund_amount: None,
            penalty_amount: None,
            unused_option: None,
            substitutions: Vec::new(),
        };

        let mut bookings = self.bookings.lock().unwrap();
        bookings.insert(booking.id, booking.clone());

        Ok(booking)
    }

    pub fn get_booking(&self, id: Uuid) -> Option<Booking> {
        let bookings = self.bookings.lock().unwrap();
        bookings.get(&id).cloned()
    }

    pub fn list_bookings(&self) -> Vec<Booking> {
        let bookings = self.bookings.lock().unwrap();
        bookings.values().cloned().collect()
    }

    pub fn adjust_tables(&self, req: AdjustTablesRequest) -> Result<Booking, BookingError> {
        let mut bookings = self.bookings.lock().unwrap();
        let booking = bookings.get_mut(&req.booking_id)
            .ok_or(BookingError::BookingNotFound)?;

        if booking.status != BookingStatus::Confirmed {
            return Err(BookingError::BookingInactive);
        }

        if req.new_tables == 0 {
            return Err(BookingError::InvalidTableCount);
        }

        let now = Utc::now();
        let hours_to_event = booking.event_date.signed_duration_since(now).num_hours();

        if req.new_tables < booking.booked_tables {
            let reduced = booking.booked_tables - req.new_tables;
            
            if hours_to_event < 24 {
                let menu_sets = self.menu_sets.lock().unwrap();
                let menu_set = menu_sets.get(&booking.menu_set_id)
                    .ok_or(BookingError::MenuSetNotFound)?;
                
                let penalty = menu_set.price_per_table * Decimal::from(reduced) * Decimal::from_f64(0.5).unwrap();
                booking.penalty_amount = Some(
                    booking.penalty_amount.unwrap_or(Decimal::ZERO) + penalty
                );
            }
        }

        booking.booked_tables = req.new_tables;

        Ok(booking.clone())
    }

    pub fn substitute_dish(&self, req: SubstituteDishRequest) -> Result<Booking, BookingError> {
        let mut bookings = self.bookings.lock().unwrap();
        let booking = bookings.get_mut(&req.booking_id)
            .ok_or(BookingError::BookingNotFound)?;

        if booking.status != BookingStatus::Confirmed {
            return Err(BookingError::BookingInactive);
        }

        let menu_items = self.menu_items.lock().unwrap();
        let original = menu_items.get(&req.original_item_id)
            .ok_or(BookingError::MenuItemNotFound)?;
        let replacement = menu_items.get(&req.replacement_item_id)
            .ok_or(BookingError::MenuItemNotFound)?;

        if original.category != replacement.category {
            return Err(BookingError::DifferentCategory);
        }

        let diff = (original.price - replacement.price).abs();
        let max_allowed = original.price * Decimal::from_f64(0.3).unwrap();
        if diff > max_allowed {
            return Err(BookingError::PriceDifferenceExceeded);
        }

        let substitution = Substitution {
            original_item_id: original.id,
            replacement_item_id: replacement.id,
            original_name: original.name.clone(),
            replacement_name: replacement.name.clone(),
            price_difference: replacement.price - original.price,
        };

        booking.substitutions.push(substitution);

        Ok(booking.clone())
    }

    pub fn complete_booking(&self, req: CompleteBookingRequest) -> Result<Booking, BookingError> {
        let mut bookings = self.bookings.lock().unwrap();
        let booking = bookings.get_mut(&req.booking_id)
            .ok_or(BookingError::BookingNotFound)?;

        if booking.status != BookingStatus::Confirmed {
            return Err(BookingError::BookingInactive);
        }

        if req.actual_tables > booking.booked_tables {
            return Err(BookingError::ActualTablesExceedBooked);
        }

        booking.actual_tables = Some(req.actual_tables);
        booking.unused_option = Some(req.unused_option);
        booking.status = BookingStatus::Completed;

        if req.unused_option == UnusedOption::Refund && req.actual_tables < booking.booked_tables {
            let menu_sets = self.menu_sets.lock().unwrap();
            let menu_set = menu_sets.get(&booking.menu_set_id)
                .ok_or(BookingError::MenuSetNotFound)?;
            
            let unused_tables = booking.booked_tables - req.actual_tables;
            let refund = menu_set.price_per_table * Decimal::from(unused_tables) * Decimal::from_f64(0.6).unwrap();
            booking.refund_amount = Some(refund);
        }

        Ok(booking.clone())
    }

    pub fn cancel_booking(&self, req: CancelBookingRequest) -> Result<Booking, BookingError> {
        let mut bookings = self.bookings.lock().unwrap();
        let booking = bookings.get_mut(&req.booking_id)
            .ok_or(BookingError::BookingNotFound)?;

        if booking.status != BookingStatus::Confirmed {
            return Err(BookingError::BookingInactive);
        }

        let now = Utc::now();
        let days_to_event = booking.event_date.signed_duration_since(now).num_days();

        let refund_rate = if days_to_event >= 15 {
            Decimal::ONE
        } else if days_to_event >= 7 {
            Decimal::from_f64(0.5).unwrap()
        } else {
            Decimal::ZERO
        };

        booking.status = BookingStatus::Cancelled;
        booking.cancelled_at = Some(now);
        booking.refund_amount = Some(booking.deposit * refund_rate);

        Ok(booking.clone())
    }

    pub fn calculate_total(&self, booking: &Booking) -> Result<Decimal, BookingError> {
        let menu_sets = self.menu_sets.lock().unwrap();
        let menu_set = menu_sets.get(&booking.menu_set_id)
            .ok_or(BookingError::MenuSetNotFound)?;

        let tables = booking.actual_tables.unwrap_or(booking.booked_tables);
        let base_total = menu_set.price_per_table * Decimal::from(tables);
        
        let min_charge = booking.banquet_type.min_charge_per_table() * Decimal::from(tables);
        
        let substitution_diff: Decimal = booking.substitutions.iter()
            .map(|s| s.price_difference * Decimal::from(tables))
            .sum();
        
        let mut total = base_total + substitution_diff;
        
        if total < min_charge {
            total = min_charge;
        }

        Ok(total)
    }

    pub fn get_booking_summary(&self, booking_id: Uuid) -> Result<BookingSummary, BookingError> {
        let booking = self.get_booking(booking_id).ok_or(BookingError::BookingNotFound)?;
        let total = self.calculate_total(&booking)?;
        
        let balance = total - booking.deposit + booking.penalty_amount.unwrap_or(Decimal::ZERO) 
            - booking.refund_amount.unwrap_or(Decimal::ZERO);

        Ok(BookingSummary {
            booking_id: booking.id,
            customer_name: booking.customer_name,
            banquet_type: booking.banquet_type.as_str().to_string(),
            booked_tables: booking.booked_tables,
            event_date: booking.event_date,
            status: booking.status,
            total_amount: total,
            deposit: booking.deposit,
            balance,
            penalty_amount: booking.penalty_amount,
            refund_amount: booking.refund_amount,
        })
    }
}

impl Default for BookingService {
    fn default() -> Self {
        Self::new()
    }
}
