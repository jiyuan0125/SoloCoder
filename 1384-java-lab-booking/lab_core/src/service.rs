use chrono::{DateTime, Utc};
use uuid::Uuid;

use crate::errors::{AppError, AppResult};
use crate::models::{
    Booking, BookingStatus, Consumable, CreateBookingRequest,
    CreateConsumableRequest, CreateEquipmentRequest, CreateUserRequest, Equipment,
    LowStockAlert, UpdateConsumableStockRequest, User, UserRole,
};
use crate::state::AppState;

pub struct BookingService;

impl BookingService {
    pub fn check_time_overlap(
        start1: &DateTime<Utc>,
        end1: &DateTime<Utc>,
        start2: &DateTime<Utc>,
        end2: &DateTime<Utc>,
    ) -> bool {
        start1 < end2 && end1 > start2
    }

    pub fn create_booking(state: &AppState, req: CreateBookingRequest) -> AppResult<Booking> {
        if req.start_time >= req.end_time {
            return Err(AppError::InvalidTimeRange);
        }

        let mut inner = state
            .inner
            .write()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        if !inner.equipment.contains_key(&req.equipment_id) {
            return Err(AppError::EquipmentNotFound(req.equipment_id.to_string()));
        }

        if !inner.users.contains_key(&req.researcher_id) {
            return Err(AppError::UserNotFound(req.researcher_id.to_string()));
        }

        for usage in &req.consumables {
            if !inner.consumables.contains_key(&usage.consumable_id) {
                return Err(AppError::ConsumableNotFound(usage.consumable_id.to_string()));
            }
        }

        for booking in inner.bookings.values() {
            if booking.equipment_id == req.equipment_id
                && booking.status != BookingStatus::Cancelled
                && booking.status != BookingStatus::Completed
                && Self::check_time_overlap(
                    &booking.start_time,
                    &booking.end_time,
                    &req.start_time,
                    &req.end_time,
                )
            {
                return Err(AppError::TimeSlotConflict);
            }
        }

        let booking = Booking {
            id: Uuid::new_v4(),
            equipment_id: req.equipment_id,
            researcher_id: req.researcher_id,
            start_time: req.start_time,
            end_time: req.end_time,
            status: BookingStatus::Pending,
            consumables: req.consumables,
            created_at: Utc::now(),
        };

        inner.bookings.insert(booking.id, booking.clone());

        Ok(booking)
    }

    pub fn confirm_booking(state: &AppState, booking_id: Uuid, admin_id: Uuid) -> AppResult<Booking> {
        let mut inner = state
            .inner
            .write()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        let admin = inner
            .users
            .get(&admin_id)
            .ok_or_else(|| AppError::UserNotFound(admin_id.to_string()))?;

        if admin.role != UserRole::Admin {
            return Err(AppError::Unauthorized);
        }

        let booking = inner
            .bookings
            .get_mut(&booking_id)
            .ok_or_else(|| AppError::BookingNotFound(booking_id.to_string()))?;

        if booking.status != BookingStatus::Pending {
            return Err(AppError::InvalidStatusTransition {
                from: format!("{:?}", booking.status),
                to: "Confirmed".to_string(),
            });
        }

        booking.status = BookingStatus::Confirmed;

        Ok(booking.clone())
    }

    pub fn start_booking(state: &AppState, booking_id: Uuid) -> AppResult<Booking> {
        let mut inner = state
            .inner
            .write()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        let booking = inner
            .bookings
            .get(&booking_id)
            .ok_or_else(|| AppError::BookingNotFound(booking_id.to_string()))?
            .clone();

        if booking.status != BookingStatus::Confirmed {
            return Err(AppError::InvalidStatusTransition {
                from: format!("{:?}", booking.status),
                to: "InProgress".to_string(),
            });
        }

        for usage in &booking.consumables {
            let consumable = inner
                .consumables
                .get(&usage.consumable_id)
                .ok_or_else(|| AppError::ConsumableNotFound(usage.consumable_id.to_string()))?;

            if consumable.stock < usage.quantity {
                return Err(AppError::InsufficientStock {
                    name: consumable.name.clone(),
                    required: usage.quantity,
                    available: consumable.stock,
                });
            }
        }

        for usage in &booking.consumables {
            let consumable = inner
                .consumables
                .get_mut(&usage.consumable_id)
                .ok_or_else(|| AppError::ConsumableNotFound(usage.consumable_id.to_string()))?;
            consumable.stock -= usage.quantity;
        }

        let booking = inner
            .bookings
            .get_mut(&booking_id)
            .ok_or_else(|| AppError::BookingNotFound(booking_id.to_string()))?;

        booking.status = BookingStatus::InProgress;

        Ok(booking.clone())
    }

    pub fn complete_booking(state: &AppState, booking_id: Uuid) -> AppResult<Booking> {
        let mut inner = state
            .inner
            .write()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        let booking = inner
            .bookings
            .get_mut(&booking_id)
            .ok_or_else(|| AppError::BookingNotFound(booking_id.to_string()))?;

        if booking.status != BookingStatus::InProgress {
            return Err(AppError::InvalidStatusTransition {
                from: format!("{:?}", booking.status),
                to: "Completed".to_string(),
            });
        }

        booking.status = BookingStatus::Completed;

        Ok(booking.clone())
    }

    pub fn cancel_booking(state: &AppState, booking_id: Uuid) -> AppResult<Booking> {
        let mut inner = state
            .inner
            .write()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        let booking = inner
            .bookings
            .get(&booking_id)
            .ok_or_else(|| AppError::BookingNotFound(booking_id.to_string()))?
            .clone();

        if booking.status == BookingStatus::InProgress || booking.status == BookingStatus::Completed {
            return Err(AppError::CannotCancelStartedBooking);
        }

        if booking.status == BookingStatus::Cancelled {
            return Err(AppError::InvalidStatusTransition {
                from: "Cancelled".to_string(),
                to: "Cancelled".to_string(),
            });
        }

        if booking.status == BookingStatus::InProgress {
            for usage in &booking.consumables {
                let consumable = inner
                    .consumables
                    .get_mut(&usage.consumable_id)
                    .ok_or_else(|| AppError::ConsumableNotFound(usage.consumable_id.to_string()))?;
                consumable.stock += usage.quantity;
            }
        }

        let booking = inner
            .bookings
            .get_mut(&booking_id)
            .ok_or_else(|| AppError::BookingNotFound(booking_id.to_string()))?;

        booking.status = BookingStatus::Cancelled;

        Ok(booking.clone())
    }

    pub fn get_booking(state: &AppState, booking_id: Uuid) -> AppResult<Booking> {
        let inner = state
            .inner
            .read()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        inner
            .bookings
            .get(&booking_id)
            .cloned()
            .ok_or_else(|| AppError::BookingNotFound(booking_id.to_string()))
    }

    pub fn list_bookings(state: &AppState) -> AppResult<Vec<Booking>> {
        let inner = state
            .inner
            .read()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        Ok(inner.bookings.values().cloned().collect())
    }
}

pub struct EquipmentService;

impl EquipmentService {
    pub fn create(state: &AppState, req: CreateEquipmentRequest) -> AppResult<Equipment> {
        let mut inner = state
            .inner
            .write()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        let equipment = Equipment {
            id: Uuid::new_v4(),
            name: req.name,
            description: req.description,
        };

        inner.equipment.insert(equipment.id, equipment.clone());

        Ok(equipment)
    }

    pub fn get(state: &AppState, id: Uuid) -> AppResult<Equipment> {
        let inner = state
            .inner
            .read()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        inner
            .equipment
            .get(&id)
            .cloned()
            .ok_or_else(|| AppError::EquipmentNotFound(id.to_string()))
    }

    pub fn list(state: &AppState) -> AppResult<Vec<Equipment>> {
        let inner = state
            .inner
            .read()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        Ok(inner.equipment.values().cloned().collect())
    }
}

pub struct ConsumableService;

impl ConsumableService {
    pub fn create(state: &AppState, req: CreateConsumableRequest) -> AppResult<Consumable> {
        let mut inner = state
            .inner
            .write()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        let consumable = Consumable {
            id: Uuid::new_v4(),
            name: req.name,
            stock: req.initial_stock,
            safety_stock: req.safety_stock,
        };

        inner.consumables.insert(consumable.id, consumable.clone());

        Ok(consumable)
    }

    pub fn get(state: &AppState, id: Uuid) -> AppResult<Consumable> {
        let inner = state
            .inner
            .read()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        inner
            .consumables
            .get(&id)
            .cloned()
            .ok_or_else(|| AppError::ConsumableNotFound(id.to_string()))
    }

    pub fn list(state: &AppState) -> AppResult<Vec<Consumable>> {
        let inner = state
            .inner
            .read()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        Ok(inner.consumables.values().cloned().collect())
    }

    pub fn restock(
        state: &AppState,
        id: Uuid,
        req: UpdateConsumableStockRequest,
    ) -> AppResult<Consumable> {
        let mut inner = state
            .inner
            .write()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        let consumable = inner
            .consumables
            .get_mut(&id)
            .ok_or_else(|| AppError::ConsumableNotFound(id.to_string()))?;

        consumable.stock = consumable.stock.saturating_add(req.quantity);

        Ok(consumable.clone())
    }

    pub fn get_low_stock_alerts(state: &AppState) -> AppResult<Vec<LowStockAlert>> {
        let inner = state
            .inner
            .read()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        let alerts = inner
            .consumables
            .values()
            .filter(|c| c.is_below_safety())
            .map(|c| LowStockAlert {
                consumable_id: c.id,
                consumable_name: c.name.clone(),
                current_stock: c.stock,
                safety_stock: c.safety_stock,
            })
            .collect();

        Ok(alerts)
    }
}

pub struct UserService;

impl UserService {
    pub fn create(state: &AppState, req: CreateUserRequest) -> AppResult<User> {
        let mut inner = state
            .inner
            .write()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        let user = User {
            id: Uuid::new_v4(),
            name: req.name,
            role: req.role,
        };

        inner.users.insert(user.id, user.clone());

        Ok(user)
    }

    pub fn get(state: &AppState, id: Uuid) -> AppResult<User> {
        let inner = state
            .inner
            .read()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        inner
            .users
            .get(&id)
            .cloned()
            .ok_or_else(|| AppError::UserNotFound(id.to_string()))
    }

    pub fn list(state: &AppState) -> AppResult<Vec<User>> {
        let inner = state
            .inner
            .read()
            .map_err(|e| AppError::Internal(e.to_string()))?;

        Ok(inner.users.values().cloned().collect())
    }
}
