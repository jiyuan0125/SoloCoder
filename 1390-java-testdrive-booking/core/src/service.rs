use chrono::Local;
use std::collections::HashSet;
use uuid::Uuid;

use crate::error::{BookingError, Result};
use crate::model::*;
use crate::store::InMemoryStore;

#[derive(Clone)]
pub struct BookingService {
    store: InMemoryStore,
}

impl BookingService {
    pub fn new(store: InMemoryStore) -> Self {
        Self { store }
    }

    pub fn store(&self) -> &InMemoryStore {
        &self.store
    }

    pub async fn create_booking(&self, req: CreateBookingRequest) -> Result<Booking> {
        self.store
            .with_write_lock(|store| {
                let today = Local::now().date_naive();
                
                if req.date < today {
                    return Err(BookingError::Internal("不能预约过去的日期".to_string()));
                }

                if !store.car_models.contains_key(&req.model_id) {
                    return Err(BookingError::CarModelNotFound);
                }

                let has_booking_today = store.bookings.values().any(|b| {
                    b.customer_id == req.customer_id 
                        && b.date == req.date 
                        && !matches!(b.status, BookingStatus::Cancelled | BookingStatus::Expired)
                });
                if has_booking_today {
                    return Err(BookingError::AlreadyBookedToday);
                }

                let booked_car_ids: HashSet<Uuid> = store.bookings
                    .values()
                    .filter(|b| {
                        matches!(b.status, BookingStatus::Reserved | BookingStatus::InProgress)
                            && b.model_id == req.model_id 
                            && b.date == req.date 
                            && b.time_slot == req.time_slot
                    })
                    .map(|b| b.car_id)
                    .collect();

                let available_cars: Vec<Car> = store.cars
                    .values()
                    .filter(|c| c.model_id == req.model_id && !booked_car_ids.contains(&c.id))
                    .cloned()
                    .collect();

                if available_cars.is_empty() {
                    return Err(BookingError::TimeSlotUnavailable);
                }

                let selected_car = &available_cars[0];

                let on_duty_advisors: Vec<SalesAdvisor> = store.advisors
                    .values()
                    .filter(|a| a.on_duty)
                    .cloned()
                    .collect();

                if on_duty_advisors.is_empty() {
                    return Err(BookingError::AdvisorNotFound);
                }

                let selected_advisor = on_duty_advisors
                    .iter()
                    .min_by(|a, b| {
                        a.booking_count
                            .cmp(&b.booking_count)
                            .then_with(|| a.id.cmp(&b.id))
                    })
                    .ok_or(BookingError::AdvisorNotFound)?;

                let mut updated_advisor = selected_advisor.clone();
                updated_advisor.booking_count += 1;
                store.advisors.insert(updated_advisor.id, updated_advisor);

                let booking = Booking {
                    id: Uuid::new_v4(),
                    customer_id: req.customer_id,
                    car_id: selected_car.id,
                    model_id: req.model_id,
                    advisor_id: selected_advisor.id,
                    date: req.date,
                    time_slot: req.time_slot,
                    status: BookingStatus::Reserved,
                    created_at: Local::now(),
                };

                store.bookings.insert(booking.id, booking.clone());
                Ok(booking)
            })
            .await
    }

    pub async fn start_booking(&self, booking_id: &Uuid) -> Result<Booking> {
        self.check_and_expire_old_bookings().await;

        let mut booking = self
            .store
            .get_booking(booking_id)
            .await
            .ok_or(BookingError::BookingNotFound)?;

        match booking.status {
            BookingStatus::Reserved => {
                booking.status = BookingStatus::InProgress;
                self.store.update_booking(booking.clone()).await;
                Ok(booking)
            }
            BookingStatus::Expired => Err(BookingError::BookingExpired),
            _ => Err(BookingError::InvalidStateTransition),
        }
    }

    pub async fn complete_booking(&self, booking_id: &Uuid) -> Result<Booking> {
        let mut booking = self
            .store
            .get_booking(booking_id)
            .await
            .ok_or(BookingError::BookingNotFound)?;

        if booking.status != BookingStatus::InProgress {
            return Err(BookingError::InvalidStateTransition);
        }

        booking.status = BookingStatus::Completed;
        self.store.update_booking(booking.clone()).await;
        Ok(booking)
    }

    pub async fn cancel_booking(&self, booking_id: &Uuid) -> Result<Booking> {
        let mut booking = self
            .store
            .get_booking(booking_id)
            .await
            .ok_or(BookingError::BookingNotFound)?;

        if !matches!(booking.status, BookingStatus::Reserved) {
            return Err(BookingError::InvalidStateTransition);
        }

        booking.status = BookingStatus::Cancelled;
        self.store.update_booking(booking.clone()).await;
        Ok(booking)
    }

    pub async fn submit_feedback(&self, booking_id: &Uuid, req: SubmitFeedbackRequest) -> Result<Feedback> {
        let booking = self
            .store
            .get_booking(booking_id)
            .await
            .ok_or(BookingError::BookingNotFound)?;

        if booking.status != BookingStatus::Completed {
            return Err(BookingError::InvalidStateTransition);
        }

        let feedback = Feedback {
            booking_id: *booking_id,
            satisfaction: req.satisfaction,
            purchase_intent: req.purchase_intent,
            created_at: Local::now(),
        };

        self.store.insert_feedback(feedback.clone()).await;
        Ok(feedback)
    }

    pub async fn check_and_expire_old_bookings(&self) {
        let today = Local::now().date_naive();
        let bookings = self.store.list_bookings().await;

        for booking in bookings {
            if booking.status == BookingStatus::Reserved && booking.date < today {
                let mut updated = booking.clone();
                updated.status = BookingStatus::Expired;
                self.store.update_booking(updated).await;
            }
        }
    }
}
