use chrono::{DateTime, Utc};
use std::sync::Arc;
use uuid::Uuid;

use crate::errors::TuitionError;
use crate::models::*;
use crate::store::InMemoryStore;

pub struct TuitionService {
    store: Arc<InMemoryStore>,
}

impl TuitionService {
    pub fn new(store: Arc<InMemoryStore>) -> Self {
        Self { store }
    }

    pub fn create_student(&self, name: String) -> Student {
        let student = Student {
            id: Uuid::new_v4(),
            name,
        };
        self.store.add_student(student.clone());
        student
    }

    pub fn get_student(&self, id: Uuid) -> Option<Student> {
        self.store.get_student(id)
    }

    pub fn list_students(&self) -> Vec<Student> {
        self.store.list_students()
    }

    pub fn create_course(&self, name: String, capacity: u32, semester: Semester) -> Course {
        let course = Course {
            id: Uuid::new_v4(),
            name,
            capacity,
            semester,
        };
        self.store.add_course(course.clone());
        course
    }

    pub fn get_course(&self, id: Uuid) -> Option<Course> {
        self.store.get_course(id)
    }

    pub fn list_courses(&self, semester: &Semester) -> Vec<CourseWithEnrollment> {
        let courses = self.store.list_courses(semester);
        courses
            .into_iter()
            .map(|c| {
                let enrolled = self.store.get_enrollment_count(c.id, &c.semester);
                CourseWithEnrollment {
                    course: c.clone(),
                    enrolled_count: enrolled,
                    is_full: enrolled >= c.capacity,
                }
            })
            .collect()
    }

    pub fn set_tuition(&self, student_id: Uuid, semester: Semester, amount: f64) -> Result<TuitionRecord, TuitionError> {
        if self.store.get_student(student_id).is_none() {
            return Err(TuitionError::StudentNotFound);
        }
        let record = TuitionRecord::new(student_id, semester, amount);
        self.store.set_tuition_record(record.clone());
        Ok(record)
    }

    pub fn get_tuition_record(&self, student_id: Uuid, semester: &Semester) -> Result<TuitionRecord, TuitionError> {
        if self.store.get_student(student_id).is_none() {
            return Err(TuitionError::StudentNotFound);
        }
        self.store
            .get_tuition_record(student_id, semester)
            .ok_or_else(|| TuitionError::TuitionRecordNotFound(semester.to_string()))
    }

    pub fn pay_first_installment(&self, student_id: Uuid, semester: &Semester, amount: f64) -> Result<TuitionRecord, TuitionError> {
        if amount <= 0.0 {
            return Err(TuitionError::InvalidPaymentAmount(amount));
        }

        let record = self.get_tuition_record(student_id, semester)?;
        
        if record.first_installment > 0.0 {
            return Err(TuitionError::FirstInstallmentAlreadyPaid);
        }

        let owed = record.owed_amount();
        if amount > owed {
            return Err(TuitionError::InvalidPaymentAmount(amount));
        }

        let updated = self.store.update_tuition_record(student_id, semester, |r| {
            r.first_installment = amount;
            r.update_paid_status();
        });

        if !updated {
            return Err(TuitionError::TuitionRecordNotFound(semester.to_string()));
        }

        self.get_tuition_record(student_id, semester)
    }

    pub fn pay_second_installment(&self, student_id: Uuid, semester: &Semester, amount: f64) -> Result<TuitionRecord, TuitionError> {
        if amount <= 0.0 {
            return Err(TuitionError::InvalidPaymentAmount(amount));
        }

        let record = self.get_tuition_record(student_id, semester)?;
        
        if record.first_installment <= 0.0 {
            return Err(TuitionError::MustPayFirstInstallmentFirst);
        }

        if record.second_installment > 0.0 {
            return Err(TuitionError::SecondInstallmentAlreadyPaid);
        }

        let owed = record.owed_amount();
        if amount > owed {
            return Err(TuitionError::InvalidPaymentAmount(amount));
        }

        let updated = self.store.update_tuition_record(student_id, semester, |r| {
            r.second_installment = amount;
            r.update_paid_status();
        });

        if !updated {
            return Err(TuitionError::TuitionRecordNotFound(semester.to_string()));
        }

        self.get_tuition_record(student_id, semester)
    }

    pub fn set_enrollment_period(&self, semester: Semester, start: DateTime<Utc>, end: DateTime<Utc>) {
        let period = EnrollmentPeriod {
            semester,
            start_time: start,
            end_time: end,
        };
        self.store.set_enrollment_period(period);
    }

    pub fn enroll(&self, student_id: Uuid, course_id: Uuid, semester: &Semester) -> Result<Enrollment, TuitionError> {
        if self.store.get_student(student_id).is_none() {
            return Err(TuitionError::StudentNotFound);
        }

        let course = self.store
            .get_course(course_id)
            .ok_or(TuitionError::CourseNotFound)?;

        if &course.semester != semester {
            return Err(TuitionError::CourseNotFound);
        }

        let period = self.store
            .get_enrollment_period(semester)
            .ok_or_else(|| TuitionError::EnrollmentPeriodNotSet(semester.to_string()))?;

        if !period.is_active() {
            return Err(TuitionError::EnrollmentPeriodNotActive(semester.to_string()));
        }

        let tuition = self.get_tuition_record(student_id, semester)?;
        if !tuition.is_paid {
            return Err(TuitionError::UnpaidTuition(semester.to_string()));
        }

        if self.store.is_enrolled(course_id, student_id, semester) {
            return Err(TuitionError::AlreadyEnrolled(course_id.to_string()));
        }

        let enrolled_count = self.store.get_enrollment_count(course_id, semester);
        if enrolled_count >= course.capacity {
            return Err(TuitionError::CourseFull(course.name.clone()));
        }

        let enrollment = Enrollment {
            course_id,
            student_id,
            semester: semester.clone(),
        };

        if !self.store.add_enrollment(enrollment.clone()) {
            return Err(TuitionError::AlreadyEnrolled(course_id.to_string()));
        }

        Ok(enrollment)
    }

    pub fn drop_course(&self, student_id: Uuid, course_id: Uuid, semester: &Semester) -> Result<(), TuitionError> {
        if self.store.get_student(student_id).is_none() {
            return Err(TuitionError::StudentNotFound);
        }

        if self.store.get_course(course_id).is_none() {
            return Err(TuitionError::CourseNotFound);
        }

        if !self.store.is_enrolled(course_id, student_id, semester) {
            return Err(TuitionError::NotEnrolled(course_id.to_string()));
        }

        if !self.store.remove_enrollment(course_id, student_id, semester) {
            return Err(TuitionError::NotEnrolled(course_id.to_string()));
        }

        Ok(())
    }

    pub fn list_student_enrollments(&self, student_id: Uuid, semester: &Semester) -> Result<Vec<Course>, TuitionError> {
        if self.store.get_student(student_id).is_none() {
            return Err(TuitionError::StudentNotFound);
        }

        let course_ids = self.store.list_student_enrollments(student_id, semester);
        Ok(course_ids
            .into_iter()
            .filter_map(|id| self.store.get_course(id))
            .collect())
    }

    pub fn list_owing_students(&self, semester: &Semester) -> Vec<StudentOwingInfo> {
        let students: Vec<_> = self.store.list_students();
        let mut owing: Vec<StudentOwingInfo> = students
            .into_iter()
            .filter_map(|s| {
                self.store
                    .get_tuition_record(s.id, semester)
                    .filter(|r| !r.is_paid && r.owed_amount() > 0.0)
                    .map(|r| StudentOwingInfo {
                        student: s,
                        semester: semester.clone(),
                        owed_amount: r.owed_amount(),
                    })
            })
            .collect();

        owing.sort_by(|a, b| b.owed_amount.partial_cmp(&a.owed_amount).unwrap());
        owing
    }
}
