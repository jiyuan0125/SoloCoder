use std::collections::HashMap;
use chrono::{DateTime, Local, Duration};
use uuid::Uuid;

use crate::models::{Hall, Movie, Schedule, ScheduleCreate, ScheduleUpdate, HallType};
use crate::errors::ScheduleError;

pub const CLEANING_MINUTES: i64 = 15;
pub const MAX_PRIME_TIME_SCHEDULES_PER_DAY: usize = 2;

pub struct Scheduler {
    halls: HashMap<Uuid, Hall>,
    movies: HashMap<Uuid, Movie>,
    schedules: HashMap<Uuid, Schedule>,
}

impl Default for Scheduler {
    fn default() -> Self {
        Self::new()
    }
}

impl Scheduler {
    pub fn new() -> Self {
        let mut scheduler = Self {
            halls: HashMap::new(),
            movies: HashMap::new(),
            schedules: HashMap::new(),
        };
        scheduler.initialize_defaults();
        scheduler
    }

    fn initialize_defaults(&mut self) {
        self.add_hall(Hall::new("1号厅".to_string(), HallType::Normal));
        self.add_hall(Hall::new("2号厅".to_string(), HallType::Normal));
        self.add_hall(Hall::new("IMAX厅".to_string(), HallType::IMAX));
        self.add_hall(Hall::new("VIP厅".to_string(), HallType::VIP));

        self.add_movie(Movie::new("流浪地球3".to_string(), 173));
        self.add_movie(Movie::new("哪吒2".to_string(), 110));
        self.add_movie(Movie::new("复仇者联盟5".to_string(), 145));
    }

    pub fn add_hall(&mut self, hall: Hall) {
        self.halls.insert(hall.id, hall);
    }

    pub fn add_movie(&mut self, movie: Movie) {
        self.movies.insert(movie.id, movie);
    }

    pub fn get_halls(&self) -> Vec<&Hall> {
        self.halls.values().collect()
    }

    pub fn get_movies(&self) -> Vec<&Movie> {
        self.movies.values().collect()
    }

    pub fn get_schedules(&self) -> Vec<&Schedule> {
        self.schedules.values().collect()
    }

    pub fn get_schedule(&self, id: Uuid) -> Option<&Schedule> {
        self.schedules.get(&id)
    }

    pub fn get_hall(&self, id: Uuid) -> Option<&Hall> {
        self.halls.get(&id)
    }

    pub fn get_movie(&self, id: Uuid) -> Option<&Movie> {
        self.movies.get(&id)
    }

    fn get_hall_schedules(&self, hall_id: Uuid, exclude_id: Option<Uuid>) -> Vec<&Schedule> {
        self.schedules
            .values()
            .filter(|s| s.hall_id == hall_id && exclude_id.map_or(true, |e| e != s.id))
            .collect()
    }

    fn check_conflicts(
        &self,
        hall_id: Uuid,
        start_time: DateTime<Local>,
        end_time: DateTime<Local>,
        exclude_schedule_id: Option<Uuid>,
    ) -> Result<(), ScheduleError> {
        let hall_schedules = self.get_hall_schedules(hall_id, exclude_schedule_id);
        let temp_schedule = Schedule {
            id: Uuid::nil(),
            hall_id,
            movie_id: Uuid::nil(),
            start_time,
            end_time,
        };

        for existing in &hall_schedules {
            if temp_schedule.overlaps_with_cleaning(existing, CLEANING_MINUTES) {
                let movie = self.movies.get(&existing.movie_id);
                let conflict_desc = format!(
                    "{} ({}-{})",
                    movie.map(|m| m.title.as_str()).unwrap_or("未知影片"),
                    existing.start_time.format("%Y-%m-%d %H:%M"),
                    existing.end_time.format("%Y-%m-%d %H:%M")
                );
                return Err(ScheduleError::Conflict(conflict_desc));
            }
        }

        Ok(())
    }

    fn check_prime_time_limit(
        &self,
        hall_id: Uuid,
        start_time: DateTime<Local>,
        end_time: DateTime<Local>,
        exclude_schedule_id: Option<Uuid>,
    ) -> Result<(), ScheduleError> {
        let temp_schedule = Schedule {
            id: Uuid::nil(),
            hall_id,
            movie_id: Uuid::nil(),
            start_time,
            end_time,
        };

        let affected_dates = temp_schedule.get_prime_hours();
        let hall_schedules = self.get_hall_schedules(hall_id, exclude_schedule_id);

        for date in affected_dates {
            let mut count = 0;
            for schedule in &hall_schedules {
                let schedule_dates = schedule.get_prime_hours();
                if schedule_dates.contains(&date) {
                    count += 1;
                }
            }
            if count >= MAX_PRIME_TIME_SCHEDULES_PER_DAY {
                return Err(ScheduleError::PrimeTimeLimitExceeded);
            }
        }

        Ok(())
    }

    pub fn create_schedule(
        &mut self,
        create: ScheduleCreate,
    ) -> Result<Schedule, ScheduleError> {
        if !self.halls.contains_key(&create.hall_id) {
            return Err(ScheduleError::HallNotFound);
        }

        let movie = self.movies
            .get(&create.movie_id)
            .ok_or(ScheduleError::MovieNotFound)?;

        let duration = Duration::minutes(movie.duration_minutes);
        let end_time = create.start_time + duration;

        self.check_conflicts(create.hall_id, create.start_time, end_time, None)?;
        self.check_prime_time_limit(create.hall_id, create.start_time, end_time, None)?;

        let schedule = Schedule::new(
            create.hall_id,
            create.movie_id,
            create.start_time,
            duration,
        );

        let schedule_clone = schedule.clone();
        self.schedules.insert(schedule.id, schedule);
        Ok(schedule_clone)
    }

    pub fn update_schedule(
        &mut self,
        schedule_id: Uuid,
        update: ScheduleUpdate,
    ) -> Result<Schedule, ScheduleError> {
        let existing = self.schedules
            .get(&schedule_id)
            .cloned()
            .ok_or(ScheduleError::ScheduleNotFound)?;

        let hall_id = update.hall_id.unwrap_or(existing.hall_id);
        let movie_id = update.movie_id.unwrap_or(existing.movie_id);
        let start_time = update.start_time.unwrap_or(existing.start_time);

        if !self.halls.contains_key(&hall_id) {
            return Err(ScheduleError::HallNotFound);
        }

        let movie = self.movies
            .get(&movie_id)
            .ok_or(ScheduleError::MovieNotFound)?;

        let duration = Duration::minutes(movie.duration_minutes);
        let end_time = start_time + duration;

        self.check_conflicts(hall_id, start_time, end_time, Some(schedule_id))?;
        self.check_prime_time_limit(hall_id, start_time, end_time, Some(schedule_id))?;

        let updated = Schedule {
            id: schedule_id,
            hall_id,
            movie_id,
            start_time,
            end_time,
        };

        self.schedules.insert(schedule_id, updated.clone());
        Ok(updated)
    }

    pub fn delete_schedule(&mut self, schedule_id: Uuid) -> Result<(), ScheduleError> {
        if self.schedules.remove(&schedule_id).is_some() {
            Ok(())
        } else {
            Err(ScheduleError::ScheduleNotFound)
        }
    }
}
