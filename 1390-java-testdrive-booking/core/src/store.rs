use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;
use uuid::Uuid;

use crate::model::*;

#[derive(Clone, Default)]
pub struct InMemoryStore {
    inner: Arc<RwLock<InnerStore>>,
}

#[derive(Default)]
pub struct InnerStore {
    pub car_models: HashMap<Uuid, CarModel>,
    pub cars: HashMap<Uuid, Car>,
    pub customers: HashMap<Uuid, Customer>,
    pub advisors: HashMap<Uuid, SalesAdvisor>,
    pub bookings: HashMap<Uuid, Booking>,
    pub feedbacks: HashMap<Uuid, Feedback>,
}

impl InMemoryStore {
    pub fn new() -> Self {
        Self::default()
    }

    pub async fn get_car_model(&self, id: &Uuid) -> Option<CarModel> {
        self.inner.read().await.car_models.get(id).cloned()
    }

    pub async fn list_car_models(&self) -> Vec<CarModel> {
        self.inner.read().await.car_models.values().cloned().collect()
    }

    pub async fn get_cars_by_model(&self, model_id: &Uuid) -> Vec<Car> {
        self.inner
            .read()
            .await
            .cars
            .values()
            .filter(|c| c.model_id == *model_id)
            .cloned()
            .collect()
    }

    pub async fn get_car(&self, id: &Uuid) -> Option<Car> {
        self.inner.read().await.cars.get(id).cloned()
    }

    pub async fn list_cars(&self) -> Vec<Car> {
        self.inner.read().await.cars.values().cloned().collect()
    }

    pub async fn get_customer(&self, id: &Uuid) -> Option<Customer> {
        self.inner.read().await.customers.get(id).cloned()
    }

    pub async fn list_customers(&self) -> Vec<Customer> {
        self.inner.read().await.customers.values().cloned().collect()
    }

    pub async fn get_advisor(&self, id: &Uuid) -> Option<SalesAdvisor> {
        self.inner.read().await.advisors.get(id).cloned()
    }

    pub async fn list_advisors(&self) -> Vec<SalesAdvisor> {
        self.inner.read().await.advisors.values().cloned().collect()
    }

    pub async fn update_advisor(&self, advisor: SalesAdvisor) {
        self.inner.write().await.advisors.insert(advisor.id, advisor);
    }

    pub async fn get_booking(&self, id: &Uuid) -> Option<Booking> {
        self.inner.read().await.bookings.get(id).cloned()
    }

    pub async fn list_bookings(&self) -> Vec<Booking> {
        self.inner.read().await.bookings.values().cloned().collect()
    }

    pub async fn get_active_bookings(&self) -> Vec<Booking> {
        self.inner
            .read()
            .await
            .bookings
            .values()
            .filter(|b| matches!(b.status, BookingStatus::Reserved | BookingStatus::InProgress))
            .cloned()
            .collect()
    }

    pub async fn insert_booking(&self, booking: Booking) {
        self.inner.write().await.bookings.insert(booking.id, booking);
    }

    pub async fn update_booking(&self, booking: Booking) {
        self.inner.write().await.bookings.insert(booking.id, booking);
    }

    pub async fn get_feedback(&self, booking_id: &Uuid) -> Option<Feedback> {
        self.inner.read().await.feedbacks.get(booking_id).cloned()
    }

    pub async fn insert_feedback(&self, feedback: Feedback) {
        self.inner.write().await.feedbacks.insert(feedback.booking_id, feedback);
    }

    pub async fn with_write_lock<T>(&self, f: impl FnOnce(&mut InnerStore) -> T) -> T {
        let mut write = self.inner.write().await;
        f(&mut write)
    }
}

impl InMemoryStore {
    pub async fn seed_sample_data(&self) {
        let mut store = self.inner.write().await;

        let model1 = CarModel { id: Uuid::new_v4(), name: "特斯拉Model 3".to_string() };
        let model2 = CarModel { id: Uuid::new_v4(), name: "比亚迪汉".to_string() };
        let model3 = CarModel { id: Uuid::new_v4(), name: "蔚来ET5".to_string() };

        store.car_models.insert(model1.id, model1.clone());
        store.car_models.insert(model2.id, model2.clone());
        store.car_models.insert(model3.id, model3.clone());

        for i in 1..=3 {
            let car = Car { id: Uuid::new_v4(), model_id: model1.id, plate_number: format!("沪A0000{}", i) };
            store.cars.insert(car.id, car);
        }
        for i in 4..=6 {
            let car = Car { id: Uuid::new_v4(), model_id: model2.id, plate_number: format!("沪A0000{}", i) };
            store.cars.insert(car.id, car);
        }
        for i in 7..=9 {
            let car = Car { id: Uuid::new_v4(), model_id: model3.id, plate_number: format!("沪A0000{}", i) };
            store.cars.insert(car.id, car);
        }

        let customer1 = Customer { id: Uuid::new_v4(), name: "张三".to_string(), phone: "13800138001".to_string() };
        let customer2 = Customer { id: Uuid::new_v4(), name: "李四".to_string(), phone: "13800138002".to_string() };
        store.customers.insert(customer1.id, customer1);
        store.customers.insert(customer2.id, customer2);

        let advisor1 = SalesAdvisor { id: Uuid::new_v4(), name: "王销售".to_string(), on_duty: true, booking_count: 0 };
        let advisor2 = SalesAdvisor { id: Uuid::new_v4(), name: "李销售".to_string(), on_duty: true, booking_count: 0 };
        let advisor3 = SalesAdvisor { id: Uuid::new_v4(), name: "张销售".to_string(), on_duty: false, booking_count: 0 };
        store.advisors.insert(advisor1.id, advisor1);
        store.advisors.insert(advisor2.id, advisor2);
        store.advisors.insert(advisor3.id, advisor3);
    }
}
